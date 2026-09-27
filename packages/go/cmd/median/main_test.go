package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/b4moss/median/packages/go/db"
)

func TestMigrateDumpMatchesCreateMediaSQL(t *testing.T) {
	dialects := []string{"sqlite3", "postgres", "mysql"}
	for _, d := range dialects {
		t.Run(d, func(t *testing.T) {
			var out, errBuf bytes.Buffer
			code := run([]string{"migrate", "dump", "--dialect", d}, &out, &errBuf)
			if code != 0 {
				t.Fatalf("exit %d stderr=%q", code, errBuf.String())
			}
			want, err := db.CreateMediaSQL(d, db.IDAutoIncrement, "")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.TrimSpace(out.String())
			want = strings.TrimSpace(want)
			if got != want {
				t.Fatalf("stdout mismatch for %s\ngot:\n%s\nwant:\n%s", d, got, want)
			}
			if !strings.Contains(got, "CREATE TABLE") || !strings.Contains(got, "media") {
				t.Fatalf("expected CREATE TABLE media, got %q", got)
			}
			if !strings.Contains(got, "file_path") || !strings.Contains(got, "file_name") {
				t.Fatalf("expected path columns, got %q", got)
			}
			if strings.Contains(errBuf.String(), "CREATE TABLE") {
				t.Fatalf("SQL leaked to stderr: %q", errBuf.String())
			}
		})
	}
}

func TestMigrateDumpIDStrategyAndTable(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{
		"migrate", "dump",
		"--dialect", "sqlite3",
		"--id-strategy", "uuid_v4",
		"--table", "medias",
	}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, errBuf.String())
	}
	want, err := db.CreateMediaSQL("sqlite3", db.IDUUIDv4, "medias")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != strings.TrimSpace(want) {
		t.Fatalf("mismatch\ngot %s\nwant %s", out.String(), want)
	}
	if !strings.Contains(out.String(), "CREATE TABLE medias") {
		t.Fatalf("expected medias table: %s", out.String())
	}
	if !strings.Contains(out.String(), "TEXT PRIMARY KEY") {
		t.Fatalf("expected text PK: %s", out.String())
	}
}

func TestMigrateDumpErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"no dialect", []string{"migrate", "dump"}},
		{"bad dialect", []string{"migrate", "dump", "--dialect", "oracle"}},
		{"bad id", []string{"migrate", "dump", "--dialect", "sqlite3", "--id-strategy", "nope"}},
		{"bad table", []string{"migrate", "dump", "--dialect", "sqlite3", "--table", "bad-name"}},
		{"unknown top", []string{"foo"}},
		{"migrate only", []string{"migrate"}},
		{"migrate other", []string{"migrate", "other"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errBuf bytes.Buffer
			code := run(tc.args, &out, &errBuf)
			if code == 0 {
				t.Fatalf("expected non-zero exit, stdout=%q", out.String())
			}
			if strings.Contains(out.String(), "CREATE TABLE") {
				t.Fatalf("SQL on stdout for error case: %q", out.String())
			}
			if errBuf.Len() == 0 {
				t.Fatal("expected stderr message")
			}
		})
	}
}
