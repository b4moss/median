package core_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/b4moss/median/go/core"
	"github.com/b4moss/median/go/storage"
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
	m, err := core.New(testConfig(t, root))
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Fatal("nil median")
	}

	cfg := testConfig(t, root)
	cfg.Storages["alt"] = core.StorageConfig{Driver: core.DriverLocal, Path: t.TempDir()}
	if _, err := core.New(cfg); err != nil {
		t.Fatal(err)
	}

	bad := testConfig(t, root)
	bad.DefaultKey = "missing"
	if _, err := core.New(bad); err == nil {
		t.Fatal("defaultKey")
	}
	bad = testConfig(t, root)
	bad.Storages = nil
	if _, err := core.New(bad); err == nil {
		t.Fatal("empty storages")
	}
	bad = testConfig(t, root)
	bad.DirLetterCount = nil
	if _, err := core.New(bad); err == nil {
		t.Fatal("DirLetterCount")
	}
	bad = testConfig(t, root)
	bad.DirNestDepth = nil
	if _, err := core.New(bad); err == nil {
		t.Fatal("DirNestDepth")
	}
	bad = testConfig(t, root)
	bad.Storages["local"] = core.StorageConfig{Driver: core.DriverLocal, Path: ""}
	if _, err := core.New(bad); err == nil {
		t.Fatal("empty root")
	}
	bad = testConfig(t, root)
	bad.Storages["local"] = core.StorageConfig{Driver: "s3", Path: root}
	if _, err := core.New(bad); err == nil {
		t.Fatal("unknown driver")
	}
}

func TestStoreDeleteGet(t *testing.T) {
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
		MIME:     "text/plain",
		Filename: "note.txt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.StorageKey != "local" || res.MIME != "text/plain" || res.Size != int64(len(payload)) || res.Hash != wantHash {
		t.Fatalf("result: %+v", res)
	}
	if res.Path == "" {
		t.Fatal("empty path")
	}
	abs := filepath.Join(root, filepath.FromSlash(res.Path))
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("file missing: %v", err)
	}

	meta, err := m.Get(ctx, res.Path, core.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if meta.Path != res.Path || meta.Size != res.Size || meta.StorageKey != "local" || meta.Body != nil {
		t.Fatalf("meta: %+v", meta)
	}
	full, err := m.Get(ctx, res.Path, core.GetOptions{WithBody: true})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(full.Body, payload) || full.Hash != wantHash {
		t.Fatalf("body/hash: %q %s", full.Body, full.Hash)
	}

	if err := m.Delete(ctx, res.Path, core.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(ctx, res.Path, core.GetOptions{}); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("after delete: %v", err)
	}
}

func TestStoreExplicitKeyAndPreserve(t *testing.T) {
	rootA := t.TempDir()
	rootB := t.TempDir()
	cfg := core.Config{
		Storages: map[string]core.StorageConfig{
			"a": {Driver: core.DriverLocal, Path: rootA},
			"b": {Driver: core.DriverLocal, Path: rootB},
		},
		DefaultKey:     "a",
		DirLetterCount: intPtr(1),
		DirNestDepth:   intPtr(2),
		FilenameMode:   core.NamePreserve,
	}
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	mode := core.NamePreserve
	res, err := m.Store(context.Background(), bytes.NewReader([]byte("x")), 1, core.StoreOptions{
		MIME:         "application/octet-stream",
		Filename:     "keep-me.bin",
		StorageKey:   "b",
		FilenameMode: &mode,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.StorageKey != "b" {
		t.Fatalf("key=%s", res.StorageKey)
	}
	if filepath.Base(res.Path) != "keep-me.bin" {
		t.Fatalf("path=%s", res.Path)
	}
	if _, err := os.Stat(filepath.Join(rootB, filepath.FromSlash(res.Path))); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(rootA, filepath.FromSlash(res.Path))); !os.IsNotExist(err) {
		t.Fatal("should not write to default root")
	}
}

func TestStoreMIMEAndSize(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)
	cfg.MIMEDeny = []string{"text/html"}
	cfg.MaxSize = 4
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := m.Store(ctx, bytes.NewReader([]byte("abcd")), 4, core.StoreOptions{MIME: "text/html", Filename: "a.txt"}); err == nil {
		t.Fatal("deny mime")
	}
	if _, err := m.Store(ctx, bytes.NewReader([]byte("abcde")), 5, core.StoreOptions{MIME: "text/plain", Filename: "a.txt"}); !errors.Is(err, core.ErrSizeExceeded) {
		t.Fatalf("size: %v", err)
	}
	if _, err := m.Store(ctx, bytes.NewReader(nil), 0, core.StoreOptions{MIME: "text/plain", Filename: "empty.txt"}); err != nil {
		t.Fatalf("empty file: %v", err)
	}
	if _, err := m.Store(ctx, bytes.NewReader([]byte("x")), -1, core.StoreOptions{MIME: "text/plain"}); !errors.Is(err, core.ErrNegativeSize) {
		t.Fatalf("neg: %v", err)
	}
	if _, err := m.Store(ctx, bytes.NewReader([]byte("x")), 1, core.StoreOptions{MIME: "", Filename: "a"}); err == nil {
		t.Fatal("empty mime")
	}
	if _, err := m.Store(ctx, bytes.NewReader([]byte("x")), 1, core.StoreOptions{MIME: "text/plain", StorageKey: "nope"}); !errors.Is(err, core.ErrUnknownKey) {
		t.Fatalf("key: %v", err)
	}
}

