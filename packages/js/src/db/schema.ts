export const ID_AUTO_INCREMENT = "auto_increment";
export const ID_UUID_V4 = "uuid_v4";
export const ID_UUID_V7 = "uuid_v7";
export const ID_ULID = "ulid";

export const DEFAULT_TABLE_NAME = "media";

const TABLE_NAME_RE = /^[A-Za-z_][A-Za-z0-9_]*$/;

export function idStrategyFromEnv(): string {
  const s = (process.env.MEDIAN_ID_STRATEGY ?? "").toLowerCase().trim();
  return s || ID_AUTO_INCREMENT;
}

export function normalizeTableName(tableName?: string): string {
  const name = (tableName ?? "").trim() || DEFAULT_TABLE_NAME;
  if (!TABLE_NAME_RE.test(name)) {
    throw new Error(`invalid table name "${name}"`);
  }
  return name;
}

export function createMediaSQL(dialect: string, idStrategy: string, tableName?: string): string {
  const table = normalizeTableName(tableName);
  const textID =
    idStrategy === ID_UUID_V4 || idStrategy === ID_UUID_V7 || idStrategy === ID_ULID;
  if (!textID && idStrategy !== ID_AUTO_INCREMENT) {
    throw new Error(
      `unsupported MEDIAN_ID_STRATEGY "${idStrategy}" (want auto_increment|uuid_v4|uuid_v7|ulid)`,
    );
  }
  switch (dialect) {
    case "sqlite3":
      return textID ? sqliteMediaTextID(table) : sqliteMediaAutoInc(table);
    case "mysql":
      return textID ? mysqlMediaTextID(table) : mysqlMediaAutoInc(table);
    case "postgres":
      return textID ? postgresMediaTextID(table) : postgresMediaAutoInc(table);
    default:
      throw new Error(`unsupported goose dialect "${dialect}" (want mysql|postgres|sqlite3)`);
  }
}

function sqliteMediaAutoInc(t: string): string {
  return `
CREATE TABLE ${t} (
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
  CONSTRAINT ${t}_file_path_unique UNIQUE (file_path),
  CONSTRAINT ${t}_original_variant_unique UNIQUE (original_id, variant_key),
  FOREIGN KEY (original_id) REFERENCES ${t} (id)
);
CREATE INDEX ${t}_hash_idx ON ${t} (hash);
CREATE INDEX ${t}_original_id_idx ON ${t} (original_id);
CREATE INDEX ${t}_owned_by_status_idx ON ${t} (owned_by, status);
CREATE INDEX ${t}_created_at_idx ON ${t} (created_at);
`;
}

function sqliteMediaTextID(t: string): string {
  return `
CREATE TABLE ${t} (
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
  CONSTRAINT ${t}_file_path_unique UNIQUE (file_path),
  CONSTRAINT ${t}_original_variant_unique UNIQUE (original_id, variant_key),
  FOREIGN KEY (original_id) REFERENCES ${t} (id)
);
CREATE INDEX ${t}_hash_idx ON ${t} (hash);
CREATE INDEX ${t}_original_id_idx ON ${t} (original_id);
CREATE INDEX ${t}_owned_by_status_idx ON ${t} (owned_by, status);
CREATE INDEX ${t}_created_at_idx ON ${t} (created_at);
`;
}

function mysqlMediaAutoInc(t: string): string {
  return `
CREATE TABLE ${t} (
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
  UNIQUE KEY ${t}_file_path_unique (file_path),
  UNIQUE KEY ${t}_original_variant_unique (original_id, variant_key),
  KEY ${t}_hash_idx (hash),
  KEY ${t}_original_id_idx (original_id),
  KEY ${t}_owned_by_status_idx (owned_by, status),
  KEY ${t}_created_at_idx (created_at),
  CONSTRAINT ${t}_original_id_fk FOREIGN KEY (original_id) REFERENCES ${t} (id)
)`;
}

function mysqlMediaTextID(t: string): string {
  return `
CREATE TABLE ${t} (
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
  UNIQUE KEY ${t}_file_path_unique (file_path),
  UNIQUE KEY ${t}_original_variant_unique (original_id, variant_key),
  KEY ${t}_hash_idx (hash),
  KEY ${t}_original_id_idx (original_id),
  KEY ${t}_owned_by_status_idx (owned_by, status),
  KEY ${t}_created_at_idx (created_at),
  CONSTRAINT ${t}_original_id_fk FOREIGN KEY (original_id) REFERENCES ${t} (id)
)`;
}

function postgresMediaAutoInc(t: string): string {
  return `
CREATE TABLE ${t} (
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
  CONSTRAINT ${t}_file_path_unique UNIQUE (file_path),
  CONSTRAINT ${t}_original_variant_unique UNIQUE (original_id, variant_key),
  CONSTRAINT ${t}_original_id_fk FOREIGN KEY (original_id) REFERENCES ${t} (id)
);
CREATE INDEX ${t}_hash_idx ON ${t} (hash);
CREATE INDEX ${t}_original_id_idx ON ${t} (original_id);
CREATE INDEX ${t}_owned_by_status_idx ON ${t} (owned_by, status);
CREATE INDEX ${t}_created_at_idx ON ${t} (created_at);
`;
}

function postgresMediaTextID(t: string): string {
  return `
CREATE TABLE ${t} (
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
  CONSTRAINT ${t}_file_path_unique UNIQUE (file_path),
  CONSTRAINT ${t}_original_variant_unique UNIQUE (original_id, variant_key),
  CONSTRAINT ${t}_original_id_fk FOREIGN KEY (original_id) REFERENCES ${t} (id)
);
CREATE INDEX ${t}_hash_idx ON ${t} (hash);
CREATE INDEX ${t}_original_id_idx ON ${t} (original_id);
CREATE INDEX ${t}_owned_by_status_idx ON ${t} (owned_by, status);
CREATE INDEX ${t}_created_at_idx ON ${t} (created_at);
`;
}
