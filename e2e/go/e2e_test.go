package e2e_test

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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/b4moss/median/go/core"
	"github.com/b4moss/median/go/db"
	"github.com/b4moss/median/go/pipeline"
	"github.com/b4moss/median/go/storage"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func mustLocalRoot(t *testing.T) string {
	t.Helper()
	root := envOr("MEDIAN_E2E_LOCAL_ROOT", "")
	if root == "" {
		t.Fatal("MEDIAN_E2E_LOCAL_ROOT is required (Docker POSIX volume mount)")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	// Isolate per test under the shared volume.
	dir := filepath.Join(root, t.Name()+"-"+strconv.FormatInt(time.Now().UnixNano(), 36))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func s3Endpoint() string {
	return envOr("MEDIAN_E2E_S3_ENDPOINT", "http://127.0.0.1:9000")
}

func s3Creds() (access, secret, bucket, region string) {
	return envOr("MEDIAN_E2E_S3_ACCESS_KEY", "median_e2e"),
		envOr("MEDIAN_E2E_S3_SECRET_KEY", "median_e2e_secret"),
		envOr("MEDIAN_E2E_S3_BUCKET", "median-e2e"),
		envOr("MEDIAN_E2E_S3_REGION", "us-east-1")
}

func intPtr(v int) *int { return &v }

func ensureBucket(t *testing.T) {
	t.Helper()
	access, secret, bucket, region := s3Creds()
	client := s3.NewFromConfig(aws.Config{
		Region:      region,
		Credentials: credentials.NewStaticCredentialsProvider(access, secret, ""),
	}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3Endpoint())
		o.UsePathStyle = true
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return
	}
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		t.Fatalf("create bucket %s: %v", bucket, err)
	}
}

