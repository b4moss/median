import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import Database from "better-sqlite3";
import sharp from "sharp";
import {
  createMedian,
  ErrPDFRendererRequired,
  migrateUp,
  ID_AUTO_INCREMENT,
  thumbnailsFromImage,
  type PDFRenderer,
} from "../src/index.js";
import { baseConfig, makePNG, tempRoot } from "./helpers.js";

function fakeRenderer(err?: Error): PDFRenderer & { calls: number; page: number } {
  const state = { calls: 0, page: -1 };
  return {
    get calls() {
      return state.calls;
    },
    get page() {
      return state.page;
    },
    async renderPage(_pdf, page) {
      state.calls++;
      state.page = page;
      if (err) throw err;
      return sharp({
        create: {
          width: 80,
          height: 60,
          channels: 3,
          background: { r: 0, g: 200, b: 0 },
        },
      })
        .png()
        .toBuffer();
    },
  };
}

describe("PDF renderer + store", () => {
  it("thumbnailsFromImage produces PNG variants; page 0 passed", async () => {
    const img = await sharp({
      create: { width: 100, height: 80, channels: 3, background: { r: 0, g: 0, b: 255 } },
    })
      .png()
      .toBuffer();
    const vars = await thumbnailsFromImage(
      img,
      { presets: { sm: { mode: "longEdge", size: 40, quality: 0.8 } } },
      ["sm"],
    );
    assert.equal(vars.length, 1);
    assert.equal(vars[0]!.key, "sm");
    assert.ok(vars[0]!.bytes.length > 0);
    assert.equal(vars[0]!.mime, "image/png");
    assert.equal(vars[0]!.width, 40);
    await assert.rejects(() =>
      thumbnailsFromImage(img, { presets: { sm: { mode: "longEdge", size: 40 } } }, ["missing"]),
    );
    await assert.rejects(() =>
      thumbnailsFromImage(null, { presets: { sm: { mode: "longEdge", size: 40 } } }, ["sm"]),
    );
  });

  it("store PDF without thumbnail keeps original hash", async () => {
    const root = tempRoot();
    const m = await createMedian(baseConfig(root));
    const pdf = Buffer.from("%PDF-1.4 fake");
    const sum = createHash("sha256").update(pdf).digest("hex");
    const res = await m.store(pdf, pdf.length, { mime: "application/pdf", filename: "a.pdf" });
    assert.equal(res.hash, sum);
    assert.equal(res.variants?.length ?? 0, 0);
  });

  it("store PDF with pdfThumbnail + fake renderer", async () => {
    const root = tempRoot();
    const fake = fakeRenderer();
    const m = await createMedian(
      baseConfig(root, {
        pdfRenderer: fake,
        thumbnails: { presets: { sm: { mode: "longEdge", size: 40, quality: 0.8 } } },
      }),
    );
    const pdf = Buffer.from("%PDF-1.4 content");
    const sum = createHash("sha256").update(pdf).digest("hex");
    const res = await m.store(pdf, pdf.length, {
      mime: "application/pdf",
      filename: "a.pdf",
      pdfThumbnail: true,
      thumbnailKeys: ["sm"],
    });
    assert.equal(res.hash, sum);
    assert.equal(fake.calls, 1);
    assert.equal(fake.page, 0);
    assert.equal(res.variants?.length, 1);
    assert.ok(existsSync(join(root, res.variants![0]!.path)));
    assert.deepEqual(readFileSync(join(root, res.path)), pdf);
  });

  it("requires renderer; compensates on render error; rejects keys without flag", async () => {
    const root = tempRoot();
    const m = await createMedian(
      baseConfig(root, {
        thumbnails: { presets: { sm: { mode: "longEdge", size: 40 } } },
      }),
    );
    await assert.rejects(
      () =>
        m.store(Buffer.from("%PDF"), 4, {
          mime: "application/pdf",
          pdfThumbnail: true,
          thumbnailKeys: ["sm"],
        }),
      () => true,
    );
    assert.equal(ErrPDFRendererRequired.message.includes("PDFRenderer"), true);
    assert.equal(readdirSync(root).length, 0);

    const root2 = tempRoot();
    const m2 = await createMedian(
      baseConfig(root2, {
        pdfRenderer: fakeRenderer(new Error("boom")),
        thumbnails: { presets: { sm: { mode: "longEdge", size: 40 } } },
      }),
    );
    await assert.rejects(() =>
      m2.store(Buffer.from("%PDF"), 4, {
        mime: "application/pdf",
        pdfThumbnail: true,
        thumbnailKeys: ["sm"],
      }),
    );
    assert.equal(readdirSync(root2).length, 0);

    const m3 = await createMedian(baseConfig(tempRoot()));
    await assert.rejects(
      () =>
        m3.store(Buffer.from("%PDF"), 4, {
          mime: "application/pdf",
          thumbnailKeys: ["sm"],
        }),
      /invalid image/,
    );
  });

  it("PDF thumbs with DB create parent/child rows", async () => {
    const root = tempRoot();
    const db = new Database(":memory:");
    migrateUp(db, "sqlite3", ID_AUTO_INCREMENT);
    const m = await createMedian(
      baseConfig(root, {
        db: { db },
        defaultActor: "actor-1",
        pdfRenderer: fakeRenderer(),
        thumbnails: {
          presets: { sm: { mode: "longEdge", size: 32, quality: 0.8 } },
          defaultKeys: ["sm"],
        },
      }),
    );
    const pdf = Buffer.from("%PDF-db-test");
    const res = await m.store(pdf, pdf.length, {
      mime: "application/pdf",
      filename: "doc.pdf",
      pdfThumbnail: true,
    });
    assert.ok(res.id !== undefined);
    assert.equal(res.variants?.length, 1);
    assert.ok(res.variants![0]!.id !== undefined);
    const parent = await m.get("", { id: res.id });
    assert.equal(parent.mime, "application/pdf");
  });
});

describe("regression smoke", () => {
  it("non-image and JPEG/PNG and SVG/PDF do not break each other", async () => {
    const root = tempRoot();
    const fake = fakeRenderer();
    const m = await createMedian(
      baseConfig(root, {
        pdfRenderer: fake,
        thumbnails: {
          presets: { sm: { mode: "longEdge", size: 32, quality: 0.8 } },
          defaultKeys: ["sm"],
        },
      }),
    );
    const plain = await m.store(Buffer.from("hello"), 5, { mime: "text/plain", filename: "a.txt" });
    assert.ok(plain.path);
    await m.delete(plain.path);

    const png = await makePNG(64, 64);
    const img = await m.store(png, png.length, { mime: "image/png", filename: "p.png", compress: true });
    assert.equal(img.variants?.length, 1);

    const svg = Buffer.from(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>`);
    await m.store(svg, svg.length, { mime: "image/svg+xml", filename: "s.svg" });

    await m.store(Buffer.from("%PDF-x"), 6, {
      mime: "application/pdf",
      filename: "d.pdf",
      pdfThumbnail: true,
      thumbnailKeys: ["sm"],
    });

    const again = await m.store(png, png.length, { mime: "image/png", filename: "p2.png" });
    assert.equal(again.variants?.length, 1);
  });
});
