import { describe, it, before } from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdirSync, rmSync } from "node:fs";
import { join } from "node:path";
import Database from "better-sqlite3";
import sharp from "sharp";
import { CreateBucketCommand, HeadBucketCommand, S3Client } from "@aws-sdk/client-s3";
import {
  createMedian,
  migrateUp,
  ID_AUTO_INCREMENT,
  StorageNotFoundError,
} from "@b4moss/median";

function envOr(key: string, def: string): string {
  const v = process.env[key]?.trim();
  return v || def;
}

function localRoot(testName: string): string {
  const root = envOr("MEDIAN_E2E_LOCAL_ROOT", "");
  if (!root) {
    throw new Error("MEDIAN_E2E_LOCAL_ROOT is required (Docker POSIX volume mount)");
  }
  const dir = join(root, `${testName}-${Date.now().toString(36)}`);
  mkdirSync(dir, { recursive: true });
  return dir;
}

function s3Endpoint(): string {
  return envOr("MEDIAN_E2E_S3_ENDPOINT", "http://127.0.0.1:9000");
}

function s3Env() {
  return {
    accessKey: envOr("MEDIAN_E2E_S3_ACCESS_KEY", "median_e2e"),
    secretKey: envOr("MEDIAN_E2E_S3_SECRET_KEY", "median_e2e_secret"),
    bucket: envOr("MEDIAN_E2E_S3_BUCKET", "median-e2e"),
    region: envOr("MEDIAN_E2E_S3_REGION", "us-east-1"),
    endpoint: s3Endpoint(),
  };
}

function hashOf(buf: Buffer | Uint8Array): string {
  return createHash("sha256").update(buf).digest("hex");
}

async function waitRustFS(): Promise<void> {
  const endpoint = s3Endpoint().replace(/\/$/, "");
  const deadline = Date.now() + 60_000;
  let last = "";
  while (Date.now() < deadline) {
    try {
      const res = await fetch(`${endpoint}/health`);
      if (res.ok) return;
      last = `status ${res.status}`;
    } catch (e) {
      last = String(e);
    }
    await new Promise((r) => setTimeout(r, 500));
  }
  throw new Error(`rustfs not healthy at ${endpoint}: ${last}`);
}

async function ensureBucket(): Promise<void> {
  const { accessKey, secretKey, bucket, region, endpoint } = s3Env();
  const client = new S3Client({
    region,
    endpoint,
    forcePathStyle: true,
    credentials: { accessKeyId: accessKey, secretAccessKey: secretKey },
  });
  try {
    await client.send(new HeadBucketCommand({ Bucket: bucket }));
    return;
  } catch {
    /* create */
  }
  await client.send(new CreateBucketCommand({ Bucket: bucket }));
}

function openMemDB(): Database.Database {
  const db = new Database(":memory:");
  migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
  return db;
}

function findRow(db: Database.Database, id: string | number) {
  return db.prepare("SELECT * FROM media WHERE id = ?").get(id) as
    | {
        id: number;
        file_path: string;
        file_name: string;
        original_file_name: string | null;
        mime: string;
        size: number;
        hash: string;
        created_by: string;
        owned_by: string;
        original_id: number | null;
        variant_key: string | null;
      }
    | undefined;
}

function listChildren(db: Database.Database, parentId: string | number) {
  return db
    .prepare("SELECT * FROM media WHERE original_id = ?")
    .all(parentId) as Array<{
    id: number;
    file_path: string;
    hash: string;
    size: number;
    variant_key: string | null;
    original_id: number | null;
  }>;
}

function localCfg(root: string, extra: Record<string, unknown> = {}) {
  return {
    storages: { local: { driver: "local", path: root } },
    defaultKey: "local",
    dirLetterCount: 1,
    dirNestDepth: 2,
    ...extra,
  };
}

function s3Cfg(extra: Record<string, unknown> = {}) {
  const s = s3Env();
  return {
    storages: {
      s3: {
        driver: "s3",
        bucket: s.bucket,
        region: s.region,
        endpoint: s.endpoint,
        forcePathStyle: true,
        accessKey: s.accessKey,
        secretKey: s.secretKey,
        prefix: `e2e/${Date.now()}`,
      },
    },
    defaultKey: "s3",
    dirLetterCount: 1,
    dirNestDepth: 2,
    ...extra,
  };
}

async function makePNG(w: number, h: number): Promise<Buffer> {
  return sharp({
    create: { width: w, height: h, channels: 3, background: { r: 200, g: 40, b: 40 } },
  })
    .png()
    .toBuffer();
}

async function httpGetBody(url: string): Promise<Buffer> {
  const res = await fetch(url);
  const buf = Buffer.from(await res.arrayBuffer());
  if (!res.ok) {
    throw new Error(`presign GET ${res.status}: ${buf.subarray(0, 200).toString()}`);
  }
  return buf;
}

describe("E2E L1 Local no-DB", () => {
  it("Store → Get → Delete", async () => {
    const root = localRoot("L1");
    try {
      const m = await createMedian(localCfg(root));
      const payload = Buffer.from("hello-median-e2e-l1");
      const res = await m.store(payload, payload.length, {
        mime: "application/octet-stream",
        filename: "l1.bin",
      });
      assert.equal(res.hash, hashOf(payload));
      assert.equal(res.size, payload.length);
      const got = await m.get(res.path, { withBody: true });
      assert.deepEqual(got.body, payload);
      await m.delete(res.path);
      await assert.rejects(() => m.get(res.path), StorageNotFoundError);
    } finally {
      rmSync(root, { recursive: true, force: true });
    }
  });
});

