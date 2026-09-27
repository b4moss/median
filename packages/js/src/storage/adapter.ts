export class StorageNotFoundError extends Error {
  constructor(message = "storage: not found") {
    super(message);
    this.name = "StorageNotFoundError";
  }
}

export class StorageInvalidPathError extends Error {
  constructor(message = "storage: invalid path") {
    super(message);
    this.name = "StorageInvalidPathError";
  }
}

export const ErrNotFound = new StorageNotFoundError();
export const ErrInvalidPath = new StorageInvalidPathError();

export function isNotFound(err: unknown): boolean {
  return err instanceof StorageNotFoundError || (err instanceof Error && err.message.includes("storage: not found"));
}

export function isInvalidPath(err: unknown): boolean {
  return err instanceof StorageInvalidPathError || (err instanceof Error && err.message.includes("storage: invalid path"));
}

export type Adapter = {
  put(relPath: string, data: Uint8Array | Buffer, size?: number): Promise<void>;
  get(relPath: string): Promise<{ body: Buffer; size: number }>;
  delete(relPath: string): Promise<void>;
  exists(relPath: string): Promise<boolean>;
};
