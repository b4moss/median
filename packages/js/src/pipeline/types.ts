export type Fit = "cover" | "contain" | "stretch";
export type ResizeMode = "longEdge" | "shortEdge" | "fixed";
export type ContainBackground = "black" | "white" | "transparent" | string;

export type ResizeConstraint = {
  mode: ResizeMode;
  size?: number;
  width?: number;
  height?: number;
  fit?: Fit;
  containBackground?: ContainBackground;
};

export type ThumbnailPreset = {
  mode: ResizeMode;
  size?: number;
  width?: number;
  height?: number;
  fit?: Fit;
  containBackground?: ContainBackground;
  quality?: number;
};

export type ThumbnailsConfig = {
  presets: Record<string, ThumbnailPreset>;
  defaultKeys?: string[];
};

export type ProcessOptions = {
  mime: string;
  compress?: boolean;
  quality?: number;
  keepOriginal?: boolean;
  resize?: ResizeConstraint;
  thumbnails?: ThumbnailsConfig;
  thumbnailKeys?: string[];
};

export type Variant = {
  key: string;
  mime: string;
  bytes: Buffer;
  width: number;
  height: number;
};

export type ProcessResult = {
  bytes: Buffer;
  mime: string;
  width: number;
  height: number;
  variants: Variant[];
};

export class PipelineError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "PipelineError";
  }
}

export const ErrInvalidImage = new PipelineError("pipeline: invalid image");
export const ErrInvalidFit = new PipelineError("pipeline: invalid fit");
export const ErrInvalidParam = new PipelineError("pipeline: invalid parameter");
export const ErrUnknownPreset = new PipelineError("pipeline: unknown thumbnail key");
export const ErrBadBackground = new PipelineError("pipeline: invalid containBackground");
export const ErrInvalidSVG = new PipelineError("pipeline: invalid svg");
