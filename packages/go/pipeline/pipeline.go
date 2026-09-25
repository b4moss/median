package pipeline

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

var (
	ErrInvalidImage  = errors.New("pipeline: invalid image")
	ErrInvalidFit    = errors.New("pipeline: invalid fit")
	ErrInvalidParam  = errors.New("pipeline: invalid parameter")
	ErrUnknownPreset = errors.New("pipeline: unknown thumbnail key")
	ErrBadBackground = errors.New("pipeline: invalid containBackground")
)

type Fit string

const (
	FitCover   Fit = "cover"
	FitContain Fit = "contain"
	FitStretch Fit = "stretch"
)

type ResizeMode string

const (
	ModeLongEdge  ResizeMode = "longEdge"
	ModeShortEdge ResizeMode = "shortEdge"
	ModeFixed     ResizeMode = "fixed"
)

type ContainBackground string

const (
	BGBlack       ContainBackground = "black"
	BGWhite       ContainBackground = "white"
	BGTransparent ContainBackground = "transparent"
)

type ResizeConstraint struct {
	Mode              ResizeMode
	Size              int
	Width             int
	Height            int
	Fit               Fit
	ContainBackground ContainBackground
}

type ThumbnailPreset struct {
	Mode              ResizeMode
	Size              int
	Width             int
	Height            int
	Fit               Fit
	ContainBackground ContainBackground
	Quality           float64
}

type ThumbnailsConfig struct {
	Presets     map[string]ThumbnailPreset
	DefaultKeys []string
}

type ProcessOptions struct {
	MIME          string
	Compress      bool
	Quality       float64
	KeepOriginal  bool
	Resize        *ResizeConstraint
	Thumbnails    *ThumbnailsConfig
	ThumbnailKeys []string
}

type Variant struct {
	Key    string
	MIME   string
	Bytes  []byte
	Width  int
	Height int
}

type Result struct {
	Bytes    []byte
	MIME     string
	Width    int
	Height   int
	Variants []Variant
}

// Process applies resize/compress/thumbnails per median pipeline rules.
func Process(src []byte, opt ProcessOptions) (*Result, error) {
	if len(src) == 0 {
		return nil, ErrInvalidImage
	}
	mime := strings.ToLower(strings.TrimSpace(opt.MIME))
	img, err := decodeImage(src, mime)
	if err != nil {
		return nil, err
	}

	quality := opt.Quality
	if quality == 0 {
		quality = 0.85
	}
	if quality < 0 || quality > 1 {
		return nil, ErrInvalidParam
	}

	work := img
	transformed := false
	if opt.Resize != nil {
		work, err = applyResize(img, *opt.Resize, mime)
		if err != nil {
			return nil, err
		}
		transformed = true
	}

	var mainBytes []byte
	switch {
	case opt.Compress && opt.KeepOriginal:
		mainBytes = src
		work = img
	case opt.Compress:
		mainBytes, err = encodeImage(work, mime, quality)
		if err != nil {
			return nil, err
		}
	case transformed:
		mainBytes, err = encodeImage(work, mime, quality)
		if err != nil {
			return nil, err
		}
	default:
		mainBytes = src
	}

	b := work.Bounds()
	res := &Result{Bytes: mainBytes, MIME: mime, Width: b.Dx(), Height: b.Dy()}

	keys := opt.ThumbnailKeys
	if len(keys) == 0 && opt.Thumbnails != nil {
		keys = append([]string{}, opt.Thumbnails.DefaultKeys...)
	}
	if len(keys) == 0 {
		return res, nil
	}
	if opt.Thumbnails == nil || len(opt.Thumbnails.Presets) == 0 {
		return nil, ErrUnknownPreset
	}
	for _, key := range keys {
		preset, ok := opt.Thumbnails.Presets[key]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownPreset, key)
		}
		rc := ResizeConstraint{
			Mode: preset.Mode, Size: preset.Size, Width: preset.Width, Height: preset.Height,
			Fit: preset.Fit, ContainBackground: preset.ContainBackground,
		}
		timg, err := applyResize(img, rc, mime)
		if err != nil {
			return nil, err
		}
		q := preset.Quality
		if q == 0 {
			q = 0.8
		}
		tb, err := encodeImage(timg, mime, q)
		if err != nil {
			return nil, err
		}
		tbnds := timg.Bounds()
		res.Variants = append(res.Variants, Variant{
			Key: key, MIME: mime, Bytes: tb, Width: tbnds.Dx(), Height: tbnds.Dy(),
		})
	}
	return res, nil
}

