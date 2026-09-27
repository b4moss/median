import { randomBytes } from "node:crypto";
import { extname } from "node:path";

export const DEFAULT_RANDOM_HEX_LEN = 32;
export const MAX_RANDOM_HEX_LEN = 128;

export type NameMode = "preserve" | "random";

export class FilenameError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "FilenameError";
  }
}

export function sanitizePreserve(original: string): string {
  let name = original.replace(/\0/g, "").replace(/[/\\]/g, "_");
  while (name.includes("..")) {
    name = name.replaceAll("..", ".");
  }
  name = name.replace(/^[\s.]+|[\s.]+$/g, "");
  if (!name || name === "." || name === "..") {
    throw new FilenameError("filename: empty after sanitize");
  }
  if (/[/\\]/.test(name) || name.includes("\0")) {
    throw new FilenameError("filename: empty after sanitize");
  }
  return name;
}

export function randomFilename(original: string, hexLen: number): string {
  if (hexLen <= 0 || hexLen > MAX_RANDOM_HEX_LEN || hexLen % 2 !== 0) {
    throw new FilenameError("filename: invalid random length");
  }
  const base = randomBytes(hexLen / 2).toString("hex");
  let ext = extname(original);
  if (ext === ".") {
    ext = "";
  }
  return base + ext;
}

export function resolveFilename(original: string, mode: NameMode, randomLen = 0): string {
  switch (mode) {
    case "preserve":
    case undefined as unknown as NameMode:
      return sanitizePreserve(original);
    case "random":
      return randomFilename(original, randomLen === 0 ? DEFAULT_RANDOM_HEX_LEN : randomLen);
    default:
      throw new FilenameError("filename: unknown mode");
  }
}
