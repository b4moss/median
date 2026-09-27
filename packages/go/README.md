# packages/go

Go module **`github.com/b4moss/median/go`**（現行 `packages/go/VERSION` = **0.4.0**、タグ `packages/go/v0.4.0`）。

## できること

- `Store` / `Delete` / `Get` / `PresignGet`
- Local FS・S3 Adapter
- 任意 DB（crudian）と media 行・variants
- 画像パイプライン（resize / compress / thumbnails）
- SVG サニタイズ、PDF 先頭ページサムネ（`PDFRenderer` 注入、本番は go-fitz / CGO）

## レイアウト

| パス | 役割 |
| --- | --- |
| `core` | Median API・Config |
| `storage` / `local` / `s3` | Adapter |
| `db` | migrate・MediaRepo |
| `pipeline` | 画像・SVG・PDF |
| `internal` | MIME・ファイル名・shardian path |
| `third_party/crudian` | vendored crudian（`go.mod` の `replace`） |

## テスト

```bash
go test ./...
```

仕様: [docs/specs/v0.2.0/](../../docs/specs/v0.2.0/)〜[v0.4.0/](../../docs/specs/v0.4.0/)  
テスト仕様: [docs/tests/v0.2.0.md](../../docs/tests/v0.2.0.md)〜[v0.4.0.md](../../docs/tests/v0.4.0.md)
