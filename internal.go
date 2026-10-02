package anthias

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
)

// ErrNoInternalSecret is returned by player-internal endpoint methods when
// the client was created without [WithInternalSecret].
var ErrNoInternalSecret = errors.New("anthias: internal endpoint requires WithInternalSecret")

const internalTokenHeader = "X-Anthias-Internal-Token"

// internalToken derives the player's internal API token:
// hex(HMAC-SHA256(key=secret, msg="anthias-internal-api-v1")).
func internalToken(secret string) string {
	if secret == "" {
		return ""
	}
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte("anthias-internal-api-v1"))
	return hex.EncodeToString(m.Sum(nil))
}

// doInternal performs a request against a player-internal endpoint.
func (c *Client) doInternal(ctx context.Context, method, path string, out any) error {
	if c.internalToken == "" {
		return ErrNoInternalSecret
	}
	hdr := http.Header{internalTokenHeader: {c.internalToken}}
	return c.send(ctx, method, path, hdr, "", nil, -1, out)
}

// GetViewerPlaylist returns the currently active assets in play order and
// the moment the playlist should be fetched again, evaluated against the
// player's clock. Player-internal endpoint (built for the Anthias viewer);
// requires [WithInternalSecret]. A wrong secret yields a 403 [*APIError].
func (c *Client) GetViewerPlaylist(ctx context.Context) (*ViewerPlaylist, error) {
	var p ViewerPlaylist
	if err := c.doInternal(ctx, http.MethodGet, "/api/v2/viewer/playlist", &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetViewerSettings returns the subset of device settings the viewer uses.
// Player-internal endpoint; requires [WithInternalSecret].
func (c *Client) GetViewerSettings(ctx context.Context) (*ViewerSettings, error) {
	var s ViewerSettings
	if err := c.doInternal(ctx, http.MethodGet, "/api/v2/viewer/settings", &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// RecheckAsset queues a reachability check of a URL asset; the result shows
// up in [Asset.IsReachable] and [Asset.LastReachabilityCheck], typically
// within seconds. Disabled assets, assets still processing and assets with
// SkipAssetCheck are not probed, and the player debounces and rate-limits
// checks, so an accepted request does not guarantee a new probe. An unknown
// asset is a 404 [*APIError]. Player-internal endpoint; requires
// [WithInternalSecret].
func (c *Client) RecheckAsset(ctx context.Context, assetID string) error {
	return c.doInternal(ctx, http.MethodPost, "/api/v2/assets/"+url.PathEscape(assetID)+"/recheck", nil)
}
