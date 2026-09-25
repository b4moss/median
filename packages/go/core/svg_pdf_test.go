package core_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/b4moss/median/go/core"
	"github.com/b4moss/median/go/db"
	"github.com/b4moss/median/go/pipeline"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type fakePDFRenderer struct {
	err   error
	calls int
	page  int
}

func (f *fakePDFRenderer) RenderPage(pdf []byte, page int) (image.Image, error) {
	f.calls++
	f.page = page
	if f.err != nil {
		return nil, f.err
	}
	img := image.NewNRGBA(image.Rect(0, 0, 80, 60))
	for y := 0; y < 60; y++ {
		for x := 0; x < 80; x++ {
			img.Set(x, y, color.NRGBA{G: 200, A: 255})
		}
	}
	return img, nil
}

func TestStoreSVGSanitizesAndHashesCleanBytes(t *testing.T) {
	root := t.TempDir()
	m, err := core.New(testConfig(t, root))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	raw := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script><rect x="0" y="0" width="5" height="5"/></svg>`)
	clean, err := pipeline.SanitizeSVG(raw)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(clean)
	wantHash := hex.EncodeToString(sum[:])

	res, err := m.Store(ctx, bytes.NewReader(raw), int64(len(raw)), core.StoreOptions{
		MIME: "image/svg+xml", Filename: "x.svg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.MIME != "image/svg+xml" {
		t.Fatalf("mime %s", res.MIME)
	}
	if res.Hash != wantHash || res.Size != int64(len(clean)) {
		t.Fatalf("hash/size got %s %d want %s %d", res.Hash, res.Size, wantHash, len(clean))
	}
	body, err := os.ReadFile(filepath.Join(root, res.Path))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(body)), "<script") {
		t.Fatalf("script in stored file: %s", body)
	}
	if !bytes.Equal(body, clean) {
		t.Fatal("stored bytes != sanitized")
	}
}

func TestStoreSVGInvalidLeavesNoFile(t *testing.T) {
	root := t.TempDir()
	m, err := core.New(testConfig(t, root))
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Store(context.Background(), bytes.NewReader([]byte("<svg><rect>")), 11, core.StoreOptions{
		MIME: "image/svg+xml", Filename: "bad.svg",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Fatalf("leftover %#v", entries)
	}
}

func TestStoreSVGDeniedByMIME(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)
	cfg.MIMEDeny = []string{"image/svg+xml"}
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>`)
	_, err = m.Store(context.Background(), bytes.NewReader(raw), int64(len(raw)), core.StoreOptions{
		MIME: "image/svg+xml", Filename: "x.svg",
	})
	if err == nil {
		t.Fatal("expected mime deny")
	}
}

