import {
  DeleteObjectCommand,
  GetObjectCommand,
  HeadObjectCommand,
  PutObjectCommand,
  S3Client,
  type S3ClientConfig,
} from "@aws-sdk/client-s3";
import { getSignedUrl } from "@aws-sdk/s3-request-presigner";
import { StorageInvalidPathError, StorageNotFoundError, type Adapter } from "../adapter.js";

export const DRIVER_S3 = "s3";

export type S3Config = {
  bucket: string;
  region?: string;
  endpoint?: string;
  forcePathStyle?: boolean;
  prefix?: string;
  accessKey?: string;
  secretKey?: string;
  presignTTLMs?: number;
};

export type S3API = {
  send(command: unknown): Promise<unknown>;
};

export type S3Presigner = {
  presignGet(bucket: string, key: string, ttlMs: number): Promise<string>;
};

export type S3FS = Adapter & {
  readonly config: S3Config;
  presignGet(relPath: string, ttlMs?: number): Promise<string>;
};

const DEFAULT_PRESIGN_TTL_MS = 60 * 60 * 1000;

function cleanRel(relPath: string): string {
  if (!relPath) {
    throw new StorageInvalidPathError();
  }
  const p = relPath.replace(/\\/g, "/");
  if (p.startsWith("/")) {
    throw new StorageInvalidPathError();
  }
  for (const part of p.split("/")) {
    if (part === "..") {
      throw new StorageInvalidPathError();
    }
  }
  return p;
}

function objectKey(prefix: string | undefined, relPath: string): string {
  const clean = cleanRel(relPath);
  if (!prefix) {
    return clean;
  }
  return `${prefix.replace(/^\/+|\/+$/g, "")}/${clean}`;
}

function isNotFoundErr(err: unknown): boolean {
  if (!err) return false;
  const s = err instanceof Error ? err.message : String(err);
  const name = err instanceof Error ? err.name : "";
  return (
    name.includes("NotFound") ||
    name.includes("NoSuchKey") ||
    s.includes("NotFound") ||
    s.includes("404") ||
    s.includes("NoSuchKey") ||
    (err as { $metadata?: { httpStatusCode?: number } }).$metadata?.httpStatusCode === 404
  );
}

export async function createS3FS(cfg: S3Config, client?: S3API, presigner?: S3Presigner): Promise<S3FS> {
  if (!cfg.bucket?.trim()) {
    throw new Error("s3: invalid config: bucket required");
  }
  const presignTTLMs = cfg.presignTTLMs && cfg.presignTTLMs > 0 ? cfg.presignTTLMs : DEFAULT_PRESIGN_TTL_MS;

  let api = client;
  let ps = presigner;

  if (!api) {
    const clientCfg: S3ClientConfig = {
      region: cfg.region || "us-east-1",
      forcePathStyle: cfg.forcePathStyle,
    };
    if (cfg.endpoint) {
      clientCfg.endpoint = cfg.endpoint;
    }
    if (cfg.accessKey) {
      clientCfg.credentials = {
        accessKeyId: cfg.accessKey,
        secretAccessKey: cfg.secretKey ?? "",
      };
    }
    const sdk = new S3Client(clientCfg);
    api = sdk;
    if (!ps) {
      ps = {
        async presignGet(bucket, key, ttlMs) {
          return getSignedUrl(
            sdk,
            new GetObjectCommand({ Bucket: bucket, Key: key }),
            { expiresIn: Math.max(1, Math.floor(ttlMs / 1000)) },
          );
        },
      };
    }
  }

  if (!api) {
    throw new Error("s3: invalid config: client required");
  }

  const fs: S3FS = {
    config: { ...cfg, presignTTLMs },
    async put(relPath, data, size) {
      const key = objectKey(cfg.prefix, relPath);
      const buf = Buffer.isBuffer(data) ? data : Buffer.from(data);
      const expected = size !== undefined && size >= 0 ? size : buf.length;
      if (buf.length !== expected) {
        throw new Error("Unexpected end of stream");
      }
      await api!.send(
        new PutObjectCommand({
          Bucket: cfg.bucket,
          Key: key,
          Body: buf,
        }),
      );
    },
    async get(relPath) {
      const key = objectKey(cfg.prefix, relPath);
      try {
        const out = (await api!.send(
          new GetObjectCommand({ Bucket: cfg.bucket, Key: key }),
        )) as {
          Body?: { transformToByteArray(): Promise<Uint8Array> };
          ContentLength?: number;
        };
        if (!out.Body) {
          throw new StorageNotFoundError();
        }
        const bytes = await out.Body.transformToByteArray();
        const body = Buffer.from(bytes);
        return { body, size: out.ContentLength ?? body.length };
      } catch (err) {
        if (isNotFoundErr(err)) {
          throw new StorageNotFoundError();
        }
        throw err;
      }
    },
    async delete(relPath) {
      const key = objectKey(cfg.prefix, relPath);
      await api!.send(new DeleteObjectCommand({ Bucket: cfg.bucket, Key: key }));
    },
    async exists(relPath) {
      const key = objectKey(cfg.prefix, relPath);
      try {
        await api!.send(new HeadObjectCommand({ Bucket: cfg.bucket, Key: key }));
        return true;
      } catch (err) {
        if (isNotFoundErr(err)) {
          return false;
        }
        throw err;
      }
    },
    async presignGet(relPath, ttlMs) {
      if (!ps) {
        throw new Error("s3: cannot presign");
      }
      const ttl = ttlMs && ttlMs > 0 ? ttlMs : presignTTLMs;
      const key = objectKey(cfg.prefix, relPath);
      return ps.presignGet(cfg.bucket, key, ttl);
    },
  };
  return fs;
}
