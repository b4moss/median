# packages/go

Go module `github.com/b4moss/median/go`.

## Layout

- `core` — Store / Delete / Get / PresignGet
- `storage` / `storage/local` / `storage/s3` — adapters
- `db` — migrations helpers + MediaRepo (crudian)
- `pipeline` — resize / compress / thumbnails / SVG sanitize / PDF page render (go-fitz, CGO)
- `third_party/crudian` — vendored crudian Go module (`replace` in go.mod; upstream tag/path mismatch)

## Test

```bash
go test ./...
```
