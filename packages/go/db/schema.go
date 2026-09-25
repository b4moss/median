package db

import (
	"fmt"
	"os"
	"strings"
)

// ID strategy values (MEDIAN_ID_STRATEGY / Config).
const (
	IDAutoIncrement = "auto_increment"
	IDUUIDv4        = "uuid_v4"
	IDUUIDv7        = "uuid_v7"
	IDULID          = "ulid"
)

// IDStrategyFromEnv reads MEDIAN_ID_STRATEGY (default auto_increment).
func IDStrategyFromEnv() string {
	s := strings.ToLower(strings.TrimSpace(os.Getenv("MEDIAN_ID_STRATEGY")))
	if s == "" {
		return IDAutoIncrement
	}
	return s
}

// CreateMediaSQL returns DDL for media for goose dialect + id strategy.
func CreateMediaSQL(dialect, idStrategy string) (string, error) {
	textID := idStrategy == IDUUIDv4 || idStrategy == IDUUIDv7 || idStrategy == IDULID
	if !textID && idStrategy != IDAutoIncrement {
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
  width INT NULL,
  height INT NULL,
  hash CHAR(64) NOT NULL,
  original_id BIGINT NULL,
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NOT NULL,
  owned_by VARCHAR(255) NOT NULL,
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
  width INT NULL,
  height INT NULL,
  hash CHAR(64) NOT NULL,
  original_id VARCHAR(64) NULL,
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NOT NULL,
  owned_by VARCHAR(255) NOT NULL,
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
  width INT NULL,
  height INT NULL,
  hash CHAR(64) NOT NULL,
  original_id BIGINT NULL,
  variant_key VARCHAR(64) NULL,
  created_by VARCHAR(255) NOT NULL,
  owned_by VARCHAR(255) NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT media_path_unique UNIQUE (path),
  CONSTRAINT media_original_variant_unique UNIQUE (original_id, variant_key),
  CONSTRAINT media_original_id_fk FOREIGN KEY (original_id) REFERENCES media (id)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at);
`
}

func postgresMediaTextID() string {
	return `
CREATE TABLE media (
  id VARCHAR(64) PRIMARY KEY,
  path VARCHAR(2048) NOT NULL,
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
  CONSTRAINT media_path_unique UNIQUE (path),
  CONSTRAINT media_original_variant_unique UNIQUE (original_id, variant_key),
  CONSTRAINT media_original_id_fk FOREIGN KEY (original_id) REFERENCES media (id)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at);
`
}

func sqliteMediaAutoInc() string {
	return `
CREATE TABLE media (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path TEXT NOT NULL,
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
  CONSTRAINT media_path_unique UNIQUE (path),
  CONSTRAINT media_original_variant_unique UNIQUE (original_id, variant_key),
  FOREIGN KEY (original_id) REFERENCES media (id)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at);
`
}

func sqliteMediaTextID() string {
	return `
CREATE TABLE media (
  id TEXT PRIMARY KEY,
  path TEXT NOT NULL,
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
  CONSTRAINT media_path_unique UNIQUE (path),
  CONSTRAINT media_original_variant_unique UNIQUE (original_id, variant_key),
  FOREIGN KEY (original_id) REFERENCES media (id)
);
CREATE INDEX media_hash_idx ON media (hash);
CREATE INDEX media_original_id_idx ON media (original_id);
CREATE INDEX media_owned_by_status_idx ON media (owned_by, status);
CREATE INDEX media_created_at_idx ON media (created_at);
`
}
