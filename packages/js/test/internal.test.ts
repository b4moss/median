import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  buildStoragePath,
  checkMIME,
  randomFilename,
  resolveFilename,
  sanitizePreserve,
} from "../src/index.js";

describe("filename / shardian path", () => {
  it("preserve keeps unicode and strips dangerous chars", () => {
    assert.equal(sanitizePreserve("hello.txt"), "hello.txt");
    assert.equal(sanitizePreserve("写真.png"), "写真.png");
    assert.equal(sanitizePreserve("  .foo.  "), "foo");
    const sep = sanitizePreserve("a/b\\c.txt");
    assert.ok(!/[\\/]/.test(sep));
  });

  it("random hex + extension", () => {
    const got = randomFilename("img.PNG", 32);
    assert.ok(got.endsWith(".PNG"));
    const base = got.slice(0, -4);
    assert.equal(base.length, 32);
    assert.match(base, /^[0-9a-f]+$/);
  });

  it("buildStoragePath is deterministic and has no leading slash / ..", () => {
    const p = buildStoragePath("abc1234.jpg", 1, 4);
    assert.ok(!p.startsWith("/"));
    assert.ok(p.endsWith("abc1234.jpg"));
    assert.ok(!p.includes(".."));
    assert.equal(buildStoragePath("abc1234.jpg", 1, 4), p);
  });

  it("rejects empty sanitize / invalid random len / empty path name", () => {
    assert.throws(() => sanitizePreserve("..."));
    assert.throws(() => sanitizePreserve("\0"));
    assert.throws(() => randomFilename("x", 0));
    assert.throws(() => randomFilename("x", 31));
    assert.throws(() => randomFilename("x", 200));
    assert.throws(() => buildStoragePath("", 1, 4));
    assert.throws(() => resolveFilename("x", "weird" as never));
  });
});

describe("MIME check", () => {
  it("allow/deny rules", () => {
    checkMIME("image/png", [], []);
    checkMIME("image/png", ["image/png"], []);
    checkMIME("image/png", [], ["application/x-msdownload"]);
    assert.throws(() => checkMIME("text/html", [], ["text/html"]));
    assert.throws(() => checkMIME("image/gif", ["image/png"], []));
    assert.throws(() => checkMIME("", [], []));
  });
});