describe("E2E S1 RustFS no-DB", () => {
  before(async () => {
    await waitRustFS();
    await ensureBucket();
  });

  it("Store → Get → PresignGet → Delete", async () => {
    const m = await createMedian(s3Cfg());
    const payload = Buffer.from("hello-median-e2e-s1");
    const res = await m.store(payload, payload.length, {
      mime: "application/octet-stream",
      filename: "s1.bin",
    });
    assert.equal(res.hash, hashOf(payload));
    const got = await m.get(res.path, { withBody: true });
    assert.deepEqual(got.body, payload);
    const url = await m.presignGet(res.path, res.storageKey);
    assert.ok(url);
    assert.deepEqual(await httpGetBody(url), payload);
    await m.delete(res.path);
    await assert.rejects(() => m.get(res.path), StorageNotFoundError);
  });
});

async function runCRUD(cfgBase: Record<string, unknown>, withPresign: boolean) {
  const sqlite = openMemDB();
  const m = await createMedian({
    ...cfgBase,
    db: { db: sqlite, idStrategy: ID_AUTO_INCREMENT },
    defaultActor: "e2e-actor",
  });
  const a = Buffer.from("crud-fixture-A");
  const b = Buffer.from("crud-fixture-B-different");

  const res = await m.store(a, a.length, { mime: "application/octet-stream", filename: "a.bin" });
  assert.ok(res.id !== undefined);
  const row = findRow(sqlite, res.id!);
  assert.ok(row);
  assert.equal(row.file_path, res.path);
  assert.equal(row.hash, res.hash);
  assert.equal(row.size, res.size);
  assert.equal(row.created_by, "e2e-actor");
  assert.equal(row.owned_by, "e2e-actor");

  const got = await m.get("", { id: res.id, withBody: true });
  assert.deepEqual(got.body, a);

  if (withPresign) {
    const url = await m.presignGet(res.path, res.storageKey);
    assert.deepEqual(await httpGetBody(url), a);
  }

  const res2 = await m.store(a, a.length, { mime: "application/octet-stream", filename: "a2.bin" });
  assert.equal(String(res2.id), String(res.id));
  const row2 = findRow(sqlite, res.id!);
  assert.ok(row2);
  assert.equal(row2.hash, res.hash);
  assert.equal(row2.file_path, res.path);

  await m.delete("", { id: res.id });
  assert.equal(findRow(sqlite, res.id!), undefined);

  const resB = await m.store(b, b.length, { mime: "application/octet-stream", filename: "b.bin" });
  assert.notEqual(String(resB.id), String(res.id));
  const rowB = findRow(sqlite, resB.id!);
  assert.ok(rowB);
  assert.equal(rowB.hash, hashOf(b));
  const gotB = await m.get("", { id: resB.id, withBody: true });
  assert.deepEqual(gotB.body, b);

  await m.delete("", { id: resB.id });
  assert.equal(findRow(sqlite, resB.id!), undefined);
  await assert.rejects(() => m.get("", { id: resB.id }));
}

describe("E2E C-L Local + DB CRUD", () => {
  it("CRUD with metadata", async () => {
    const root = localRoot("CL");
    try {
      await runCRUD(localCfg(root), false);
    } finally {
      rmSync(root, { recursive: true, force: true });
    }
  });
});

describe("E2E C-S RustFS + DB CRUD", () => {
  before(async () => {
    await waitRustFS();
    await ensureBucket();
  });

  it("CRUD with metadata + Presign", async () => {
    await runCRUD(s3Cfg(), true);
  });
});

async function runThumb(cfgBase: Record<string, unknown>) {
  const sqlite = openMemDB();
  const m = await createMedian({
    ...cfgBase,
    db: { db: sqlite, idStrategy: ID_AUTO_INCREMENT },
    defaultActor: "e2e-actor",
    thumbnails: {
      presets: { sm: { mode: "longEdge", size: 32, quality: 0.8 } },
      defaultKeys: ["sm"],
    },
  });
  const payload = await makePNG(64, 64);
  const res = await m.store(payload, payload.length, { mime: "image/png", filename: "pic.png" });
  assert.ok(res.id !== undefined);
  assert.equal(res.variants?.length, 1);
  assert.equal(res.variants![0]!.key, "sm");
  const v = res.variants![0]!;

  const parent = await m.get("", { id: res.id, withBody: true });
  assert.equal(parent.body!.length, parent.size);
  const childGet = await m.get("", { id: v.id, withBody: true });
  assert.equal(childGet.body!.length, childGet.size);
  assert.equal(childGet.size, v.size);

  const prow = findRow(sqlite, res.id!);
  assert.ok(prow);
  assert.equal(prow.original_id, null);
  assert.equal(prow.variant_key, null);
  const children = listChildren(sqlite, res.id!);
  assert.equal(children.length, 1);
  assert.equal(children[0]!.variant_key, "sm");
  assert.equal(children[0]!.file_path, v.path);
  assert.equal(children[0]!.hash, v.hash);

  await m.delete("", { id: res.id });
  assert.equal(findRow(sqlite, res.id!), undefined);
  assert.equal(findRow(sqlite, v.id!), undefined);
  await assert.rejects(() => m.get("", { id: res.id }));
}

describe("E2E T-L Local + DB thumbnails", () => {
  it("variants + cascade", async () => {
    const root = localRoot("TL");
    try {
      await runThumb(localCfg(root));
    } finally {
      rmSync(root, { recursive: true, force: true });
    }
  });
});

describe("E2E T-S RustFS + DB thumbnails", () => {
  before(async () => {
    await waitRustFS();
    await ensureBucket();
  });

  it("variants + cascade", async () => {
    await runThumb(s3Cfg());
  });
});
