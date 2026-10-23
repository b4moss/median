import { describe, it } from "node:test";
import assert from "node:assert/strict";
import Database from "better-sqlite3";
import {
  createMedian,
  createMediaRepo,
  ErrDBRequired,
  ErrIDRequired,
  ErrUnknownKey,
  migrateUp,
  ID_AUTO_INCREMENT,
  DBErrNotFound,
} from "../src/index.js";
import { baseConfig, makePNG, tempRoot } from "./helpers.js";

async function withFamilyDB(
  root: string,
  extra: Record<string, unknown> = {},
  tableName?: string,
) {
  const db = new Database(":memory:");
  migrateUp(db, "sqlite3", ID_AUTO_INCREMENT, tableName);
  const m = await createMedian(
    baseConfig(root, {
      db: { db, idStrategy: ID_AUTO_INCREMENT, tableName },
      defaultActor: "actor-1",
      thumbnails: {
        presets: {
          sm: { mode: "longEdge", size: 32, quality: 0.8 },
          md: { mode: "longEdge", size: 48, quality: 0.8 },
          aa: { mode: "longEdge", size: 40, quality: 0.8 },
          zz: { mode: "longEdge", size: 40, quality: 0.8 },
        },
        defaultKeys: ["sm", "md"],
      },
      ...extra,
    }),
  );
  return { m, db };
}

function setStatus(db: Database.Database, table: string, id: string | number, status: string) {
  db.prepare(`UPDATE ${table} SET status = ? WHERE id = ?`).run(status, id);
}

function setWidth(db: Database.Database, table: string, id: string | number, width: number | null) {
  db.prepare(`UPDATE ${table} SET width = ? WHERE id = ?`).run(width, id);
}

