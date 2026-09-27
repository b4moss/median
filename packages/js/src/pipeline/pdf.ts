import type { ThumbnailsConfig, Variant } from "./types.js";
import { thumbnailsFromImage } from "./process.js";

/** Rasterizes a PDF page. Page 0 is the first page. Returns image bytes (PNG/JPEG) for the pipeline. */
export type PDFRenderer = {
  renderPage(pdf: Uint8Array | Buffer, page: number): Promise<Buffer | Uint8Array>;
};

export async function thumbnailsFromPdfPage(
  renderer: PDFRenderer,
  pdf: Uint8Array | Buffer,
  thumbs: ThumbnailsConfig | undefined,
  keys: string[],
): Promise<Variant[]> {
  const img = await renderer.renderPage(pdf, 0);
  return thumbnailsFromImage(img, thumbs, keys);
}

export { thumbnailsFromImage };
