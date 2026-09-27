import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  createS3FS,
  ErrNotS3,
  StorageInvalidPathError,
  StorageNotFoundError,
  createMedian,
  type S3API,
  type S3Presigner,
} from "../src/index.js";
import { baseConfig, tempRoot } from "./helpers.js";

function mockS3() {
  const store = new Map<string, Buffer>();
  const client: S3API = {
    async send(command: unknown) {
      const name = (command as { constructor: { name: string } }).constructor.name;
      const input = (command as { input: Record<string, unknown> }).input;
      const key = String(input.Key);
      if (name === "PutObjectCommand") {
        store.set(key, Buffer.from(input.Body as Buffer));
        return {};
      }
      if (name === "GetObjectCommand") {
        const body = store.get(key);
        if (!body) {
          const err = new Error("NoSuchKey");
          err.name = "NoSuchKey";
          throw err;
        }
        return {
          Body: {
            async transformToByteArray() {
              return new Uint8Array(body);
            },
          },
          ContentLength: body.length,
        };
      }
      if (name === "DeleteObjectCommand") {
        store.delete(key);
        return {};
      }
      if (name === "HeadObjectCommand") {
        if (!store.has(key)) {
          const err = new Error("NotFound");
          err.name = "NotFound";
          (err as { $metadata: { httpStatusCode: number } }).$metadata = { httpStatusCode: 404 };
          throw err;
        }
        return {};
      }
      if (name === "FailCommand") {
        throw new Error("5xx boom");
      }
      throw new Error(`unexpected ${name}`);
    },
  };
  const lastTTL: { ms?: number } = {};
  const presigner: S3Presigner = {
    async presignGet(bucket, key, ttlMs) {
      lastTTL.ms = ttlMs;
      return `https://example.test/${bucket}/${key}?ttl=${ttlMs}`;
    },
  };
  return { client, presigner, store, lastTTL };
}

describe("storage/s3 + presignGet", () => {
  it("put/get/delete/exists and presign TTL", async () => {
    const { client, presigner, lastTTL } = mockS3();
    const fs = await createS3FS(
      { bucket: "b", prefix: "pfx", presignTTLMs: 3600_000 },
      client,
      presigner,
    );
    await fs.put("a/b.txt", Buffer.from("hello"));
    assert.equal(await fs.exists("a/b.txt"), true);
    const got = await fs.get("a/b.txt");
    assert.equal(got.body.toString(), "hello");
    const url = await fs.presignGet("a/b.txt");
    assert.match(url, /ttl=3600000/);
    assert.equal(lastTTL.ms, 3600_000);
    const url2 = await fs.presignGet("a/b.txt", 120_000);
    assert.match(url2, /ttl=120000/);
    await fs.delete("a/b.txt");
    assert.equal(await fs.exists("a/b.txt"), false);
  });

  it("rejects invalid path / not found / propagates 5xx", async () => {
    const { client, presigner } = mockS3();
    const fs = await createS3FS({ bucket: "b" }, client, presigner);
    await assert.rejects(() => fs.put("../x", Buffer.from("x")), StorageInvalidPathError);
    await assert.rejects(() => fs.get("missing"), StorageNotFoundError);
    await assert.rejects(async () => {
      await client.send({ constructor: { name: "FailCommand" }, input: {} });
    }, /5xx/);
  });

  it("median.presignGet uses defaultKey and rejects local", async () => {
    const { client, presigner, lastTTL } = mockS3();
    const s3fs = await createS3FS({ bucket: "b", presignTTLMs: 3600_000 }, client, presigner);
    const root = tempRoot();
    const m = await createMedian(
      baseConfig(root, {
        storages: {
          local: { driver: "local", path: root },
          s3: { driver: "s3", bucket: "b" },
        },
        defaultKey: "s3",
        adapterOverrides: { s3: s3fs },
        s3Overrides: { s3: s3fs },
      }),
    );
    const url = await m.presignGet("obj.txt");
    assert.match(url, /obj\.txt/);
    assert.equal(lastTTL.ms, 3600_000);

    const mLocal = await createMedian(baseConfig(root));
    await assert.rejects(() => mLocal.presignGet("x"), () => true);
    assert.equal(ErrNotS3.message.includes("s3"), true);
  });
});
