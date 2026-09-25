package storage

import (
	"context"
	"errors"
	"io"
)

// ErrNotFound indicates the relative path does not exist.
var ErrNotFound = errors.New("storage: not found")

// ErrInvalidPath indicates a relative path is empty, absolute, or escapes the root.
var ErrInvalidPath = errors.New("storage: invalid path")

// Adapter writes and reads objects under a storage root using relative paths.
type Adapter interface {
	Put(ctx context.Context, relPath string, r io.Reader, size int64) error
	Get(ctx context.Context, relPath string) (io.ReadCloser, int64, error)
	Delete(ctx context.Context, relPath string) error
	Exists(ctx context.Context, relPath string) (bool, error)
}