func TestDeleteGetValidation(t *testing.T) {
	m, err := core.New(testConfig(t, t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := m.Delete(ctx, "", core.DeleteOptions{}); !errors.Is(err, core.ErrEmptyPath) {
		t.Fatalf("%v", err)
	}
	if _, err := m.Get(ctx, "", core.GetOptions{}); !errors.Is(err, core.ErrEmptyPath) {
		t.Fatalf("%v", err)
	}
	if err := m.Delete(ctx, "x", core.DeleteOptions{StorageKey: "nope"}); !errors.Is(err, core.ErrUnknownKey) {
		t.Fatalf("%v", err)
	}
}

func TestConcurrentStoreLimit(t *testing.T) {
	root := t.TempDir()
	cfg := testConfig(t, root)
	cfg.MaxConcurrentStores = 2
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	started := make(chan struct{}, 3)
	release := make(chan struct{})
	var wg sync.WaitGroup
	run := func() {
		defer wg.Done()
		ctx := context.Background()
		// Block inside Store by using a slow reader after acquire — simpler: hold slots via blocking on release after started.
		pr, pw := io.Pipe()
		go func() {
			started <- struct{}{}
			<-release
			_, _ = pw.Write([]byte("ab"))
			_ = pw.Close()
		}()
		_, err := m.Store(ctx, pr, 2, core.StoreOptions{MIME: "text/plain", Filename: "c.bin"})
		if err != nil {
			t.Errorf("store: %v", err)
		}
	}
	wg.Add(2)
	go run()
	go run()
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting start")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		_, err := m.Store(ctx, bytes.NewReader([]byte("zz")), 2, core.StoreOptions{MIME: "text/plain", Filename: "d.bin"})
		errCh <- err
	}()
	select {
	case err := <-errCh:
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			// may be deadline
			if err == nil {
				t.Fatal("third store should wait/fail")
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("third store did not return")
	}
	close(release)
	wg.Wait()
}

func TestDecodeBase64AndPart(t *testing.T) {
	raw := []byte("hi")
	enc := "aGk="
	data, mime, err := core.DecodeBase64(enc)
	if err != nil || mime != "" || !bytes.Equal(data, raw) {
		t.Fatalf("%q %q %v", data, mime, err)
	}
	data, mime, err = core.DecodeBase64("data:text/plain;base64,aGk=")
	if err != nil || mime != "text/plain" || !bytes.Equal(data, raw) {
		t.Fatalf("data url: %q %q %v", data, mime, err)
	}
	if _, _, err := core.DecodeBase64(""); err == nil {
		t.Fatal("empty")
	}
	if _, _, err := core.DecodeBase64("data:text/plain,aGk="); err == nil {
		t.Fatal("non-base64 data url")
	}
	if _, _, err := core.DecodeBase64("!!!!"); err == nil {
		t.Fatal("bad b64")
	}

	r, size, opt, err := core.PartToStoreOptions(core.MultipartPart{
		Filename: "p.txt",
		MIME:     "text/plain",
		Reader:   bytes.NewReader(raw),
		Size:     int64(len(raw)),
	})
	if err != nil || size != 2 || opt.Filename != "p.txt" || opt.MIME != "text/plain" {
		t.Fatalf("%v %d %+v", err, size, opt)
	}
	m, err := core.New(testConfig(t, t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	res, err := m.Store(context.Background(), r, size, opt)
	if err != nil {
		t.Fatal(err)
	}
	if res.Size != 2 {
		t.Fatalf("%+v", res)
	}
	if _, _, _, err := core.PartToStoreOptions(core.MultipartPart{MIME: "text/plain", Size: 1}); err == nil {
		t.Fatal("nil reader")
	}
	if _, _, _, err := core.PartToStoreOptions(core.MultipartPart{Reader: bytes.NewReader(raw), Size: 2}); err == nil {
		t.Fatal("mime required")
	}
}
