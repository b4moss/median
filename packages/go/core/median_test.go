package core_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
	"testing"

	"github.com/b4moss/median/packages/go/core"
	"github.com/b4moss/median/packages/go/db"
	"github.com/b4moss/median/packages/go/pipeline"
	"github.com/b4moss/median/packages/go/storage"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func intPtr(v int) *int { return &v }

func testConfig(t *testing.T, root string) core.Config {
	t.Helper()
	return core.Config{
		Storages: map[string]core.StorageConfig{
			"local": {Driver: core.DriverLocal, Path: root},
		},
		DefaultKey:     "local",
		DirLetterCount: intPtr(1),
		DirNestDepth:   intPtr(2),
	}
}

func TestNewDefaultsAndValidation(t *testing.T) {
	root := t.TempDir()
	if _, err := core.New(testConfig(t, root)); err != nil {
		t.Fatal(err)
	}
	bad := testConfig(t, root)
	bad.DefaultKey = "missing"
	if _, err := core.New(bad); err == nil {
		t.Fatal("defaultKey")
	}
	bad = testConfig(t, root)
	bad.DirLetterCount = nil
	if _, err := core.New(bad); err == nil {
		t.Fatal("DirLetterCount")
	}
	bad = testConfig(t, root)
	bad.Storages["s3"] = core.StorageConfig{Driver: core.DriverS3}
	bad.DefaultKey = "s3"
	if _, err := core.New(bad); err == nil {
		t.Fatal("s3 bucket required")
	}
}

func TestStoreDeleteGetNoDB(t *testing.T) {
	root := t.TempDir()
	m, err := core.New(testConfig(t, root))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payload := []byte("payload-bytes")
	sum := sha256.Sum256(payload)
	wantHash := hex.EncodeToString(sum[:])
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "text/plain", Filename: "note.txt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Hash != wantHash || res.StorageKey != "local" {
		t.Fatalf("%+v", res)
	}
	meta, err := m.Get(ctx, res.Path, core.GetOptions{})
	if err != nil || meta.Path != res.Path {
		t.Fatal(err)
	}
	full, err := m.Get(ctx, res.Path, core.GetOptions{WithBody: true})
	if err != nil || !bytes.Equal(full.Body, payload) {
		t.Fatal(err)
	}
	if err := m.Delete(ctx, res.Path, core.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(ctx, res.Path, core.GetOptions{}); !errors.Is(err, storage.ErrNotFound) {
		t.Fatal(err)
	}
}

func TestStoreWithDBAndCascade(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "m.db")
	gdb, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := gdb.DB()
	if err := db.MigrateUp(context.Background(), sqlDB, "sqlite3", db.IDAutoIncrement, ""); err != nil {
		t.Fatal(err)
	}
	cfg := testConfig(t, root)
	cfg.DB = &core.DBConfig{Gorm: gdb, IDStrategy: db.IDAutoIncrement}
	cfg.DefaultActor = "actor-1"
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
	ctx := context.Background()
	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	payload := buf.Bytes()

	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ID == nil || len(res.Variants) != 1 {
		t.Fatalf("%+v", res)
	}
	// duplicate hash returns existing
	res2, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic2.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(res2.ID) != fmt.Sprint(res.ID) {
		t.Fatalf("dup %v vs %v", res2.ID, res.ID)
	}
	got, err := m.Get(ctx, "", core.GetOptions{ID: res.ID, WithBody: true})
	if err != nil || len(got.Body) == 0 {
		t.Fatal(err)
	}
	if err := m.Delete(ctx, "", core.DeleteOptions{ID: res.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(ctx, "", core.GetOptions{ID: res.ID}); !errors.Is(err, db.ErrNotFound) {
		t.Fatal(err)
	}
}

func TestPresignRequiresS3(t *testing.T) {
	m, err := core.New(testConfig(t, t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.PresignGet(context.Background(), "a", "local", 0); !errors.Is(err, core.ErrNotS3) {
		t.Fatalf("%v", err)
	}
}
