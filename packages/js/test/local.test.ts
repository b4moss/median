import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { createLocalFS, StorageInvalidPathError, StorageNotFoundError } from "../src/index.js";
import { tempRoot } from "./helpers.js";

describe("storage/local", () => {
  it("put/get round-trip, nested path, overwrite", async () => {
    const fs = await createLocalFS(tempRoot());
    const payload = Buffer.from("hello-world");
    await fs.put("a/b/c.txt", payload);
    const got = await fs.get("a/b/c.txt");
    assert.deepEqual(got.body, payload);
    await fs.put("a/b/c.txt", Buffer.from("overwrite"));
    assert.equal((await fs.get("a/b/c.txt")).body.toString(), "overwrite");
  });

  it("delete and exists", async () => {
    const fs = await createLocalFS(tempRoot());
    assert.equal(await fs.exists("missing.txt"), false);
    await fs.put("f.txt", Buffer.from("x"));
    assert.equal(await fs.exists("f.txt"), true);
    await fs.delete("f.txt");
    assert.equal(await fs.exists("f.txt"), false);
  });

  it("rejects .. / absolute / empty path", async () => {
    const fs = await createLocalFS(tempRoot());
    await assert.rejects(() => fs.put("..", Buffer.from("x")), StorageInvalidPathError);
    await assert.rejects(() => fs.put("/abs", Buffer.from("x")), StorageInvalidPathError);
    await assert.rejects(() => fs.put("", Buffer.from("x")), StorageInvalidPathError);
  });

  it("get missing is not found; delete missing is not found", async () => {
    const fs = await createLocalFS(tempRoot());
    await assert.rejects(() => fs.get("nope"), StorageNotFoundError);
    await assert.rejects(() => fs.delete("nope"), StorageNotFoundError);
  });
});
