package core

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/b4moss/median/go/internal"
	"github.com/b4moss/median/go/storage"
)

// StoreResult is returned by Store (v0.2.0 fields).
type StoreResult struct {
	Path       string
	MIME       string
	Size       int64
	Hash       string
	StorageKey string
}

// StoreOptions controls a Store call.
type StoreOptions struct {
	MIME         string
	Filename     string
	StorageKey   string
	FilenameMode *NameMode
	RandomHexLen int
	Actor        string // ignored without DB
}

// GetOptions controls a Get call.
type GetOptions struct {
	StorageKey string
	WithBody   bool
}

// GetResult is returned by Get.
type GetResult struct {
	Path       string
	MIME       string
	Size       int64
	Hash       string
	StorageKey string
	Body       []byte // set when WithBody
}

// DeleteOptions controls a Delete call.
type DeleteOptions struct {
	StorageKey string
}

var (
	ErrNegativeSize = errors.New("median: negative size")
	ErrSizeExceeded = errors.New("median: size exceeds MaxSize")
	ErrEmptyPath    = errors.New("median: path is empty")
	ErrUnknownKey   = errors.New("median: unknown storage key")
)

// Store saves content under a resolved storage path.
func (m *Median) Store(ctx context.Context, r io.Reader, size int64, opt StoreOptions) (*StoreResult, error) {
	if size < 0 {
		return nil, ErrNegativeSize
	}
	if err := m.checkSize(size); err != nil {
		return nil, err
	}
	if err := internal.CheckMIME(opt.MIME, m.cfg.MIMEAllow, m.cfg.MIMEDeny); err != nil {
		return nil, err
	}

	if err := m.acquire(ctx); err != nil {
		return nil, err
	}
	defer m.release()

	key, ad, err := m.resolveKey(opt.StorageKey)
	if err != nil {
		return nil, err
	}

	mode := toInternalMode(m.cfg.FilenameMode)
	if opt.FilenameMode != nil {
		mode = toInternalMode(*opt.FilenameMode)
	}
	randLen := m.cfg.RandomHexLen
	if opt.RandomHexLen != 0 {
		randLen = opt.RandomHexLen
	}
	baseName, err := internal.ResolveFilename(opt.Filename, mode, randLen)
	if err != nil {
		return nil, err
	}
	relPath, err := internal.BuildStoragePath(baseName, *m.cfg.DirLetterCount, *m.cfg.DirNestDepth)
	if err != nil {
		return nil, err
	}

	// Hash while buffering for Put (size known).
	data, err := readExact(r, size)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])

	if err := ad.Put(ctx, relPath, bytes.NewReader(data), size); err != nil {
		return nil, err
	}

	return &StoreResult{
		Path:       relPath,
		MIME:       opt.MIME,
		Size:       size,
		Hash:       hash,
		StorageKey: key,
	}, nil
}

// Delete removes an object by path (no DB).
func (m *Median) Delete(ctx context.Context, objectPath string, opt DeleteOptions) error {
	if objectPath == "" {
		return ErrEmptyPath
	}
	_, ad, err := m.resolveKey(opt.StorageKey)
	if err != nil {
		return err
	}
	return ad.Delete(ctx, objectPath)
}

// Get loads metadata and optionally body by path (no DB).
func (m *Median) Get(ctx context.Context, objectPath string, opt GetOptions) (*GetResult, error) {
	if objectPath == "" {
		return nil, ErrEmptyPath
	}
	key, ad, err := m.resolveKey(opt.StorageKey)
	if err != nil {
		return nil, err
	}
	rc, size, err := ad.Get(ctx, objectPath)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	out := &GetResult{
		Path:       objectPath,
		Size:       size,
		StorageKey: key,
		MIME:       "", // not persisted without DB
		Hash:       "",
	}
	if opt.WithBody {
		body, err := io.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		out.Body = body
		sum := sha256.Sum256(body)
		out.Hash = hex.EncodeToString(sum[:])
	}
	return out, nil
}

func (m *Median) checkSize(size int64) error {
	if size > m.cfg.MaxSize {
		return ErrSizeExceeded
	}
	return nil
}

func (m *Median) resolveKey(key string) (string, storage.Adapter, error) {
	if key == "" {
		key = m.cfg.DefaultKey
	}
	ad, ok := m.adapters[key]
	if !ok {
		return "", nil, fmt.Errorf("%w: %s", ErrUnknownKey, key)
	}
	return key, ad, nil
}

func (m *Median) acquire(ctx context.Context) error {
	select {
	case m.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *Median) release() {
	select {
	case <-m.slots:
	default:
	}
}

func readExact(r io.Reader, size int64) ([]byte, error) {
	if size == 0 {
		// Allow empty; drain nothing required.
		buf := make([]byte, 0)
		// Ensure reader is not required to provide data.
		return buf, nil
	}
	buf := make([]byte, size)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// BaseName is a small helper for tests/helpers.
func BaseName(p string) string {
	return path.Base(p)
}
