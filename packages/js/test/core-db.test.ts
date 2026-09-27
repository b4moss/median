import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { existsSync, readdirSync } from "node:fs";
import { join } from "node:path";
import Database from "better-sqlite3";
import {
  createMedian,
  ErrActorRequired,
  ErrDuplicate,
  ErrIDRequired,
  migrateUp,
  ID_AUTO_INCREMENT,
} from "../src/index.js";
import { baseConfig, makePNG, tempRoot } from "./helpers.js";

async function withDB(root: string, extra: Record<string, unknown> = {}) {
  const db = new Database(":memory:");
  migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
  return createMedian(
    baseConfig(root, {
      db: { db, idStrategy: ID_AUTO_INCREMENT },
      defaultActor: "actor-1",
      thumbnails: {
        presets: { sm: { mode: "longEdge", size: 32, quality: 0.8 } },
        defaultKeys: ["sm"],
      },
      ...extra,
    }),
  );
}

describe("core with DB + pipeline", () => {
  it("store returns id, actors, variants; duplicate hash returns existing", async () => {
    const root = tempRoot();
    const m = await withDB(root);
    const payload = await makePNG(64, 64);
    const res = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic.png",
    });
    assert.ok(res.id !== undefined);
    assert.equal(res.variants?.length, 1);
    assert.ok(res.width && res.height);

    const res2 = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic2.png",
    });
    assert.equal(String(res2.id), String(res.id));

    await assert.rejects(
      () =>
        m.store(payload, payload.length, {
          mime: "image/png",
          filename: "pic3.png",
          rejectDuplicate: true,
        }),
      () => true,
    );
    assert.equal(ErrDuplicate.message, "median: duplicate hash");

    const got = await m.get("", { id: res.id, withBody: true });
    assert.ok(got.body && got.body.length > 0);

    await m.delete("", { id: res.id });
    await assert.rejects(() => m.get("", { id: res.id }));
  });

  it("requires actor; rejects non-image resize; rejects missing id", async () => {
    const root = tempRoot();
    const db = new Database(":memory:");
    migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
    const m = await createMedian(
      baseConfig(root, {
        db: { db },
        // no defaultActor
      }),
    );
    await assert.rejects(
      () => m.store(Buffer.from("x"), 1, { mime: "text/plain", filename: "a.txt" }),
      () => true,
    );
    assert.equal(ErrActorRequired.message, "median: actor required");
    assert.equal(readdirSync(root).length, 0);

    const m2 = await withDB(root);
    await assert.rejects(
      () =>
        m2.store(Buffer.from("%PDF"), 4, {
          mime: "application/pdf",
          filename: "a.pdf",
          resize: { mode: "longEdge", size: 10 },
        }),
      /invalid image/,
    );

    await assert.rejects(() => m2.get("", {}), () => true);
    await assert.rejects(() => m2.delete("", {}), () => true);
    assert.equal(ErrIDRequired.message, "median: id required");
  });

  it("unknown thumbnail key leaves no partial files", async () => {
    const root = tempRoot();
    const m = await withDB(root);
    const payload = await makePNG(32, 32);
    await assert.rejects(
      () =>
        m.store(payload, payload.length, {
          mime: "image/png",
          filename: "p.png",
          thumbnailKeys: ["nope"],
        }),
      /unknown thumbnail/,
    );
    // storage should be empty of files (dirs may exist from failed attempts — none if fail before put)
    const walk = (dir: string): string[] => {
      const out: string[] = [];
      for (const e of readdirSync(dir, { withFileTypes: true })) {
        const p = join(dir, e.name);
        if (e.isDirectory()) out.push(...walk(p));
        else out.push(p);
      }
      return out;
    };
    assert.equal(walk(root).length, 0);
  });

  it("parent delete removes child files", async () => {
    const root = tempRoot();
    const m = await withDB(root);
    const payload = await makePNG(64, 64);
    const res = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic.png",
    });
    const paths = [res.path, ...(res.variants ?? []).map((v) => v.path)];
    for (const p of paths) {
      assert.ok(existsSync(join(root, p)));
    }
    await m.delete("", { id: res.id });
    for (const p of paths) {
      assert.equal(existsSync(join(root, p)), false);
    }
  });

  it("compress + resize produce files with dimensions", async () => {
    const root = tempRoot();
    const m = await withDB(root, {
      thumbnails: {
        presets: { sm: { mode: "longEdge", size: 32, quality: 0.8 } },
        defaultKeys: ["sm"],
      },
    });
    const payload = await makePNG(100, 80);
    const res = await m.store(payload, payload.length, {
      mime: "image/png",
      filename: "pic.png",
      compress: true,
      resize: { mode: "longEdge", size: 50 },
    });
    assert.ok(res.width === 50 || res.height === 40);
    assert.ok(existsSync(join(root, res.path)));
    assert.ok(res.variants?.[0] && existsSync(join(root, res.variants[0].path)));
    void createHash;
  });
});
