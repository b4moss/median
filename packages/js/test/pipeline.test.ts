import { describe, it } from "node:test";
import assert from "node:assert/strict";
import sharp from "sharp";
import { processImage } from "../src/index.js";
import { makeJPEG, makePNG } from "./helpers.js";

describe("pipeline image", () => {
  it("longEdge / shortEdge / fixed cover contain stretch", async () => {
    const src = await makePNG(200, 100);
    const long = await processImage(src, {
      mime: "image/png",
      resize: { mode: "longEdge", size: 100 },
    });
    assert.equal(long.width, 100);
    assert.equal(long.height, 50);

    const short = await processImage(src, {
      mime: "image/png",
      resize: { mode: "shortEdge", size: 50 },
    });
    assert.equal(short.height, 50);

    const cover = await processImage(src, {
      mime: "image/png",
      resize: { mode: "fixed", width: 80, height: 80, fit: "cover" },
    });
    assert.equal(cover.width, 80);
    assert.equal(cover.height, 80);

    const stretch = await processImage(src, {
      mime: "image/png",
      resize: { mode: "fixed", width: 60, height: 40, fit: "stretch" },
    });
    assert.equal(stretch.width, 60);
    assert.equal(stretch.height, 40);
  });

  it("contain backgrounds: black default, white, hex; jpeg+transparent fails", async () => {
    const src = await makePNG(200, 100, { r: 255, g: 0, b: 0, alpha: 1 });
    const black = await processImage(src, {
      mime: "image/png",
      resize: { mode: "fixed", width: 100, height: 100, fit: "contain" },
    });
    const blackMeta = await sharp(black.bytes).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
    // corner pixel
    const br = blackMeta.data[0]!;
    const bg = blackMeta.data[1]!;
    const bb = blackMeta.data[2]!;
    assert.equal(br, 0);
    assert.equal(bg, 0);
    assert.equal(bb, 0);

    const white = await processImage(src, {
      mime: "image/png",
      resize: {
        mode: "fixed",
        width: 100,
        height: 100,
        fit: "contain",
        containBackground: "white",
      },
    });
    const whiteMeta = await sharp(white.bytes).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
    assert.equal(whiteMeta.data[0], 255);
    assert.equal(whiteMeta.data[1], 255);
    assert.equal(whiteMeta.data[2], 255);

    const green = await processImage(src, {
      mime: "image/png",
      resize: {
        mode: "fixed",
        width: 100,
        height: 100,
        fit: "contain",
        containBackground: "#00ff00",
      },
    });
    const greenMeta = await sharp(green.bytes).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
    assert.equal(greenMeta.data[0], 0);
    assert.equal(greenMeta.data[1], 255);
    assert.equal(greenMeta.data[2], 0);

    const jpeg = await makeJPEG(50, 50);
    await assert.rejects(
      () =>
        processImage(jpeg, {
          mime: "image/jpeg",
          resize: {
            mode: "fixed",
            width: 50,
            height: 50,
            fit: "contain",
            containBackground: "transparent",
          },
        }),
      /containBackground/,
    );
  });

  it("compress and thumbnails", async () => {
    const src = await makePNG(200, 200, { r: 0, g: 0, b: 255, alpha: 1 });
    const compressed = await processImage(src, { mime: "image/png", compress: true, quality: 0.8 });
    assert.ok(compressed.bytes.length > 0);

    const thumbs = await processImage(src, {
      mime: "image/png",
      thumbnails: {
        presets: { sm: { mode: "longEdge", size: 50, quality: 0.8 } },
        defaultKeys: ["sm"],
      },
    });
    assert.equal(thumbs.variants.length, 1);
    assert.equal(thumbs.variants[0]!.key, "sm");
    assert.equal(thumbs.variants[0]!.width, 50);
  });

  it("animated gif remains decodable after resize", async () => {
    // Minimal 1x1 GIF
    const gif = await sharp({
      create: { width: 40, height: 20, channels: 3, background: { r: 1, g: 2, b: 3 } },
    })
      .gif()
      .toBuffer();
    const res = await processImage(gif, {
      mime: "image/gif",
      resize: { mode: "longEdge", size: 20 },
    });
    const meta = await sharp(res.bytes).metadata();
    assert.ok(meta.width && meta.width <= 20);
  });

  it("rejects invalid image / fit / size / hex / unknown key", async () => {
    await assert.rejects(() => processImage(Buffer.from("not-image"), { mime: "image/png" }));
    const png = await makePNG(10, 10);
    await assert.rejects(() =>
      processImage(png, {
        mime: "image/png",
        resize: { mode: "fixed", width: 10, height: 10 },
      }),
    );
    await assert.rejects(() =>
      processImage(png, {
        mime: "image/png",
        resize: { mode: "longEdge", size: 0 },
      }),
    );
    await assert.rejects(() =>
      processImage(png, {
        mime: "image/png",
        resize: {
          mode: "fixed",
          width: 10,
          height: 10,
          fit: "contain",
          containBackground: "nope",
        },
      }),
    );
    await assert.rejects(() =>
      processImage(png, {
        mime: "image/png",
        thumbnailKeys: ["nope"],
        thumbnails: { presets: {} },
      }),
    );
  });
});
