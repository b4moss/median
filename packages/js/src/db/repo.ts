import { randomUUID } from "node:crypto";
import { v7 as uuidv7 } from "uuid";
import type Database from "better-sqlite3";
import {
  ID_AUTO_INCREMENT,
  ID_ULID,
  ID_UUID_V4,
  ID_UUID_V7,
  normalizeTableName,
} from "./schema.js";
import { withColumnDefaults, type ColumnMap, type ResolvedColumnMap } from "./columns.js";

export class DBError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "DBError";
  }
}

export const ErrNotFound = new DBError("db: media not found");
export const ErrInvalidInput = new DBError("db: invalid input");

export type Media = {
  id?: string | number;
  filePath: string;
  fileName: string;
  originalFileName?: string | null;
  mime: string;
  size: number;
  width?: number | null;
  height?: number | null;
  hash: string;
  originalId?: string | number | null;
  variantKey?: string | null;
  createdBy: string;
  ownedBy: string;
  status?: string;
  createdAt?: string;
};

export type RepoOptions = {
  idStrategy?: string;
  tableName?: string;
  columns?: ColumnMap;
};

type Row = Record<string, unknown>;

function asString(v: unknown): string {
  if (v == null) return "";
  if (typeof v === "string") return v;
  if (Buffer.isBuffer(v)) return v.toString();
  return String(v);
}

function asNumber(v: unknown): number {
  if (typeof v === "number") return v;
  if (typeof v === "bigint") return Number(v);
  if (typeof v === "string") return Number(v);
  return 0;
}

function rowToMedia(row: Row, c: ResolvedColumnMap): Media {
  return {
    id: row[c.id] as string | number,
    filePath: asString(row[c.filePath]),
    fileName: asString(row[c.fileName]),
    originalFileName: row[c.originalFileName] == null ? null : asString(row[c.originalFileName]),
    mime: asString(row[c.mime]),
    size: asNumber(row[c.size]),
    width: row[c.width] == null ? null : asNumber(row[c.width]),
    height: row[c.height] == null ? null : asNumber(row[c.height]),
    hash: asString(row[c.hash]),
    originalId: (row[c.originalId] as string | number | null) ?? null,
    variantKey: row[c.variantKey] == null ? null : asString(row[c.variantKey]),
    createdBy: asString(row[c.createdBy]),
    ownedBy: asString(row[c.ownedBy]),
    status: asString(row[c.status]),
    createdAt: asString(row[c.createdAt]),
  };
}

export class MediaRepo {
  private readonly table: string;
  private readonly cols: ResolvedColumnMap;
  private readonly idStrategy: string;

  constructor(
    private readonly db: Database.Database,
    options: RepoOptions = {},
  ) {
    this.idStrategy = options.idStrategy || ID_AUTO_INCREMENT;
    this.table = normalizeTableName(options.tableName);
    this.cols = withColumnDefaults(options.columns);
  }

  create(m: Media): Media {
    if (!m) {
      throw new DBError("db: invalid input");
    }
    if (!m.filePath || !m.fileName || !m.mime || !m.hash) {
      throw new DBError("db: invalid input: file_path/file_name/mime/hash required");
    }
    if (!m.createdBy || !m.ownedBy) {
      throw new DBError("db: invalid input: actor required");
    }
    const status = m.status || "active";
    const c = this.cols;
    let id: string | number | undefined;

    switch (this.idStrategy) {
      case ID_UUID_V4:
        id = randomUUID();
        break;
      case ID_UUID_V7:
      case ID_ULID:
        id = uuidv7();
        break;
      case ID_AUTO_INCREMENT:
        break;
      default:
        throw new DBError(`unsupported id strategy: ${this.idStrategy}`);
    }

    const cols = [
      c.filePath,
      c.fileName,
      c.originalFileName,
      c.mime,
      c.size,
      c.width,
      c.height,
      c.hash,
      c.originalId,
      c.variantKey,
      c.createdBy,
      c.ownedBy,
      c.status,
    ];
    const values = [
      m.filePath,
      m.fileName,
      m.originalFileName ?? null,
      m.mime,
      m.size,
      m.width ?? null,
      m.height ?? null,
      m.hash,
      m.originalId ?? null,
      m.variantKey ?? null,
      m.createdBy,
      m.ownedBy,
      status,
    ];

    if (id !== undefined) {
      this.db
        .prepare(
          `INSERT INTO ${this.table} (${c.id}, ${cols.join(", ")})
           VALUES (?, ${cols.map(() => "?").join(", ")})`,
        )
        .run(id, ...values);
      return this.findById(id);
    }

    const info = this.db
      .prepare(
        `INSERT INTO ${this.table} (${cols.join(", ")})
         VALUES (${cols.map(() => "?").join(", ")})`,
      )
      .run(...values);
    return this.findById(Number(info.lastInsertRowid));
  }

  findById(id: string | number): Media {
    if (id === undefined || id === null || id === "") {
      throw new DBError("db: invalid input: empty id");
    }
    const row = this.db
      .prepare(`SELECT * FROM ${this.table} WHERE ${this.cols.id} = ?`)
      .get(id) as Row | undefined;
    if (!row) {
      throw new DBError("db: media not found");
    }
    return rowToMedia(row, this.cols);
  }

  findByHash(hash: string): Media {
    if (!hash) {
      throw new DBError("db: invalid input: empty hash");
    }
    const c = this.cols;
    let row = this.db
      .prepare(
        `SELECT * FROM ${this.table} WHERE ${c.hash} = ? AND ${c.variantKey} IS NULL LIMIT 1`,
      )
      .get(hash) as Row | undefined;
    if (!row) {
      row = this.db
        .prepare(`SELECT * FROM ${this.table} WHERE ${c.hash} = ? LIMIT 1`)
        .get(hash) as Row | undefined;
    }
    if (!row) {
      throw new DBError("db: media not found");
    }
    return rowToMedia(row, this.cols);
  }

  listChildren(parentId: string | number): Media[] {
    const rows = this.db
      .prepare(`SELECT * FROM ${this.table} WHERE ${this.cols.originalId} = ?`)
      .all(parentId) as Row[];
    return rows.map((row) => rowToMedia(row, this.cols));
  }

  deleteById(id: string | number): void {
    if (id === undefined || id === null || id === "") {
      throw new DBError("db: invalid input: empty id");
    }
    const children = this.listChildren(id);
    const del = this.db.prepare(`DELETE FROM ${this.table} WHERE ${this.cols.id} = ?`);
    for (const ch of children) {
      del.run(ch.id!);
    }
    const info = del.run(id);
    if (info.changes === 0) {
      throw new DBError("db: media not found");
    }
  }
}

export function createMediaRepo(db: Database.Database, options?: RepoOptions | string): MediaRepo {
  if (typeof options === "string") {
    return new MediaRepo(db, { idStrategy: options });
  }
  return new MediaRepo(db, options || {});
}

export function isNotFound(err: unknown): boolean {
  return err instanceof DBError && err.message.includes("not found");
}
