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
	return c.send(ctx, http.MethodPost, "/api/v2/recover", contentType, body, total, nil)
}

// Reboot asks the player to reboot.
func (c *Client) Reboot(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/api/v2/reboot", nil, nil)
}

// Shutdown asks the player to shut down.
func (c *Client) Shutdown(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/api/v2/shutdown", nil, nil)
}
