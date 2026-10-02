package anthias

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestInternalEndpointsSendDerivedToken(t *testing.T) {
	// Reference value: hex(HMAC-SHA256(key="test-secret", msg="anthias-internal-api-v1")),
	// computed independently with Python's hmac module.
	const want = "d8f01b17a399f522cc39eb39851f0d44eb30bfc6bf41b833af1fa5e1f0f191c4"
	var got []string
	c := newTestClientOpts(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.URL.Path+"="+r.Header.Get("X-Anthias-Internal-Token"))
		switch r.URL.Path {
		case "/api/v2/viewer/settings":
			writeJSON(t, w, http.StatusOK, ViewerSettings{ScreenRotation: 90})
		case "/api/v2/assets":
			writeJSON(t, w, http.StatusOK, []Asset{})
		default:
			w.WriteHeader(http.StatusAccepted)
		}
	}), WithInternalSecret("test-secret"))

	if _, err := c.GetViewerSettings(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := c.RecheckAsset(context.Background(), "a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ListAssets(context.Background()); err != nil {
		t.Fatal(err)
	}
	wantCalls := []string{
		"/api/v2/viewer/settings=" + want,
		"/api/v2/assets/a1/recheck=" + want,
		"/api/v2/assets=", // the token is not sent to operator endpoints
	}
	if len(got) != len(wantCalls) {
		t.Fatalf("calls = %v", got)
	}
	for i := range wantCalls {
		if got[i] != wantCalls[i] {
			t.Fatalf("call %d = %q, want %q", i, got[i], wantCalls[i])
		}
	}
}

func TestInternalEndpointsWithoutSecret(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no request expected without an internal secret")
	}))
	if _, err := c.GetViewerPlaylist(context.Background()); !errors.Is(err, ErrNoInternalSecret) {
		t.Fatalf("error = %v, want ErrNoInternalSecret", err)
	}
}
