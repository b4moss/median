# packages/go

Go module **`github.com/b4moss/median/go`** (current `packages/go/VERSION` = **0.4.0**, tag `packages/go/v0.4.0`).

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
| `pipeline` | Image / SVG / PDF |
| `internal` | MIME, filename, shardian path |
| `third_party/crudian` | Vendored crudian (`replace` in `go.mod`) |

## Test

```bash
go test ./...
```

Specs: [docs/specs/v0.2.0/](../../docs/specs/v0.2.0/)–[v0.4.0/](../../docs/specs/v0.4.0/)  
Acceptance tests: [docs/tests/v0.2.0.md](../../docs/tests/v0.2.0.md)–[v0.4.0.md](../../docs/tests/v0.4.0.md)

Repository hub: [README.md](../../README.md) / [README-ja.md](../../README-ja.md)
