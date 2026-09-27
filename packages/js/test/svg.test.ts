import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFileSync, readdirSync } from "node:fs";
import { join } from "node:path";
import { createMedian, sanitizeSvg } from "../src/index.js";
import { baseConfig, tempRoot } from "./helpers.js";

describe("SVG sanitize + store", () => {
  it("keeps shapes and style; strips script/on*/javascript:/external", () => {
    const keep = sanitizeSvg(`<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10">
  <style>.a{fill:red}</style>
  <rect class="a" x="0" y="0" width="10" height="10"/>
</svg>`);
    const ks = keep.toString();
    assert.match(ks, /<rect/i);
    assert.match(ks, /<style/i);
    assert.match(ks, /fill:red/);

    const dirty = sanitizeSvg(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
  <script>alert(1)</script>
  <rect onclick="evil()" x="0" y="0" width="1" height="1"/>
  <a href="javascript:alert(1)"><text>x</text></a>
  <image xlink:href="https://evil.example/x.png"/>
  <circle cx="1" cy="1" r="1"/>
</svg>`);
    const s = dirty.toString().toLowerCase();
    assert.ok(!s.includes("<script"));
    assert.ok(!s.includes("alert(1)"));
    assert.ok(!s.includes("onclick"));
    assert.ok(!s.includes("javascript:"));
    assert.ok(!s.includes("https://evil.example"));
    assert.match(s, /<circle/);
  });

  it("rejects broken SVG", () => {
    assert.throws(() => sanitizeSvg("<svg><rect>"));
    assert.throws(() => sanitizeSvg(""));
  });

  it("store sanitizes and hashes clean bytes; no leftovers on invalid", async () => {
    const root = tempRoot();
    const m = await createMedian(baseConfig(root));
    const raw = Buffer.from(
      `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script><rect x="0" y="0" width="5" height="5"/></svg>`,
    );
    const clean = sanitizeSvg(raw);
    const wantHash = createHash("sha256").update(clean).digest("hex");
    const res = await m.store(raw, raw.length, { mime: "image/svg+xml", filename: "x.svg" });
    assert.equal(res.hash, wantHash);
    assert.equal(res.size, clean.length);
    const body = readFileSync(join(root, res.path));
    assert.ok(!body.toString().toLowerCase().includes("<script"));
    assert.deepEqual(body, clean);

    const root2 = tempRoot();
    const m2 = await createMedian(baseConfig(root2));
    await assert.rejects(() =>
      m2.store(Buffer.from("<svg><rect>"), 11, { mime: "image/svg+xml", filename: "bad.svg" }),
    );
    assert.equal(readdirSync(root2).length, 0);
  });

  it("MIME deny rejects before sanitize", async () => {
    const root = tempRoot();
    const m = await createMedian(baseConfig(root, { mimeDeny: ["image/svg+xml"] }));
    const raw = Buffer.from(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>`);
    await assert.rejects(() =>
      m.store(raw, raw.length, { mime: "image/svg+xml", filename: "x.svg" }),
    );
  });
});
