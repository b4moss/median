import { optimize, type PluginConfig } from "svgo";
import { ErrInvalidSVG, PipelineError } from "./types.js";

const DANGEROUS_URL = /^(?:javascript:|data:text\/html|https?:|\/\/)/i;

function stripDangerousAttrs(value: string): string {
  return value
    .replace(/\s+on[a-z]+\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)/gi, "")
    .replace(
      /\s+(?:xlink:)?href\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))/gi,
      (full, d, s, u) => {
        const v = (d ?? s ?? u ?? "").trim();
        if (DANGEROUS_URL.test(v)) {
          return "";
        }
        return full;
      },
    )
    .replace(/\s+src\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))/gi, (full, d, s, u) => {
      const v = (d ?? s ?? u ?? "").trim();
      if (DANGEROUS_URL.test(v)) {
        return "";
      }
      return full;
    });
}

const removeScriptElements: PluginConfig = {
  name: "medianRemoveScripts",
  fn: () => ({
    element: {
      enter: (node, parentNode) => {
        if (node.name === "script" || node.name === "foreignObject") {
          parentNode.children = parentNode.children.filter((child) => child !== node);
        }
      },
    },
  }),
};

/** Sanitize SVG via svgo + attribute cleanup matching Go median contracts. */
export function sanitizeSvg(src: Uint8Array | Buffer | string): Buffer {
  const input = typeof src === "string" ? src : Buffer.from(src).toString("utf8");
  if (!input.trim()) {
    throw new PipelineError("pipeline: invalid svg");
  }
  let result;
  try {
    result = optimize(input, {
      multipass: false,
      plugins: [
        {
          name: "preset-default",
          params: {
            overrides: {
              removeViewBox: false,
              removeTitle: false,
              removeDesc: false,
              inlineStyles: false,
              minifyStyles: false,
              convertShapeToPath: false,
            },
          },
        },
        removeScriptElements,
        {
          name: "removeAttrs",
          params: {
            attrs: "on*",
          },
        },
      ],
    });
  } catch {
    throw new PipelineError("pipeline: invalid svg");
  }
  if ("error" in result && result.error) {
    throw new PipelineError("pipeline: invalid svg");
  }
  const data = "data" in result ? result.data : "";
  if (!data || !data.trim()) {
    throw new PipelineError("pipeline: invalid svg");
  }
  let cleaned = stripDangerousAttrs(data);
  cleaned = cleaned
    .replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, "")
    .replace(/<script\b[^>]*\/>/gi, "")
    .replace(/<foreignobject\b[^>]*>[\s\S]*?<\/foreignobject>/gi, "");
  if (!cleaned.trim()) {
    throw new PipelineError("pipeline: invalid svg");
  }
  return Buffer.from(cleaned, "utf8");
}

export { ErrInvalidSVG };
