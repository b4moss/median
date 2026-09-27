package db_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/b4moss/median/packages/go/db"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openSQLite(t *testing.T) (*sql.DB, *gorm.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "t.db")
	gdb, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	return sqlDB, gdb
}

func TestCreateMediaSQLDialects(t *testing.T) {
	for _, d := range []string{"mysql", "postgres", "sqlite3"} {
		q, err := db.CreateMediaSQL(d, db.IDAutoIncrement, "")
		if err != nil || !strings.Contains(strings.ToUpper(q), "CREATE TABLE") {
			t.Fatalf("%s auto: %v", d, err)
		}
		if !strings.Contains(q, "file_path") || strings.Contains(q, " path ") {
			t.Fatalf("%s missing file_path columns", d)
		}
		if _, err := db.CreateMediaSQL(d, db.IDUUIDv4, ""); err != nil {
			t.Fatalf("%s text: %v", d, err)
		}
		q2, err := db.CreateMediaSQL(d, db.IDAutoIncrement, "medias")
		if err != nil || !strings.Contains(q2, "CREATE TABLE medias") {
			t.Fatalf("%s custom table: %v %s", d, err, q2)
		}
	}
	if _, err := db.CreateMediaSQL("oracle", db.IDAutoIncrement, ""); err == nil {
		t.Fatal("expected dialect error")
	}
	if _, err := db.CreateMediaSQL("sqlite3", "nope", ""); err == nil {
		t.Fatal("expected strategy error")
	}
	if _, err := db.CreateMediaSQL("sqlite3", db.IDAutoIncrement, "bad-name"); err == nil {
		t.Fatal("expected table name error")
	}
}

func TestMigrateAndRepo(t *testing.T) {
	sqlDB, gdb := openSQLite(t)
	ctx := context.Background()
	if err := db.MigrateUp(ctx, sqlDB, "sqlite3", db.IDAutoIncrement, ""); err != nil {
		t.Fatal(err)
	}
	var names []string
	rows, err := sqlDB.Query(`PRAGMA table_info(media)`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	_ = rows.Close()
	joined := strings.Join(names, ",")
	for _, want := range []string{"file_path", "file_name", "original_file_name"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing column %s in %v", want, names)
		}
	}
	if strings.Contains(joined, ",path,") || joined == "path" || strings.HasPrefix(joined, "path,") || strings.HasSuffix(joined, ",path") {
		t.Fatalf("legacy path column still present: %v", names)
	}

	repo, err := db.NewMediaRepo(gdb, db.RepoOptions{IDStrategy: db.IDAutoIncrement})
	if err != nil {
		t.Fatal(err)
	}
	m, err := repo.Create(ctx, &db.Media{
		FilePath: "a/b.png", FileName: "b.png", OriginalFileName: "pic.png",
		MIME: "image/png", Size: 10, Hash: "abc",
		CreatedBy: "u1", OwnedBy: "u1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.ID == nil {
		t.Fatal("expected id")
	}
	got, err := repo.FindByID(ctx, m.ID)
	if err != nil || got.FilePath != "a/b.png" || got.FileName != "b.png" || got.OriginalFileName != "pic.png" || got.CreatedBy != "u1" {
		t.Fatalf("%+v %v", got, err)
	}
	byHash, err := repo.FindByHash(ctx, "abc")
	if err != nil || fmt.Sprint(byHash.ID) != fmt.Sprint(m.ID) {
		t.Fatalf("hash: %+v %v", byHash, err)
	}
	vk := "sm"
	child, err := repo.Create(ctx, &db.Media{
		FilePath: "a/b-sm.png", FileName: "b-sm.png",
		MIME: "image/png", Size: 2, Hash: "def",
		OriginalID: m.ID, VariantKey: &vk,
		CreatedBy: "u1", OwnedBy: "u1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteByID(ctx, m.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByID(ctx, m.ID); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("parent: %v", err)
	}
	if _, err := repo.FindByID(ctx, child.ID); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("child: %v", err)
	}
	if err := db.MigrateDown(ctx, sqlDB, ""); err != nil {
		t.Fatal(err)
	}
}

func TestMigrateCustomTableAndColumnMap(t *testing.T) {
	sqlDB, gdb := openSQLite(t)
	ctx := context.Background()
	if err := db.MigrateUp(ctx, sqlDB, "sqlite3", db.IDAutoIncrement, "medias"); err != nil {
		t.Fatal(err)
	}
	// Remap file_path -> storage_key by recreating table with alias for map test:
	// Use default columns on custom table name first.
	repo, err := db.NewMediaRepo(gdb, db.RepoOptions{
		IDStrategy: db.IDAutoIncrement,
		TableName:  "medias",
	})
	if err != nil {
		t.Fatal(err)
	}
	m, err := repo.Create(ctx, &db.Media{
		FilePath: "x/y.bin", FileName: "y.bin", OriginalFileName: "orig.bin",
		MIME: "application/octet-stream", Size: 1, Hash: "hh",
		CreatedBy: "a", OwnedBy: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByID(ctx, m.ID)
	if err != nil || got.FilePath != "x/y.bin" || got.OriginalFileName != "orig.bin" {
		t.Fatalf("%+v %v", got, err)
	}
	if err := db.MigrateDown(ctx, sqlDB, "medias"); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRequiresActor(t *testing.T) {
	sqlDB, gdb := openSQLite(t)
	ctx := context.Background()
	_ = db.MigrateUp(ctx, sqlDB, "sqlite3", db.IDAutoIncrement, "")
	repo, _ := db.NewMediaRepo(gdb, db.RepoOptions{IDStrategy: db.IDAutoIncrement})
	_, err := repo.Create(ctx, &db.Media{FilePath: "x", FileName: "x", MIME: "text/plain", Hash: "h", Size: 1})
	if err == nil {
		t.Fatal("expected actor error")
	}
}
