import { describe, it } from "node:test";
import assert from "node:assert/strict";
import Database from "better-sqlite3";
import {
  createMediaRepo,
  createMediaSQL,
  ID_AUTO_INCREMENT,
  ID_UUID_V4,
  migrateDown,
  migrateUp,
} from "../src/index.js";

describe("db / migrations", () => {
  it("migrateUp creates media table; migrateDown drops it", () => {
    const db = new Database(":memory:");
    migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
    const cols = db.prepare("PRAGMA table_info(media)").all() as { name: string }[];
    const names = cols.map((c) => c.name);
    for (const need of [
      "id",
      "file_path",
      "file_name",
      "original_file_name",
      "mime",
      "size",
      "hash",
      "original_id",
      "variant_key",
      "created_by",
      "owned_by",
    ]) {
      assert.ok(names.includes(need), need);
    }
    assert.ok(!names.includes("path"));
    migrateDown(db);
    const tables = db.prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='media'").all();
    assert.equal(tables.length, 0);
  });

  it("auto_increment integer PK and uuid text PK", () => {
    const db = new Database(":memory:");
    migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
    const repo = createMediaRepo(db, ID_AUTO_INCREMENT);
    const row = repo.create({
      filePath: "a/b.png",
      fileName: "b.png",
      originalFileName: "pic.png",
      mime: "image/png",
      size: 1,
      hash: "a".repeat(64),
      createdBy: "u",
      ownedBy: "u",
    });
    assert.equal(typeof row.id, "number");
    assert.equal(row.fileName, "b.png");
    assert.equal(row.originalFileName, "pic.png");

    const db2 = new Database(":memory:");
    migrateUp(db2, "sqlite3", ID_UUID_V4);
    const repo2 = createMediaRepo(db2, ID_UUID_V4);
    const row2 = repo2.create({
      filePath: "a/b.png",
      fileName: "b.png",
      mime: "image/png",
      size: 1,
      hash: "b".repeat(64),
      createdBy: "u",
      ownedBy: "u",
    });
    assert.equal(typeof row2.id, "string");
  });

  it("Create / FindByID / FindByHash / ListChildren / DeleteByID cascade", () => {
    const db = new Database(":memory:");
    migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
    const repo = createMediaRepo(db);
    const parent = repo.create({
      filePath: "p.png",
      fileName: "p.png",
      mime: "image/png",
      size: 10,
      hash: "c".repeat(64),
      createdBy: "a",
      ownedBy: "a",
    });
    const child = repo.create({
      filePath: "c.png",
      fileName: "c.png",
      mime: "image/png",
      size: 2,
      hash: "d".repeat(64),
      originalId: parent.id,
      variantKey: "sm",
      createdBy: "a",
      ownedBy: "a",
    });
    assert.equal(repo.findById(parent.id!).id, parent.id);
    assert.equal(repo.findByHash("c".repeat(64)).id, parent.id);
    assert.equal(repo.listChildren(parent.id!).length, 1);
    assert.equal(repo.listChildren(parent.id!)[0]!.id, child.id);
    repo.deleteById(parent.id!);
    assert.throws(() => repo.findById(parent.id!));
    assert.throws(() => repo.findById(child.id!));
  });

  it("custom table name", () => {
    const db = new Database(":memory:");
    migrateUp(db, "sqlite3", ID_AUTO_INCREMENT, "medias");
    const repo = createMediaRepo(db, { tableName: "medias" });
    const row = repo.create({
      filePath: "x.bin",
      fileName: "x.bin",
      originalFileName: "orig.bin",
      mime: "application/octet-stream",
      size: 1,
      hash: "e".repeat(64),
      createdBy: "a",
      ownedBy: "a",
    });
    assert.equal(repo.findById(row.id!).originalFileName, "orig.bin");
    migrateDown(db, "medias");
  });

  it("rejects unsupported dialect / strategy / missing fields / not found", () => {
    assert.throws(() => createMediaSQL("oracle", ID_AUTO_INCREMENT));
    assert.throws(() => createMediaSQL("sqlite3", "bad"));
    const db = new Database(":memory:");
    migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
    const repo = createMediaRepo(db);
    assert.throws(() =>
      repo.create({
        filePath: "",
        fileName: "x",
        mime: "x",
        size: 0,
        hash: "h",
        createdBy: "a",
        ownedBy: "a",
      }),
    );
    assert.throws(() => repo.findById(999));
  });
});
