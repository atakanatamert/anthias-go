package anthias

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// GetDeviceSettings fetches the player's device settings.
func (c *Client) GetDeviceSettings(ctx context.Context) (*DeviceSettings, error) {
	var s DeviceSettings
	if err := c.do(ctx, http.MethodGet, "/api/v2/device_settings", nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateDeviceSettings partially updates device settings.
func (c *Client) UpdateDeviceSettings(ctx context.Context, req UpdateDeviceSettingsRequest) error {
	return c.do(ctx, http.MethodPatch, "/api/v2/device_settings", req, nil)
}

// Backup triggers a backup on the player and returns the resulting backup
// filename.
func (c *Client) Backup(ctx context.Context) (string, error) {
	var filename string
	if err := c.do(ctx, http.MethodPost, "/api/v2/backup", nil, &filename); err != nil {
		return "", err
	}
	return filename, nil
}

// Recover restores a backup by uploading it (multipart field
// "backup_upload"). filename sets the uploaded part's filename. size must be
// the exact payload length so the request uses Content-Length rather than
// chunked transfer encoding. The payload is streamed and never buffered.
//
// Recover honors ctx; pass a generous context.WithTimeout for large backups.
func (c *Client) Recover(ctx context.Context, r io.Reader, filename string, size int64) error {
	head, tail, contentType, err := buildMultipartHeader("backup_upload", filename, "application/octet-stream")
	if err != nil {
		return err
	}
	total := int64(len(head)) + size + int64(len(tail))
	body := io.MultiReader(bytes.NewReader(head), r, bytes.NewReader(tail))
	return c.send(ctx, http.MethodPost, "/api/v2/recover", nil, contentType, body, total, nil)
}

// Reboot asks the player to reboot.
func (c *Client) Reboot(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/api/v2/reboot", nil, nil)
}

// Shutdown asks the player to shut down.
func (c *Client) Shutdown(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/api/v2/shutdown", nil, nil)
}

// SetDisplayPower turns the attached display on or off over HDMI-CEC and
// returns the player's status message. Experimental on the player side.
// Errors are [*APIError]s: 503 when the player has no CEC adapter, 502 when
// the display did not respond; [APIError.Message] holds the reason.
func (c *Client) SetDisplayPower(ctx context.Context, on bool) (string, error) {
	state := "off"
	if on {
		state = "on"
	}
	var out struct {
		Message string `json:"message"`
	}
	if err := c.do(ctx, http.MethodPost, "/api/v2/display/"+state, nil, &out); err != nil {
		return "", err
	}
	return out.Message, nil
}

// GetIPAddresses returns the player's addresses as URLs (e.g.
// "http://10.0.0.108"), as shown on its splash screen; empty while the
// player has no address yet. Unlike [Client.GetInfo] this endpoint needs
// no credentials and is cheap enough to poll.
func (c *Client) GetIPAddresses(ctx context.Context) ([]string, error) {
	var out struct {
		IPAddresses []string `json:"ip_addresses"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/v2/network/ip-addresses", nil, &out); err != nil {
		return nil, err
	}
	return out.IPAddresses, nil
}
