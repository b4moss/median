export class MedianError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "MedianError";
  }
}

export const ErrNegativeSize = new MedianError("median: negative size");
export const ErrSizeExceeded = new MedianError("median: size exceeds MaxSize");
export const ErrEmptyPath = new MedianError("median: path is empty");
export const ErrUnknownKey = new MedianError("median: unknown storage key");
export const ErrActorRequired = new MedianError("median: actor required");
export const ErrIDRequired = new MedianError("median: id required");
export const ErrDuplicate = new MedianError("median: duplicate hash");
export const ErrNotS3 = new MedianError("median: storage is not s3");
export const ErrPDFRendererRequired = new MedianError("median: PDFRenderer required");
export const ErrInvalidBase64 = new MedianError("median: invalid base64");
export const ErrInvalidPart = new MedianError("median: invalid multipart part");

export function isMedianError(err: unknown, message: string): boolean {
  return err instanceof Error && err.message === message;
}
