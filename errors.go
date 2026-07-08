package anthias

import (
	"fmt"
	"net/http"
)

// APIError represents a non-2xx response from the Anthias API.
type APIError struct {
	StatusCode int
	Method     string
	URL        string
	Body       []byte // raw response body (possibly truncated to 64 KiB)
}

func (e *APIError) Error() string {
	return fmt.Sprintf("anthias: %s %s: %d %s: %s",
		e.Method, e.URL, e.StatusCode, http.StatusText(e.StatusCode), e.Body)
}

// IsNotFound reports whether the error is a 404 response.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}