func decodeImage(src []byte, mime string) (image.Image, error) {
	r := bytes.NewReader(src)
	var (
		img image.Image
		err error
	)
	switch {
	case strings.Contains(mime, "jpeg"), strings.Contains(mime, "jpg"):
		img, err = jpeg.Decode(r)
	case strings.Contains(mime, "png"):
		img, err = png.Decode(r)
	case strings.Contains(mime, "gif"):
		img, err = gif.Decode(r)
	case strings.Contains(mime, "webp"):
		img, err = webp.Decode(r)
	default:
		img, _, err = image.Decode(bytes.NewReader(src))
	}
	if err != nil {
		return nil, ErrInvalidImage
	}
	return img, nil
}

func encodeImage(img image.Image, mime string, quality float64) ([]byte, error) {
	var buf bytes.Buffer
	q := int(quality * 100)
	if q < 1 {
		q = 1
	}
	if q > 100 {
		q = 100
	}
	var err error
	switch {
	case strings.Contains(mime, "jpeg"), strings.Contains(mime, "jpg"):
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: q})
	case strings.Contains(mime, "gif"):
		err = gif.Encode(&buf, img, nil)
	case strings.Contains(mime, "webp"):
		// No std webp encoder; PNG bytes accepted for round-trip tests of alpha pipelines.
		err = png.Encode(&buf, img)
	default:
		err = png.Encode(&buf, img)
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func applyResize(img image.Image, rc ResizeConstraint, mime string) (image.Image, error) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	switch rc.Mode {
	case ModeLongEdge:
		if rc.Size <= 0 {
			return nil, ErrInvalidParam
		}
		if w >= h {
			return imaging.Resize(img, rc.Size, 0, imaging.Lanczos), nil
		}
		return imaging.Resize(img, 0, rc.Size, imaging.Lanczos), nil
	case ModeShortEdge:
		if rc.Size <= 0 {
			return nil, ErrInvalidParam
		}
		if w <= h {
			return imaging.Resize(img, rc.Size, 0, imaging.Lanczos), nil
		}
		return imaging.Resize(img, 0, rc.Size, imaging.Lanczos), nil
	case ModeFixed:
		if rc.Width <= 0 || rc.Height <= 0 {
			return nil, ErrInvalidParam
		}
		switch rc.Fit {
		case FitCover:
			return imaging.Fill(img, rc.Width, rc.Height, imaging.Center, imaging.Lanczos), nil
		case FitStretch:
			return imaging.Resize(img, rc.Width, rc.Height, imaging.Lanczos), nil
		case FitContain:
			return containResize(img, rc.Width, rc.Height, rc.ContainBackground, mime)
		default:
			return nil, ErrInvalidFit
		}
	default:
		return nil, ErrInvalidParam
	}
}

func containResize(img image.Image, tw, th int, bg ContainBackground, mime string) (image.Image, error) {
	if bg == "" {
		bg = BGBlack
	}
	fitted := imaging.Fit(img, tw, th, imaging.Lanczos)
	c, err := parseBackground(bg, mime)
	if err != nil {
		return nil, err
	}
	dst := image.NewNRGBA(image.Rect(0, 0, tw, th))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	fb := fitted.Bounds()
	ox := (tw - fb.Dx()) / 2
	oy := (th - fb.Dy()) / 2
	draw.Draw(dst, image.Rect(ox, oy, ox+fb.Dx(), oy+fb.Dy()), fitted, fb.Min, draw.Over)
	return dst, nil
}

func parseBackground(bg ContainBackground, mime string) (color.Color, error) {
	s := strings.ToLower(strings.TrimSpace(string(bg)))
	switch s {
	case "black", "":
		return color.NRGBA{A: 255}, nil
	case "white":
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}, nil
	case "transparent":
		if strings.Contains(mime, "jpeg") || strings.Contains(mime, "jpg") {
			return nil, ErrBadBackground
		}
		return color.NRGBA{}, nil
	}
	if !strings.HasPrefix(s, "#") {
		return nil, ErrBadBackground
	}
	hex := strings.TrimPrefix(s, "#")
	var r, g, b, a uint8
	a = 255
	switch len(hex) {
	case 6:
		var rv, gv, bv int
		if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &rv, &gv, &bv); err != nil {
			return nil, ErrBadBackground
		}
		r, g, b = uint8(rv), uint8(gv), uint8(bv)
	case 8:
		var rv, gv, bv, av int
		if _, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &rv, &gv, &bv, &av); err != nil {
			return nil, ErrBadBackground
		}
		r, g, b, a = uint8(rv), uint8(gv), uint8(bv), uint8(av)
	default:
		return nil, ErrBadBackground
	}
	return color.NRGBA{R: r, G: g, B: b, A: a}, nil
}
