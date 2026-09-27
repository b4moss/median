import sharp, { type Sharp } from "sharp";
import {
  ErrBadBackground,
  ErrInvalidFit,
  ErrInvalidImage,
  ErrInvalidParam,
  ErrUnknownPreset,
  PipelineError,
  type ContainBackground,
  type ProcessOptions,
  type ProcessResult,
  type ResizeConstraint,
  type ThumbnailPreset,
  type ThumbnailsConfig,
  type Variant,
} from "./types.js";

function parseBackground(
  bg: ContainBackground | undefined,
  mime: string,
): { r: number; g: number; b: number; alpha: number } {
  const s = (bg ?? "black").toString().trim().toLowerCase();
  if (s === "black" || s === "") {
    return { r: 0, g: 0, b: 0, alpha: 1 };
  }
  if (s === "white") {
    return { r: 255, g: 255, b: 255, alpha: 1 };
  }
  if (s === "transparent") {
    if (mime.includes("jpeg") || mime.includes("jpg")) {
      throw new PipelineError("pipeline: invalid containBackground");
    }
    return { r: 0, g: 0, b: 0, alpha: 0 };
  }
  if (!s.startsWith("#")) {
    throw new PipelineError("pipeline: invalid containBackground");
  }
  const hex = s.slice(1);
  if (hex.length === 6) {
    const n = Number.parseInt(hex, 16);
    if (Number.isNaN(n)) {
      throw new PipelineError("pipeline: invalid containBackground");
    }
    return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255, alpha: 1 };
  }
  if (hex.length === 8) {
    const n = Number.parseInt(hex, 16);
    if (Number.isNaN(n)) {
      throw new PipelineError("pipeline: invalid containBackground");
    }
    return {
      r: (n >> 24) & 255,
      g: (n >> 16) & 255,
      b: (n >> 8) & 255,
      alpha: ((n & 255) / 255),
    };
  }
  throw new PipelineError("pipeline: invalid containBackground");
}

async function loadImage(src: Buffer, mime: string): Promise<{ img: Sharp; width: number; height: number }> {
  if (!src.length) {
    throw new PipelineError("pipeline: invalid image");
  }
  try {
    const img = sharp(src, { animated: mime.includes("gif"), failOn: "error" });
    const meta = await img.metadata();
    if (!meta.width || !meta.height) {
      throw new PipelineError("pipeline: invalid image");
    }
    return { img, width: meta.width, height: meta.height };
  } catch (err) {
    if (err instanceof PipelineError) throw err;
    throw new PipelineError("pipeline: invalid image");
  }
}

async function applyResize(
  src: Buffer,
  rc: ResizeConstraint,
  mime: string,
): Promise<{ buffer: Buffer; width: number; height: number }> {
  const { width: w, height: h } = await loadImage(src, mime);
  let pipeline = sharp(src, { animated: mime.includes("gif"), failOn: "error" });

  switch (rc.mode) {
    case "longEdge": {
      if (!rc.size || rc.size <= 0) {
        throw new PipelineError("pipeline: invalid parameter");
      }
      if (w >= h) {
        pipeline = pipeline.resize({ width: rc.size, withoutEnlargement: false });
      } else {
        pipeline = pipeline.resize({ height: rc.size, withoutEnlargement: false });
      }
      break;
    }
    case "shortEdge": {
      if (!rc.size || rc.size <= 0) {
        throw new PipelineError("pipeline: invalid parameter");
      }
      if (w <= h) {
        pipeline = pipeline.resize({ width: rc.size, withoutEnlargement: false });
      } else {
        pipeline = pipeline.resize({ height: rc.size, withoutEnlargement: false });
      }
      break;
    }
    case "fixed": {
      if (!rc.width || !rc.height || rc.width <= 0 || rc.height <= 0) {
        throw new PipelineError("pipeline: invalid parameter");
      }
      switch (rc.fit) {
        case "cover":
          pipeline = pipeline.resize({
            width: rc.width,
            height: rc.height,
            fit: "cover",
            position: "centre",
          });
          break;
        case "stretch":
          pipeline = pipeline.resize({
            width: rc.width,
            height: rc.height,
            fit: "fill",
          });
          break;
        case "contain": {
          const bg = parseBackground(rc.containBackground, mime);
          pipeline = pipeline.resize({
            width: rc.width,
            height: rc.height,
            fit: "contain",
            background: bg,
          });
          break;
        }
        default:
          throw new PipelineError("pipeline: invalid fit");
      }
      break;
    }
    default:
      throw new PipelineError("pipeline: invalid parameter");
  }

  const buffer = await encodeSharp(pipeline, mime, 0.85);
  const meta = await sharp(buffer).metadata();
  return { buffer, width: meta.width ?? 0, height: meta.height ?? 0 };
}

