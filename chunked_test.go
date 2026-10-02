package anthias

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"testing"
)

// chunkServer emulates the player's resumable file_asset endpoint: it
// writes each chunk at its Content-Range offset into the session named by
// X-Upload-Id (minting one on the first chunk), mirroring Anthias'
// validation of the range against the chunk length.
type chunkServer struct {
	t        *testing.T
	mu       sync.Mutex
	files    map[string][]byte
	next     int
	failCall int // 1-based request number to fail with 500; 0 = never
	calls    int
}

var rangeRe = regexp.MustCompile(`^bytes (\d+)-(\d+)/(\d+)$`)

func (s *chunkServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.calls == s.failCall {
		http.Error(w, `{"detail":"boom"}`, http.StatusInternalServerError)
		return
	}
	m := rangeRe.FindStringSubmatch(r.Header.Get("Content-Range"))
	if m == nil {
		http.Error(w, "missing Content-Range", http.StatusBadRequest)
		return
	}
	start, _ := strconv.ParseInt(m[1], 10, 64)
	end, _ := strconv.ParseInt(m[2], 10, 64)
	total, _ := strconv.ParseInt(m[3], 10, 64)
	f, _, err := r.FormFile("file_upload")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, _ := io.ReadAll(f)
	if int64(len(data)) != end-start+1 || end >= total {
		http.Error(w, "range/length mismatch", http.StatusBadRequest)
		return
	}
	id := r.Header.Get("X-Upload-Id")
	if id == "" {
		s.next++
		id = fmt.Sprintf("u%d", s.next)
		s.files[id] = make([]byte, total)
	}
	buf, ok := s.files[id]
	if !ok {
		http.Error(w, "unknown upload id", http.StatusBadRequest)
		return
	}
	copy(buf[start:], data)
	writeJSON(s.t, w, http.StatusOK, FileUpload{URI: "/data/" + id + ".tmp", Ext: ".mp4", UploadID: id})
}

func TestUploadFileReaderChunkedReassembles(t *testing.T) {
	srv := &chunkServer{t: t, files: map[string][]byte{}}
	c := newTestClient(t, srv)
	payload := bytes.Repeat([]byte("0123456789abcdefghij"), 51) // 1020 bytes: 3 full 256-byte chunks + a 252-byte one

	var lastSent, lastTotal int64
	fu, err := c.UploadFileReaderChunked(context.Background(), bytes.NewReader(payload), "clip.mp4", int64(len(payload)),
		WithChunkSize(256), WithProgress(func(sent, total int64) { lastSent, lastTotal = sent, total }))
	if err != nil {
		t.Fatal(err)
	}
	if srv.calls != 4 {
		t.Fatalf("requests = %d, want 4", srv.calls)
	}
	if got := srv.files[fu.UploadID]; !bytes.Equal(got, payload) {
		t.Fatalf("reassembled %d bytes differ from payload", len(got))
	}
	if fu.URI != "/data/u1.tmp" {
		t.Fatalf("URI = %q", fu.URI)
	}
	if lastSent != int64(len(payload)) || lastTotal != int64(len(payload)) {
		t.Fatalf("final progress = %d/%d, want %d/%d", lastSent, lastTotal, len(payload), len(payload))
	}
}

func TestUploadFileReaderChunkedResume(t *testing.T) {
	srv := &chunkServer{t: t, files: map[string][]byte{}, failCall: 3}
	c := newTestClient(t, srv)
	payload := bytes.Repeat([]byte("x1y2z3"), 100) // 600 bytes, chunks of 200

	_, err := c.UploadFileReaderChunked(context.Background(), bytes.NewReader(payload), "clip.mp4", int64(len(payload)), WithChunkSize(200))
	var cerr *ChunkedUploadError
	if !errors.As(err, &cerr) {
		t.Fatalf("error = %v, want *ChunkedUploadError", err)
	}
	if cerr.UploadID != "u1" || cerr.Offset != 400 {
		t.Fatalf("ChunkedUploadError = %+v, want upload u1 at offset 400", cerr)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("wrapped error = %v, want 500 *APIError", err)
	}

	fu, err := c.UploadFileReaderChunked(context.Background(), bytes.NewReader(payload), "clip.mp4", int64(len(payload)),
		WithChunkSize(200), WithResume(cerr.UploadID, cerr.Offset))
	if err != nil {
		t.Fatal(err)
	}
	if fu.UploadID != "u1" || !bytes.Equal(srv.files["u1"], payload) {
		t.Fatalf("resumed upload %q did not reassemble the payload", fu.UploadID)
	}
	if srv.calls != 4 {
		t.Fatalf("requests = %d, want 4 (2 ok + 1 failed + 1 resumed)", srv.calls)
	}
}

func TestUploadFileReaderChunkedRejectsBadResume(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no request expected")
	}))
	r := bytes.NewReader(make([]byte, 10))
	for name, opts := range map[string][]UploadOption{
		"offset past end":    {WithResume("u1", 10)},
		"offset without id":  {WithResume("", 4)},
		"negative offset":    {WithResume("u1", -1)},
		"non-positive chunk": {WithChunkSize(0)},
	} {
		if _, err := c.UploadFileReaderChunked(context.Background(), r, "a.png", 10, opts...); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
