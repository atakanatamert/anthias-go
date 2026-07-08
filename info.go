package anthias

import (
	"context"
	"net/http"
)

// GetInfo fetches player runtime information.
func (c *Client) GetInfo(ctx context.Context) (*Info, error) {
	var i Info
	if err := c.do(ctx, http.MethodGet, "/api/v2/info", nil, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

// GetIntegrations fetches player integration information (e.g. balena).
func (c *Client) GetIntegrations(ctx context.Context) (*Integrations, error) {
	var i Integrations
	if err := c.do(ctx, http.MethodGet, "/api/v2/integrations", nil, &i); err != nil {
		return nil, err
	}
	return &i, nil
}
