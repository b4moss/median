# packages/go

Go module **`github.com/b4moss/median/packages/go`**  
Current: `packages/go/VERSION` → **`0.9.0`** (tag `packages/go/v0.9.0`).

Language-port hub for Go. Cross-language product overview and host workflow: [root README](../../README.md). Specs: [docs/specs/](../../docs/specs/).

## Install

```bash
go get github.com/b4moss/median/packages/go@v0.9.0
```

```bash
go install github.com/b4moss/median/packages/go/cmd/median@v0.9.0
# or: go run github.com/b4moss/median/packages/go/cmd/median@v0.9.0 …
```

## Layout

| Path | Role |
| --- | --- |
| `core` | `New` / Store / Delete / Get / PresignGet / Config |
| `storage` / `local` / `s3` | Adapters |
| `db` | `CreateMediaSQL` / `MigrateUp` / MediaRepo |
| `cmd/median` | CLI (`migrate dump`) |
| `pipeline` | Image / SVG / PDF |
| `internal` | MIME, filename, shardian path |
| `third_party/crudian` | Vendored crudian (`replace` in `go.mod`) |

## Step-by-step (Go host)

### 1. Add the module

```bash
go get github.com/b4moss/median/packages/go@v0.9.0
```

### 2. Create the media table (production)

Dump DDL and commit it into **your** migrations (goose / Flyway / raw SQL — your tool):

```bash
go run github.com/b4moss/median/packages/go/cmd/median@v0.9.0 \
  migrate dump --dialect postgres --id-strategy auto_increment --table media \
  > migrations/XXXX_median_media.sql
```

| Flag | Required | Default |
| --- | --- | --- |
| `--dialect` | yes | — (`mysql` \| `postgres` \| `sqlite3`) |
| `--id-strategy` | no | `auto_increment` (`uuid_v4` \| `uuid_v7` \| `ulid`) |
| `--table` | no | `media` |

SQL goes to **stdout only**. Match these flags with runtime `DBConfig` below.

For tests / throwaway SQLite you may call `db.MigrateUp` instead of the host migrate pipeline.

Existing tables with different column names: **do not** use the CLI — configure `DBConfig.Columns` (column map). See [db-media](../../docs/specs/db-media/).

### 3. Wire `core.New`

```go
package main

import (
	"bytes"
	"context"
	"log"
	"os"

	"github.com/b4moss/median/packages/go/core"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	gdb, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := core.New(core.Config{
		Storages: map[string]core.StorageConfig{
			"local": {Driver: core.DriverLocal, Path: "/var/median"},
		},
		DefaultKey:   "local",
		DefaultActor: "system",
		DB: &core.DBConfig{
			Gorm:       gdb,
			IDStrategy: "auto_increment", // same as dump
			TableName:  "media",          // same as dump
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	payload := []byte("hello")
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME:     "text/plain",
		Filename: "hello.txt",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("stored path=%s hash=%s", res.Path, res.Hash)

	got, err := m.Get(ctx, res.Path, core.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer got.Body.Close()

	if err := m.Delete(ctx, res.Path, core.DeleteOptions{}); err != nil {
		log.Fatal(err)
	}
}
```

S3 example storage entry:

```go
"s3": {
  Driver:         core.DriverS3,
  Bucket:         "my-bucket",
  Region:         "ap-northeast-1",
  Endpoint:       "https://s3.example", // optional (MinIO / RustFS)
  ForcePathStyle: true,
  Prefix:         "uploads/",
  AccessKey:      "...",
  SecretKey:      "...",
},
```

### 4. Library DDL (optional)

```go
sql, err := db.CreateMediaSQL("postgres", db.IDAutoIncrement, "media")
```

Same SQL as the CLI for the same arguments.

## Develop / test

```bash
cd packages/go
go test ./...
go test ./cmd/median/
```

## Release

Tag **`packages/go/vX.Y.Z`** matching `VERSION`. Root tag `vX.Y.Z` alone does **not** publish Go. See [`.github/CI.md`](../../.github/CI.md).

## Docs

- Specs: [core-api](../../docs/specs/core-api/), [storage](../../docs/specs/storage/), [db-media](../../docs/specs/db-media/), [ddl-cli](../../docs/specs/ddl-cli/), [go-module-path](../../docs/specs/go-module-path/), …
- Tests: [docs/tests/](../../docs/tests/)
- E2E: [e2e/](../../e2e/) / [docs/specs/e2e/](../../docs/specs/e2e/)
- Migrations reference: [migrations/README.md](../../migrations/README.md)
