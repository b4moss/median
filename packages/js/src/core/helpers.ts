import { ErrInvalidBase64, ErrInvalidPart } from "./errors.js";
import type { StoreOptions } from "./types.js";

export function decodeBase64(s: string): { data: Buffer; mime?: string } {
  const trimmed = s.trim();
  if (!trimmed) {
    throw ErrInvalidBase64;
  }
  let payload = trimmed;
  let mime: string | undefined;
  if (trimmed.startsWith("data:")) {
    const comma = trimmed.indexOf(",");
    if (comma < 0) {
      throw ErrInvalidBase64;
    }
    const meta = trimmed.slice("data:".length, comma);
    payload = trimmed.slice(comma + 1);
    if (!meta.includes(";base64")) {
      throw ErrInvalidBase64;
    }
    let media = meta.replace(/;base64$/, "").replace(/;$/, "");
    if (media) {
      mime = media;
    }
  }
  try {
    return { data: Buffer.from(payload, "base64"), mime };
  } catch {
    throw ErrInvalidBase64;
  }
}

export type MultipartPart = {
  filename?: string;
  mime: string;
  body: Uint8Array | Buffer;
  size?: number;
};

export function partToStoreInput(part: MultipartPart): {
  data: Buffer;
  size: number;
  options: StoreOptions;
} {
  if (!part.body) {
    throw ErrInvalidPart;
  }
  const size = part.size ?? (Buffer.isBuffer(part.body) ? part.body.length : part.body.byteLength);
  if (size < 0) {
    throw ErrInvalidPart;
  }
  let mimeType = (part.mime ?? "").trim();
  if (!mimeType) {
    throw ErrInvalidPart;
  }
  const semi = mimeType.indexOf(";");
  if (semi >= 0) {
    mimeType = mimeType.slice(0, semi).trim();
  }
  return {
    data: Buffer.isBuffer(part.body) ? part.body : Buffer.from(part.body),
    size,
    options: {
      mime: mimeType,
      filename: part.filename,
    },
  };
}
