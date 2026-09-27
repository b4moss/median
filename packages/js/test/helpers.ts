import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import sharp from "sharp";
import type { Config } from "../src/index.js";

export function tempRoot(prefix = "median-"): string {
  return mkdtempSync(join(tmpdir(), prefix));
}

export function baseConfig(root: string, overrides: Partial<Config> = {}): Config {
  return {
    storages: {
      local: { driver: "local", path: root },
    },
    defaultKey: "local",
    dirLetterCount: 1,
    dirNestDepth: 2,
    ...overrides,
  };
}

export async function makePNG(w: number, h: number, color = { r: 255, g: 0, b: 0, alpha: 1 }): Promise<Buffer> {
  return sharp({
    create: {
      width: w,
      height: h,
      channels: 4,
      background: color,
    },
  })
    .png()
    .toBuffer();
}

export async function makeJPEG(w: number, h: number): Promise<Buffer> {
  return sharp({
    create: {
      width: w,
      height: h,
      channels: 3,
      background: { r: 0, g: 0, b: 255 },
    },
  })
    .jpeg()
    .toBuffer();
}
