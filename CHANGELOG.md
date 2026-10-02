# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-10-02

Compatibility release for Anthias v2026.09.0; covers every endpoint of its v2 API.

### Changed

- **Breaking:** `ReplaceAsset` takes a new `ReplaceAssetRequest`. The player ignores `uri`, `mimetype` and `ext` on `PUT /api/v2/assets/{id}`, so those fields are gone instead of being silently dropped.
- **Breaking:** removed `UpdateAssetRequest.URI` and `UpdateAssetRequest.Mimetype`; the player ignores them on `PATCH`, so the update reported success but changed nothing.
- The client no longer follows HTTP redirects. With authentication enabled, Anthias redirects requests without credentials to its HTML login page; following that made write calls (`DeleteAsset`, `Reboot`, `SetPlaylistOrder`, …) report success without doing anything, and read calls fail with a JSON decode error. Clients passed via `WithHTTPClient` are copied, not modified.

### Added

- `APIError.IsUnauthorized` (401, or a redirect to the login page), `APIError.Location`, and `APIError.Message` for the player's `error`/`message`/`detail` text.
- `SetDisplayPower` (HDMI-CEC) and `GetIPAddresses` (unauthenticated).
- Resumable chunked uploads: `UploadFileChunked`, `UploadFileReaderChunked`, `WithChunkSize`, `WithResume`, `ChunkedUploadError`.
- Content import and Screenly migration: `ValidateImportToken`, `ImportItem` (`ImportYodeck`, `ImportScreenCloud`, `ImportPiSignage`, `ImportXibo`), `ValidateScreenlyToken`, `MigrateAssetToScreenly`.
- Player-internal endpoints behind `WithInternalSecret`: `GetViewerPlaylist`, `GetViewerSettings`, `RecheckAsset`.
- Asset scheduling and webpage fields: `SkipSSLVerify`, `PlayDays`, `PlayTimeFrom`, `PlayTimeTo`, `RefreshIntervalS`, `CustomHeaders`; read-only `IsReachable`, `LastReachabilityCheck`, `Metadata`.
- `UpdateAssetRequest.PlayOrder`, `ClearPlayTimeWindow` and `ClearCustomHeaders`.
- Device settings: `Timezone`, `PreferDarkMode`, `VerifySSL`, `ScreenRotation` and the display power schedule (`DisplayPowerScheduleEnabled`, `DisplayPowerOnTime`, `DisplayPowerOffTime`, `DisplayPowerDays`).
- Info: `Storage` health, device `Time`, `UnderVoltage`, `Memory.LowRAM`.
- `FileUpload.UploadID`.

### Fixed

- `examples/basic` and the README upload example uploaded a `.txt` file, which the player now rejects (only images and videos are accepted).

## [0.1.0] - 2026-07-08

### Added

- Client for the Anthias v2 REST API with functional options: `WithHTTPClient`, `WithTimeout`, `WithBasicAuth`, `WithUserAgent`.
- Asset operations: list, get, create, update, replace, delete, playlist ordering, playback control, and asset content retrieval.
- Streaming file uploads with explicit `Content-Length` and an optional progress callback.
- Device operations: settings get and update, backup, recover, reboot, and shutdown.
- System endpoints: info and integrations.
- Typed `APIError` for non-2xx responses.

[Unreleased]: https://github.com/atakanatamert/anthias-go/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/atakanatamert/anthias-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/atakanatamert/anthias-go/releases/tag/v0.1.0