func TestStorePDFWithoutThumbnail(t *testing.T) {
	root := t.TempDir()
	m, err := core.New(testConfig(t, root))
	if err != nil {
		t.Fatal(err)
	}
	pdf := []byte("%PDF-1.4 fake")
	sum := sha256.Sum256(pdf)
	res, err := m.Store(context.Background(), bytes.NewReader(pdf), int64(len(pdf)), core.StoreOptions{
		MIME: "application/pdf", Filename: "a.pdf",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Hash != hex.EncodeToString(sum[:]) || len(res.Variants) != 0 {
		t.Fatalf("%+v", res)
	}
}

func TestStorePDFThumbnailWithFakeRenderer(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)
	fake := &fakePDFRenderer{}
	cfg.PDFRenderer = fake
	cfg.Thumbnails = &pipeline.ThumbnailsConfig{
		Presets: map[string]pipeline.ThumbnailPreset{
			"sm": {Mode: pipeline.ModeLongEdge, Size: 40, Quality: 0.8},
		},
	}
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pdf := []byte("%PDF-1.4 content")
	sum := sha256.Sum256(pdf)
	res, err := m.Store(context.Background(), bytes.NewReader(pdf), int64(len(pdf)), core.StoreOptions{
		MIME: "application/pdf", Filename: "a.pdf",
		PDFThumbnail:  true,
		ThumbnailKeys: []string{"sm"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Hash != hex.EncodeToString(sum[:]) {
		t.Fatalf("hash %s", res.Hash)
	}
	if fake.calls != 1 || fake.page != 0 {
		t.Fatalf("renderer calls=%d page=%d", fake.calls, fake.page)
	}
	if len(res.Variants) != 1 || res.Variants[0].Key != "sm" {
		t.Fatalf("%+v", res.Variants)
	}
	if _, err := os.Stat(filepath.Join(root, res.Variants[0].Path)); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(root, res.Path))
	if !bytes.Equal(body, pdf) {
		t.Fatal("pdf body changed")
	}
}

func TestStorePDFThumbnailRequiresRenderer(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)
	cfg.Thumbnails = &pipeline.ThumbnailsConfig{
		Presets: map[string]pipeline.ThumbnailPreset{
			"sm": {Mode: pipeline.ModeLongEdge, Size: 40},
		},
	}
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pdf := []byte("%PDF")
	_, err = m.Store(context.Background(), bytes.NewReader(pdf), int64(len(pdf)), core.StoreOptions{
		MIME: "application/pdf", PDFThumbnail: true, ThumbnailKeys: []string{"sm"},
	})
	if !errors.Is(err, core.ErrPDFRendererRequired) {
		t.Fatalf("got %v", err)
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Fatalf("leftover %#v", entries)
	}
}

func TestStorePDFThumbnailKeysWithoutFlagRejected(t *testing.T) {
	root := t.TempDir()
	m, err := core.New(testConfig(t, root))
	if err != nil {
		t.Fatal(err)
	}
	pdf := []byte("%PDF")
	_, err = m.Store(context.Background(), bytes.NewReader(pdf), int64(len(pdf)), core.StoreOptions{
		MIME: "application/pdf", ThumbnailKeys: []string{"sm"},
	})
	if !errors.Is(err, pipeline.ErrInvalidImage) {
		t.Fatalf("got %v", err)
	}
}

func TestStorePDFRenderErrorCompensates(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)
	cfg.PDFRenderer = &fakePDFRenderer{err: errors.New("boom")}
	cfg.Thumbnails = &pipeline.ThumbnailsConfig{
		Presets: map[string]pipeline.ThumbnailPreset{
			"sm": {Mode: pipeline.ModeLongEdge, Size: 40},
		},
	}
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pdf := []byte("%PDF")
	_, err = m.Store(context.Background(), bytes.NewReader(pdf), int64(len(pdf)), core.StoreOptions{
		MIME: "application/pdf", PDFThumbnail: true, ThumbnailKeys: []string{"sm"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 0 {
		t.Fatalf("leftover %#v", entries)
	}
}

func TestStorePDFThumbnailWithDB(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "m.db")
	gdb, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := gdb.DB()
	if err := db.MigrateUp(context.Background(), sqlDB, "sqlite3", db.IDAutoIncrement); err != nil {
		t.Fatal(err)
	}
	cfg := testConfig(t, root)
	cfg.DB = &core.DBConfig{Gorm: gdb, IDStrategy: db.IDAutoIncrement}
	cfg.DefaultActor = "actor-1"
	cfg.PDFRenderer = &fakePDFRenderer{}
	cfg.Thumbnails = &pipeline.ThumbnailsConfig{
		Presets: map[string]pipeline.ThumbnailPreset{
			"sm": {Mode: pipeline.ModeLongEdge, Size: 32, Quality: 0.8},
		},
		DefaultKeys: []string{"sm"},
	}
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pdf := []byte("%PDF-db-test")
	res, err := m.Store(context.Background(), bytes.NewReader(pdf), int64(len(pdf)), core.StoreOptions{
		MIME: "application/pdf", Filename: "doc.pdf", PDFThumbnail: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ID == nil || len(res.Variants) != 1 || res.Variants[0].ID == nil {
		t.Fatalf("%+v", res)
	}
	parent, err := m.Get(context.Background(), "", core.GetOptions{ID: res.ID})
	if err != nil || parent.MIME != "application/pdf" {
		t.Fatal(err)
	}
}

func TestStoreJPEGStillWorks(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)
	cfg.Thumbnails = &pipeline.ThumbnailsConfig{
		Presets: map[string]pipeline.ThumbnailPreset{
			"sm": {Mode: pipeline.ModeLongEdge, Size: 32, Quality: 0.8},
		},
		DefaultKeys: []string{"sm"},
	}
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	payload := buf.Bytes()
	res, err := m.Store(context.Background(), bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "p.png",
	})
	if err != nil || len(res.Variants) != 1 {
		t.Fatalf("%v %+v", err, res)
	}
}
