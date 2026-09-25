package internal_test

import (
	"path"
	"strings"
	"testing"

	"github.com/b4moss/median/go/internal"
)

func TestSanitizePreserve(t *testing.T) {
	got, err := internal.SanitizePreserve("hello.txt")
	if err != nil || got != "hello.txt" {
		t.Fatalf("got %q err=%v", got, err)
	}
	got, err = internal.SanitizePreserve("写真.png")
	if err != nil || got != "写真.png" {
		t.Fatalf("unicode: %q err=%v", got, err)
	}
	got, err = internal.SanitizePreserve("  .foo.  ")
	if err != nil || got != "foo" {
		t.Fatalf("trim: %q err=%v", got, err)
	}
	got, err = internal.SanitizePreserve("a/b\\c.txt")
	if err != nil || strings.ContainsAny(got, `/\`) {
		t.Fatalf("sep: %q err=%v", got, err)
	}
	if _, err := internal.SanitizePreserve("..."); err == nil {
		t.Fatal("expected empty error")
	}
	if _, err := internal.SanitizePreserve("\x00"); err == nil {
		t.Fatal("expected empty error for nul")
	}
}

func TestRandomFilename(t *testing.T) {
	got, err := internal.RandomFilename("img.PNG", 32)
	if err != nil {
		t.Fatal(err)
	}
	if path.Ext(got) != ".PNG" {
		t.Fatalf("ext: %q", got)
	}
	base := strings.TrimSuffix(got, ".PNG")
	if len(base) != 32 {
		t.Fatalf("len=%d got=%q", len(base), got)
	}
	for _, c := range base {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Fatalf("non-hex: %q", got)
		}
	}
	if _, err := internal.RandomFilename("x", 0); err == nil {
		t.Fatal("expected len error")
	}
	if _, err := internal.RandomFilename("x", 200); err == nil {
		t.Fatal("expected max len error")
	}
	if _, err := internal.RandomFilename("x", 31); err == nil {
		t.Fatal("expected odd len error")
	}
}

func TestBuildStoragePath(t *testing.T) {
	p, err := internal.BuildStoragePath("abc1234.jpg", 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(p, "/") {
		t.Fatalf("leading slash: %q", p)
	}
	if !strings.HasSuffix(p, "abc1234.jpg") {
		t.Fatalf("suffix: %q", p)
	}
	p2, err := internal.BuildStoragePath("abc1234.jpg", 1, 4)
	if err != nil || p2 != p {
		t.Fatalf("deterministic: %q vs %q", p, p2)
	}
	if _, err := internal.BuildStoragePath("", 1, 4); err == nil {
		t.Fatal("expected empty name error")
	}
}

func TestCheckMIME(t *testing.T) {
	if err := internal.CheckMIME("image/png", nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := internal.CheckMIME("image/png", []string{"image/png"}, nil); err != nil {
		t.Fatal(err)
	}
	if err := internal.CheckMIME("image/png", nil, []string{"application/x-msdownload"}); err != nil {
		t.Fatal(err)
	}
	if err := internal.CheckMIME("text/html", nil, []string{"text/html"}); err == nil {
		t.Fatal("deny")
	}
	if err := internal.CheckMIME("image/gif", []string{"image/png"}, nil); err == nil {
		t.Fatal("allow miss")
	}
	if err := internal.CheckMIME("", nil, nil); err == nil {
		t.Fatal("empty")
	}
}
