package anthias

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

var (
	testStart = time.Date(2026, 7, 8, 9, 0, 0, 0, time.UTC)
	testEnd   = time.Date(2026, 7, 9, 9, 0, 0, 0, time.UTC)
)

func testAsset(id string) Asset {
	return Asset{
		AssetID:        id,
		Name:           "Menu",
		URI:            "https://example.com/menu.png",
		StartDate:      testStart,
		EndDate:        testEnd,
		Duration:       10,
		Mimetype:       "image/png",
		IsEnabled:      true,
		NoCache:        true,
		PlayOrder:      3,
		SkipAssetCheck: true,
		IsActive:       true,
		IsProcessing:   false,
	}
}

func TestListAndGetAssets(t *testing.T) {
	calls := 0
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertCommonHeaders(t, r)
		switch calls {
		case 0:
			assertMethodPath(t, r, http.MethodGet, "/api/v2/assets")
			writeJSON(t, w, http.StatusOK, []Asset{testAsset("asset-1")})
		case 1:
			assertMethodPath(t, r, http.MethodGet, "/api/v2/assets/asset-1")
			writeJSON(t, w, http.StatusOK, testAsset("asset-1"))
		default:
			t.Fatalf("unexpected call %d", calls)
		}
		calls++
	}))

	assets, err := c.ListAssets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 || assets[0].AssetID != "asset-1" || !assets[0].StartDate.Equal(testStart) {
		t.Fatalf("assets = %#v", assets)
	}
	asset, err := c.GetAsset(context.Background(), "asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if asset.AssetID != "asset-1" || asset.Mimetype != "image/png" {
		t.Fatalf("asset = %#v", asset)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestAssetWriteBodies(t *testing.T) {
	calls := 0
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertCommonHeaders(t, r)
		assertContentType(t, r, "application/json")
		switch calls {
		case 0:
			assertMethodPath(t, r, http.MethodPost, "/api/v2/assets")
			var got CreateAssetRequest
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got.Name != "Menu" || got.Ext != ".png" || got.PlayOrder == nil || *got.PlayOrder != 4 {
				t.Fatalf("create body = %#v", got)
			}
			writeJSON(t, w, http.StatusCreated, testAsset("asset-1"))
		case 1:
			assertMethodPath(t, r, http.MethodPatch, "/api/v2/assets/asset-1")
			var got UpdateAssetRequest
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got.Name == nil || *got.Name != "Updated" || got.IsEnabled == nil || *got.IsEnabled {
				t.Fatalf("update body = %#v", got)
			}
			writeJSON(t, w, http.StatusOK, testAsset("asset-1"))
		case 2:
			assertMethodPath(t, r, http.MethodPut, "/api/v2/assets/asset-1")
			var got CreateAssetRequest
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got.URI != "https://example.com/replacement.png" || got.Duration != 20 {
				t.Fatalf("replace body = %#v", got)
			}
			writeJSON(t, w, http.StatusOK, testAsset("asset-1"))
		default:
			t.Fatalf("unexpected call %d", calls)
		}
		calls++
	}))

	createReq := CreateAssetRequest{
		Name:      "Menu",
		URI:       "https://example.com/menu.png",
		Ext:       ".png",
		StartDate: testStart,
		EndDate:   testEnd,
		Duration:  10,
		Mimetype:  "image/png",
		IsEnabled: true,
		NoCache:   Ptr(true),
		PlayOrder: Ptr(4),
	}
	if _, err := c.CreateAsset(context.Background(), createReq); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UpdateAsset(context.Background(), "asset-1", UpdateAssetRequest{
		Name:      Ptr("Updated"),
		IsEnabled: Ptr(false),
	}); err != nil {
		t.Fatal(err)
	}
	replaceReq := createReq
	replaceReq.URI = "https://example.com/replacement.png"
	replaceReq.Duration = 20
	if _, err := c.ReplaceAsset(context.Background(), "asset-1", replaceReq); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}

func TestAssetActions(t *testing.T) {
	calls := 0
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertCommonHeaders(t, r)
		switch calls {
		case 0:
			assertMethodPath(t, r, http.MethodDelete, "/api/v2/assets/asset-1")
			w.WriteHeader(http.StatusNoContent)
		case 1:
			assertMethodPath(t, r, http.MethodPost, "/api/v2/assets/order")
			assertContentType(t, r, "application/x-www-form-urlencoded")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			vals, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatal(err)
			}
			if vals.Get("ids") != "asset-2,asset-1" {
				t.Fatalf("ids = %q", vals.Get("ids"))
			}
			w.WriteHeader(http.StatusNoContent)
		case 2:
			assertMethodPath(t, r, http.MethodGet, "/api/v2/assets/control/asset&asset-1")
			w.WriteHeader(http.StatusNoContent)
		case 3:
			assertMethodPath(t, r, http.MethodGet, "/api/v2/assets/asset-1/content")
			writeJSON(t, w, http.StatusOK, AssetContent{
				Type:     "file",
				Filename: "menu.png",
				Mimetype: "image/png",
				Content:  "ZmFrZQ==",
			})
		default:
			t.Fatalf("unexpected call %d", calls)
		}
		calls++
	}))

	if err := c.DeleteAsset(context.Background(), "asset-1"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetPlaylistOrder(context.Background(), []string{"asset-2", "asset-1"}); err != nil {
		t.Fatal(err)
	}
	if err := c.ControlPlayback(context.Background(), PlaybackAsset("asset-1")); err != nil {
		t.Fatal(err)
	}
	content, err := c.GetAssetContent(context.Background(), "asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if content.Type != "file" || !strings.HasPrefix(content.Content, "Zm") {
		t.Fatalf("content = %#v", content)
	}
	if calls != 4 {
		t.Fatalf("calls = %d, want 4", calls)
	}
}
