package anthias

import (
	"bytes"
	"context"
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

// UploadProgress receives cumulative bytes sent and the total request body
// size during an upload.
type UploadProgress func(sent, total int64)

// UploadOption configures an upload.
type UploadOption func(*uploadOptions)

type uploadOptions struct {
	progress UploadProgress
}

// WithProgress registers a progress callback. It is invoked on the first
// read, at most every ~500ms thereafter, and once at completion.
func WithProgress(fn UploadProgress) UploadOption {
	return func(o *uploadOptions) { o.progress = fn }
}

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
	if err := c.send(ctx, http.MethodPost, "/api/v2/file_asset", contentType, body, total, &fu); err != nil {
		return nil, err
	}
	return &fu, nil
}
