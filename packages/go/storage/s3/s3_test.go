package s3store_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/b4moss/median/go/storage"
	s3store "github.com/b4moss/median/go/storage/s3"
)

type memS3 struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newMem() *memS3 { return &memS3{data: map[string][]byte{}} }

func (m *memS3) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	b, _ := io.ReadAll(params.Body)
	m.mu.Lock()
	m.data[aws.ToString(params.Key)] = b
	m.mu.Unlock()
	return &s3.PutObjectOutput{}, nil
}

func (m *memS3) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	m.mu.Lock()
	b, ok := m.data[aws.ToString(params.Key)]
	m.mu.Unlock()
	if !ok {
		return nil, errors.New("NoSuchKey")
	}
	cl := int64(len(b))
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(b)), ContentLength: &cl}, nil
}

func (m *memS3) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.mu.Lock()
	delete(m.data, aws.ToString(params.Key))
	m.mu.Unlock()
	return &s3.DeleteObjectOutput{}, nil
}

func (m *memS3) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	m.mu.Lock()
	_, ok := m.data[aws.ToString(params.Key)]
	m.mu.Unlock()
	if !ok {
		return nil, errors.New("NotFound")
	}
	return &s3.HeadObjectOutput{}, nil
}

type fakePresign struct{}

func (fakePresign) PresignGET(ctx context.Context, bucket, key string, ttl time.Duration) (string, error) {
	return "https://example.test/" + bucket + "/" + key + "?ttl=" + ttl.String(), nil
}

func TestS3RoundTripAndPresign(t *testing.T) {
	fs, err := s3store.NewFromClient(s3store.Config{Bucket: "b", Prefix: "p"}, newMem(), fakePresign{})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	data := []byte("hello-s3")
	if err := fs.Put(ctx, "a/b.txt", bytes.NewReader(data), int64(len(data))); err != nil {
		t.Fatal(err)
	}
	rc, size, err := fs.Get(ctx, "a/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	got, _ := io.ReadAll(rc)
	if size != int64(len(data)) || !bytes.Equal(got, data) {
		t.Fatalf("%q %d", got, size)
	}
	ok, err := fs.Exists(ctx, "a/b.txt")
	if err != nil || !ok {
		t.Fatal(err)
	}
	url, err := fs.PresignGet(ctx, "a/b.txt", time.Hour)
	if err != nil || !strings.Contains(url, "ttl=") {
		t.Fatalf("%s %v", url, err)
	}
	if err := fs.Delete(ctx, "a/b.txt"); err != nil {
		t.Fatal(err)
	}
	ok, _ = fs.Exists(ctx, "a/b.txt")
	if ok {
		t.Fatal("still exists")
	}
	if err := fs.Put(ctx, "../x", bytes.NewReader(data), int64(len(data))); !errors.Is(err, storage.ErrInvalidPath) {
		t.Fatalf("%v", err)
	}
	if _, err := s3store.NewFromClient(s3store.Config{}, newMem(), nil); err == nil {
		t.Fatal("bucket required")
	}
}