describe("getWithAllThumbnails", () => {
  it("returns family by original id with width/key order", async () => {
    const root = tempRoot();
    const { m } = await withFamilyDB(root);
    const payload = await makePNG(80, 80);
    const res = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic.png",
      thumbnailKeys: ["sm", "md", "aa", "zz"],
    });
    assert.equal(res.variants?.length, 4);

    const got = await m.getWithAllThumbnails(res.id!);
    assert.equal(String(got.id), String(res.id));
    assert.equal(got.path, res.path);
    assert.equal(got.hash, res.hash);
    assert.equal(got.storageKey, "local");
    assert.ok(Array.isArray(got.variants));
    assert.equal(got.variants!.length, 4);
    assert.deepEqual(
      got.variants!.map((v) => v.key),
      ["sm", "aa", "zz", "md"],
    );
    for (const v of got.variants!) {
      assert.ok(v.key && v.path && v.mime && v.hash && v.id !== undefined);
      assert.ok(v.size > 0);
      assert.equal(v.storageKey, "local");
    }

    const childId = res.variants![0]!.id!;
    const fromChild = await m.getWithAllThumbnails(childId);
    assert.equal(String(fromChild.id), String(res.id));
    assert.equal(fromChild.path, res.path);
    assert.equal(fromChild.variants!.length, 4);
  });

  it("resolves inactive derived id and filters inactive children", async () => {
    const root = tempRoot();
    const { m, db } = await withFamilyDB(root);
    const payload = await makePNG(64, 64);
    const res = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic.png",
      thumbnailKeys: ["sm", "md"],
    });
    const inactiveId = res.variants![0]!.id!;
    const inactiveKey = res.variants![0]!.key;
    setStatus(db, "media", inactiveId, "inactive");

    const got = await m.getWithAllThumbnails(inactiveId);
    assert.equal(String(got.id), String(res.id));
    assert.equal(got.variants!.length, 1);
    assert.notEqual(got.variants![0]!.key, inactiveKey);
  });

  it("returns empty variants array when no thumbnails", async () => {
    const root = tempRoot();
    const db = new Database(":memory:");
    migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
    const m = await createMedian(
      baseConfig(root, {
        db: { db, idStrategy: ID_AUTO_INCREMENT },
        defaultActor: "actor-1",
      }),
    );
    const payload = Buffer.from("plain-no-thumbs");
    const res = await m.store(payload, payload.length, {
      mime: "text/plain",
      filename: "note.txt",
    });
    const got = await m.getWithAllThumbnails(res.id!);
    assert.ok(Array.isArray(got.variants));
    assert.equal(got.variants!.length, 0);
  });

  it("orders width 0 before numeric and null last", async () => {
    const root = tempRoot();
    const { m, db } = await withFamilyDB(root);
    const payload = await makePNG(64, 64);
    const res = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic.png",
      thumbnailKeys: ["sm", "md"],
    });
    let smId: string | number | undefined;
    let mdId: string | number | undefined;
    for (const v of res.variants!) {
      if (v.key === "sm") smId = v.id;
      if (v.key === "md") mdId = v.id;
    }
    setWidth(db, "media", smId!, 0);
    setWidth(db, "media", mdId!, null);

    const repo = createMediaRepo(db, ID_AUTO_INCREMENT);
    repo.create({
      filePath: "x/mid.png",
      fileName: "mid.png",
      mime: "image/png",
      size: 1,
      hash: "midhash",
      width: 10,
      originalId: res.id,
      variantKey: "mid",
      createdBy: "actor-1",
      ownedBy: "actor-1",
    });

    const got = await m.getWithAllThumbnails(res.id!);
    assert.deepEqual(
      got.variants!.map((v) => v.key),
      ["sm", "mid", "md"],
    );
  });

  it("storageKey omit uses default; explicit uses that key", async () => {
    const root = tempRoot();
    const root2 = tempRoot();
    const db = new Database(":memory:");
    migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
    const m = await createMedian(
      baseConfig(root, {
        storages: {
          local: { driver: "local", path: root },
          alt: { driver: "local", path: root2 },
        },
        db: { db, idStrategy: ID_AUTO_INCREMENT },
        defaultActor: "actor-1",
      }),
    );
    const payload = Buffer.from("x");
    const res = await m.store(payload, payload.length, {
      mime: "text/plain",
      filename: "a.txt",
      storageKey: "alt",
    });
    const def = await m.getWithAllThumbnails(res.id!);
    assert.equal(def.storageKey, "local");
    const alt = await m.getWithAllThumbnails(res.id!, { storageKey: "alt" });
    assert.equal(alt.storageKey, "alt");
  });

  it("works with custom table name", async () => {
    const root = tempRoot();
    const { m } = await withFamilyDB(root, {}, "medias");
    const payload = await makePNG(48, 48);
    const res = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic.png",
    });
    const got = await m.getWithAllThumbnails(res.id!);
    assert.equal(got.variants!.length, 2);
    assert.ok(got.variants![0]!.width <= got.variants![1]!.width);
  });

  it("errors: db required, not found, inactive parent, id, unknown key", async () => {
    const root = tempRoot();
    const noDB = await createMedian(baseConfig(root));
    await assert.rejects(() => noDB.getWithAllThumbnails(1), (err: Error) => {
      assert.equal(err.message, ErrDBRequired.message);
      assert.notEqual(err.message, DBErrNotFound.message);
      return true;
    });

    const { m, db } = await withFamilyDB(root);
    const payload = await makePNG(40, 40);
    const res = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic.png",
    });

    await assert.rejects(() => m.getWithAllThumbnails(999999), (err: Error) => {
      assert.match(err.message, /not found/);
      return true;
    });

    const childId = res.variants![0]!.id!;
    setStatus(db, "media", res.id!, "inactive");
    await assert.rejects(() => m.getWithAllThumbnails(childId), (err: Error) => {
      assert.match(err.message, /not found/);
      return true;
    });
    setStatus(db, "media", res.id!, "active");

    await assert.rejects(() => m.getWithAllThumbnails("" as unknown as string), () => true);
    assert.equal(ErrIDRequired.message, "median: id required");

    await assert.rejects(
      () => m.getWithAllThumbnails(res.id!, { storageKey: "nope" }),
      (err: Error) => {
        assert.match(err.message, /unknown storage key/);
        assert.equal(ErrUnknownKey.message, "median: unknown storage key");
        return true;
      },
    );
  });

  it("does not change get for inactive rows (U9)", async () => {
    const root = tempRoot();
    const { m, db } = await withFamilyDB(root);
    const payload = await makePNG(32, 32);
    const res = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic.png",
    });
    setStatus(db, "media", res.id!, "inactive");
    const got = await m.get("", { id: res.id });
    assert.equal(String(got.id), String(res.id));
    await assert.rejects(() => m.getWithAllThumbnails(res.id!), /not found/);
  });
});
