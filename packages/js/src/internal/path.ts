import { shardian } from "@b4moss/shardian";

export class PathError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "PathError";
  }
}

/** Build shardian storage path (no leading slash). */
export function buildStoragePath(fileName: string, dirLetterCount: number, dirNestDepth: number): string {
  if (!fileName) {
    throw new PathError("path: empty filename");
  }
  try {
    return shardian(fileName, {
      dirLetterCount,
      dirNestDepth,
      stripHeadSlash: true,
    });
  } catch (err) {
    throw new PathError(err instanceof Error ? err.message : String(err));
  }
}
