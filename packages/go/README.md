# packages/go

Go module **`github.com/b4moss/median/packages/go`** (current `packages/go/VERSION` = **0.9.0**, tag `packages/go/v0.9.0`).

Module path matches the repo subdirectory and tag prefix (`packages/go/vX.Y.Z`) so `proxy.golang.org` can resolve versions.

## CLI

```bash
go run github.com/b4moss/median/packages/go/cmd/median@v0.9.0 \
  migrate dump --dialect postgres --id-strategy auto_increment --table media
```

SQL goes to stdout (redirect into the host app's migrations). See [docs/specs/v0.9.0/](../../docs/specs/v0.9.0/).

## Features

- `Store` / `Delete` / `Get` / `PresignGet`
- Local FS and S3 adapters
- Optional DB (crudian) with media rows and variants
- Image pipeline (resize / compress / thumbnails)
- SVG sanitize; PDF first-page thumbnails (`PDFRenderer` injection; production default go-fitz / CGO)

## Layout

| Path | Role |
| --- | --- |
| `core` | Median API and Config |
| `storage` / `local` / `s3` | Adapters |
| `db` | migrate helpers and MediaRepo |
| `cmd/median` | CLI (`migrate dump`) |
| `pipeline` | Image / SVG / PDF |
| `internal` | MIME, filename, shardian path |
| `third_party/crudian` | Vendored crudian (`replace` in `go.mod`) |

## Test

```bash
go test ./...
```

Specs: [docs/specs/v0.2.0/](../../docs/specs/v0.2.0/)–[v0.4.0/](../../docs/specs/v0.4.0/), [v0.7.0/](../../docs/specs/v0.7.0/)–[v0.9.0/](../../docs/specs/v0.9.0/)  
Acceptance tests: [docs/tests/v0.2.0.md](../../docs/tests/v0.2.0.md)–[v0.4.0.md](../../docs/tests/v0.4.0.md), [v0.7.0.md](../../docs/tests/v0.7.0.md)–[v0.9.0.md](../../docs/tests/v0.9.0.md)  
E2E (shared): [e2e/](../../e2e/) / [docs/specs/v0.6.0/](../../docs/specs/v0.6.0/)

Repository hub: [README.md](../../README.md) / [README-ja.md](../../README-ja.md)
