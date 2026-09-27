package local

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/b4moss/median/packages/go/storage"
)

func TestPutGetRoundTrip(t *testing.T) {
	fs, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	data := []byte("hello median")
	if err := fs.Put(ctx, "a/b/c.txt", bytes.NewReader(data), int64(len(data))); err != nil {
		t.Fatal(err)
	}
	rc, size, err := fs.Get(ctx, "a/b/c.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if size != int64(len(data)) || !bytes.Equal(got, data) {
		t.Fatalf("got size=%d data=%q", size, got)
	}
}

func TestPutNestedAndOverwrite(t *testing.T) {
	fs, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := fs.Put(ctx, "x/y/z.bin", bytes.NewReader([]byte("one")), 3); err != nil {
		t.Fatal(err)
	}
	if err := fs.Put(ctx, "x/y/z.bin", bytes.NewReader([]byte("two!")), 4); err != nil {
		t.Fatal(err)
	}
	rc, _, err := fs.Get(ctx, "x/y/z.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	got, _ := io.ReadAll(rc)
	if string(got) != "two!" {
		t.Fatalf("overwrite failed: %q", got)
	}
}

func TestRejectTraversalAndAbsolute(t *testing.T) {
	fs, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, p := range []string{"../x", "a/../../b", "/etc/passwd", ""} {
		if err := fs.Put(ctx, p, bytes.NewReader([]byte("x")), 1); !errors.Is(err, storage.ErrInvalidPath) {
			t.Fatalf("Put(%q): want ErrInvalidPath, got %v", p, err)
		}
		if _, _, err := fs.Get(ctx, p); !errors.Is(err, storage.ErrInvalidPath) {
			t.Fatalf("Get(%q): want ErrInvalidPath, got %v", p, err)
		}
		if _, err := fs.Exists(ctx, p); !errors.Is(err, storage.ErrInvalidPath) {
			t.Fatalf("Exists(%q): want ErrInvalidPath, got %v", p, err)
		}
	}
}

func TestDeleteAndExists(t *testing.T) {
	fs, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := fs.Put(ctx, "f.txt", bytes.NewReader([]byte("z")), 1); err != nil {
		t.Fatal(err)
	}
	ok, err := fs.Exists(ctx, "f.txt")
	if err != nil || !ok {
		t.Fatalf("exists after put: ok=%v err=%v", ok, err)
	}
	if err := fs.Delete(ctx, "f.txt"); err != nil {
		t.Fatal(err)
	}
	ok, err = fs.Exists(ctx, "f.txt")
	if err != nil || ok {
		t.Fatalf("exists after delete: ok=%v err=%v", ok, err)
	}
	if _, _, err := fs.Get(ctx, "f.txt"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("get missing: %v", err)
	}
	if err := fs.Delete(ctx, "f.txt"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("delete missing: %v", err)
	}
	ok, err = fs.Exists(ctx, "never.txt")
	if err != nil || ok {
		t.Fatalf("exists never: ok=%v err=%v", ok, err)
	}
}

func TestNewEmptyRoot(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("expected error")
	}
	if _, err := New("   "); err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveStaysInRoot(t *testing.T) {
	root := t.TempDir()
	fs, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	full, err := fs.resolve("ok/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(full, filepath.Clean(root)+string(filepath.Separator)) && full != filepath.Join(root, "ok", "file.txt") {
		// still under root
		rel, relErr := filepath.Rel(root, full)
		if relErr != nil || strings.HasPrefix(rel, "..") {
			t.Fatalf("escaped root: %s", full)
		}
	}
}