async function encodeSharp(img: Sharp, mime: string, quality: number): Promise<Buffer> {
  let q = Math.round(quality * 100);
  if (q < 1) q = 1;
  if (q > 100) q = 100;
  const m = mime.toLowerCase();
  if (m.includes("jpeg") || m.includes("jpg")) {
    return img.jpeg({ quality: q, mozjpeg: true }).toBuffer();
  }
  if (m.includes("webp")) {
    return img.webp({ quality: q }).toBuffer();
  }
  if (m.includes("gif")) {
    return img.gif().toBuffer();
  }
  return img.png().toBuffer();
}

async function encodeBuffer(src: Buffer, mime: string, quality: number): Promise<Buffer> {
  return encodeSharp(sharp(src, { animated: mime.includes("gif"), failOn: "error" }), mime, quality);
}

export async function processImage(src: Uint8Array | Buffer, opt: ProcessOptions): Promise<ProcessResult> {
  const data = Buffer.isBuffer(src) ? src : Buffer.from(src);
  if (!data.length) {
    throw new PipelineError("pipeline: invalid image");
  }
  const mime = (opt.mime ?? "").toLowerCase().trim();
  const { width: origW, height: origH } = await loadImage(data, mime);

  let quality = opt.quality ?? 0;
  if (quality === 0) quality = 0.85;
  if (quality < 0 || quality > 1) {
    throw new PipelineError("pipeline: invalid parameter");
  }

  let workBuf = data;
  let workW = origW;
  let workH = origH;
  let transformed = false;

  if (opt.resize) {
    const resized = await applyResize(data, opt.resize, mime);
    workBuf = resized.buffer;
    workW = resized.width;
    workH = resized.height;
    transformed = true;
  }

  let mainBytes: Buffer;
  if (opt.compress && opt.keepOriginal) {
    mainBytes = data;
    workBuf = data;
    workW = origW;
    workH = origH;
  } else if (opt.compress) {
    mainBytes = await encodeBuffer(workBuf, mime, quality);
    const meta = await sharp(mainBytes).metadata();
    workW = meta.width ?? workW;
    workH = meta.height ?? workH;
  } else if (transformed) {
    mainBytes = workBuf;
  } else {
    mainBytes = data;
  }

  const res: ProcessResult = {
    bytes: mainBytes,
    mime,
    width: workW,
    height: workH,
    variants: [],
  };

  let keys = opt.thumbnailKeys ?? [];
  if (keys.length === 0 && opt.thumbnails?.defaultKeys?.length) {
    keys = [...opt.thumbnails.defaultKeys];
  }
  if (keys.length === 0) {
    return res;
  }
  if (!opt.thumbnails?.presets || Object.keys(opt.thumbnails.presets).length === 0) {
    throw new PipelineError("pipeline: unknown thumbnail key");
  }

  for (const key of keys) {
    const preset = opt.thumbnails.presets[key];
    if (!preset) {
      throw new PipelineError(`pipeline: unknown thumbnail key: ${key}`);
    }
    const rc: ResizeConstraint = {
      mode: preset.mode,
      size: preset.size,
      width: preset.width,
      height: preset.height,
      fit: preset.fit,
      containBackground: preset.containBackground,
    };
    const timg = await applyResize(data, rc, mime);
    const q = preset.quality && preset.quality !== 0 ? preset.quality : 0.8;
    const tb = await encodeBuffer(timg.buffer, mime, q);
    const meta = await sharp(tb).metadata();
    res.variants.push({
      key,
      mime,
      bytes: tb,
      width: meta.width ?? timg.width,
      height: meta.height ?? timg.height,
    });
  }
  return res;
}

export async function thumbnailsFromImage(
  image: Buffer | Uint8Array | null | undefined,
  thumbs: ThumbnailsConfig | undefined,
  keys: string[],
): Promise<Variant[]> {
  if (!image) {
    throw new PipelineError("pipeline: invalid image");
  }
  const data = Buffer.isBuffer(image) ? image : Buffer.from(image);
  let useKeys = keys;
  if (useKeys.length === 0 && thumbs?.defaultKeys?.length) {
    useKeys = [...thumbs.defaultKeys];
  }
  if (useKeys.length === 0) {
    return [];
  }
  if (!thumbs?.presets || Object.keys(thumbs.presets).length === 0) {
    throw new PipelineError("pipeline: unknown thumbnail key");
  }
  const out: Variant[] = [];
  for (const key of useKeys) {
    const preset: ThumbnailPreset | undefined = thumbs.presets[key];
    if (!preset) {
      throw new PipelineError(`pipeline: unknown thumbnail key: ${key}`);
    }
    const rc: ResizeConstraint = {
      mode: preset.mode,
      size: preset.size,
      width: preset.width,
      height: preset.height,
      fit: preset.fit,
      containBackground: preset.containBackground,
    };
    const timg = await applyResize(data, rc, "image/png");
    const png = await sharp(timg.buffer).png().toBuffer();
    const meta = await sharp(png).metadata();
    out.push({
      key,
      mime: "image/png",
      bytes: png,
      width: meta.width ?? timg.width,
      height: meta.height ?? timg.height,
    });
  }
  return out;
}

export {
  ErrInvalidImage,
  ErrInvalidFit,
  ErrInvalidParam,
  ErrUnknownPreset,
  ErrBadBackground,
};
