package pipeline_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/b4moss/median/packages/go/pipeline"
)

func makePNG(t *testing.T, w, h int, c color.Color) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestLongEdgeAndContainBackgrounds(t *testing.T) {
	src := makePNG(t, 200, 100, color.NRGBA{R: 255, A: 255})
	res, err := pipeline.Process(src, pipeline.ProcessOptions{
		MIME: "image/png",
		Resize: &pipeline.ResizeConstraint{Mode: pipeline.ModeLongEdge, Size: 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Width != 100 || res.Height != 50 {
		t.Fatalf("longEdge got %dx%d", res.Width, res.Height)
	}

	res, err = pipeline.Process(src, pipeline.ProcessOptions{
		MIME: "image/png",
		Resize: &pipeline.ResizeConstraint{
			Mode: pipeline.ModeFixed, Width: 100, Height: 100, Fit: pipeline.FitContain,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Width != 100 || res.Height != 100 {
		t.Fatalf("contain size %dx%d", res.Width, res.Height)
	}
	img, err := png.Decode(bytes.NewReader(res.Bytes))
	if err != nil {
		t.Fatal(err)
	}
	// corner should be black by default
	r, g, b, a := img.At(0, 0).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Fatalf("expected black corner got %d %d %d %d", r>>8, g>>8, b>>8, a>>8)
	}

	res, err = pipeline.Process(src, pipeline.ProcessOptions{
		MIME: "image/png",
		Resize: &pipeline.ResizeConstraint{
			Mode: pipeline.ModeFixed, Width: 100, Height: 100, Fit: pipeline.FitContain,
			ContainBackground: pipeline.BGWhite,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	img, _ = png.Decode(bytes.NewReader(res.Bytes))
	r, g, b, a = img.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Fatalf("expected white corner")
	}

	res, err = pipeline.Process(src, pipeline.ProcessOptions{
		MIME: "image/png",
		Resize: &pipeline.ResizeConstraint{
			Mode: pipeline.ModeFixed, Width: 100, Height: 100, Fit: pipeline.FitContain,
			ContainBackground: "#00ff00",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	img, _ = png.Decode(bytes.NewReader(res.Bytes))
	r, g, b, _ = img.At(0, 0).RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 {
		t.Fatalf("expected green corner %d %d %d", r>>8, g>>8, b>>8)
	}

	_, err = pipeline.Process(src, pipeline.ProcessOptions{
		MIME: "image/jpeg",
		Resize: &pipeline.ResizeConstraint{
			Mode: pipeline.ModeFixed, Width: 50, Height: 50, Fit: pipeline.FitContain,
			ContainBackground: pipeline.BGTransparent,
		},
	})
	if err == nil {
		t.Fatal("jpeg+transparent should fail")
	}
}

func TestThumbnails(t *testing.T) {
	src := makePNG(t, 200, 200, color.NRGBA{B: 255, A: 255})
	res, err := pipeline.Process(src, pipeline.ProcessOptions{
		MIME: "image/png",
		Thumbnails: &pipeline.ThumbnailsConfig{
			Presets: map[string]pipeline.ThumbnailPreset{
				"sm": {Mode: pipeline.ModeLongEdge, Size: 50, Quality: 0.8},
			},
			DefaultKeys: []string{"sm"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Variants) != 1 || res.Variants[0].Key != "sm" || res.Variants[0].Width != 50 {
		t.Fatalf("%+v", res.Variants)
	}
	_, err = pipeline.Process(src, pipeline.ProcessOptions{
		MIME:          "image/png",
		ThumbnailKeys: []string{"nope"},
		Thumbnails:    &pipeline.ThumbnailsConfig{Presets: map[string]pipeline.ThumbnailPreset{}},
	})
	if err == nil {
		t.Fatal("unknown key")
	}
}
