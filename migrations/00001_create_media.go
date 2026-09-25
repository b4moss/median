//go:build ignore

// Dialect- and id-strategy-aware goose migration for media.
// Imported/embedded from packages/go at v0.3.0 without the ignore constraint.
//
// One migration definition:
//   - goose dialect selects MySQL/MariaDB | Postgres | SQLite DDL
//   - id strategy (same definition) selects integer AI vs text PK
// Targets aligned with crudian.

package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(up00001CreateMedia, down00001CreateMedia)
}

func up00001CreateMedia(ctx context.Context, tx *sql.Tx) error {
	q, err := createMediaSQL(goose.GetDialect(), idStrategyFromEnv())
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, q)
	return err
}

func down00001CreateMedia(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS media`)
	return err
}

// idStrategyFromEnv reads MEDIAN_ID_STRATEGY.
// Values: auto_increment (default) | uuid_v4 | uuid_v7 | ulid
// uuid_* / ulid share the text PK DDL branch inside this same migration.
func idStrategyFromEnv() string {
	s := strings.ToLower(strings.TrimSpace(os.Getenv("MEDIAN_ID_STRATEGY")))
	if s == "" {
		return "auto_increment"
	}
	return s
}

func createMediaSQL(dialect, idStrategy string) (string, error) {
	textID := idStrategy == "uuid_v4" || idStrategy == "uuid_v7" || idStrategy == "ulid"
	if !textID && idStrategy != "auto_increment" {
		return "", fmt.Errorf("unsupported MEDIAN_ID_STRATEGY %q (want auto_increment|uuid_v4|uuid_v7|ulid)", idStrategy)
	}

	switch dialect {
	case "mysql":
		if textID {
			return mysqlMediaTextID(), nil
		}
		return mysqlMediaAutoInc(), nil
	case "postgres":
		if textID {
			return postgresMediaTextID(), nil
		}
		return postgresMediaAutoInc(), nil
	case "sqlite3":
		if textID {
			return sqliteMediaTextID(), nil
		}
		return sqliteMediaAutoInc(), nil
	default:
		return "", fmt.Errorf("unsupported goose dialect %q (want mysql|postgres|sqlite3)", dialect)
	}
}

func mysqlMediaAutoInc() string {
	return `
CREATE TABLE media (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  path VARCHAR(2048) NOT NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  hash CHAR(64) NULL,
  original_id BIGINT NULL,
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NULL,
  owned_by VARCHAR(255) NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY media_path_unique (path),
  UNIQUE KEY media_original_variant_unique (original_id, variant_key),
  KEY media_hash_idx (hash),
  KEY media_original_id_idx (original_id),
  KEY media_owned_by_status_idx (owned_by, status),
  KEY media_created_at_idx (created_at),
  CONSTRAINT media_original_id_fk FOREIGN KEY (original_id) REFERENCES media (id)
)`
}

func mysqlMediaTextID() string {
	return `
CREATE TABLE media (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  path VARCHAR(2048) NOT NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  hash CHAR(64) NULL,
  original_id VARCHAR(64) NULL,
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NULL,
  owned_by VARCHAR(255) NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY media_path_unique (path),
  UNIQUE KEY media_original_variant_unique (original_id, variant_key),
  KEY media_hash_idx (hash),
  KEY media_original_id_idx (original_id),
  KEY media_owned_by_status_idx (owned_by, status),
  KEY media_created_at_idx (created_at),
  CONSTRAINT media_original_id_fk FOREIGN KEY (original_id) REFERENCES media (id)
)`
}

func postgresMediaAutoInc() string {
	return `
CREATE TABLE media (
  id BIGSERIAL PRIMARY KEY,
  path VARCHAR(2048) NOT NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  hash CHAR(64) NULL,
  original_id BIGINT NULL REFERENCES media (id),
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NULL,
  owned_by VARCHAR(255) NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT media_path_unique UNIQUE (path),
  CONSTRAINT media_original_variant_unique UNIQUE (original_id, variant_key)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at)`
}

func postgresMediaTextID() string {
	return `
CREATE TABLE media (
  id VARCHAR(64) PRIMARY KEY,
  path VARCHAR(2048) NOT NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  hash CHAR(64) NULL,
  original_id VARCHAR(64) NULL REFERENCES media (id),
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NULL,
  owned_by VARCHAR(255) NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT media_path_unique UNIQUE (path),
  CONSTRAINT media_original_variant_unique UNIQUE (original_id, variant_key)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at)`
}

func sqliteMediaAutoInc() string {
	return `
CREATE TABLE media (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  path TEXT NOT NULL,
  mime TEXT NOT NULL,
  size INTEGER NOT NULL,
  hash TEXT NULL,
  original_id INTEGER NULL REFERENCES media (id),
  variant_key TEXT NULL,
  created_by TEXT NULL,
  owned_by TEXT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  CONSTRAINT media_path_unique UNIQUE (path),
  CONSTRAINT media_original_variant_unique UNIQUE (original_id, variant_key)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at)`
}

func sqliteMediaTextID() string {
	return `
CREATE TABLE media (
  id TEXT NOT NULL PRIMARY KEY,
  path TEXT NOT NULL,
  mime TEXT NOT NULL,
  size INTEGER NOT NULL,
  hash TEXT NULL,
  original_id TEXT NULL REFERENCES media (id),
  variant_key TEXT NULL,
  created_by TEXT NULL,
  owned_by TEXT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  CONSTRAINT media_path_unique UNIQUE (path),
  CONSTRAINT media_original_variant_unique UNIQUE (original_id, variant_key)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at)`
}
