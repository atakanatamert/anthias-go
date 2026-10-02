package anthias

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testUA = "anthias-go-test"

func newTestClient(t *testing.T, h http.Handler) *Client {
	t.Helper()
	return newTestClientOpts(t, h)
}

func newTestClientOpts(t *testing.T, h http.Handler, opts ...Option) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	c, err := New(srv.URL, append([]Option{
		WithBasicAuth("user", "pass"),
		WithUserAgent(testUA),
	}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func assertCommonHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("User-Agent"); got != testUA {
		t.Fatalf("User-Agent = %q, want %q", got, testUA)
	}
	user, pass, ok := r.BasicAuth()
	if !ok || user != "user" || pass != "pass" {
		t.Fatalf("BasicAuth = %q/%q/%v, want user/pass/true", user, pass, ok)
	}
}

func assertMethodPath(t *testing.T, r *http.Request, method, path string) {
	t.Helper()
	if r.Method != method {
		t.Fatalf("method = %s, want %s", r.Method, method)
	}
	if r.URL.Path != path {
		t.Fatalf("path = %s, want %s", r.URL.Path, path)
	}
}

func assertContentType(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Header.Get("Content-Type"); got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatal(err)
	}
}

func TestNewRejectsBaseURLWithoutScheme(t *testing.T) {
	if _, err := New("player.local"); err == nil {
		t.Fatal("New without scheme succeeded")
	}
}

func TestAPIError(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethodPath(t, r, http.MethodGet, "/api/v2/assets/missing")
		assertCommonHeaders(t, r)
		http.Error(w, "not here", http.StatusNotFound)
	}))

	_, err := c.GetAsset(context.Background(), "missing")
	if err == nil {
		t.Fatal("GetAsset succeeded")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
	if apiErr.Method != http.MethodGet {
		t.Fatalf("Method = %s, want %s", apiErr.Method, http.MethodGet)
	}
	if apiErr.URL != "/api/v2/assets/missing" {
		t.Fatalf("URL = %s, want /api/v2/assets/missing", apiErr.URL)
	}
	if string(apiErr.Body) == "" {
		t.Fatal("Body is empty")
	}
	if !apiErr.IsNotFound() {
		t.Fatal("IsNotFound returned false")
	}
}

func TestContextCancellation(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server received request for canceled context")
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.ListAssets(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

// With auth enabled, Anthias redirects unauthenticated API calls to its HTML
// login page. Following that redirect turned writes into silent successes.
func TestLoginRedirectIsUnauthorizedError(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/" {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html>login</html>"))
			return
		}
		http.Redirect(w, r, "/login/?next="+r.URL.Path, http.StatusFound)
	}))

	err := c.DeleteAsset(context.Background(), "asset-1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("DeleteAsset error = %v, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusFound || !apiErr.IsUnauthorized() {
		t.Fatalf("APIError = %+v, want 302 with IsUnauthorized", apiErr)
	}
	if _, err := c.ListAssets(context.Background()); !errors.As(err, &apiErr) || !apiErr.IsUnauthorized() {
		t.Fatalf("ListAssets error = %v, want unauthorized *APIError", err)
	}
}

func TestIsUnauthorized(t *testing.T) {
	cases := []struct {
		err  APIError
		want bool
	}{
		{APIError{StatusCode: http.StatusUnauthorized}, true},
		{APIError{StatusCode: http.StatusFound, Location: "/login/?next=/api/v2/info"}, true},
		{APIError{StatusCode: http.StatusMovedPermanently, Location: "/api/v2/assets/"}, false},
		{APIError{StatusCode: http.StatusForbidden}, false},
		{APIError{StatusCode: http.StatusNotFound}, false},
	}
	for _, tc := range cases {
		if got := tc.err.IsUnauthorized(); got != tc.want {
			t.Errorf("IsUnauthorized(%d %q) = %v, want %v", tc.err.StatusCode, tc.err.Location, got, tc.want)
		}
	}
}

func TestAPIErrorMessage(t *testing.T) {
	cases := map[string]string{
		`{"message":"No HDMI-CEC adapter detected on this device."}`: "No HDMI-CEC adapter detected on this device.",
		`{"valid":false,"error":"Could not reach Screenly."}`:        "Could not reach Screenly.",
		`{"detail":"No Asset matches the given query."}`:             "No Asset matches the given query.",
		`{"success":false,"asset_id":null,"error":null}`:             "",
		`{"play_time_to":["must be set together"]}`:                  "",
		`<html>login</html>`: "",
	}
	for body, want := range cases {
		e := &APIError{Body: []byte(body)}
		if got := e.Message(); got != want {
			t.Errorf("Message(%s) = %q, want %q", body, got, want)
		}
	}
}