func waitRustFS(t *testing.T) {
	t.Helper()
	endpoint := strings.TrimRight(s3Endpoint(), "/")
	deadline := time.Now().Add(60 * time.Second)
	var last error
	for time.Now().Before(deadline) {
		resp, err := http.Get(endpoint + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return
			}
			last = fmt.Errorf("status %d", resp.StatusCode)
		} else {
			last = err
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("rustfs not healthy at %s: %v", endpoint, last)
}

func openMemDB(t *testing.T) (*gorm.DB, *db.MediaRepo) {
	t.Helper()
	dsn := fmt.Sprintf("file:median_e2e_%d?mode=memory&cache=shared", time.Now().UnixNano())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.MigrateUp(context.Background(), sqlDB, "sqlite3", db.IDAutoIncrement); err != nil {
		t.Fatal(err)
	}
	repo, err := db.NewMediaRepo(gdb, db.IDAutoIncrement)
	if err != nil {
		t.Fatal(err)
	}
	return gdb, repo
}

func localConfig(root string) core.Config {
	return core.Config{
		Storages: map[string]core.StorageConfig{
			"local": {Driver: core.DriverLocal, Path: root},
		},
		DefaultKey:     "local",
		DirLetterCount: intPtr(1),
		DirNestDepth:   intPtr(2),
	}
}

func s3Config() core.Config {
	access, secret, bucket, region := s3Creds()
	return core.Config{
		Storages: map[string]core.StorageConfig{
			"s3": {
				Driver:         core.DriverS3,
				Bucket:         bucket,
				Region:         region,
				Endpoint:       s3Endpoint(),
				ForcePathStyle: true,
				AccessKey:      access,
				SecretKey:      secret,
				Prefix:         fmt.Sprintf("e2e/%d", time.Now().UnixNano()),
			},
		},
		DefaultKey:     "s3",
		DirLetterCount: intPtr(1),
		DirNestDepth:   intPtr(2),
	}
}

func makePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{R: 200, G: 40, B: 40, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func hashOf(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func httpGetBody(t *testing.T, url string) []byte {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("presign GET status %d body=%q", resp.StatusCode, string(body[:min(len(body), 200)]))
	}
	return body
}

func TestL1_LocalNoDB(t *testing.T) {
	root := mustLocalRoot(t)
	m, err := core.New(localConfig(root))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payload := []byte("hello-median-e2e-l1")
	want := hashOf(payload)
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "application/octet-stream", Filename: "l1.bin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Hash != want || res.Size != int64(len(payload)) {
		t.Fatalf("store meta: %+v", res)
	}
	got, err := m.Get(ctx, res.Path, core.GetOptions{WithBody: true})
	if err != nil || !bytes.Equal(got.Body, payload) {
		t.Fatalf("get: %v", err)
	}
	if err := m.Delete(ctx, res.Path, core.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(ctx, res.Path, core.GetOptions{}); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestS1_RustFSNoDB(t *testing.T) {
	waitRustFS(t)
	ensureBucket(t)
	m, err := core.New(s3Config())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	payload := []byte("hello-median-e2e-s1")
	want := hashOf(payload)
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "application/octet-stream", Filename: "s1.bin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Hash != want {
		t.Fatalf("hash %s want %s", res.Hash, want)
	}
	got, err := m.Get(ctx, res.Path, core.GetOptions{WithBody: true})
	if err != nil || !bytes.Equal(got.Body, payload) {
		t.Fatalf("get: %v", err)
	}
	url, err := m.PresignGet(ctx, res.Path, res.StorageKey, 0)
	if err != nil || url == "" {
		t.Fatal(err)
	}
	if body := httpGetBody(t, url); !bytes.Equal(body, payload) {
		t.Fatalf("presign body mismatch")
	}
	if err := m.Delete(ctx, res.Path, core.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(ctx, res.Path, core.GetOptions{}); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func runCRUD(t *testing.T, cfg core.Config, withPresign bool) {
	t.Helper()
	gdb, repo := openMemDB(t)
	cfg.DB = &core.DBConfig{Gorm: gdb, IDStrategy: db.IDAutoIncrement}
	cfg.DefaultActor = "e2e-actor"
	m, err := core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a := []byte("crud-fixture-A")
	b := []byte("crud-fixture-B-different")

	res, err := m.Store(ctx, bytes.NewReader(a), int64(len(a)), core.StoreOptions{
		MIME: "application/octet-stream", Filename: "a.bin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ID == nil {
		t.Fatal("expected id")
	}
	row, err := repo.FindByID(ctx, res.ID)
	if err != nil {
		t.Fatal(err)
	}
	if row.Path != res.Path || row.Hash != res.Hash || row.Size != res.Size {
		t.Fatalf("db row mismatch: %+v vs %+v", row, res)
	}
	if row.CreatedBy != "e2e-actor" || row.OwnedBy != "e2e-actor" {
		t.Fatalf("actor: %+v", row)
	}

	got, err := m.Get(ctx, "", core.GetOptions{ID: res.ID, WithBody: true})
	if err != nil || !bytes.Equal(got.Body, a) {
		t.Fatalf("get: %v", err)
	}

	if withPresign {
		url, err := m.PresignGet(ctx, res.Path, res.StorageKey, 0)
		if err != nil {
			t.Fatal(err)
		}
		if body := httpGetBody(t, url); !bytes.Equal(body, a) {
			t.Fatal("presign mismatch")
		}
	}

	res2, err := m.Store(ctx, bytes.NewReader(a), int64(len(a)), core.StoreOptions{
		MIME: "application/octet-stream", Filename: "a2.bin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(res2.ID) != fmt.Sprint(res.ID) {
		t.Fatalf("duplicate hash should return same id: %v vs %v", res2.ID, res.ID)
	}
	row2, err := repo.FindByID(ctx, res.ID)
	if err != nil || row2.Hash != res.Hash || row2.Path != res.Path {
		t.Fatalf("meta corrupted after re-store: %v %+v", err, row2)
	}

	if err := m.Delete(ctx, "", core.DeleteOptions{ID: res.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByID(ctx, res.ID); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("old id should be gone: %v", err)
	}

	resB, err := m.Store(ctx, bytes.NewReader(b), int64(len(b)), core.StoreOptions{
		MIME: "application/octet-stream", Filename: "b.bin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(resB.ID) == fmt.Sprint(res.ID) {
		t.Fatal("expected new id after replace")
	}
	rowB, err := repo.FindByID(ctx, resB.ID)
	if err != nil || rowB.Hash != hashOf(b) {
		t.Fatalf("new row: %v %+v", err, rowB)
	}
	gotB, err := m.Get(ctx, "", core.GetOptions{ID: resB.ID, WithBody: true})
	if err != nil || !bytes.Equal(gotB.Body, b) {
		t.Fatalf("get B: %v", err)
	}

	if err := m.Delete(ctx, "", core.DeleteOptions{ID: resB.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(ctx, "", core.GetOptions{ID: resB.ID}); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected db not found: %v", err)
	}
	if _, err := repo.FindByID(ctx, resB.ID); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("expected repo not found: %v", err)
	}
}

func TestCL_LocalCRUD(t *testing.T) {
	root := mustLocalRoot(t)
	runCRUD(t, localConfig(root), false)
}

func TestCS_RustFSCRUD(t *testing.T) {
	waitRustFS(t)
	ensureBucket(t)
	runCRUD(t, s3Config(), true)
}

func runThumb(t *testing.T, cfg core.Config) {
	t.Helper()
	gdb, repo := openMemDB(t)
	cfg.DB = &core.DBConfig{Gorm: gdb, IDStrategy: db.IDAutoIncrement}
	cfg.DefaultActor = "e2e-actor"
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
	payload := makePNG(t, 64, 64)
	res, err := m.Store(ctx, bytes.NewReader(payload), int64(len(payload)), core.StoreOptions{
		MIME: "image/png", Filename: "pic.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ID == nil || len(res.Variants) != 1 || res.Variants[0].Key != "sm" {
		t.Fatalf("variants: %+v", res)
	}
	v := res.Variants[0]

	parent, err := m.Get(ctx, "", core.GetOptions{ID: res.ID, WithBody: true})
	if err != nil || int64(len(parent.Body)) != parent.Size {
		t.Fatalf("parent get: %v size=%d len=%d", err, parent.Size, len(parent.Body))
	}
	childGet, err := m.Get(ctx, "", core.GetOptions{ID: v.ID, WithBody: true})
	if err != nil || int64(len(childGet.Body)) != childGet.Size || childGet.Size != v.Size {
		t.Fatalf("variant get: %v %+v", err, v)
	}

	prow, err := repo.FindByID(ctx, res.ID)
	if err != nil {
		t.Fatal(err)
	}
	if prow.OriginalID != nil || prow.VariantKey != nil {
		t.Fatalf("parent should not be variant: %+v", prow)
	}
	children, err := repo.ListChildren(ctx, res.ID)
	if err != nil || len(children) != 1 {
		t.Fatalf("children: %v %d", err, len(children))
	}
	ch := children[0]
	if ch.VariantKey == nil || *ch.VariantKey != "sm" || ch.Path != v.Path || ch.Hash != v.Hash {
		t.Fatalf("child row: %+v vs %+v", ch, v)
	}

	if err := m.Delete(ctx, "", core.DeleteOptions{ID: res.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByID(ctx, res.ID); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("parent db: %v", err)
	}
	if _, err := repo.FindByID(ctx, v.ID); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("child db: %v", err)
	}
	if _, err := m.Get(ctx, "", core.GetOptions{ID: res.ID}); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("parent get: %v", err)
	}
}

func TestTL_LocalThumbnails(t *testing.T) {
	root := mustLocalRoot(t)
	runThumb(t, localConfig(root))
}

func TestTS_RustFSThumbnails(t *testing.T) {
	waitRustFS(t)
	ensureBucket(t)
	runThumb(t, s3Config())
}
