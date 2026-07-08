package anthias

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// Version is the current SDK version.
const Version = "0.1.0"

const (
	maxErrorBody  = 64 * 1024
	defaultUA     = "anthias-go/" + Version
	progressDelay = 500 * time.Millisecond
)

// Client is an Anthias v2 API client. Construct one with [New].
type Client struct {
	baseURL   string
	httpc     *http.Client
	username  string
	password  string
	userAgent string
}

// Option configures a [Client].
type Option func(*Client)

// WithHTTPClient sets the underlying *http.Client used for requests.
// Default: &http.Client{Timeout: 30s}.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpc = hc }
}

// WithTimeout sets the timeout on the client's *http.Client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpc.Timeout = d }
}

// WithBasicAuth enables HTTP Basic Auth (Anthias auth_basic backend).
func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// New creates a Client for the player at baseURL. baseURL must include a
// scheme (e.g. "http://192.168.1.50" or "http://player.local:8080"); a
// trailing slash is trimmed.
func New(baseURL string, opts ...Option) (*Client, error) {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return nil, errors.New("anthias: baseURL must include a scheme (http:// or https://)")
	}
	c := &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		httpc:     &http.Client{Timeout: 30 * time.Second},
		userAgent: defaultUA,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// do JSON-encodes body (if non-nil) and decodes a JSON response into out (if
// non-nil). Non-2xx responses become *APIError.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	if body == nil {
		return c.send(ctx, method, path, "", nil, -1, out)
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.send(ctx, method, path, "application/json", bytes.NewReader(buf), int64(len(buf)), out)
}

// send is the low-level request helper. body may be nil; contentLength of -1
// leaves Content-Length unset. Non-2xx responses become *APIError with the
// body capped at maxErrorBody bytes.
func (c *Client) send(ctx context.Context, method, path, contentType string, body io.Reader, contentLength int64, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if contentLength >= 0 {
		req.ContentLength = contentLength
	}
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			URL:        path,
			Body:       errBody,
		}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxErrorBody))
	return nil
}
