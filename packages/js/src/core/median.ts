import { createHash } from "node:crypto";
import { basename, extname } from "node:path";
import Database from "better-sqlite3";
import { checkMIME } from "../internal/mime.js";
import { resolveFilename, type NameMode, DEFAULT_RANDOM_HEX_LEN } from "../internal/filename.js";
import { buildStoragePath } from "../internal/path.js";
import {
  createMediaRepo,
  isNotFound as isDBNotFound,
  type Media,
  type MediaRepo,
} from "../db/index.js";
import {
  processImage,
  sanitizeSvg,
  thumbnailsFromImage,
  type Variant,
} from "../pipeline/index.js";
import { isNotFound as isStorageNotFound, type Adapter } from "../storage/adapter.js";
import { createLocalFS } from "../storage/local/local.js";
import { createS3FS, DRIVER_S3, type S3FS } from "../storage/s3/s3.js";
import {
  ErrActorRequired,
  ErrDuplicate,
  ErrEmptyPath,
  ErrIDRequired,
  ErrNegativeSize,
  ErrNotS3,
  ErrPDFRendererRequired,
  ErrSizeExceeded,
  ErrUnknownKey,
  MedianError,
} from "./errors.js";
import { Semaphore } from "./semaphore.js";
import type {
  Config,
  DeleteOptions,
  GetOptions,
  GetResult,
  Median,
  StoreOptions,
  StoreResult,
  VariantResult,
} from "./types.js";

export const DRIVER_LOCAL = "local";
export const DEFAULT_MAX_SIZE = 20 << 20;
export const DEFAULT_MAX_CONCURRENT_STORES = 20;
export const DEFAULT_PRESIGN_TTL_MS = 60 * 60 * 1000;

function sha256Hex(data: Uint8Array | Buffer): string {
  return createHash("sha256").update(data).digest("hex");
}

function isImageMIME(mime: string): boolean {
  switch (mime) {
    case "image/jpeg":
    case "image/jpg":
    case "image/png":
    case "image/gif":
    case "image/webp":
      return true;
    default:
      return false;
  }
}

function isSVGMIME(mime: string): boolean {
  return mime === "image/svg+xml";
}

function isPDFMIME(mime: string): boolean {
  return mime === "application/pdf";
}

async function readInput(
  data: Uint8Array | Buffer | AsyncIterable<Uint8Array> | ReadableStream<Uint8Array>,
  size: number,
): Promise<Buffer> {
  if (Buffer.isBuffer(data)) {
    if (data.length !== size && size >= 0) {
      // allow exact
    }
    return size >= 0 && size < data.length ? data.subarray(0, size) : data;
  }
  if (data instanceof Uint8Array) {
    const buf = Buffer.from(data);
    return size >= 0 && size < buf.length ? buf.subarray(0, size) : buf;
  }
  if (typeof (data as ReadableStream).getReader === "function") {
    const reader = (data as ReadableStream<Uint8Array>).getReader();
    const chunks: Buffer[] = [];
    let total = 0;
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      chunks.push(Buffer.from(value));
      total += value.byteLength;
      if (size >= 0 && total >= size) break;
    }
    const buf = Buffer.concat(chunks);
    if (size >= 0 && buf.length < size) {
      throw new Error("Unexpected end of stream");
    }
    return size >= 0 ? buf.subarray(0, size) : buf;
  }
  const chunks: Buffer[] = [];
  let total = 0;
  for await (const chunk of data as AsyncIterable<Uint8Array>) {
    chunks.push(Buffer.from(chunk));
    total += chunk.byteLength;
    if (size >= 0 && total >= size) break;
  }
  const buf = Buffer.concat(chunks);
  if (size >= 0 && buf.length < size) {
    throw new Error("Unexpected end of stream");
  }
  return size >= 0 ? buf.subarray(0, size) : buf;
}

function mediaToStoreResult(row: Media, storageKey: string, defaultKey: string): StoreResult {
  return {
    id: row.id,
    path: row.filePath,
    mime: row.mime,
    size: row.size,
    hash: row.hash,
    storageKey: storageKey || defaultKey,
    width: row.width ?? undefined,
    height: row.height ?? undefined,
  };
}

type InternalConfig = Required<
  Pick<
    Config,
    | "defaultKey"
    | "maxSize"
    | "maxConcurrentStores"
    | "filenameMode"
    | "randomHexLen"
    | "presignTTLMs"
  >
> &
  Config & {
    dirLetterCount: number;
    dirNestDepth: number;
  };

