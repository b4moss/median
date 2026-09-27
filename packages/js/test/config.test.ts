import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { createMedian, DRIVER_S3 } from "../src/index.js";
import { baseConfig, tempRoot } from "./helpers.js";

describe("createMedian", () => {
  it("creates instance with valid local config and applies defaults", async () => {
    const root = tempRoot();
    const m = await createMedian(baseConfig(root));
    assert.equal(m.config.maxSize, 20 << 20);
    assert.equal(m.config.maxConcurrentStores, 20);
    assert.equal(m.config.filenameMode, "random");
    assert.equal(m.config.presignTTLMs, 60 * 60 * 1000);
  });

  it("accepts multiple storage keys and thumbnails", async () => {
    const root = tempRoot();
    const root2 = tempRoot();
    const m = await createMedian(
      baseConfig(root, {
        storages: {
          local: { driver: "local", path: root },
          other: { driver: "local", path: root2 },
        },
        thumbnails: {
          presets: { sm: { mode: "longEdge", size: 32 } },
          defaultKeys: ["sm"],
        },
      }),
    );
    assert.ok(m);
  });

  it("rejects missing defaultKey / empty storages / missing shardian", async () => {
    const root = tempRoot();
    await assert.rejects(() => createMedian({ ...baseConfig(root), defaultKey: "" }), /defaultKey/);
    await assert.rejects(
      () =>
        createMedian({
          defaultKey: "local",
          dirLetterCount: 1,
          dirNestDepth: 2,
          storages: {},
        }),
      /storages is empty/,
    );
    await assert.rejects(
      () =>
        createMedian({
          storages: { local: { driver: "local", path: root } },
          defaultKey: "local",
        } as never),
      /DirLetterCount/,
    );
  });

  it("rejects empty local root / unknown driver / s3 without bucket", async () => {
    await assert.rejects(
      () =>
        createMedian({
          storages: { local: { driver: "local", path: "" } },
          defaultKey: "local",
          dirLetterCount: 1,
          dirNestDepth: 2,
        }),
      /empty/,
    );
    await assert.rejects(
      () =>
        createMedian({
          storages: { x: { driver: "weird" } },
          defaultKey: "x",
          dirLetterCount: 1,
          dirNestDepth: 2,
        }),
      /unknown driver/,
    );
    await assert.rejects(
      () =>
        createMedian({
          storages: { s3: { driver: DRIVER_S3 } },
          defaultKey: "s3",
          dirLetterCount: 1,
          dirNestDepth: 2,
        }),
      /bucket/,
    );
  });

  it("rejects thumbnail defaultKeys missing from presets", async () => {
    const root = tempRoot();
    await assert.rejects(
      () =>
        createMedian(
          baseConfig(root, {
            thumbnails: { presets: {}, defaultKeys: ["sm"] },
          }),
        ),
      /missing from presets/,
    );
  });
});
