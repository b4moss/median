package core_test

import (
	"bytes"
	"context"
	"database/sql"
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
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openCoreDB(t *testing.T, table string, strategy string) (*gorm.DB, *sql.DB) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "m.db")
	gdb, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	if strategy == "" {
		strategy = db.IDAutoIncrement
	}
	if err := db.MigrateUp(context.Background(), sqlDB, "sqlite3", strategy, table); err != nil {
		t.Fatal(err)
	}
	return gdb, sqlDB
}

func makePNGBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func setStatus(t *testing.T, sqlDB *sql.DB, table string, id any, status string) {
	t.Helper()
	if table == "" {
		table = "media"
	}
	q := fmt.Sprintf("UPDATE %s SET status = ? WHERE id = ?", table)
	if _, err := sqlDB.Exec(q, status, id); err != nil {
		t.Fatal(err)
	}
}

func setWidth(t *testing.T, sqlDB *sql.DB, table string, id any, width any) {
	t.Helper()
	if table == "" {
		table = "media"
	}
	q := fmt.Sprintf("UPDATE %s SET width = ? WHERE id = ?", table)
	if _, err := sqlDB.Exec(q, width, id); err != nil {
		t.Fatal(err)
	}
}

func familyMedian(t *testing.T, root string, gdb *gorm.DB, strategy, table string) *core.Median {
	t.Helper()
	if strategy == "" {
		strategy = db.IDAutoIncrement
	}
	cfg := testConfig(t, root)
	cfg.DB = &core.DBConfig{Gorm: gdb, IDStrategy: strategy, TableName: table}
	cfg.DefaultActor = "actor-1"
	cfg.Thumbnails = &pipeline.ThumbnailsConfig{
		Presets: map[string]pipeline.ThumbnailPreset{
			"sm": {Mode: pipeline.ModeLongEdge, Size: 32, Quality: 0.8},
			"md": {Mode: pipeline.ModeLongEdge, Size: 48, Quality: 0.8},
			"aa": {Mode: pipeline.ModeLongEdge, Size: 40, Quality: 0.8},
			"zz": {Mode: pipeline.ModeLongEdge, Size: 40, Quality: 0.8},
		},
		DefaultKeys: []string{"sm", "md"},
	}
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func assertVariantShape(t *testing.T, v core.VariantResult, storageKey string) {
	t.Helper()
	if v.Key == "" || v.Path == "" || v.MIME == "" || v.Hash == "" || v.ID == nil {
		t.Fatalf("incomplete variant: %+v", v)
	}
	if v.Size <= 0 {
		t.Fatalf("variant size: %+v", v)
	}
	if v.StorageKey != storageKey {
		t.Fatalf("storageKey want %s got %s", storageKey, v.StorageKey)
	}
}

func TestGetWithAllThumbnails_HappyPathOriginalAndDerived(t *testing.T) {
	root := t.TempDir()
	gdb, _ := openCoreDB(t, "", db.IDAutoIncrement)
	m := familyMedian(t, root, gdb, db.IDAutoIncrement, "")
	ctx := context.Background()
	payload := makePNGBytes(t, 80, 80)

	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic.png", ThumbnailKeys: []string{"sm", "md", "aa", "zz"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Variants) != 4 {
		t.Fatalf("store variants: %d", len(res.Variants))
	}

	got, err := m.GetWithAllThumbnails(ctx, res.ID, core.GetWithAllThumbnailsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(got.ID) != fmt.Sprint(res.ID) || got.Path != res.Path || got.Hash != res.Hash {
		t.Fatalf("parent meta: %+v vs %+v", got, res)
	}
	if got.StorageKey != "local" {
		t.Fatalf("default storageKey: %s", got.StorageKey)
	}
	if got.Variants == nil || len(got.Variants) != 4 {
		t.Fatalf("variants len=%d nil=%v", len(got.Variants), got.Variants == nil)
	}
	// width asc: sm(32) < aa(40)=zz(40) < md(48); same width aa before zz
	wantKeys := []string{"sm", "aa", "zz", "md"}
	for i, k := range wantKeys {
		if got.Variants[i].Key != k {
			t.Fatalf("order[%d]=%s want %s (full=%v)", i, got.Variants[i].Key, k, keysOf(got.Variants))
		}
		assertVariantShape(t, got.Variants[i], "local")
	}
	for i := 1; i < len(got.Variants); i++ {
		if got.Variants[i].Width < got.Variants[i-1].Width {
			t.Fatalf("not width-asc: %+v", got.Variants)
		}
	}

	// derived active id → same family, top-level is parent
	childID := res.Variants[0].ID
	fromChild, err := m.GetWithAllThumbnails(ctx, childID, core.GetWithAllThumbnailsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(fromChild.ID) != fmt.Sprint(res.ID) || fromChild.Path != res.Path {
		t.Fatalf("from child top-level should be parent: %+v", fromChild)
	}
	if len(fromChild.Variants) != 4 {
		t.Fatalf("from child variants: %d", len(fromChild.Variants))
	}
}

func keysOf(vs []core.VariantResult) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.Key
	}
	return out
}

