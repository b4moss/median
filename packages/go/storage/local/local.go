package local

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/b4moss/median/packages/go/storage"
)

// FS stores objects on the local filesystem under Root.
type FS struct {
	Root string
}

// New returns a Local FS adapter rooted at root.
func New(root string) (*FS, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("local: root path is empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, err
	}
	return &FS{Root: abs}, nil
}

func (f *FS) resolve(relPath string) (string, error) {
	clean, err := cleanRel(relPath)
	if err != nil {
		return "", err
	}
	full := filepath.Join(f.Root, filepath.FromSlash(clean))
	rel, err := filepath.Rel(f.Root, full)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return "", storage.ErrInvalidPath
	}
	return full, nil
}

func cleanRel(relPath string) (string, error) {
	if relPath == "" {
		return "", storage.ErrInvalidPath
	}
	p := filepath.ToSlash(relPath)
	if strings.HasPrefix(p, "/") || filepath.IsAbs(relPath) {
		return "", storage.ErrInvalidPath
	}
	if strings.Contains(p, "\x00") {
		return "", storage.ErrInvalidPath
	}
	parts := strings.Split(p, "/")
	for _, part := range parts {
		if part == ".." {
			return "", storage.ErrInvalidPath
		}
	}
	clean := pathCleanSlash(p)
	if clean == "." || clean == "" {
		return "", storage.ErrInvalidPath
	}
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return "", storage.ErrInvalidPath
	}
	return clean, nil
}

func pathCleanSlash(p string) string {
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(p)))
}

// Put writes r to relPath under the root, creating parent directories as needed.
func (f *FS) Put(ctx context.Context, relPath string, r io.Reader, size int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	full, err := f.resolve(relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(full), ".median-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	var written int64
	if size >= 0 {
		written, err = io.CopyN(tmp, r, size)
		if err != nil && !errors.Is(err, io.EOF) {
			_ = tmp.Close()
			return err
		}
		if size > 0 && written != size && err == nil {
			// CopyN returns EOF when fewer bytes; treat short read as error when size known and positive.
			if written < size {
				_ = tmp.Close()
				return io.ErrUnexpectedEOF
			}
		}
	} else {
		written, err = io.Copy(tmp, r)
		if err != nil {
			_ = tmp.Close()
			return err
		}
	}
	_ = written
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.Rename(tmpName, full)
}

// Get opens relPath for reading.
func (f *FS) Get(ctx context.Context, relPath string) (io.ReadCloser, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	full, err := f.resolve(relPath)
	if err != nil {
		return nil, 0, err
	}
	st, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, storage.ErrNotFound
		}
		return nil, 0, err
	}
	if st.IsDir() {
		return nil, 0, storage.ErrNotFound
	}
	file, err := os.Open(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, storage.ErrNotFound
		}
		return nil, 0, err
	}
	return file, st.Size(), nil
}

// Delete removes relPath. Missing paths return ErrNotFound.
func (f *FS) Delete(ctx context.Context, relPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	full, err := f.resolve(relPath)
	if err != nil {
		return err
	}
	err = os.Remove(full)
	if err != nil {
		if os.IsNotExist(err) {
			return storage.ErrNotFound
		}
		return err
	}
	return nil
}

// Exists reports whether relPath exists as a file.
func (f *FS) Exists(ctx context.Context, relPath string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	full, err := f.resolve(relPath)
	if err != nil {
		return false, err
	}
	st, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if st.IsDir() {
		return false, nil
	}
	return true, nil
}
