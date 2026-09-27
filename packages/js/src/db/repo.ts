import { randomUUID } from "node:crypto";
import { v7 as uuidv7 } from "uuid";
import type Database from "better-sqlite3";
import {
  ID_AUTO_INCREMENT,
  ID_ULID,
  ID_UUID_V4,
  ID_UUID_V7,
} from "./schema.js";

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
  path: string;
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

type Row = {
  id: string | number;
  path: string;
  mime: string;
  size: number;
  width: number | null;
  height: number | null;
  hash: string;
  original_id: string | number | null;
  variant_key: string | null;
  created_by: string;
  owned_by: string;
  status: string;
  created_at: string;
};

function rowToMedia(row: Row): Media {
  return {
    id: row.id,
    path: row.path,
    mime: row.mime,
    size: Number(row.size),
    width: row.width ?? null,
    height: row.height ?? null,
    hash: row.hash,
    originalId: row.original_id ?? null,
    variantKey: row.variant_key ?? null,
    createdBy: row.created_by,
    ownedBy: row.owned_by,
    status: row.status,
    createdAt: row.created_at,
  };
}

export class MediaRepo {
  constructor(
    private readonly db: Database.Database,
    private readonly idStrategy: string = ID_AUTO_INCREMENT,
  ) {}

  create(m: Media): Media {
    if (!m) {
      throw new DBError("db: invalid input");
    }
    if (!m.path || !m.mime || !m.hash) {
      throw new DBError("db: invalid input: path/mime/hash required");
    }
    if (!m.createdBy || !m.ownedBy) {
      throw new DBError("db: invalid input: actor required");
    }
    const status = m.status || "active";
    let id: string | number | undefined;

    switch (this.idStrategy) {
      case ID_UUID_V4:
        id = randomUUID();
        break;
      case ID_UUID_V7:
        id = uuidv7();
        break;
      case ID_ULID:
        id = uuidv7();
        break;
      case ID_AUTO_INCREMENT:
        break;
      default:
        throw new DBError(`unsupported id strategy: ${this.idStrategy}`);
    }

    if (id !== undefined) {
      this.db
        .prepare(
          `INSERT INTO media (id, path, mime, size, width, height, hash, original_id, variant_key, created_by, owned_by, status)
           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        )
        .run(
          id,
          m.path,
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
        );
      return this.findById(id);
    }

    const info = this.db
      .prepare(
        `INSERT INTO media (path, mime, size, width, height, hash, original_id, variant_key, created_by, owned_by, status)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      )
      .run(
        m.path,
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
      );
    return this.findById(Number(info.lastInsertRowid));
  }

  findById(id: string | number): Media {
    if (id === undefined || id === null || id === "") {
      throw new DBError("db: invalid input: empty id");
    }
    const row = this.db.prepare("SELECT * FROM media WHERE id = ?").get(id) as Row | undefined;
    if (!row) {
      throw new DBError("db: media not found");
    }
    return rowToMedia(row);
  }

  findByHash(hash: string): Media {
    if (!hash) {
      throw new DBError("db: invalid input: empty hash");
    }
    let row = this.db
      .prepare("SELECT * FROM media WHERE hash = ? AND variant_key IS NULL LIMIT 1")
      .get(hash) as Row | undefined;
    if (!row) {
      row = this.db
        .prepare("SELECT * FROM media WHERE hash = ? LIMIT 1")
        .get(hash) as Row | undefined;
    }
    if (!row) {
      throw new DBError("db: media not found");
    }
    return rowToMedia(row);
  }

  listChildren(parentId: string | number): Media[] {
    const rows = this.db
      .prepare("SELECT * FROM media WHERE original_id = ?")
      .all(parentId) as Row[];
    return rows.map(rowToMedia);
  }

  deleteById(id: string | number): void {
    if (id === undefined || id === null || id === "") {
      throw new DBError("db: invalid input: empty id");
    }
    const children = this.listChildren(id);
    const del = this.db.prepare("DELETE FROM media WHERE id = ?");
    for (const ch of children) {
      del.run(ch.id!);
    }
    const info = del.run(id);
    if (info.changes === 0) {
      throw new DBError("db: media not found");
    }
  }
}

export function createMediaRepo(db: Database.Database, idStrategy?: string): MediaRepo {
  return new MediaRepo(db, idStrategy || ID_AUTO_INCREMENT);
}

export function isNotFound(err: unknown): boolean {
  return err instanceof DBError && err.message.includes("not found");
}
