package pipeline

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
)

// PDFRenderer rasterizes a PDF page. Page 0 is the first page.
type PDFRenderer interface {
	RenderPage(pdf []byte, page int) (image.Image, error)
}

// ThumbnailsFromImage applies thumbnail presets to an already-decoded image.
// Variants are encoded as image/png.
func ThumbnailsFromImage(img image.Image, thumbs *ThumbnailsConfig, keys []string) ([]Variant, error) {
	if img == nil {
		return nil, ErrInvalidImage
	}
	if len(keys) == 0 {
		if thumbs != nil {
			keys = append([]string{}, thumbs.DefaultKeys...)
		}
	}
	if len(keys) == 0 {
		return nil, nil
	}
	if thumbs == nil || len(thumbs.Presets) == 0 {
		return nil, ErrUnknownPreset
	}
	var out []Variant
	for _, key := range keys {
		preset, ok := thumbs.Presets[key]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownPreset, key)
		}
		rc := ResizeConstraint{
			Mode: preset.Mode, Size: preset.Size, Width: preset.Width, Height: preset.Height,
			Fit: preset.Fit, ContainBackground: preset.ContainBackground,
		}
		timg, err := applyResize(img, rc, "image/png")
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, timg); err != nil {
			return nil, err
		}
		b := timg.Bounds()
		out = append(out, Variant{
			Key: key, MIME: "image/png", Bytes: buf.Bytes(),
			Width: b.Dx(), Height: b.Dy(),
		})
	}
	return out, nil
}
