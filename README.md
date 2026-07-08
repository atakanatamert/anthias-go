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

## Upload With Progress

Uploads stream the body and set `Content-Length`; pass a generous context timeout for large files.

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
defer cancel()

upload, err := client.UploadFileReader(
	ctx,
	strings.NewReader("hello"),
	"hello.txt",
	5,
	anthias.WithProgress(func(sent, total int64) {
		_ = float64(sent) / float64(total)
	}),
)
if err != nil {
	return err
}
_ = upload
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
| `GetDeviceSettings(ctx)` | `GET /api/v2/device_settings` |
| `UpdateDeviceSettings(ctx, req)` | `PATCH /api/v2/device_settings` |
| `Backup(ctx)` | `POST /api/v2/backup` |
| `Recover(ctx, r, filename, size)` | `POST /api/v2/recover` |
| `Reboot(ctx)` | `POST /api/v2/reboot` |
| `Shutdown(ctx)` | `POST /api/v2/shutdown` |
| `GetInfo(ctx)` | `GET /api/v2/info` |
| `GetIntegrations(ctx)` | `GET /api/v2/integrations` |

## Compatibility

`anthias-go` targets the Anthias v2 API and uses only the Go standard library.

## Acknowledgements

[Anthias](https://github.com/Screenly/Anthias) is an open source digital signage platform maintained by [Screenly](https://www.screenly.io/). This project is an independent client library and is not affiliated with or endorsed by Screenly.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). This project is released under the [MIT License](LICENSE).
