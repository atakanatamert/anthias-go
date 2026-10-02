# anthias-go

[![CI](https://github.com/atakanatamert/anthias-go/actions/workflows/ci.yml/badge.svg)](https://github.com/atakanatamert/anthias-go/actions/workflows/ci.yml)

Go SDK for the [Anthias](https://github.com/Screenly/Anthias) digital signage player v2 REST API.

## Install

```sh
go get github.com/atakanatamert/anthias-go
```

## Quickstart

```go
ctx := context.Background()

client, err := anthias.New("http://192.168.1.50")
if err != nil {
	return err
}

assets, err := client.ListAssets(ctx)
if err != nil {
	return err
}
_ = assets
```

## Authentication

Use HTTP Basic Auth when the player has the `auth_basic` backend enabled.

```go
client, err := anthias.New(
	"http://192.168.1.50",
	anthias.WithBasicAuth("admin", "password"),
)
```

Anthias still accepts Basic Auth but logs it as deprecated. Missing credentials are answered with a redirect to the player's login page and wrong ones with `401`; the client never follows redirects, so both surface as an `*APIError`:

```go
var apiErr *anthias.APIError
if errors.As(err, &apiErr) && apiErr.IsUnauthorized() {
	// configure or fix credentials
}
```

## Upload With Progress

The player accepts only image and video uploads, judged by the file name's extension. Uploads stream the body and set `Content-Length`; pass a generous context timeout for large files.

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
defer cancel()

upload, err := client.UploadFile(
	ctx,
	"/path/to/menu.png",
	anthias.WithProgress(func(sent, total int64) {
		_ = float64(sent) / float64(total)
	}),
)
if err != nil {
	return err
}
_ = upload
```

### Resumable uploads

For large videos or unreliable links, `UploadFileChunked` sends the file in chunks (8 MiB by default). If a chunk fails, the error says where to continue:

```go
upload, err := client.UploadFileChunked(ctx, "/path/to/video.mp4")
var cerr *anthias.ChunkedUploadError
if errors.As(err, &cerr) {
	upload, err = client.UploadFileChunked(ctx, "/path/to/video.mp4",
		anthias.WithResume(cerr.UploadID, cerr.Offset))
}
```

## Editing Assets

An asset's `uri` and `mimetype` are fixed when it is created; to change them, create a new asset. `UpdateAsset` sends only non-nil fields; use `ClearPlayTimeWindow` / `ClearCustomHeaders` to remove a daily play window or custom headers.

```go
_, err := client.UpdateAsset(ctx, assetID, anthias.UpdateAssetRequest{
	PlayDays:            []int{1, 2, 3, 4, 5}, // Monday–Friday
	ClearPlayTimeWindow: true,
})
```

## Playback Control

```go
if err := client.ControlPlayback(ctx, anthias.PlaybackNext); err != nil {
	return err
}

if err := client.ControlPlayback(ctx, anthias.PlaybackAsset("asset-id")); err != nil {
	return err
}
```

## Display, Network and Integrations

```go
// HDMI-CEC; a 503 *APIError means the player has no CEC adapter.
msg, err := client.SetDisplayPower(ctx, false)

// No credentials needed; returns URLs such as "http://10.0.0.108".
urls, err := client.GetIPAddresses(ctx)

// Import media from another signage platform (Yodeck, ScreenCloud, piSignage, Xibo).
v, err := client.ValidateImportToken(ctx, anthias.ImportYodeck, token)
for _, item := range v.Items {
	if item.Importable {
		res, err := client.ImportItem(ctx, anthias.ImportYodeck, anthias.ImportItemRequest{Token: token, RemoteID: item.RemoteID})
		_, _ = res, err
	}
}
```

Per-item outcomes (`Valid`, `Success`, `Skipped`) come back as results; transport failures and unknown providers or assets are `*APIError`s whose `Message()` carries the player's explanation.

## Player-Internal Endpoints

`GetViewerPlaylist`, `GetViewerSettings` and `RecheckAsset` use the player's internal token instead of user credentials. Pass the player's `django_secret_key` from `~/.anthias/anthias.conf` on the device. It is Django's secret key, so keep it private. Anthias builds these endpoints for its own viewer, so they may change between player releases with less notice.

```go
client, err := anthias.New("http://192.168.1.50", anthias.WithInternalSecret(secret))
playlist, err := client.GetViewerPlaylist(ctx) // active assets in play order + next change
```

## Methods

| Method | Anthias endpoint |
|---|---|
| `ListAssets(ctx)` | `GET /api/v2/assets` |
| `GetAsset(ctx, assetID)` | `GET /api/v2/assets/{id}` |
| `CreateAsset(ctx, req)` | `POST /api/v2/assets` |
| `UpdateAsset(ctx, assetID, req)` | `PATCH /api/v2/assets/{id}` |
| `ReplaceAsset(ctx, assetID, req)` | `PUT /api/v2/assets/{id}` |
| `DeleteAsset(ctx, assetID)` | `DELETE /api/v2/assets/{id}` |
| `SetPlaylistOrder(ctx, assetIDs)` | `POST /api/v2/assets/order` |
| `ControlPlayback(ctx, command)` | `GET /api/v2/assets/control/{command}` |
| `GetAssetContent(ctx, assetID)` | `GET /api/v2/assets/{id}/content` |
| `UploadFile(ctx, path, opts...)` | `POST /api/v2/file_asset` |
| `UploadFileReader(ctx, r, filename, size, opts...)` | `POST /api/v2/file_asset` |
| `UploadFileChunked(ctx, path, opts...)` | `POST /api/v2/file_asset` (chunked, resumable) |
| `UploadFileReaderChunked(ctx, r, filename, size, opts...)` | `POST /api/v2/file_asset` (chunked, resumable) |
| `GetDeviceSettings(ctx)` | `GET /api/v2/device_settings` |
| `UpdateDeviceSettings(ctx, req)` | `PATCH /api/v2/device_settings` |
| `Backup(ctx)` | `POST /api/v2/backup` |
| `Recover(ctx, r, filename, size)` | `POST /api/v2/recover` |
| `Reboot(ctx)` | `POST /api/v2/reboot` |
| `Shutdown(ctx)` | `POST /api/v2/shutdown` |
| `GetInfo(ctx)` | `GET /api/v2/info` |
| `GetIntegrations(ctx)` | `GET /api/v2/integrations` |
| `SetDisplayPower(ctx, on)` | `POST /api/v2/display/{on,off}` |
| `GetIPAddresses(ctx)` | `GET /api/v2/network/ip-addresses` |
| `ValidateScreenlyToken(ctx, token)` | `POST /api/v2/integrations/screenly/validate` |
| `MigrateAssetToScreenly(ctx, req)` | `POST /api/v2/integrations/screenly/migrate` |
| `ValidateImportToken(ctx, provider, token)` | `POST /api/v2/integrations/import/{provider}/validate` |
| `ImportItem(ctx, provider, req)` | `POST /api/v2/integrations/import/{provider}/item` |
| `GetViewerPlaylist(ctx)` | `GET /api/v2/viewer/playlist` (internal) |
| `GetViewerSettings(ctx)` | `GET /api/v2/viewer/settings` (internal) |
| `RecheckAsset(ctx, assetID)` | `POST /api/v2/assets/{id}/recheck` (internal) |

## Compatibility

`anthias-go` targets the Anthias v2 API (schema 2.0.0) and uses only the Go standard library. It covers every endpoint of Anthias v2026.09.0 and was verified against a live v2026.09.0 player.

Against older players, everything the player lacks degrades instead of breaking. Missing response fields decode as zero values, and calling an endpoint the player doesn't have returns a 404 `*APIError`. Minimum player versions, from the published API schemas:

| Feature | Minimum Anthias version |
|---|---|
| Device settings, info, integrations | v0.20.0 |
| Asset scheduling (`PlayDays`, play window, `RefreshIntervalS`), reachability fields, `GetIPAddresses`, `RecheckAsset` | v2026.05.0 |
| `SetDisplayPower`, viewer endpoints, Screenly migration, `ScreenRotation`, `Memory.LowRAM` | v2026.05.1 |
| `PreferDarkMode` | v2026.07.0 |
| Resuming chunked uploads (`UploadID`), `SkipSSLVerify`, `CustomHeaders`, content import, `Timezone`, `VerifySSL`, `Info.Time` | v2026.07.2 |
| Display power schedule (`DisplayPower*` settings) | v2026.08.1 |
| `Info.Storage`, `Info.UnderVoltage` | v2026.08.2 |

When adding fields to requests, prefer setting only what you need: older players ignore request fields they don't know. They don't reject them, so a setting sent to a player that predates it is silently not applied.

## Acknowledgements

[Anthias](https://github.com/Screenly/Anthias) is an open source digital signage platform maintained by [Screenly](https://www.screenly.io/). This project is an independent client library and is not affiliated with or endorsed by Screenly.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). This project is released under the [MIT License](LICENSE).
