package pipeline_test

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/b4moss/median/packages/go/pipeline"
)

func TestSanitizeSVGKeepsShapesAndStyle(t *testing.T) {
	in := []byte(`<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10">
  <style>.a{fill:red}</style>
  <rect class="a" x="0" y="0" width="10" height="10"/>
</svg>`)
	out, err := pipeline.SanitizeSVG(in)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "<rect") {
		t.Fatalf("rect missing: %s", s)
	}
	if !strings.Contains(s, "<style") {
		t.Fatalf("style missing: %s", s)
	}
	if !strings.Contains(s, "fill:red") {
		t.Fatalf("style body missing: %s", s)
	}
}

func TestSanitizeSVGStripsScriptHandlersAndExternalRefs(t *testing.T) {
	in := []byte(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
  <script>alert(1)</script>
  <rect onclick="evil()" x="0" y="0" width="1" height="1"/>
  <a href="javascript:alert(1)"><text>x</text></a>
  <image xlink:href="https://evil.example/x.png"/>
  <circle cx="1" cy="1" r="1"/>
</svg>`)
	out, err := pipeline.SanitizeSVG(in)
	if err != nil {
		t.Fatal(err)
	}
	s := strings.ToLower(string(out))
	if strings.Contains(s, "<script") || strings.Contains(s, "alert(1)") {
		t.Fatalf("script leaked: %s", s)
	}
	if strings.Contains(s, "onclick") {
		t.Fatalf("handler leaked: %s", s)
	}
	if strings.Contains(s, "javascript:") || strings.Contains(s, "https://evil.example") {
		t.Fatalf("external ref leaked: %s", s)
	}
	if !strings.Contains(s, "<circle") {
		t.Fatalf("safe circle missing: %s", s)
	}
}

func TestSanitizeSVGInvalidXML(t *testing.T) {
	_, err := pipeline.SanitizeSVG([]byte(`<svg><rect>`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestThumbnailsFromImage(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 100, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.NRGBA{B: 255, A: 255})
		}
	}
	thumbs := &pipeline.ThumbnailsConfig{
		Presets: map[string]pipeline.ThumbnailPreset{
			"sm": {Mode: pipeline.ModeLongEdge, Size: 40, Quality: 0.8},
		},
	}
	vars, err := pipeline.ThumbnailsFromImage(img, thumbs, []string{"sm"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vars) != 1 || vars[0].Key != "sm" || len(vars[0].Bytes) == 0 {
		t.Fatalf("%+v", vars)
	}
	if vars[0].Width != 40 {
		t.Fatalf("width %d", vars[0].Width)
	}
	if vars[0].MIME != "image/png" {
		t.Fatalf("mime %s", vars[0].MIME)
	}

	_, err = pipeline.ThumbnailsFromImage(img, thumbs, []string{"missing"})
	if err == nil {
		t.Fatal("unknown key")
	}
	_, err = pipeline.ThumbnailsFromImage(nil, thumbs, []string{"sm"})
	if err == nil {
		t.Fatal("nil image")
	}
}

type fakePDF struct {
	page int
	err  error
	img  image.Image
}

func (f *fakePDF) RenderPage(pdf []byte, page int) (image.Image, error) {
	f.page = page
	if f.err != nil {
		return nil, f.err
	}
	if f.img != nil {
		return f.img, nil
	}
	img := image.NewNRGBA(image.Rect(0, 0, 60, 40))
	return img, nil
}

func TestFakePDFRendererPageZero(t *testing.T) {
	f := &fakePDF{}
	img, err := f.RenderPage([]byte("%PDF"), 0)
	if err != nil || img == nil {
		t.Fatal(err)
	}
	if f.page != 0 {
		t.Fatalf("page %d", f.page)
	}
}
