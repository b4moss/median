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

| Port | Package | Language hub (install, Config, CLI details) |
|------|---------|-----------------------------------------------|
| **Go** | `github.com/b4moss/median/packages/go` | [`packages/go/README.md`](./packages/go/README.md) |
| **TypeScript** | `@b4moss/median` | [`packages/js/README.md`](./packages/js/README.md) |
| **PHP** | — | Not scheduled — [`docs/plans/unscheduled/packages-php.md`](./docs/plans/unscheduled/packages-php.md) |

This root README is the **shared host workflow**. Port-specific install commands, import paths, `New` / `createMedian` examples, and full CLI flags live only in the language hubs above.

---

## Prerequisites

| | Go host | JS / Node host |
|--|---------|----------------|
| Runtime | Go **1.26+** | Node.js **24+** |
| Package | `github.com/b4moss/median/packages/go@v0.9.0` | `@b4moss/median@0.9.0` |
| DB (optional) | GORM-backed dialect via crudian | better-sqlite3 at runtime; DDL dump still supports mysql / postgres / sqlite3 |
| Storage | Local path and/or S3-compatible endpoint | Same |

You do **not** need this monorepo cloned to use published packages. Clone only for contributing or running E2E.

---

## Host usage (step by step)

Recommended production path when **your app owns migrations**.

### Step 1 — Pick a port and install

- Go → [`packages/go/README.md`](./packages/go/README.md)  
  `go get github.com/b4moss/median/packages/go@v0.9.0`
- JS → [`packages/js/README.md`](./packages/js/README.md)  
  `npm install @b4moss/median`

### Step 2 — Decide table name, ID strategy, dialect

Use the **same values** for the DDL dump and for runtime config.

| Setting | Default | Alternatives |
|---------|---------|--------------|
| Table | `media` | any identifier accepted by `NormalizeTableName` |
| ID strategy | `auto_increment` | `uuid_v4` · `uuid_v7` · `ulid` |
| Dialect | — (required for dump) | `mysql` \| `postgres` \| `sqlite3` |

Skip this step only if you run **without DB** (storage-only) — see [Without a media DB](#without-a-media-db) below.

### Step 3 — Dump DDL into *your* migrations

median does **not** run production migrate for you. Generate SQL and commit it to the host migrate tool (goose, Prisma, Flyway, knex, …).

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

| Stream | Content |
|--------|---------|
| **stdout** | SQL only (safe to redirect into a migration file) |
| **stderr** | errors / usage |
| Exit code | `0` on success, non-zero on bad flags |

- Spec: [`docs/specs/ddl-cli/`](./docs/specs/ddl-cli/)
- Schema reference: [`migrations/README.md`](./migrations/README.md)

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

## Without a media DB

Omit DB config entirely. median still Store / Get / Delete against Local FS or S3; there is no metadata row and no `migrate dump` step. Wiring examples: language hubs (leave `DB` / `db` unset).

---

## Common pitfalls

| Pitfall | Fix |
|---------|-----|
| Dump flags ≠ runtime Config | Match `--table` / `--id-strategy` / dialect with `TableName` / `IDStrategy` |
| Expecting median to `migrate up` in prod | Use host migrate + `migrate dump`; `MigrateUp` / `migrateUp` are for tests / throwaway DBs |
| Redirecting stderr into the SQL file | Capture **stdout only**; check stderr separately |
| Go `go get` 404 from proxy | Module path is `…/packages/go`; tag is `packages/go/vX.Y.Z` — see [`go-module-path`](./docs/specs/go-module-path/) |
| Confusing public `path` with DB column | API field stays `path`; physical column is `file_path` |
| PDF thumbs with no renderer | Inject a renderer in Config; median does not ship a PDF engine |

---

## Quick CLI cheat sheet

```text
median migrate dump --dialect <mysql|postgres|sqlite3> [--id-strategy …] [--table …]
```

| | Go | JS |
|--|----|----|
| Run | `go run …/cmd/median@v0.9.0 migrate dump …` | `npx @b4moss/median migrate dump …` |
| Install binary | `go install …/cmd/median@v0.9.0` | `npm i -g @b4moss/median` (or local `npx`) |
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
| [`packages/go/README.md`](./packages/go/README.md) | Go language hub |
| [`packages/js/README.md`](./packages/js/README.md) | JS language hub |

## License

MIT © Bicycle for Mind LLC. — see [`LICENSE`](./LICENSE).
