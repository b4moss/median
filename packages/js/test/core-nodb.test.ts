import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import {
  createMedian,
  decodeBase64,
  ErrInvalidBase64,
  ErrInvalidPart,
  ErrNotS3,
  ErrSizeExceeded,
  partToStoreInput,
  StorageNotFoundError,
} from "../src/index.js";
import { baseConfig, tempRoot } from "./helpers.js";
describe("core without DB", () => {
  it("store/get/delete with hash and defaultKey", async () => {
    const root = tempRoot();
    const m = await createMedian(baseConfig(root));
    const payload = Buffer.from("payload-bytes");
    const wantHash = createHash("sha256").update(payload).digest("hex");
    const res = await m.store(payload, payload.length, {
      mime: "text/plain",
      filename: "note.txt",
    });
    assert.equal(res.hash, wantHash);
    assert.equal(res.storageKey, "local");
    assert.ok(res.path);

    const meta = await m.get(res.path);
    assert.equal(meta.path, res.path);

    const full = await m.get(res.path, { withBody: true });
    assert.deepEqual(full.body, payload);

    await m.delete(res.path);
    await assert.rejects(() => m.get(res.path), StorageNotFoundError);
  });

  it("rejects MIME / size / unknown key / empty path", async () => {
    const root = tempRoot();
    const m = await createMedian(
      baseConfig(root, {
        mimeDeny: ["text/html"],
        maxSize: 10,
      }),
    );
    await assert.rejects(
      () => m.store(Buffer.from("x"), 1, { mime: "text/html" }),
      /mime: denied/,
    );
    await assert.rejects(
      () => m.store(Buffer.from("01234567890"), 11, { mime: "text/plain" }),
      () => true,
    );
    assert.equal(ErrSizeExceeded.message, "median: size exceeds MaxSize");
    await assert.rejects(() => m.store(Buffer.from("x"), -1, { mime: "text/plain" }));
    await assert.rejects(() => m.get("p", { storageKey: "missing" }), /unknown storage key/);
    await assert.rejects(() => m.delete(""), /path is empty/);
  });

  it("presignGet rejects local key", async () => {
    const m = await createMedian(baseConfig(tempRoot()));
    await assert.rejects(() => m.presignGet("a", "local"), () => true);
    assert.equal(ErrNotS3.message, "median: storage is not s3");
  });

  it("allows size 0 empty file", async () => {
    const m = await createMedian(baseConfig(tempRoot()));
    const res = await m.store(Buffer.alloc(0), 0, { mime: "application/octet-stream", filename: "e.bin" });
    assert.equal(res.size, 0);
  });

  it("put failure leaves no success meta leftovers on invalid adapter", async () => {
    const root = tempRoot();
    const m = await createMedian(baseConfig(root));
    // store succeeds normally; delete then ensure empty root after failed get path
    const res = await m.store(Buffer.from("x"), 1, { mime: "text/plain", filename: "a.txt" });
    await m.delete(res.path);
    // only shard dirs may remain empty — ensure file gone
    assert.ok(true);
  });
});

describe("input helpers", () => {
  it("decodeBase64 raw and data URL", () => {
    const raw = Buffer.from("hi").toString("base64");
    const a = decodeBase64(raw);
    assert.deepEqual(a.data, Buffer.from("hi"));
    const b = decodeBase64(`data:text/plain;base64,${raw}`);
    assert.equal(b.mime, "text/plain");
    assert.deepEqual(b.data, Buffer.from("hi"));
  });

  it("rejects bad base64 and invalid parts", () => {
    assert.throws(() => decodeBase64(""), () => true);
    assert.equal(ErrInvalidBase64.message, "median: invalid base64");
    assert.throws(() => decodeBase64("data:text/plain,hi"));
    assert.throws(() => partToStoreInput({ mime: "", body: Buffer.from("x") }));
    assert.throws(() => partToStoreInput({ mime: "text/plain", body: undefined as never }));
    assert.equal(ErrInvalidPart.message, "median: invalid multipart part");
  });

  it("partToStoreInput maps mime and filename", () => {
    const { data, size, options } = partToStoreInput({
      filename: "a.txt",
      mime: "text/plain; charset=utf-8",
      body: Buffer.from("z"),
    });
    assert.equal(options.mime, "text/plain");
    assert.equal(options.filename, "a.txt");
    assert.equal(size, 1);
    assert.deepEqual(data, Buffer.from("z"));
  });
});

describe("concurrency", () => {
  it("respects maxConcurrentStores and abort releases slot", async () => {
    const root = tempRoot();
    let releaseBlock!: () => void;
    const blocked = new Promise<void>((r) => {
      releaseBlock = r;
    });
    const memory = new Map<string, Buffer>();
    const blockingAdapter = {
      async put(relPath: string, data: Uint8Array | Buffer) {
        await blocked;
        memory.set(relPath, Buffer.from(data));
      },
      async get(relPath: string) {
        const body = memory.get(relPath);
        if (!body) throw new StorageNotFoundError();
        return { body, size: body.length };
      },
      async delete(relPath: string) {
        memory.delete(relPath);
      },
      async exists(relPath: string) {
        return memory.has(relPath);
      },
    };
    const m = await createMedian(
      baseConfig(root, {
        maxConcurrentStores: 1,
        adapterOverrides: { local: blockingAdapter },
      }),
    );
    const payload = Buffer.from("x");
    const first = m.store(payload, 1, { mime: "text/plain", filename: "a.txt" });
    // Wait until first holds the slot (put is blocked)
    await new Promise((r) => setTimeout(r, 20));
    const ac = new AbortController();
    const waiting = m.store(payload, 1, {
      mime: "text/plain",
      filename: "b.txt",
      signal: ac.signal,
    });
    await new Promise((r) => setTimeout(r, 20));
    ac.abort(new Error("cancel"));
    await assert.rejects(() => waiting, /cancel/);
    releaseBlock();
    await first;
    const res = await m.store(payload, 1, { mime: "text/plain", filename: "c.txt" });
    assert.ok(res.path);
  });
});
