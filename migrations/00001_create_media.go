//go:build ignore

// Dialect-aware goose migration for media (default auto-increment schema).
// Imported/embedded from packages/go at v0.3.0 without the ignore constraint.
//
// One migration definition; goose dialect selects DDL.
// Targets: MySQL, MariaDB (mysql driver), Postgres, SQLite — aligned with crudian.

package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(up00001CreateMedia, down00001CreateMedia)
}

func up00001CreateMedia(ctx context.Context, tx *sql.Tx) error {
	q, err := createMediaSQL(goose.GetDialect())
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

func createMediaSQL(dialect string) (string, error) {
	switch dialect {
	case "mysql":
		// MySQL and MariaDB
		return `
CREATE TABLE media (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  path VARCHAR(2048) NOT NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  hash CHAR(64) NULL,
  original_id BIGINT NULL,
  created_by VARCHAR(255) NULL,
  owned_by VARCHAR(255) NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY media_path_unique (path),
  KEY media_hash_idx (hash),
  KEY media_original_id_idx (original_id),
  KEY media_owned_by_status_idx (owned_by, status),
  KEY media_created_at_idx (created_at),
  CONSTRAINT media_original_id_fk FOREIGN KEY (original_id) REFERENCES media (id)
)`, nil
	case "postgres":
		return `
CREATE TABLE media (
  id BIGSERIAL PRIMARY KEY,
  path VARCHAR(2048) NOT NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  hash CHAR(64) NULL,
  original_id BIGINT NULL REFERENCES media (id),
  created_by VARCHAR(255) NULL,
  owned_by VARCHAR(255) NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT media_path_unique UNIQUE (path)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at)`, nil
	case "sqlite3":
		return `
CREATE TABLE media (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  path TEXT NOT NULL,
  mime TEXT NOT NULL,
  size INTEGER NOT NULL,
  hash TEXT NULL,
  original_id INTEGER NULL REFERENCES media (id),
  created_by TEXT NULL,
  owned_by TEXT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  CONSTRAINT media_path_unique UNIQUE (path)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at)`, nil
	default:
		return "", fmt.Errorf("unsupported goose dialect %q (want mysql|postgres|sqlite3)", dialect)
	}
}