func TestGetWithAllThumbnails_InactiveDerivedStillResolves(t *testing.T) {
	root := t.TempDir()
	gdb, sqlDB := openCoreDB(t, "", db.IDAutoIncrement)
	m := familyMedian(t, root, gdb, db.IDAutoIncrement, "")
	ctx := context.Background()
	payload := makePNGBytes(t, 64, 64)
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic.png", ThumbnailKeys: []string{"sm", "md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	inactiveID := res.Variants[0].ID
	inactiveKey := res.Variants[0].Key
	setStatus(t, sqlDB, "", inactiveID, "inactive")

	got, err := m.GetWithAllThumbnails(ctx, inactiveID, core.GetWithAllThumbnailsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(got.ID) != fmt.Sprint(res.ID) {
		t.Fatalf("want parent %v got %v", res.ID, got.ID)
	}
	for _, v := range got.Variants {
		if v.Key == inactiveKey || fmt.Sprint(v.ID) == fmt.Sprint(inactiveID) {
			t.Fatalf("inactive child must be excluded: %+v", got.Variants)
		}
	}
	if len(got.Variants) != 1 {
		t.Fatalf("want 1 active sibling, got %+v", got.Variants)
	}
}

func TestGetWithAllThumbnails_EmptyVariantsSlice(t *testing.T) {
	root := t.TempDir()
	gdb, _ := openCoreDB(t, "", db.IDAutoIncrement)
	cfg := testConfig(t, root)
	cfg.DB = &core.DBConfig{Gorm: gdb, IDStrategy: db.IDAutoIncrement}
	cfg.DefaultActor = "actor-1"
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payload := []byte("plain-no-thumbs")
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "text/plain", Filename: "note.txt",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.GetWithAllThumbnails(ctx, res.ID, core.GetWithAllThumbnailsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Variants == nil {
		t.Fatal("Variants must be empty slice, not nil")
	}
	if len(got.Variants) != 0 {
		t.Fatalf("want empty, got %d", len(got.Variants))
	}
}

func TestGetWithAllThumbnails_FiltersInactiveChildren(t *testing.T) {
	root := t.TempDir()
	gdb, sqlDB := openCoreDB(t, "", db.IDAutoIncrement)
	m := familyMedian(t, root, gdb, db.IDAutoIncrement, "")
	ctx := context.Background()
	payload := makePNGBytes(t, 64, 64)
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic.png", ThumbnailKeys: []string{"sm", "md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	setStatus(t, sqlDB, "", res.Variants[0].ID, "inactive")

	got, err := m.GetWithAllThumbnails(ctx, res.ID, core.GetWithAllThumbnailsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Variants) != 1 || got.Variants[0].Key != "md" {
		t.Fatalf("want only md: %+v", got.Variants)
	}
}

func TestGetWithAllThumbnails_WidthZeroAndNullOrdering(t *testing.T) {
	root := t.TempDir()
	gdb, sqlDB := openCoreDB(t, "", db.IDAutoIncrement)
	m := familyMedian(t, root, gdb, db.IDAutoIncrement, "")
	ctx := context.Background()
	payload := makePNGBytes(t, 64, 64)
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic.png", ThumbnailKeys: []string{"sm", "md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	// force widths: sm→0, md→NULL; add another via repo with width 10
	var smID, mdID any
	for _, v := range res.Variants {
		if v.Key == "sm" {
			smID = v.ID
		}
		if v.Key == "md" {
			mdID = v.ID
		}
	}
	setWidth(t, sqlDB, "", smID, 0)
	setWidth(t, sqlDB, "", mdID, nil)

	repo, err := db.NewMediaRepo(gdb, db.RepoOptions{IDStrategy: db.IDAutoIncrement})
	if err != nil {
		t.Fatal(err)
	}
	vk := "mid"
	w10 := 10
	child, err := repo.Create(ctx, &db.Media{
		FilePath: "x/mid.png", FileName: "mid.png",
		MIME: "image/png", Size: 1, Hash: "midhash",
		Width: &w10, OriginalID: res.ID, VariantKey: &vk,
		CreatedBy: "actor-1", OwnedBy: "actor-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = child

	got, err := m.GetWithAllThumbnails(ctx, res.ID, core.GetWithAllThumbnailsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Variants) != 3 {
		t.Fatalf("variants: %d %+v", len(got.Variants), keysOf(got.Variants))
	}
	// order: width 0 (sm), width 10 (mid), width NULL (md) last
	if got.Variants[0].Key != "sm" || got.Variants[1].Key != "mid" || got.Variants[2].Key != "md" {
		t.Fatalf("order: %v", keysOf(got.Variants))
	}
}

func TestGetWithAllThumbnails_StorageKey(t *testing.T) {
	root := t.TempDir()
	root2 := t.TempDir()
	gdb, _ := openCoreDB(t, "", db.IDAutoIncrement)
	cfg := testConfig(t, root)
	cfg.Storages["alt"] = core.StorageConfig{Driver: core.DriverLocal, Path: root2}
	cfg.DB = &core.DBConfig{Gorm: gdb, IDStrategy: db.IDAutoIncrement}
	cfg.DefaultActor = "actor-1"
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payload := []byte("x")
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "text/plain", Filename: "a.txt", StorageKey: "alt",
	})
	if err != nil {
		t.Fatal(err)
	}

	def, err := m.GetWithAllThumbnails(ctx, res.ID, core.GetWithAllThumbnailsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if def.StorageKey != "local" {
		t.Fatalf("omit → defaultKey local, got %s", def.StorageKey)
	}

	alt, err := m.GetWithAllThumbnails(ctx, res.ID, core.GetWithAllThumbnailsOptions{StorageKey: "alt"})
	if err != nil {
		t.Fatal(err)
	}
	if alt.StorageKey != "alt" {
		t.Fatalf("explicit alt, got %s", alt.StorageKey)
	}
}

func TestGetWithAllThumbnails_CustomTableName(t *testing.T) {
	root := t.TempDir()
	gdb, _ := openCoreDB(t, "medias", db.IDAutoIncrement)
	m := familyMedian(t, root, gdb, db.IDAutoIncrement, "medias")
	ctx := context.Background()
	payload := makePNGBytes(t, 48, 48)
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.GetWithAllThumbnails(ctx, res.ID, core.GetWithAllThumbnailsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Variants) != 2 {
		t.Fatalf("variants: %d", len(got.Variants))
	}
	if got.Variants[0].Width > got.Variants[1].Width {
		t.Fatalf("not asc: %+v", got.Variants)
	}
}

func TestGetWithAllThumbnails_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("db required", func(t *testing.T) {
		m, err := core.New(testConfig(t, t.TempDir()))
		if err != nil {
			t.Fatal(err)
		}
		_, err = m.GetWithAllThumbnails(ctx, 1, core.GetWithAllThumbnailsOptions{})
		if !errors.Is(err, core.ErrDBRequired) {
			t.Fatalf("want ErrDBRequired, got %v", err)
		}
		if errors.Is(err, db.ErrNotFound) {
			t.Fatal("must not be not found")
		}
	})

	root := t.TempDir()
	gdb, sqlDB := openCoreDB(t, "", db.IDAutoIncrement)
	m := familyMedian(t, root, gdb, db.IDAutoIncrement, "")
	payload := makePNGBytes(t, 40, 40)
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic.png",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("not found id", func(t *testing.T) {
		_, err := m.GetWithAllThumbnails(ctx, 999999, core.GetWithAllThumbnailsOptions{})
		if !errors.Is(err, db.ErrNotFound) {
			t.Fatalf("want not found, got %v", err)
		}
	})

	t.Run("inactive parent via child", func(t *testing.T) {
		childID := res.Variants[0].ID
		setStatus(t, sqlDB, "", res.ID, "inactive")
		_, err := m.GetWithAllThumbnails(ctx, childID, core.GetWithAllThumbnailsOptions{})
		if !errors.Is(err, db.ErrNotFound) {
			t.Fatalf("want not found, got %v", err)
		}
		setStatus(t, sqlDB, "", res.ID, "active")
	})

	t.Run("id required", func(t *testing.T) {
		_, err := m.GetWithAllThumbnails(ctx, nil, core.GetWithAllThumbnailsOptions{})
		if !errors.Is(err, core.ErrIDRequired) {
			t.Fatalf("nil: %v", err)
		}
		_, err = m.GetWithAllThumbnails(ctx, "", core.GetWithAllThumbnailsOptions{})
		if !errors.Is(err, core.ErrIDRequired) {
			t.Fatalf("empty: %v", err)
		}
	})

	t.Run("unknown storage key", func(t *testing.T) {
		_, err := m.GetWithAllThumbnails(ctx, res.ID, core.GetWithAllThumbnailsOptions{StorageKey: "nope"})
		if !errors.Is(err, core.ErrUnknownKey) {
			t.Fatalf("want ErrUnknownKey, got %v", err)
		}
	})
}

func TestGetWithAllThumbnails_DoesNotChangeGet(t *testing.T) {
	root := t.TempDir()
	gdb, sqlDB := openCoreDB(t, "", db.IDAutoIncrement)
	m := familyMedian(t, root, gdb, db.IDAutoIncrement, "")
	ctx := context.Background()
	payload := makePNGBytes(t, 32, 32)
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	setStatus(t, sqlDB, "", res.ID, "inactive")
	// Get still returns inactive rows (U9)
	got, err := m.Get(ctx, "", core.GetOptions{ID: res.ID})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(got.ID) != fmt.Sprint(res.ID) {
		t.Fatalf("%+v", got)
	}
	_, err = m.GetWithAllThumbnails(ctx, res.ID, core.GetWithAllThumbnailsOptions{})
	if !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("family should be not found: %v", err)
	}
}
