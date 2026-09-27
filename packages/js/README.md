# @b4moss/median

TypeScript implementation (`packages/js`). npm: **[@b4moss/median](https://www.npmjs.com/package/@b4moss/median)** (current **0.9.0**, git tag `v0.9.0`).

Same contract vocabulary as the Go package (`packages/go`): Store / Delete / Get / PresignGet, Local and S3, optional DB, image pipeline, SVG (svgo), and PDF first-page thumbnails (`PDFRenderer` injection).

## Requirements

- Node.js **24+**

## Install

```bash
npm install @b4moss/median
```

## CLI

```bash
npx @b4moss/median migrate dump --dialect postgres --id-strategy auto_increment --table media
```

SQL goes to stdout. See [docs/specs/ddl-cli/](../../docs/specs/ddl-cli/).

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
| `src/cli.ts` | CLI (`bin`: `median` → `migrate dump`) |
| `src/pipeline` | process / sanitizeSvg / PDF thumbs |
| `src/internal` | MIME / filename / shardian path |

## Docs

- Specs: [docs/specs/packages-js/](../../docs/specs/packages-js/), [db-media](../../docs/specs/db-media/), [ddl-cli](../../docs/specs/ddl-cli/)
- Acceptance tests: [docs/tests/packages-js/packages-js.md](../../docs/tests/packages-js/packages-js.md), [db-media](../../docs/tests/db-media/), [ddl-cli](../../docs/tests/ddl-cli/)
- E2E (shared): [e2e/](../../e2e/) / [docs/specs/e2e/](../../docs/specs/e2e/)
- CI / publish: [.github/CI.md](../../.github/CI.md) (`release` + Trusted Publisher)
- Repository hub: [README.md](../../README.md) / [README-ja.md](../../README-ja.md)
