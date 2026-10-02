package anthias

import (
	"context"
	"net/http"
	"net/url"
)

// ImportProvider identifies a third-party signage platform that
// [Client.ValidateImportToken] and [Client.ImportItem] can pull media from.
type ImportProvider string

// Import providers known to Anthias v2026.09.0. The token format differs
// per provider:
const (
	// ImportYodeck token: a Yodeck API token, or "<label>:<token>".
	ImportYodeck ImportProvider = "yodeck"
	// ImportScreenCloud token: a ScreenCloud Studio API token.
	ImportScreenCloud ImportProvider = "screencloud"
	// ImportPiSignage token: "<subdomain>:<email>:<password>".
	ImportPiSignage ImportProvider = "pisignage"
	// ImportXibo token: "<cms-url> <client_id> <client_secret>".
	ImportXibo ImportProvider = "xibo"
)

// ValidateScreenlyToken checks a Screenly v4.1 API token and, when valid,
// get-or-creates the asset group that migrated assets are placed in. An
// invalid token is reported as Valid == false with a nil error; network
// failures towards Screenly are a 502 [*APIError] (see [APIError.Message]).
// The token is not stored on the player.
func (c *Client) ValidateScreenlyToken(ctx context.Context, token string) (*ScreenlyValidation, error) {
	var v ScreenlyValidation
	body := map[string]string{"token": token}
	if err := c.do(ctx, http.MethodPost, "/api/v2/integrations/screenly/validate", body, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// MigrateAssetToScreenly copies one Anthias asset to the Screenly account
// of req.Token. Per-asset failures (e.g. Screenly rejecting the asset) are
// reported as Success == false with Error set and a nil error; an unknown
// asset is a 404 and a network failure a 502 [*APIError].
func (c *Client) MigrateAssetToScreenly(ctx context.Context, req ScreenlyMigrateRequest) (*ScreenlyMigration, error) {
	var m ScreenlyMigration
	if err := c.do(ctx, http.MethodPost, "/api/v2/integrations/screenly/migrate", req, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ValidateImportToken checks a provider token and lists the provider's
// media. An invalid token is Valid == false with a nil error; an unknown
// provider is a 404 and a network failure a 502 [*APIError]. The token is
// not stored on the player.
func (c *Client) ValidateImportToken(ctx context.Context, provider ImportProvider, token string) (*ImportValidation, error) {
	var v ImportValidation
	body := map[string]string{"token": token}
	path := "/api/v2/integrations/import/" + url.PathEscape(string(provider)) + "/validate"
	if err := c.do(ctx, http.MethodPost, path, body, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// ImportItem imports one media item (an [ImportMediaItem.RemoteID]) from
// the provider as a new Anthias asset. Per-item failures are reported as
// Success == false with Error set and a nil error; deliberate skips set
// Skipped and Reason. An unknown provider is a 404 and a network failure a
// 502 [*APIError].
func (c *Client) ImportItem(ctx context.Context, provider ImportProvider, req ImportItemRequest) (*ImportResult, error) {
	var r ImportResult
	path := "/api/v2/integrations/import/" + url.PathEscape(string(provider)) + "/item"
	if err := c.do(ctx, http.MethodPost, path, req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
