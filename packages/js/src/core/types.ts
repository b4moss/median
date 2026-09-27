import type Database from "better-sqlite3";
import type { Adapter } from "../storage/adapter.js";
import type { S3FS } from "../storage/s3/s3.js";
import type { NameMode } from "../internal/filename.js";
import type { PDFRenderer, ResizeConstraint, ThumbnailsConfig } from "../pipeline/index.js";

export type StorageConfig = {
  driver: string;
  path?: string;
  bucket?: string;
  region?: string;
  endpoint?: string;
  forcePathStyle?: boolean;
  prefix?: string;
  accessKey?: string;
  secretKey?: string;
};

export type DBConfig = {
  /** Existing better-sqlite3 Database instance */
  db?: Database.Database;
  /** SQLite DSN / file path (used when db not provided) */
  dsn?: string;
  idStrategy?: string;
  /** Default media */
  tableName?: string;
  columns?: import("../db/columns.js").ColumnMap;
};

export type Config = {
  storages?: Record<string, StorageConfig>;
  defaultKey: string;
  mimeAllow?: string[];
  mimeDeny?: string[];
  maxSize?: number;
  maxConcurrentStores?: number;
  filenameMode?: NameMode;
  randomHexLen?: number;
  dirLetterCount?: number;
  dirNestDepth?: number;
  db?: DBConfig;
  defaultActor?: string;
  thumbnails?: ThumbnailsConfig;
  /** Presign GET TTL in milliseconds (default 1h) */
  presignTTLMs?: number;
  pdfRenderer?: PDFRenderer;
  adapterOverrides?: Record<string, Adapter>;
  s3Overrides?: Record<string, S3FS>;
};

export type VariantResult = {
  key: string;
  path: string;
  mime: string;
  size: number;
  hash: string;
  width: number;
  height: number;
  id?: string | number;
  storageKey: string;
};

export type StoreResult = {
  id?: string | number;
  path: string;
  mime: string;
  size: number;
  hash: string;
  storageKey: string;
  width?: number;
  height?: number;
  variants?: VariantResult[];
};

export type StoreOptions = {
  mime: string;
  filename?: string;
  storageKey?: string;
  filenameMode?: NameMode;
  randomHexLen?: number;
  actor?: string;
  rejectDuplicate?: boolean;
  compress?: boolean;
  keepOriginal?: boolean;
  quality?: number;
  resize?: ResizeConstraint;
  thumbnailKeys?: string[];
  pdfThumbnail?: boolean;
  signal?: AbortSignal;
};

export type GetOptions = {
  storageKey?: string;
  withBody?: boolean;
  id?: string | number;
};

export type GetResult = {
  id?: string | number;
  path: string;
  mime?: string;
  size: number;
  hash?: string;
  storageKey: string;
  width?: number | null;
  height?: number | null;
  body?: Buffer;
};

export type DeleteOptions = {
  storageKey?: string;
  id?: string | number;
};

export type Median = {
  store(
    data: Uint8Array | Buffer | AsyncIterable<Uint8Array> | ReadableStream<Uint8Array>,
    size: number,
    opt: StoreOptions,
  ): Promise<StoreResult>;
  delete(objectPath: string, opt?: DeleteOptions): Promise<void>;
  get(objectPath: string, opt?: GetOptions): Promise<GetResult>;
  presignGet(objectPath: string, storageKey?: string, ttlMs?: number): Promise<string>;
  readonly config: Readonly<Config>;
};
