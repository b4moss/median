//go:build ignore

// Dialect- and id-strategy-aware goose migration for media.
// Runtime wiring lives in packages/go/db (CreateMediaSQL / MigrateUp).
// Keep this file as the repo-root goose reference; do not compile into packages/go.
//
// One migration definition:
//   - goose dialect selects MySQL/MariaDB | Postgres | SQLite DDL
//   - id strategy (same definition) selects integer AI vs text PK
//   - table name defaults to media (override via MEDIAN_TABLE_NAME)
// Targets aligned with crudian.

package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(up00001CreateMedia, down00001CreateMedia)
}

func up00001CreateMedia(ctx context.Context, tx *sql.Tx) error {
	q, err := createMediaSQL(goose.GetDialect(), idStrategyFromEnv(), tableNameFromEnv())
	if err != nil {
		return err
	}
	for _, stmt := range splitSQL(q) {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func down00001CreateMedia(ctx context.Context, tx *sql.Tx) error {
	table, err := normalizeTableName(tableNameFromEnv())
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s`, table))
	return err
}

func idStrategyFromEnv() string {
	s := strings.ToLower(strings.TrimSpace(os.Getenv("MEDIAN_ID_STRATEGY")))
	if s == "" {
		return "auto_increment"
	}
	return s
}

func tableNameFromEnv() string {
	return strings.TrimSpace(os.Getenv("MEDIAN_TABLE_NAME"))
}

var tableNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func normalizeTableName(tableName string) (string, error) {
	if tableName == "" {
		return "media", nil
	}
	if !tableNameRe.MatchString(tableName) {
		return "", fmt.Errorf("invalid table name %q", tableName)
	}
	return tableName, nil
}

func splitSQL(q string) []string {
	parts := strings.Split(q, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func createMediaSQL(dialect, idStrategy, tableName string) (string, error) {
	table, err := normalizeTableName(tableName)
	if err != nil {
		return "", err
	}
	textID := idStrategy == "uuid_v4" || idStrategy == "uuid_v7" || idStrategy == "ulid"
	if !textID && idStrategy != "auto_increment" {
		return "", fmt.Errorf("unsupported MEDIAN_ID_STRATEGY %q (want auto_increment|uuid_v4|uuid_v7|ulid)", idStrategy)
	}
	switch dialect {
	case "mysql":
		if textID {
			return mysqlMediaTextID(table), nil
		}
		return mysqlMediaAutoInc(table), nil
	case "postgres":
		if textID {
			return postgresMediaTextID(table), nil
		}
		return postgresMediaAutoInc(table), nil
	case "sqlite3":
		if textID {
			return sqliteMediaTextID(table), nil
		}
		return sqliteMediaAutoInc(table), nil
	default:
		return "", fmt.Errorf("unsupported goose dialect %q (want mysql|postgres|sqlite3)", dialect)
	}
}

func mysqlMediaAutoInc(t string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  file_path VARCHAR(2048) NOT NULL,
  file_name VARCHAR(512) NOT NULL,
  original_file_name VARCHAR(512) NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  width INT NULL,
  height INT NULL,
  hash CHAR(64) NOT NULL,
  original_id BIGINT NULL,
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NOT NULL,
  owned_by VARCHAR(255) NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY %s_file_path_unique (file_path),
  UNIQUE KEY %s_original_variant_unique (original_id, variant_key),
  KEY %s_hash_idx (hash),
  KEY %s_original_id_idx (original_id),
  KEY %s_owned_by_status_idx (owned_by, status),
  KEY %s_created_at_idx (created_at),
  CONSTRAINT %s_original_id_fk FOREIGN KEY (original_id) REFERENCES %s (id)
)`, t, t, t, t, t, t, t, t, t)
}

func mysqlMediaTextID(t string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
  id VARCHAR(64) NOT NULL PRIMARY KEY,
  file_path VARCHAR(2048) NOT NULL,
  file_name VARCHAR(512) NOT NULL,
  original_file_name VARCHAR(512) NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  width INT NULL,
  height INT NULL,
  hash CHAR(64) NOT NULL,
  original_id VARCHAR(64) NULL,
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NOT NULL,
  owned_by VARCHAR(255) NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY %s_file_path_unique (file_path),
  UNIQUE KEY %s_original_variant_unique (original_id, variant_key),
  KEY %s_hash_idx (hash),
  KEY %s_original_id_idx (original_id),
  KEY %s_owned_by_status_idx (owned_by, status),
  KEY %s_created_at_idx (created_at),
  CONSTRAINT %s_original_id_fk FOREIGN KEY (original_id) REFERENCES %s (id)
)`, t, t, t, t, t, t, t, t, t)
}

func postgresMediaAutoInc(t string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
  id BIGSERIAL PRIMARY KEY,
  file_path VARCHAR(2048) NOT NULL,
  file_name VARCHAR(512) NOT NULL,
  original_file_name VARCHAR(512) NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  width INT NULL,
  height INT NULL,
  hash CHAR(64) NOT NULL,
  original_id BIGINT NULL,
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NOT NULL,
  owned_by VARCHAR(255) NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT %s_file_path_unique UNIQUE (file_path),
  CONSTRAINT %s_original_variant_unique UNIQUE (original_id, variant_key),
  CONSTRAINT %s_original_id_fk FOREIGN KEY (original_id) REFERENCES %s (id)
);
CREATE INDEX %s_hash_idx ON %s (hash);
CREATE INDEX %s_original_id_idx ON %s (original_id);
CREATE INDEX %s_owned_by_status_idx ON %s (owned_by, status);
CREATE INDEX %s_created_at_idx ON %s (created_at);
`, t, t, t, t, t, t, t, t, t, t, t, t, t)
}

