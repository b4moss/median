# median

[![CI](https://github.com/b4moss/median/actions/workflows/ci.yml/badge.svg)](https://github.com/b4moss/median/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/b4moss/median)](https://codecov.io/gh/b4moss/median)
[![npm](https://img.shields.io/npm/v/@b4moss/median)](https://www.npmjs.com/package/@b4moss/median)
[![Release](https://img.shields.io/github/v/release/b4moss/median)](https://github.com/b4moss/median/releases)
[![License](https://img.shields.io/github/license/b4moss/median)](https://github.com/b4moss/median/blob/main/LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/b4moss/median/badge)](https://scorecard.dev/viewer/?uri=github.com/b4moss/median)

DDD-oriented media store: put bytes in storage and get them back under **one contract** across languages.

- [日本語版 README](./README-ja.md)

## What you get

Core operations are **Store / Delete / Get** (plus **PresignGet** where supported), with:

- **Local FS** and **S3-compatible** adapters
- Optional **media DB** metadata (table name + column map)
- Image pipeline (compress / resize / thumbnails), **SVG sanitize**, **PDF** first-page thumbnails (renderer injection)

Versions are **independent per language**. Check the current line in each port’s README / `VERSION` / `package.json`. Today Go and npm both ship **`0.9.0`** (DDL CLI).

## Choose a language port

| Port | Package | Language hub |
|------|---------|----------------|
| **Go** | `github.com/b4moss/median/packages/go` | [`packages/go/README.md`](./packages/go/README.md) |
| **TypeScript** | `@b4moss/median` | [`packages/js/README.md`](./packages/js/README.md) |
| **PHP** | — | Not scheduled — [`docs/plans/unscheduled/packages-php.md`](./docs/plans/unscheduled/packages-php.md) |

Install commands, import paths, Config examples, and CLI flags live in the **language hub**. This root README is the shared host workflow.

---

## Host usage (step by step)

This is the recommended production path for a new app that owns its own migrations.

### Step 1 — Pick a port and install

- Go: see [`packages/go/README.md`](./packages/go/README.md) (`go get …/packages/go@v0.9.0`)
- JS: see [`packages/js/README.md`](./packages/js/README.md) (`npm install @b4moss/median`)

### Step 2 — Decide table name and ID strategy

Defaults:

| Setting | Default |
|---------|---------|
| Table | `media` |
| ID strategy | `auto_increment` (`uuid_v4` / `uuid_v7` / `ulid` also supported) |
| Dialect | one of `mysql` \| `postgres` \| `sqlite3` |

You will use the **same values** for the DDL dump and for runtime config.

### Step 3 — Dump DDL into *your* migrations

median does **not** run your production migrate for you. Generate SQL and commit it to the host app’s migrate tool (goose, Prisma, Flyway, knex, …).

```bash
# Go
go run github.com/b4moss/median/packages/go/cmd/median@v0.9.0 \
  migrate dump --dialect postgres --id-strategy auto_increment --table media \
  > migrations/XXXX_median_media.sql

# JS
npx @b4moss/median migrate dump \
  --dialect postgres --id-strategy auto_increment --table media \
  > migrations/XXXX_median_media.sql
```

- **stdout** = SQL only (safe to redirect)
- **stderr** = errors / usage
- Spec: [`docs/specs/ddl-cli/`](./docs/specs/ddl-cli/)

Then run your usual `migrate up` in CI / deploy.

**Existing table** with different column names? Skip the dump. Map columns in Config instead ([`docs/specs/db-media/`](./docs/specs/db-media/)).

### Step 4 — Apply migrations, then start the app

Order matters:

1. Host migrate applies the media table  
2. App boots and constructs median with the **same** `table` / `id-strategy` / dialect assumptions  
3. Call Store / Get / Delete

Language-specific `New` / `createMedian` wiring: Go and JS hubs above.

### Step 5 — Store → Get → Delete

Conceptually (pseudo):

```text
Store(bytes, size, { mime, filename, … })
  → writes object to storage
  → optionally inserts a media row (file_path / file_name / …)
  → returns { path, hash, id?, variants?, … }

Get(path)
  → reads object (and optional DB enrichment)

Delete(path)
  → removes object (+ DB row / variants when configured)
```

Public field name `path` remains the storage-relative key; the DB column is `file_path` (v0.7+).

### Step 6 — Optional: thumbnails / SVG / PDF / S3

Configure pipeline and storages in the language Config. Details and constraints (e.g. PDF renderer must be injected) are in the language hub and [`docs/specs/`](./docs/specs/).

### Step 7 — Verify before release

- Unit tests: `packages/go` → `go test ./…` · `packages/js` → `npm test`
- E2E (RustFS + LocalFS, Go + JS): `./e2e/run.sh` or Actions [`e2e.yml`](./.github/workflows/e2e.yml) `workflow_dispatch` — see [`docs/specs/e2e/`](./docs/specs/e2e/)

---

## Quick CLI cheat sheet

```text
median migrate dump --dialect <mysql|postgres|sqlite3> [--id-strategy …] [--table …]
```

| | Go | JS |
|--|----|----|
| Run | `go run …/cmd/median@v0.9.0 migrate dump …` | `npx @b4moss/median migrate dump …` |
| Library twin | `db.CreateMediaSQL` | `createMediaSQL` |

Full flags and examples: language hubs · [`migrations/README.md`](./migrations/README.md).

---

## Docs map

| Doc | Contents |
|-----|----------|
| [`docs/README.md`](./docs/README.md) | Product pillar (purpose, scope, tech policy) |
| [`docs/roadmap.md`](./docs/roadmap.md) | Milestones |
| [`docs/specs/`](./docs/specs/) | Current specs (domain-oriented) |
| [`docs/plans/`](./docs/plans/) | Upcoming plans (PHP unscheduled) |
| [`docs/tests/`](./docs/tests/) | TDD acceptance specs |
| [`docs/charter/`](./docs/charter/) | Charter + OKF |
| [`migrations/`](./migrations/) | Schema reference (goose-oriented source of truth) |
| [`e2e/`](./e2e/) | E2E harness |
| [`.github/CI.md`](./.github/CI.md) | CI/CD / publish tags |

## License

MIT © Bicycle for Mind LLC. — see [`LICENSE`](./LICENSE).
