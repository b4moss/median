# @b4moss/median

TypeScript 実装（`packages/js`）。npm: **[@b4moss/median](https://www.npmjs.com/package/@b4moss/median)**（現行 **0.5.0**、Git タグ `v0.5.0`）。

Go 版（`packages/go`）と同じ契約語彙で Store / Delete / Get / PresignGet、Local・S3、任意 DB、画像パイプライン、SVG（svgo）、PDF 先頭ページサムネ（`PDFRenderer` 注入）を提供する。

## 要件

- Node.js **24+**

## インストール

```bash
npm install @b4moss/median
```

## 開発

```bash
npm ci
npm test
npm run test:coverage
npm run build
```

## レイアウト

| パス | 役割 |
| --- | --- |
| `src/core` | `createMedian` / store / delete / get / presignGet |
| `src/storage` | local / s3 |
| `src/db` | migrate / MediaRepo |
| `src/pipeline` | process / sanitizeSvg / PDF thumbs |
| `src/internal` | MIME / filename / shardian path |

## ドキュメント

- 仕様: [docs/specs/v0.5.0/](../../docs/specs/v0.5.0/)
- テスト仕様: [docs/tests/v0.5.0.md](../../docs/tests/v0.5.0.md)
- CI / 公開: [.github/CI.md](../../.github/CI.md)（`release` + Trusted Publisher）
