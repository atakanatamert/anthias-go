package anthias

import (
	"context"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestUploadFileReaderMultipartContentLengthAndProgress(t *testing.T) {
	payload := "hello upload"
	var finalSent, finalTotal int64
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethodPath(t, r, http.MethodPost, "/api/v2/file_asset")
		assertCommonHeaders(t, r)
		if r.ContentLength <= int64(len(payload)) {
			t.Fatalf("ContentLength = %d, want multipart length > payload", r.ContentLength)
		}
		if got, want := r.Header.Get("Content-Length"), strconv.FormatInt(r.ContentLength, 10); got != want {
			t.Fatalf("Content-Length header = %q, want %q", got, want)
		}
		if len(r.TransferEncoding) != 0 {
			t.Fatalf("TransferEncoding = %#v, want none", r.TransferEncoding)
		}
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		if mediaType != "multipart/form-data" {
			t.Fatalf("media type = %q, want multipart/form-data", mediaType)
		}
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}
		part, err := mr.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		if part.FormName() != "file_upload" {
			t.Fatalf("form name = %q, want file_upload", part.FormName())
		}
		if part.FileName() != "hello.txt" {
			t.Fatalf("filename = %q, want hello.txt", part.FileName())
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != payload {
			t.Fatalf("body = %q, want %q", body, payload)
		}
		if part, err = mr.NextPart(); err != io.EOF {
			t.Fatalf("next part = %#v, %v; want EOF", part, err)
		}
		writeJSON(t, w, http.StatusCreated, FileUpload{URI: "/asset/file/hello.txt", Ext: ".txt"})
	}))

	fu, err := c.UploadFileReader(context.Background(), strings.NewReader(payload), "hello.txt", int64(len(payload)), WithProgress(func(sent, total int64) {
		finalSent, finalTotal = sent, total
	}))
	if err != nil {
		t.Fatal(err)
	}
	if fu.URI != "/asset/file/hello.txt" || fu.Ext != ".txt" {
		t.Fatalf("upload = %#v", fu)
	}
	if finalSent != finalTotal || finalTotal <= int64(len(payload)) {
		t.Fatalf("final progress = (%d,%d), want (total,total)", finalSent, finalTotal)
	}
}
