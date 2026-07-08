# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-07-08

### Added

- Client for the Anthias v2 REST API with functional options: `WithHTTPClient`, `WithTimeout`, `WithBasicAuth`, `WithUserAgent`.
- Asset operations: list, get, create, update, replace, delete, playlist ordering, playback control, and asset content retrieval.
- Streaming file uploads with explicit `Content-Length` and an optional progress callback.
- Device operations: settings get and update, backup, recover, reboot, and shutdown.
- System endpoints: info and integrations.
- Typed `APIError` for non-2xx responses.

[Unreleased]: https://github.com/atakanatamert/anthias-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/atakanatamert/anthias-go/releases/tag/v0.1.0
