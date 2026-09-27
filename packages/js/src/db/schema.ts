export const ID_AUTO_INCREMENT = "auto_increment";
export const ID_UUID_V4 = "uuid_v4";
export const ID_UUID_V7 = "uuid_v7";
export const ID_ULID = "ulid";

export function idStrategyFromEnv(): string {
  const s = (process.env.MEDIAN_ID_STRATEGY ?? "").toLowerCase().trim();
  return s || ID_AUTO_INCREMENT;
}

export function createMediaSQL(dialect: string, idStrategy: string): string {
  const textID =
    idStrategy === ID_UUID_V4 || idStrategy === ID_UUID_V7 || idStrategy === ID_ULID;
  if (!textID && idStrategy !== ID_AUTO_INCREMENT) {
    throw new Error(
      `unsupported MEDIAN_ID_STRATEGY "${idStrategy}" (want auto_increment|uuid_v4|uuid_v7|ulid)`,
    );
  }
  switch (dialect) {
    case "sqlite3":
      return textID ? sqliteMediaTextID() : sqliteMediaAutoInc();
    case "mysql":
      return textID ? mysqlMediaTextID() : mysqlMediaAutoInc();
    case "postgres":
      return textID ? postgresMediaTextID() : postgresMediaAutoInc();
    default:
      throw new Error(`unsupported goose dialect "${dialect}" (want mysql|postgres|sqlite3)`);
  }
}

function sqliteMediaAutoInc(): string {
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
`;
}

function sqliteMediaTextID(): string {
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
`;
}

function mysqlMediaAutoInc(): string {
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
)`;
}

function mysqlMediaTextID(): string {
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
)`;
}

function postgresMediaAutoInc(): string {
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
`;
}

function postgresMediaTextID(): string {
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
`;
}
