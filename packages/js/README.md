# @b4moss/median

TypeScript implementation (`packages/js`). npm: **[@b4moss/median](https://www.npmjs.com/package/@b4moss/median)** (current **0.5.0**, git tag `v0.5.0`).

Same contract vocabulary as the Go package (`packages/go`): Store / Delete / Get / PresignGet, Local and S3, optional DB, image pipeline, SVG (svgo), and PDF first-page thumbnails (`PDFRenderer` injection).

## Requirements

- Node.js **24+**

## Install

```bash
npm install @b4moss/median
```

## Develop

```bash
npm ci
npm test
npm run test:coverage
npm run build
```

## Layout

| Path | Role |
| --- | --- |
| `src/core` | `createMedian` / store / delete / get / presignGet |
| `src/storage` | local / s3 |
| `src/db` | migrate / MediaRepo |
| `src/pipeline` | process / sanitizeSvg / PDF thumbs |
| `src/internal` | MIME / filename / shardian path |

## Docs

- Specs: [docs/specs/v0.5.0/](../../docs/specs/v0.5.0/)
- Acceptance tests: [docs/tests/v0.5.0.md](../../docs/tests/v0.5.0.md)
- CI / publish: [.github/CI.md](../../.github/CI.md) (`release` + Trusted Publisher)
- Repository hub: [README.md](../../README.md) / [README-ja.md](../../README-ja.md)
