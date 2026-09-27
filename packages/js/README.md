# @b4moss/median

TypeScript port (`packages/js`). npm: **[@b4moss/median](https://www.npmjs.com/package/@b4moss/median)**  
Current: **`0.9.0`** (git tag `v0.9.0`). Requires **Node.js 24+**.

Language-port hub for JS/TS. Cross-language product overview and host workflow: [root README](../../README.md). Specs: [docs/specs/](../../docs/specs/).

## Install

```bash
npm install @b4moss/median
```

CLI (after install / via npx):

```bash
npx @b4moss/median migrate dump --dialect postgres --table media
```

`package.json` maps `bin.median` → `./dist/cli.js` (built on publish).

## Layout

| Path | Role |
| --- | --- |
| `src/core` | `createMedian` / store / delete / get / presignGet |
| `src/storage` | local / s3 |
| `src/db` | `createMediaSQL` / `migrateUp` / MediaRepo |
| `src/cli.ts` | CLI (`migrate dump`) |
| `src/pipeline` | process / sanitizeSvg / PDF thumbs |
| `src/internal` | MIME / filename / shardian path |

**DB note:** JS does **not** use crudian. Runtime MediaRepo is **SQLite** (better-sqlite3). DDL string generation still supports mysql / postgres / sqlite3 for host dumps.

## Step-by-step (Node host)

### 1. Install

```bash
npm install @b4moss/median
```

### 2. Create the media table (production)

```bash
npx @b4moss/median migrate dump \
  --dialect postgres \
  --id-strategy auto_increment \
  --table media \
  > migrations/XXXX_median_media.sql
```

| Flag | Required | Default |
| --- | --- | --- |
| `--dialect` | yes | — (`mysql` \| `postgres` \| `sqlite3`) |
| `--id-strategy` | no | `auto_increment` |
| `--table` | no | `media` |

Commit the SQL into **your** migrate tool, then run it. Match flags with `db` config below.

For unit tests, `migrateUp(db, "sqlite3", ID_AUTO_INCREMENT)` on an in-memory better-sqlite3 is fine.

Existing tables + column maps: skip the CLI; see [db-media](../../docs/specs/db-media/).

### 3. Create a client

```ts
import Database from "better-sqlite3";
import {
  createMedian,
  migrateUp,
  ID_AUTO_INCREMENT,
} from "@b4moss/median";

const sql = new Database("app.sqlite");
// Dev/demo only — production should apply CLI dump via host migrations:
migrateUp(sql, "sqlite3", ID_AUTO_INCREMENT);

const median = await createMedian({
  storages: {
    local: { driver: "local", path: "./data/median" },
  },
  defaultKey: "local",
  defaultActor: "system",
  db: {
    db: sql,
    idStrategy: ID_AUTO_INCREMENT,
    tableName: "media",
  },
});

const body = Buffer.from("hello");
const stored = await median.store(body, body.length, {
  mime: "text/plain",
  filename: "hello.txt",
});
console.log(stored.path, stored.hash);

const got = await median.get(stored.path, { withBody: true });
console.log(got.body?.toString());

await median.delete(stored.path);
```

S3 storage entry sketch:

```ts
s3: {
  driver: "s3",
  bucket: "my-bucket",
  region: "ap-northeast-1",
  endpoint: "https://s3.example",
  forcePathStyle: true,
  prefix: "uploads/",
  accessKey: "...",
  secretKey: "...",
},
```

### 4. Library DDL (optional)

```ts
import { createMediaSQL, ID_AUTO_INCREMENT } from "@b4moss/median";

const sqlText = createMediaSQL("postgres", ID_AUTO_INCREMENT, "media");
```

Same payload as the CLI for the same arguments.

## Develop / test

```bash
cd packages/js
npm ci
npm test
npm run test:coverage
npm run build
```

## Release

Root git tag **`vX.Y.Z`** must match `package.json` `version`. Pushing that tree to `release` publishes to npm when packed content differs ([`.github/CI.md`](../../.github/CI.md)). Trusted Publishing on GitHub Actions.

## Docs

- Specs: [packages-js](../../docs/specs/packages-js/), [db-media](../../docs/specs/db-media/), [ddl-cli](../../docs/specs/ddl-cli/), …
- Tests: [docs/tests/packages-js/](../../docs/tests/packages-js/), [ddl-cli](../../docs/tests/ddl-cli/)
- E2E: [e2e/](../../e2e/) / [docs/specs/e2e/](../../docs/specs/e2e/)
- Migrations reference: [migrations/README.md](../../migrations/README.md)