func postgresMediaTextID(t string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
  id VARCHAR(64) PRIMARY KEY,
  file_path VARCHAR(2048) NOT NULL,
  file_name VARCHAR(512) NOT NULL,
  original_file_name VARCHAR(512) NULL,
  mime VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  width INT NULL,
  height INT NULL,
  hash CHAR(64) NOT NULL,
  original_id VARCHAR(64) NULL,
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NOT NULL,
  owned_by VARCHAR(255) NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT %s_file_path_unique UNIQUE (file_path),
  CONSTRAINT %s_original_variant_unique UNIQUE (original_id, variant_key),
  CONSTRAINT %s_original_id_fk FOREIGN KEY (original_id) REFERENCES %s (id)
);
CREATE INDEX %s_hash_idx ON %s (hash);
CREATE INDEX %s_original_id_idx ON %s (original_id);
CREATE INDEX %s_owned_by_status_idx ON %s (owned_by, status);
CREATE INDEX %s_created_at_idx ON %s (created_at);
`, t, t, t, t, t, t, t, t, t, t, t, t, t)
}

func sqliteMediaAutoInc(t string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  file_name TEXT NOT NULL,
  original_file_name TEXT NULL,
  mime TEXT NOT NULL,
  size INTEGER NOT NULL,
  width INTEGER NULL,
  height INTEGER NULL,
  hash TEXT NOT NULL,
  original_id INTEGER NULL,
  variant_key TEXT NULL,
  created_by TEXT NOT NULL,
  owned_by TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  CONSTRAINT %s_file_path_unique UNIQUE (file_path),
  CONSTRAINT %s_original_variant_unique UNIQUE (original_id, variant_key),
  FOREIGN KEY (original_id) REFERENCES %s (id)
);
CREATE INDEX %s_hash_idx ON %s (hash);
CREATE INDEX %s_original_id_idx ON %s (original_id);
CREATE INDEX %s_owned_by_status_idx ON %s (owned_by, status);
CREATE INDEX %s_created_at_idx ON %s (created_at);
`, t, t, t, t, t, t, t, t, t, t, t, t)
}

func sqliteMediaTextID(t string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
  id TEXT PRIMARY KEY,
  file_path TEXT NOT NULL,
  file_name TEXT NOT NULL,
  original_file_name TEXT NULL,
  mime TEXT NOT NULL,
  size INTEGER NOT NULL,
  width INTEGER NULL,
  height INTEGER NULL,
  hash TEXT NOT NULL,
  original_id TEXT NULL,
  variant_key TEXT NULL,
  created_by TEXT NOT NULL,
  owned_by TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  CONSTRAINT %s_file_path_unique UNIQUE (file_path),
  CONSTRAINT %s_original_variant_unique UNIQUE (original_id, variant_key),
  FOREIGN KEY (original_id) REFERENCES %s (id)
);
CREATE INDEX %s_hash_idx ON %s (hash);
CREATE INDEX %s_original_id_idx ON %s (original_id);
CREATE INDEX %s_owned_by_status_idx ON %s (owned_by, status);
CREATE INDEX %s_created_at_idx ON %s (created_at);
`, t, t, t, t, t, t, t, t, t, t, t, t)
}
