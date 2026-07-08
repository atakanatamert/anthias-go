package anthias

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// PlaybackCommand is a playlist control command for [Client.ControlPlayback].
type PlaybackCommand string

const (
	// PlaybackNext skips to the next asset.
	PlaybackNext PlaybackCommand = "next"
	// PlaybackPrevious returns to the previous asset.
	PlaybackPrevious PlaybackCommand = "previous"
)

// PlaybackAsset switches playback to the given asset. The resulting command
// is path-escaped before being sent.
func PlaybackAsset(assetID string) PlaybackCommand {
	return PlaybackCommand("asset&" + assetID)
}

// ListAssets lists all assets.
func (c *Client) ListAssets(ctx context.Context) ([]Asset, error) {
	var assets []Asset
	if err := c.do(ctx, http.MethodGet, "/api/v2/assets", nil, &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

// GetAsset fetches a single asset by ID.
func (c *Client) GetAsset(ctx context.Context, assetID string) (*Asset, error) {
	var a Asset
	if err := c.do(ctx, http.MethodGet, "/api/v2/assets/"+url.PathEscape(assetID), nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// CreateAsset creates a new asset.
func (c *Client) CreateAsset(ctx context.Context, req CreateAssetRequest) (*Asset, error) {
	var a Asset
	if err := c.do(ctx, http.MethodPost, "/api/v2/assets", req, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// UpdateAsset partially updates an asset.
func (c *Client) UpdateAsset(ctx context.Context, assetID string, req UpdateAssetRequest) (*Asset, error) {
	var a Asset
	if err := c.do(ctx, http.MethodPatch, "/api/v2/assets/"+url.PathEscape(assetID), req, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// ReplaceAsset replaces an asset (full update).
func (c *Client) ReplaceAsset(ctx context.Context, assetID string, req CreateAssetRequest) (*Asset, error) {
	var a Asset
	if err := c.do(ctx, http.MethodPut, "/api/v2/assets/"+url.PathEscape(assetID), req, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// DeleteAsset deletes an asset.
func (c *Client) DeleteAsset(ctx context.Context, assetID string) error {
	return c.do(ctx, http.MethodDelete, "/api/v2/assets/"+url.PathEscape(assetID), nil, nil)
}

// SetPlaylistOrder reorders assets. assetIDs is sent form-encoded as a
// comma-joined "ids" value.
func (c *Client) SetPlaylistOrder(ctx context.Context, assetIDs []string) error {
	form := url.Values{}
	form.Set("ids", strings.Join(assetIDs, ","))
	body := form.Encode()
	return c.send(ctx, http.MethodPost, "/api/v2/assets/order",
		"application/x-www-form-urlencoded", strings.NewReader(body), int64(len(body)), nil)
}

// ControlPlayback sends a playback command (next/previous/asset).
func (c *Client) ControlPlayback(ctx context.Context, command PlaybackCommand) error {
	path := "/api/v2/assets/control/" + url.PathEscape(string(command))
	return c.do(ctx, http.MethodGet, path, nil, nil)
}

// GetAssetContent fetches an asset's content (file bytes or URL).
func (c *Client) GetAssetContent(ctx context.Context, assetID string) (*AssetContent, error) {
	var content AssetContent
	if err := c.do(ctx, http.MethodGet, "/api/v2/assets/"+url.PathEscape(assetID)+"/content", nil, &content); err != nil {
		return nil, err
	}
	return &content, nil
}
