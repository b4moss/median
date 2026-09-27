import { mkdir, open, rename, rm, stat, writeFile } from "node:fs/promises";
import { dirname, isAbsolute, join, relative, resolve, sep } from "node:path";
import { randomBytes } from "node:crypto";
import {
  Adapter,
  ErrInvalidPath,
  ErrNotFound,
  StorageInvalidPathError,
  StorageNotFoundError,
} from "../adapter.js";

export type LocalFS = Adapter & { readonly root: string };

function cleanRel(relPath: string): string {
  if (!relPath) {
    throw new StorageInvalidPathError();
  }
  const p = relPath.replace(/\\/g, "/");
  if (p.startsWith("/") || isAbsolute(relPath)) {
    throw new StorageInvalidPathError();
  }
  if (p.includes("\0")) {
    throw new StorageInvalidPathError();
  }
  for (const part of p.split("/")) {
    if (part === "..") {
      throw new StorageInvalidPathError();
    }
  }
  const cleaned = p
    .split("/")
    .filter((seg) => seg !== "" && seg !== ".")
    .join("/");
  if (!cleaned || cleaned === ".") {
    throw new StorageInvalidPathError();
  }
  if (cleaned.startsWith("../") || cleaned === "..") {
    throw new StorageInvalidPathError();
  }
  return cleaned;
}

export async function createLocalFS(root: string): Promise<LocalFS> {
  if (!root || !root.trim()) {
    throw new Error("local: root path is empty");
  }
  const abs = resolve(root);
  await mkdir(abs, { recursive: true });

  function resolvePath(relPath: string): string {
    const clean = cleanRel(relPath);
    const full = join(abs, ...clean.split("/"));
    const rel = relative(abs, full);
    if (!rel || rel.startsWith("..") || rel.includes(`..${sep}`)) {
      throw new StorageInvalidPathError();
    }
    return full;
  }

  return {
    root: abs,
    async put(relPath, data, size) {
      const full = resolvePath(relPath);
      await mkdir(dirname(full), { recursive: true });
      const buf = Buffer.isBuffer(data) ? data : Buffer.from(data);
      const expected = size !== undefined && size >= 0 ? size : buf.length;
      if (buf.length !== expected) {
        throw new Error("Unexpected end of stream");
      }
      const tmp = join(dirname(full), `.median-${randomBytes(8).toString("hex")}`);
      try {
        await writeFile(tmp, buf);
        await rename(tmp, full);
      } catch (err) {
        await rm(tmp, { force: true }).catch(() => undefined);
        throw err;
      }
    },
    async get(relPath) {
      const full = resolvePath(relPath);
      try {
        const st = await stat(full);
        if (st.isDirectory()) {
          throw new StorageNotFoundError();
        }
        const fh = await open(full, "r");
        try {
          const body = await fh.readFile();
          return { body, size: st.size };
        } finally {
          await fh.close();
        }
      } catch (err) {
        if (err instanceof StorageNotFoundError || err instanceof StorageInvalidPathError) {
          throw err;
        }
        if ((err as NodeJS.ErrnoException).code === "ENOENT") {
          throw new StorageNotFoundError();
        }
        throw err;
      }
    },
    async delete(relPath) {
      const full = resolvePath(relPath);
      try {
        await rm(full);
      } catch (err) {
        if ((err as NodeJS.ErrnoException).code === "ENOENT") {
          throw new StorageNotFoundError();
        }
        throw err;
      }
    },
    async exists(relPath) {
      try {
        const full = resolvePath(relPath);
        const st = await stat(full);
        return st.isFile();
      } catch (err) {
        if (err instanceof StorageInvalidPathError) {
          throw err;
        }
        if ((err as NodeJS.ErrnoException).code === "ENOENT") {
          return false;
        }
        throw err;
      }
    },
  };
}

export { ErrInvalidPath, ErrNotFound };
