package anthias

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func testSettings() DeviceSettings {
	return DeviceSettings{
		PlayerName:               "Lobby",
		AudioOutput:              "hdmi",
		DefaultDuration:          10,
		DefaultStreamingDuration: 120,
		DateFormat:               "YYYY-MM-DD",
		AuthBackend:              "auth_basic",
		ShowSplash:               true,
		DefaultAssets:            false,
		ShufflePlaylist:          true,
		Use24HourClock:           true,
		DebugLogging:             false,
		Username:                 "admin",
	}
}

func TestDeviceSettingsBackupPower(t *testing.T) {
	calls := 0
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertCommonHeaders(t, r)
		switch calls {
		case 0:
			assertMethodPath(t, r, http.MethodGet, "/api/v2/device_settings")
			writeJSON(t, w, http.StatusOK, testSettings())
		case 1:
			assertMethodPath(t, r, http.MethodPatch, "/api/v2/device_settings")
			assertContentType(t, r, "application/json")
			var got UpdateDeviceSettingsRequest
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got.PlayerName == nil || *got.PlayerName != "Front Desk" || got.Password == nil || *got.Password != "secret" {
				t.Fatalf("settings body = %#v", got)
			}
			w.WriteHeader(http.StatusNoContent)
		case 2:
			assertMethodPath(t, r, http.MethodPost, "/api/v2/backup")
			writeJSON(t, w, http.StatusCreated, "backup-20260708.zip")
		case 3:
			assertMethodPath(t, r, http.MethodPost, "/api/v2/reboot")
			w.WriteHeader(http.StatusNoContent)
		case 4:
			assertMethodPath(t, r, http.MethodPost, "/api/v2/shutdown")
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected call %d", calls)
		}
		calls++
	}))

	settings, err := c.GetDeviceSettings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if settings.PlayerName != "Lobby" || !settings.Use24HourClock {
		t.Fatalf("settings = %#v", settings)
	}
	if err := c.UpdateDeviceSettings(context.Background(), UpdateDeviceSettingsRequest{
		PlayerName: Ptr("Front Desk"),
		Password:   Ptr("secret"),
		Password2:  Ptr("secret"),
	}); err != nil {
		t.Fatal(err)
	}
	filename, err := c.Backup(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if filename != "backup-20260708.zip" {
		t.Fatalf("backup = %q", filename)
	}
	if err := c.Reboot(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := c.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 5 {
		t.Fatalf("calls = %d, want 5", calls)
	}
}

func TestRecoverMultipartContentLength(t *testing.T) {
	payload := "backup-bytes"
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertMethodPath(t, r, http.MethodPost, "/api/v2/recover")
		assertCommonHeaders(t, r)
		if got, want := r.Header.Get("Content-Length"), strconv.FormatInt(r.ContentLength, 10); got != want {
			t.Fatalf("Content-Length header = %q, want %q", got, want)
		}
		if len(r.TransferEncoding) != 0 {
			t.Fatalf("TransferEncoding = %#v, want none", r.TransferEncoding)
		}
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		if mediaType != "multipart/form-data" {
			t.Fatalf("media type = %q, want multipart/form-data", mediaType)
		}
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}
		part, err := mr.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		if part.FormName() != "backup_upload" {
			t.Fatalf("form name = %q, want backup_upload", part.FormName())
		}
		if part.FileName() != "backup.zip" {
			t.Fatalf("filename = %q, want backup.zip", part.FileName())
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != payload {
			t.Fatalf("body = %q, want %q", body, payload)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := c.Recover(context.Background(), strings.NewReader(payload), "backup.zip", int64(len(payload))); err != nil {
		t.Fatal(err)
	}
}
