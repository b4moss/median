# median

[![CI](https://github.com/b4moss/median/actions/workflows/ci.yml/badge.svg)](https://github.com/b4moss/median/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/b4moss/median)](https://codecov.io/gh/b4moss/median)
[![npm](https://img.shields.io/npm/v/@b4moss/median)](https://www.npmjs.com/package/@b4moss/median)
[![Release](https://img.shields.io/github/v/release/b4moss/median)](https://github.com/b4moss/median/releases)
[![License](https://img.shields.io/github/license/b4moss/median)](https://github.com/b4moss/median/blob/main/LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/b4moss/median/badge)](https://scorecard.dev/viewer/?uri=github.com/b4moss/median)

DDD-oriented media store: put bytes in storage and get them back under one contract across languages.

- [日本語版 README](./README-ja.md)

Core operations are **Store / Delete / Get** (plus **PresignGet** where supported), with Local FS and S3 adapters, optional media DB metadata, and an image pipeline (compress / resize / thumbnails, SVG sanitize, PDF first-page thumbnails).

Versions are **independent per language**. A planning milestone name does not force every package to share that number (for example Go may sit on `0.4.0` while npm is `0.5.0`).

## TypeScript / JavaScript

npm package: **[@b4moss/median](https://www.npmjs.com/package/@b4moss/median)** (`packages/js`)

Current line: see `packages/js/package.json` (currently **`0.5.0`**). Requires **Node.js 24+**.

```bash
npm install @b4moss/median
```

```bash
cd packages/js
npm ci
npm test
npm run build
```

API notes: [`packages/js/README.md`](./packages/js/README.md).  
Acceptance tests: [`docs/tests/v0.5.0.md`](./docs/tests/v0.5.0.md). Specs: [`docs/specs/v0.5.0/`](./docs/specs/v0.5.0/).

**Release:** root git tag `vX.Y.Z` must match `package.json` `version`. Pushing that tree to `release` publishes to npm when packed content differs from the registry ([`.github/CI.md`](./.github/CI.md)). Trusted Publishing is configured for GitHub Actions.

## Go

Go module: **[github.com/b4moss/median/go](./packages/go)** (`packages/go`)

Current line: `packages/go/VERSION` → **`0.4.0`**, git tag **`packages/go/v0.4.0`**.

```bash
go get github.com/b4moss/median/go@v0.4.0
```

```bash
cd packages/go
go test ./...
```

| Import path | Role |
|-------------|------|
| `github.com/b4moss/median/go/core` | `New` / Store / Delete / Get / PresignGet |
| `github.com/b4moss/median/go/storage/...` | Local / S3 adapters |
| `github.com/b4moss/median/go/db` | Migrations helpers + MediaRepo |
| `github.com/b4moss/median/go/pipeline` | Image / SVG / PDF pipeline |

Usage: [`packages/go/README.md`](./packages/go/README.md).  
Specs: [`docs/specs/v0.2.0/`](./docs/specs/v0.2.0/)–[`v0.4.0/`](./docs/specs/v0.4.0/). Tests: [`docs/tests/`](./docs/tests/).

**Release:** tag `packages/go/vX.Y.Z` matching `VERSION`. Root tag `vX.Y.Z` alone does **not** publish Go.

## PHP

Not scheduled yet (`unscheduled`). Planned under `packages/php` — see [`docs/plans/unscheduled/packages-php.md`](./docs/plans/unscheduled/packages-php.md).

Next milestone is **E2E tests** (`v0.6.0`: RustFS + POSIX LocalFS, Go and JS; not in regular CI) — [`docs/plans/v0.6.0/`](./docs/plans/v0.6.0/).

## Docs

| Doc | Contents |
|-----|----------|
| [`docs/main.md`](./docs/main.md) | Product hub (purpose, scope, tech policy) |
| [`docs/roadmap.md`](./docs/roadmap.md) | Milestones |
| [`docs/specs/`](./docs/specs/) | Shipped specs (v0.1–v0.5) |
| [`docs/plans/`](./docs/plans/) | Upcoming plans (`v0.6.0` E2E; PHP unscheduled) |
| [`docs/tests/`](./docs/tests/) | TDD acceptance specs |
| [`docs/charter/`](./docs/charter/) | Charter (Git, SemVer, TDD, …) |
| [`.github/CI.md`](./.github/CI.md) | CI/CD policy |

## License

MIT © Bicycle for Mind LLC. — see [`LICENSE`](./LICENSE).
