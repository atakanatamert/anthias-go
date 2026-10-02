package anthias

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"time"
)

// DefaultChunkSize is the chunk size of chunked uploads unless
// [WithChunkSize] says otherwise.
const DefaultChunkSize = 8 << 20

// UploadProgress receives cumulative bytes sent and the total request body
// size during an upload.
type UploadProgress func(sent, total int64)

// UploadOption configures an upload.
type UploadOption func(*uploadOptions)

type uploadOptions struct {
	progress     UploadProgress
	chunkSize    int64
	uploadID     string
	resumeOffset int64
}

// WithProgress registers a progress callback. It is invoked on the first
// read, at most every ~500ms thereafter, and once at completion.
func WithProgress(fn UploadProgress) UploadOption {
	return func(o *uploadOptions) { o.progress = fn }
}

// WithChunkSize sets the chunk size in bytes for [Client.UploadFileChunked]
// and [Client.UploadFileReaderChunked]. Ignored by single-request uploads.
func WithChunkSize(n int64) UploadOption {
	return func(o *uploadOptions) { o.chunkSize = n }
}

// WithResume continues an interrupted chunked upload from the session and
// offset reported by a [*ChunkedUploadError]. The same file must be passed
// again. Ignored by single-request uploads.
func WithResume(uploadID string, offset int64) UploadOption {
	return func(o *uploadOptions) {
		o.uploadID = uploadID
		o.resumeOffset = offset
	}
}

// ChunkedUploadError reports a failed chunk of a chunked upload. Every byte
// before Offset was stored by the player; pass [WithResume](UploadID,
// Offset) to continue. UploadID is empty when the first chunk failed, in
// which case the upload simply starts over.
type ChunkedUploadError struct {
	UploadID string
	Offset   int64
	Err      error
}

func (e *ChunkedUploadError) Error() string {
	return fmt.Sprintf("anthias: chunked upload failed at byte %d (upload id %q): %v", e.Offset, e.UploadID, e.Err)
}

func (e *ChunkedUploadError) Unwrap() error { return e.Err }

// progressReader wraps an io.Reader, counting bytes read and invoking a
// progress callback subject to a rate limit.
type progressReader struct {
	r     io.Reader
	fn    UploadProgress
	sent  int64
	total int64
	last  time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.sent += int64(n)
	now := time.Now()
	if pr.last.IsZero() || now.Sub(pr.last) >= progressDelay || err != nil {
		pr.fn(pr.sent, pr.total)
		pr.last = now
	}
	return n, err
}

// buildMultipartHeader writes a single multipart field's preamble (boundary
// line + headers + blank line) and returns it along with the closing
// boundary and the multipart Content-Type. The file body itself is streamed
// by the caller between head and tail.
func buildMultipartHeader(field, filename, mimetype string) (head, tail []byte, contentType string, err error) {
	if mimetype == "" {
		mimetype = "application/octet-stream"
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf("form-data; name=%q; filename=%q", field, filename))
	h.Set("Content-Type", mimetype)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if _, err = mw.CreatePart(h); err != nil {
		return nil, nil, "", err
	}
	contentType = mw.FormDataContentType()
	tail = []byte("\r\n--" + mw.Boundary() + "--\r\n")
	return buf.Bytes(), tail, contentType, nil
}

func guessMimetype(filename string) string {
	if ext := filepath.Ext(filename); ext != "" {
		if mt := mime.TypeByExtension(ext); mt != "" {
			return mt
		}
	}
	return "application/octet-stream"
}

// UploadFile uploads the file at path to POST /api/v2/file_asset (multipart
// field "file_upload"). The file is streamed and never fully buffered.
// The player accepts only images and videos, judged by the file name's
// extension; other files are rejected with a 400 [*APIError].
//
// Uploads honor ctx; pass a generous context.WithTimeout for large files,
// since a short client-wide timeout (see [WithTimeout]) may abort long
// transfers.
func (c *Client) UploadFile(ctx context.Context, path string, opts ...UploadOption) (*FileUpload, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return c.UploadFileReader(ctx, f, filepath.Base(path), fi.Size(), opts...)
}

