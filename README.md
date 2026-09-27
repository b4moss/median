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

Versions are **independent per language**. A planning milestone name does not force every package to share that number (for example historically Go `0.4.0` vs npm `0.5.0`). The current DDL CLI line ships Go and npm both at **`0.9.0`**.

## TypeScript / JavaScript

npm package: **[@b4moss/median](https://www.npmjs.com/package/@b4moss/median)** (`packages/js`)

Current line: see `packages/js/package.json` (currently **`0.9.0`**). Requires **Node.js 24+**.

```bash
npm install @b4moss/median
```

```bash
npx @b4moss/median migrate dump --dialect postgres --table media
```

```bash
cd packages/js
npm ci
npm test
npm run build
```

API notes: [`packages/js/README.md`](./packages/js/README.md).  
Acceptance tests: [`docs/tests/ddl-cli/ddl-cli.md`](./docs/tests/ddl-cli/ddl-cli.md). Specs: [`docs/specs/ddl-cli/`](./docs/specs/ddl-cli/).

**Release:** root git tag `vX.Y.Z` must match `package.json` `version`. Pushing that tree to `release` publishes to npm when packed content differs from the registry ([`.github/CI.md`](./.github/CI.md)). Trusted Publishing is configured for GitHub Actions.

## Go

Go module: **[github.com/b4moss/median/packages/go](./packages/go)** (`packages/go`)

Current line: `packages/go/VERSION` → **`0.9.0`**, git tag **`packages/go/v0.9.0`**.

```bash
go get github.com/b4moss/median/packages/go@v0.9.0
```

```bash
go run github.com/b4moss/median/packages/go/cmd/median@v0.9.0 \
  migrate dump --dialect postgres --table media
```

```bash
cd packages/go
go test ./...
```

| Import path | Role |
|-------------|------|
| `github.com/b4moss/median/packages/go/core` | `New` / Store / Delete / Get / PresignGet |
| `github.com/b4moss/median/packages/go/storage/...` | Local / S3 adapters |
| `github.com/b4moss/median/packages/go/db` | Migrations helpers + MediaRepo |
| `github.com/b4moss/median/packages/go/pipeline` | Image / SVG / PDF pipeline |
| `github.com/b4moss/median/packages/go/cmd/median` | CLI (`migrate dump`) |

Usage: [`packages/go/README.md`](./packages/go/README.md).  
Specs: [`docs/specs/`](./docs/specs/) (domain-oriented, incl. [`ddl-cli`](./docs/specs/ddl-cli/)). Tests: [`docs/tests/`](./docs/tests/).

**Release:** tag `packages/go/vX.Y.Z` matching `VERSION`. Root tag `vX.Y.Z` alone does **not** publish Go.

## PHP

Not scheduled yet (`unscheduled`). Planned under `packages/php` — see [`docs/plans/unscheduled/packages-php.md`](./docs/plans/unscheduled/packages-php.md).

## E2E

Shipped milestone **`v0.6.0`**: RustFS + POSIX LocalFS, CRUD+DB meta, thumbnails (Go and JS). Not in regular CI — run `./e2e/run.sh` locally (especially before version bumps) or trigger [`.github/workflows/e2e.yml`](./.github/workflows/e2e.yml) via `workflow_dispatch`. Specs: [`docs/specs/e2e/`](./docs/specs/e2e/).

## Docs

| Doc | Contents |
|-----|----------|
| [`docs/README.md`](./docs/README.md) | Product pillar (purpose, scope, tech policy) |
| [`docs/roadmap.md`](./docs/roadmap.md) | Milestones |
| [`docs/specs/`](./docs/specs/) | Current specs (domain-oriented) |
| [`docs/plans/`](./docs/plans/) | Upcoming plans (PHP unscheduled) |
| [`docs/tests/`](./docs/tests/) | TDD acceptance specs |
| [`docs/charter/`](./docs/charter/) | Charter (Git, SemVer, TDD, …) |
| [`e2e/`](./e2e/) | E2E harness (RustFS + LocalFS) |
| [`.github/CI.md`](./.github/CI.md) | CI/CD policy |

## License

MIT © Bicycle for Mind LLC. — see [`LICENSE`](./LICENSE).
