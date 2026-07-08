# Contributing

Thanks for helping improve `anthias-go`.

## Workflow

1. Fork the repository.
2. Create a focused branch from `main`.
3. Keep changes small and focused.
4. Run `gofmt` on all Go changes.
5. Run the required checks before opening a pull request:

```sh
go vet ./...
go build ./...
go test -race -cover ./...
```

## Commits

Use Conventional Commits, for example:

```text
feat: add asset list endpoint
fix: preserve response body in APIError
docs: clarify basic auth option
```

## Pull Requests

Include a concise summary, note any API behavior changes, and confirm the checks above pass. Do not add third-party dependencies; this SDK is stdlib-only.