// UploadFileReader uploads the contents of r as filename (size bytes) to
// POST /api/v2/file_asset (multipart field "file_upload"). size must be the
// exact content length so the request uses Content-Length rather than
// chunked transfer encoding (some proxies mishandle chunked uploads).
// filename's extension must identify an image or video (see [Client.UploadFile]).
//
// Uploads honor ctx; pass a generous context.WithTimeout for large files.
func (c *Client) UploadFileReader(ctx context.Context, r io.Reader, filename string, size int64, opts ...UploadOption) (*FileUpload, error) {
	o := uploadOptions{}
	for _, opt := range opts {
		opt(&o)
	}

	head, tail, contentType, err := buildMultipartHeader("file_upload", filename, guessMimetype(filename))
	if err != nil {
		return nil, err
	}
	total := int64(len(head)) + size + int64(len(tail))

	var body io.Reader = io.MultiReader(bytes.NewReader(head), r, bytes.NewReader(tail))
	if o.progress != nil {
		body = &progressReader{r: body, fn: o.progress, total: total}
	}

	var fu FileUpload
	if err := c.send(ctx, http.MethodPost, "/api/v2/file_asset", nil, contentType, body, total, &fu); err != nil {
		return nil, err
	}
	return &fu, nil
}

// UploadFileChunked uploads the file at path in chunks (see
// [Client.UploadFileReaderChunked]).
func (c *Client) UploadFileChunked(ctx context.Context, path string, opts ...UploadOption) (*FileUpload, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return c.UploadFileReaderChunked(ctx, f, filepath.Base(path), fi.Size(), opts...)
}

// UploadFileReaderChunked uploads size bytes of r as filename to POST
// /api/v2/file_asset as a resumable upload: one request per chunk
// ([DefaultChunkSize] unless [WithChunkSize]), each with a Content-Range
// header, tied together by the player's upload id. Use it for large videos
// or unreliable links. If a chunk fails, the error is a
// [*ChunkedUploadError] carrying the upload id and the offset to resume
// from with [WithResume]. filename's extension must identify an image or
// video. Progress callbacks receive file bytes (sent, size).
//
// Resuming needs Anthias v2026.07.2 or newer: older players accept the
// chunks but return no upload id, so an interrupted upload must restart.
//
// Each chunk honors ctx; the client-wide timeout applies per chunk.
func (c *Client) UploadFileReaderChunked(ctx context.Context, r io.ReaderAt, filename string, size int64, opts ...UploadOption) (*FileUpload, error) {
	o := uploadOptions{chunkSize: DefaultChunkSize}
	for _, opt := range opts {
		opt(&o)
	}
	switch {
	case size <= 0:
		return nil, errors.New("anthias: chunked upload needs size > 0")
	case o.chunkSize <= 0:
		return nil, errors.New("anthias: chunk size must be > 0")
	case o.resumeOffset < 0 || o.resumeOffset >= size:
		return nil, fmt.Errorf("anthias: resume offset %d outside file of %d bytes", o.resumeOffset, size)
	case o.resumeOffset > 0 && o.uploadID == "":
		return nil, errors.New("anthias: resuming at a non-zero offset needs an upload id")
	}

	head, tail, contentType, err := buildMultipartHeader("file_upload", filename, guessMimetype(filename))
	if err != nil {
		return nil, err
	}

	uploadID, offset := o.uploadID, o.resumeOffset
	var fu FileUpload
	for offset < size {
		end := min(offset+o.chunkSize, size) - 1
		n := end - offset + 1
		var chunk io.Reader = io.NewSectionReader(r, offset, n)
		if o.progress != nil {
			chunk = &progressReader{r: chunk, fn: o.progress, sent: offset, total: size}
		}
		body := io.MultiReader(bytes.NewReader(head), chunk, bytes.NewReader(tail))
		hdr := http.Header{"Content-Range": {fmt.Sprintf("bytes %d-%d/%d", offset, end, size)}}
		if uploadID != "" {
			hdr.Set("X-Upload-Id", uploadID)
		}
		fu = FileUpload{}
		total := int64(len(head)) + n + int64(len(tail))
		if err := c.send(ctx, http.MethodPost, "/api/v2/file_asset", hdr, contentType, body, total, &fu); err != nil {
			return nil, &ChunkedUploadError{UploadID: uploadID, Offset: offset, Err: err}
		}
		uploadID = fu.UploadID
		offset = end + 1
	}
	return &fu, nil
}
