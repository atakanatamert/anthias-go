package anthias

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIError represents a non-2xx response from the Anthias API.
type APIError struct {
	StatusCode int
	Method     string
	URL        string
	// Location is the Location header of a 3xx response (redirects are
	// never followed), e.g. the player's "/login/" page.
	Location string
	Body     []byte // raw response body (possibly truncated to 64 KiB)
}

func (e *APIError) Error() string {
	if e.Location != "" {
		return fmt.Sprintf("anthias: %s %s: %d %s: redirected to %s",
			e.Method, e.URL, e.StatusCode, http.StatusText(e.StatusCode), e.Location)
	}
	return fmt.Sprintf("anthias: %s %s: %d %s: %s",
		e.Method, e.URL, e.StatusCode, http.StatusText(e.StatusCode), e.Body)
}

// IsNotFound reports whether the error is a 404 response.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnauthorized reports whether the player rejected the request's
// credentials. Anthias answers wrong Basic Auth credentials with 401 and
// missing credentials (when authentication is enabled) with a redirect to
// its login page.
func (e *APIError) IsUnauthorized() bool {
	if e.StatusCode == http.StatusUnauthorized {
		return true
	}
	return e.StatusCode >= 300 && e.StatusCode < 400 && strings.Contains(e.Location, "/login")
}

// Message returns the human-readable reason from a JSON error body (the
// player's "error", "message" or "detail" field), or "" when the body has
// none. Validation errors keyed by field name are not flattened; inspect
// Body for those.
func (e *APIError) Message() string {
	var body struct {
		Error   *string `json:"error"`
		Message *string `json:"message"`
		Detail  *string `json:"detail"`
	}
	if json.Unmarshal(e.Body, &body) != nil {
		return ""
	}
	for _, s := range []*string{body.Error, body.Message, body.Detail} {
		if s != nil && *s != "" {
			return *s
		}
	}
	return ""
}
