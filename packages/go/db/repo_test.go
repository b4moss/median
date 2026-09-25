package db_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/b4moss/median/go/db"
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
		q, err := db.CreateMediaSQL(d, db.IDAutoIncrement)
		if err != nil || !strings.Contains(strings.ToUpper(q), "CREATE TABLE") {
			t.Fatalf("%s auto: %v", d, err)
		}
		if _, err := db.CreateMediaSQL(d, db.IDUUIDv4); err != nil {
			t.Fatalf("%s text: %v", d, err)
		}
	}
	if _, err := db.CreateMediaSQL("oracle", db.IDAutoIncrement); err == nil {
		t.Fatal("expected dialect error")
	}
	if _, err := db.CreateMediaSQL("sqlite3", "nope"); err == nil {
		t.Fatal("expected strategy error")
	}
}

func TestMigrateAndRepo(t *testing.T) {
	sqlDB, gdb := openSQLite(t)
	ctx := context.Background()
	if err := db.MigrateUp(ctx, sqlDB, "sqlite3", db.IDAutoIncrement); err != nil {
		t.Fatal(err)
	}
	repo, err := db.NewMediaRepo(gdb, db.IDAutoIncrement)
	if err != nil {
		t.Fatal(err)
	}
	m, err := repo.Create(ctx, &db.Media{
		Path: "a/b.png", MIME: "image/png", Size: 10, Hash: "abc",
		CreatedBy: "u1", OwnedBy: "u1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.ID == nil {
		t.Fatal("expected id")
	}
	got, err := repo.FindByID(ctx, m.ID)
	if err != nil || got.Path != "a/b.png" || got.CreatedBy != "u1" || got.OwnedBy != "u1" {
		t.Fatalf("%+v %v", got, err)
	}
	byHash, err := repo.FindByHash(ctx, "abc")
	if err != nil || fmt.Sprint(byHash.ID) != fmt.Sprint(m.ID) {
		t.Fatalf("hash: %+v %v", byHash, err)
	}
	vk := "sm"
	child, err := repo.Create(ctx, &db.Media{
		Path: "a/b-sm.png", MIME: "image/png", Size: 2, Hash: "def",
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
	if err := db.MigrateDown(ctx, sqlDB); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRequiresActor(t *testing.T) {
	sqlDB, gdb := openSQLite(t)
	ctx := context.Background()
	_ = db.MigrateUp(ctx, sqlDB, "sqlite3", db.IDAutoIncrement)
	repo, _ := db.NewMediaRepo(gdb, db.IDAutoIncrement)
	_, err := repo.Create(ctx, &db.Media{Path: "x", MIME: "text/plain", Hash: "h", Size: 1})
	if err == nil {
		t.Fatal("expected actor error")
	}
}