export async function createMedian(cfg: Config): Promise<Median> {
  if ((!cfg.storages || Object.keys(cfg.storages).length === 0) && !cfg.adapterOverrides) {
    throw new MedianError("median: storages is empty");
  }
  if (!cfg.defaultKey) {
    throw new MedianError("median: defaultKey is empty");
  }
  if (cfg.storages) {
    if (!(cfg.defaultKey in cfg.storages) && !(cfg.adapterOverrides && cfg.defaultKey in cfg.adapterOverrides)) {
      throw new MedianError(`median: defaultKey "${cfg.defaultKey}" not in storages`);
    }
  } else if (!(cfg.adapterOverrides && cfg.defaultKey in cfg.adapterOverrides)) {
    throw new MedianError(`median: defaultKey "${cfg.defaultKey}" not in storages`);
  }
  if (cfg.dirLetterCount === undefined || cfg.dirNestDepth === undefined) {
    throw new MedianError("median: DirLetterCount and DirNestDepth are required");
  }
  if (cfg.dirLetterCount <= 0 || cfg.dirNestDepth <= 0) {
    throw new MedianError("median: DirLetterCount and DirNestDepth must be positive");
  }
  if (cfg.thumbnails?.defaultKeys) {
    for (const k of cfg.thumbnails.defaultKeys) {
      if (!cfg.thumbnails.presets?.[k]) {
        throw new MedianError(`median: thumbnail defaultKey "${k}" missing from presets`);
      }
    }
  }

  const resolved: InternalConfig = {
    ...cfg,
    defaultKey: cfg.defaultKey,
    maxSize: cfg.maxSize && cfg.maxSize > 0 ? cfg.maxSize : DEFAULT_MAX_SIZE,
    maxConcurrentStores:
      cfg.maxConcurrentStores && cfg.maxConcurrentStores > 0
        ? cfg.maxConcurrentStores
        : DEFAULT_MAX_CONCURRENT_STORES,
    filenameMode: cfg.filenameMode || "random",
    randomHexLen: cfg.randomHexLen && cfg.randomHexLen > 0 ? cfg.randomHexLen : DEFAULT_RANDOM_HEX_LEN,
    presignTTLMs: cfg.presignTTLMs && cfg.presignTTLMs > 0 ? cfg.presignTTLMs : DEFAULT_PRESIGN_TTL_MS,
    dirLetterCount: cfg.dirLetterCount,
    dirNestDepth: cfg.dirNestDepth,
    mimeAllow: cfg.mimeAllow ?? [],
    mimeDeny: cfg.mimeDeny ?? [],
  };

  const adapters: Record<string, Adapter> = {};
  const s3map: Record<string, S3FS> = {};

  for (const [key, sc] of Object.entries(cfg.storages ?? {})) {
    if (cfg.adapterOverrides?.[key]) {
      adapters[key] = cfg.adapterOverrides[key];
      if (cfg.s3Overrides?.[key]) {
        s3map[key] = cfg.s3Overrides[key];
      }
      continue;
    }
    switch (sc.driver) {
      case DRIVER_LOCAL: {
        if (!sc.path?.trim()) {
          throw new MedianError(`median: storage "${key}": local root is empty`);
        }
        adapters[key] = await createLocalFS(sc.path);
        break;
      }
      case DRIVER_S3: {
        if (!sc.bucket) {
          throw new MedianError(`median: storage "${key}": s3 bucket required`);
        }
        const fs = await createS3FS({
          bucket: sc.bucket,
          region: sc.region,
          endpoint: sc.endpoint,
          forcePathStyle: sc.forcePathStyle,
          prefix: sc.prefix,
          accessKey: sc.accessKey,
          secretKey: sc.secretKey,
          presignTTLMs: resolved.presignTTLMs,
        });
        adapters[key] = fs;
        s3map[key] = fs;
        break;
      }
      case "":
        throw new MedianError(`median: storage "${key}": driver is empty`);
      default:
        throw new MedianError(`median: storage "${key}": unknown driver "${sc.driver}"`);
    }
  }
  for (const [key, ov] of Object.entries(cfg.adapterOverrides ?? {})) {
    adapters[key] = ov;
    if (cfg.s3Overrides?.[key]) {
      s3map[key] = cfg.s3Overrides[key];
    }
  }

  let repo: MediaRepo | undefined;
  let dbEnabled = false;
  if (cfg.db) {
    let sqlite = cfg.db.db;
    if (!sqlite) {
      if (!cfg.db.dsn) {
        throw new MedianError("median: DB.db or DB.dsn required");
      }
      sqlite = new Database(cfg.db.dsn);
    }
    repo = createMediaRepo(sqlite, {
      idStrategy: cfg.db.idStrategy,
      tableName: cfg.db.tableName,
      columns: cfg.db.columns,
    });
    dbEnabled = true;
  }

  const slots = new Semaphore(resolved.maxConcurrentStores);

  function resolveKey(key?: string): { key: string; adapter: Adapter } {
    const k = key || resolved.defaultKey;
    const ad = adapters[k];
    if (!ad) {
      throw new MedianError(`median: unknown storage key: ${k}`);
    }
    return { key: k, adapter: ad };
  }

  async function rollbackPaths(ad: Adapter, paths: string[]): Promise<void> {
    for (let i = paths.length - 1; i >= 0; i--) {
      try {
        await ad.delete(paths[i]!);
      } catch (err) {
        if (!isStorageNotFound(err)) {
          // best-effort
        }
      }
    }
  }

  async function store(
    data: Uint8Array | Buffer | AsyncIterable<Uint8Array> | ReadableStream<Uint8Array>,
    size: number,
    opt: StoreOptions,
  ): Promise<StoreResult> {
    if (size < 0) {
      throw ErrNegativeSize;
    }
    if (size > resolved.maxSize) {
      throw ErrSizeExceeded;
    }
    checkMIME(opt.mime, resolved.mimeAllow, resolved.mimeDeny);

    let actor = opt.actor || resolved.defaultActor || "";
    if (dbEnabled && !actor) {
      throw ErrActorRequired;
    }

    const release = await slots.acquire(opt.signal);
    try {
      let bytes = await readInput(data, size);
      if (isSVGMIME(opt.mime)) {
        bytes = sanitizeSvg(bytes);
      }
      const hash = sha256Hex(bytes);

      if (dbEnabled && repo) {
        try {
          const existing = repo.findByHash(hash);
          if (opt.rejectDuplicate) {
            throw ErrDuplicate;
          }
          return mediaToStoreResult(existing, opt.storageKey ?? "", resolved.defaultKey);
        } catch (err) {
          if (!isDBNotFound(err)) {
            if (err === ErrDuplicate) throw err;
            throw err;
          }
        }
      }

      const { key, adapter: ad } = resolveKey(opt.storageKey);

      const runPipe =
        isImageMIME(opt.mime) &&
        (!!opt.compress ||
          !!opt.resize ||
          (opt.thumbnailKeys?.length ?? 0) > 0 ||
          (resolved.thumbnails?.defaultKeys?.length ?? 0) > 0);
      const runPDFThumbs = isPDFMIME(opt.mime) && !!opt.pdfThumbnail;

      let mainData = bytes;
      let width: number | undefined;
      let height: number | undefined;
      let variants: Variant[] = [];

      if (runPipe) {
        const pres = await processImage(bytes, {
          mime: opt.mime,
          compress: opt.compress,
          keepOriginal: opt.keepOriginal,
          quality: opt.quality,
          resize: opt.resize,
          thumbnails: resolved.thumbnails,
          thumbnailKeys: opt.thumbnailKeys,
        });
        mainData = pres.bytes;
        width = pres.width;
        height = pres.height;
        variants = pres.variants;
      } else if (runPDFThumbs) {
        let keys = opt.thumbnailKeys ?? [];
        if (keys.length === 0 && resolved.thumbnails?.defaultKeys) {
          keys = [...resolved.thumbnails.defaultKeys];
        }
        if (keys.length > 0) {
          if (!resolved.pdfRenderer) {
            throw ErrPDFRendererRequired;
          }
          const img = await resolved.pdfRenderer.renderPage(bytes, 0);
          variants = await thumbnailsFromImage(img, resolved.thumbnails, keys);
        }
      } else if (!isImageMIME(opt.mime) && (opt.resize || (opt.thumbnailKeys?.length ?? 0) > 0)) {
        throw new MedianError("pipeline: invalid image");
      }

      const mode: NameMode = opt.filenameMode ?? resolved.filenameMode;
      const randLen = opt.randomHexLen && opt.randomHexLen !== 0 ? opt.randomHexLen : resolved.randomHexLen;
      const baseName = resolveFilename(opt.filename ?? "", mode === "preserve" ? "preserve" : "random", randLen);
      const relPath = buildStoragePath(baseName, resolved.dirLetterCount, resolved.dirNestDepth);

      const written: string[] = [relPath];
      try {
        await ad.put(relPath, mainData, mainData.length);
      } catch (err) {
        await rollbackPaths(ad, written);
        throw err;
      }
      const mainHashHex = sha256Hex(mainData);

      const varResults: VariantResult[] = [];
      for (const v of variants) {
        const ext = extname(baseName);
        let vName = resolveFilename(v.key + ext, "random", randLen);
        vName = `${v.key}-${vName}`;
        const vPath = buildStoragePath(vName, resolved.dirLetterCount, resolved.dirNestDepth);
        try {
          await ad.put(vPath, v.bytes, v.bytes.length);
          written.push(vPath);
        } catch (err) {
          await rollbackPaths(ad, written);
          throw err;
        }
        varResults.push({
          key: v.key,
          path: vPath,
          mime: v.mime,
          size: v.bytes.length,
          hash: sha256Hex(v.bytes),
          width: v.width,
          height: v.height,
          storageKey: key,
        });
      }

      const out: StoreResult = {
        path: relPath,
        mime: opt.mime,
        size: mainData.length,
        hash: mainHashHex,
        storageKey: key,
        width,
        height,
        variants: varResults.length ? varResults : undefined,
      };

      if (dbEnabled && repo) {
        try {
          const created = repo.create({
            filePath: relPath,
            fileName: baseName,
            originalFileName: opt.filename ?? "",
            mime: opt.mime,
            size: out.size,
            hash,
            width: width ?? null,
            height: height ?? null,
            createdBy: actor,
            ownedBy: actor,
          });
          out.id = created.id;
          for (let i = 0; i < varResults.length; i++) {
            const vr = varResults[i]!;
            const child = repo.create({
              filePath: vr.path,
              fileName: basename(vr.path),
              originalFileName: "",
              mime: vr.mime,
              size: vr.size,
              hash: vr.hash,
              width: vr.width,
              height: vr.height,
              originalId: created.id,
              variantKey: vr.key,
              createdBy: actor,
              ownedBy: actor,
            });
            vr.id = child.id;
          }
          out.variants = varResults.length ? varResults : undefined;
        } catch (err) {
          if (out.id !== undefined) {
            try {
              repo.deleteById(out.id);
            } catch {
              /* ignore */
            }
          }
          await rollbackPaths(ad, written);
          throw err;
        }
      }
      return out;
    } finally {
      release();
    }
  }

  async function del(objectPath: string, opt: DeleteOptions = {}): Promise<void> {
    if (dbEnabled && repo) {
      if (opt.id === undefined || opt.id === null || opt.id === "") {
        throw ErrIDRequired;
      }
      const parent = repo.findById(opt.id);
      const children = repo.listChildren(opt.id);
      const { adapter: ad } = resolveKey(opt.storageKey);
      let firstErr: unknown;
      for (const ch of children) {
        try {
          await ad.delete(ch.filePath);
        } catch (err) {
          if (!isStorageNotFound(err) && !firstErr) firstErr = err;
        }
      }
      try {
        await ad.delete(parent.filePath);
      } catch (err) {
        if (!isStorageNotFound(err) && !firstErr) firstErr = err;
      }
      repo.deleteById(opt.id);
      if (firstErr) throw firstErr;
      return;
    }
    if (!objectPath) {
      throw ErrEmptyPath;
    }
    const { adapter: ad } = resolveKey(opt.storageKey);
    await ad.delete(objectPath);
  }

  async function get(objectPath: string, opt: GetOptions = {}): Promise<GetResult> {
    if (dbEnabled && repo) {
      if (opt.id === undefined || opt.id === null || opt.id === "") {
        throw ErrIDRequired;
      }
      const row = repo.findById(opt.id);
      const { key, adapter: ad } = resolveKey(opt.storageKey);
      const out: GetResult = {
        id: row.id,
        path: row.filePath,
        mime: row.mime,
        size: row.size,
        hash: row.hash,
        storageKey: key,
        width: row.width,
        height: row.height,
      };
      if (opt.withBody) {
        const { body } = await ad.get(row.filePath);
        out.body = body;
      }
      return out;
    }
    if (!objectPath) {
      throw ErrEmptyPath;
    }
    const { key, adapter: ad } = resolveKey(opt.storageKey);
    const { body, size } = await ad.get(objectPath);
    const out: GetResult = { path: objectPath, size, storageKey: key };
    if (opt.withBody) {
      out.body = body;
      out.hash = sha256Hex(body);
    }
    return out;
  }

  async function presignGet(objectPath: string, storageKey?: string, ttlMs?: number): Promise<string> {
    const key = storageKey || resolved.defaultKey;
    const fs = s3map[key];
    if (!fs) {
      throw ErrNotS3;
    }
    const ttl = ttlMs && ttlMs > 0 ? ttlMs : resolved.presignTTLMs;
    return fs.presignGet(objectPath, ttl);
  }

  return {
    store,
    delete: del,
    get,
    presignGet,
    config: resolved,
  };
}
