export type ColumnMap = {
  id?: string;
  filePath?: string;
  fileName?: string;
  originalFileName?: string;
  mime?: string;
  size?: string;
  width?: string;
  height?: string;
  hash?: string;
  originalId?: string;
  variantKey?: string;
  createdBy?: string;
  ownedBy?: string;
  status?: string;
  createdAt?: string;
};

export type ResolvedColumnMap = {
  id: string;
  filePath: string;
  fileName: string;
  originalFileName: string;
  mime: string;
  size: string;
  width: string;
  height: string;
  hash: string;
  originalId: string;
  variantKey: string;
  createdBy: string;
  ownedBy: string;
  status: string;
  createdAt: string;
};

export function defaultColumnMap(): ResolvedColumnMap {
  return {
    id: "id",
    filePath: "file_path",
    fileName: "file_name",
    originalFileName: "original_file_name",
    mime: "mime",
    size: "size",
    width: "width",
    height: "height",
    hash: "hash",
    originalId: "original_id",
    variantKey: "variant_key",
    createdBy: "created_by",
    ownedBy: "owned_by",
    status: "status",
    createdAt: "created_at",
  };
}

export function withColumnDefaults(c?: ColumnMap): ResolvedColumnMap {
  const d = defaultColumnMap();
  if (!c) return d;
  return {
    id: c.id || d.id,
    filePath: c.filePath || d.filePath,
    fileName: c.fileName || d.fileName,
    originalFileName: c.originalFileName || d.originalFileName,
    mime: c.mime || d.mime,
    size: c.size || d.size,
    width: c.width || d.width,
    height: c.height || d.height,
    hash: c.hash || d.hash,
    originalId: c.originalId || d.originalId,
    variantKey: c.variantKey || d.variantKey,
    createdBy: c.createdBy || d.createdBy,
    ownedBy: c.ownedBy || d.ownedBy,
    status: c.status || d.status,
    createdAt: c.createdAt || d.createdAt,
  };
}
