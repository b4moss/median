import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { createMediaSQL, ID_AUTO_INCREMENT, ID_UUID_V4 } from "../src/db/schema.js";
import { run } from "../src/cli.js";

function capture(args: string[]) {
  let out = "";
  let err = "";
  const code = run(
    args,
    { write: (s) => { out += s; } },
    { write: (s) => { err += s; } },
  );
  return { code, out, err };
}

describe("cli migrate dump", () => {
  for (const dialect of ["sqlite3", "postgres", "mysql"] as const) {
    it(`matches createMediaSQL for ${dialect}`, () => {
      const { code, out, err } = capture(["migrate", "dump", "--dialect", dialect]);
      assert.equal(code, 0);
      const want = createMediaSQL(dialect, ID_AUTO_INCREMENT).trim() + "\n";
      assert.equal(out, want);
      assert.match(out, /CREATE TABLE/);
      assert.match(out, /media/);
      assert.match(out, /file_path/);
      assert.doesNotMatch(err, /CREATE TABLE/);
    });
  }

  it("supports id-strategy and table", () => {
    const { code, out } = capture([
      "migrate",
      "dump",
      "--dialect",
      "sqlite3",
      "--id-strategy",
      "uuid_v4",
      "--table",
      "medias",
    ]);
    assert.equal(code, 0);
    assert.equal(out, createMediaSQL("sqlite3", ID_UUID_V4, "medias").trim() + "\n");
    assert.match(out, /CREATE TABLE medias/);
    assert.match(out, /TEXT PRIMARY KEY/);
  });

  it("rejects bad args", () => {
    const cases: string[][] = [
      ["migrate", "dump"],
      ["migrate", "dump", "--dialect", "oracle"],
      ["migrate", "dump", "--dialect", "sqlite3", "--id-strategy", "nope"],
      ["migrate", "dump", "--dialect", "sqlite3", "--table", "bad-name"],
      ["foo"],
      ["migrate"],
      ["migrate", "other"],
    ];
    for (const args of cases) {
      const { code, out, err } = capture(args);
      assert.notEqual(code, 0, `expected failure for ${args.join(" ")}`);
      assert.doesNotMatch(out, /CREATE TABLE/);
      assert.ok(err.length > 0);
    }
  });
});
