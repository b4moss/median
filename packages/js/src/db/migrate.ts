import type Database from "better-sqlite3";
import { createMediaSQL, idStrategyFromEnv, normalizeTableName } from "./schema.js";

export type SqliteDatabase = Database.Database;

function splitSQL(q: string): string[] {
  return q
    .split(";")
    .map((s) => s.trim())
    .filter(Boolean);
}

export function migrateUp(
  db: SqliteDatabase,
  dialect: string,
  idStrategy?: string,
  tableName?: string,
): void {
  const strategy = idStrategy || idStrategyFromEnv();
  const q = createMediaSQL(dialect, strategy, tableName);
  for (const stmt of splitSQL(q)) {
    db.exec(stmt);
  }
}

export function migrateDown(db: SqliteDatabase, tableName?: string): void {
  const table = normalizeTableName(tableName);
  db.exec(`DROP TABLE IF EXISTS ${table}`);
}
