package core

import (
	"errors"
	"fmt"

	"github.com/b4moss/median/go/internal"
	"github.com/b4moss/median/go/storage"
	"github.com/b4moss/median/go/storage/local"
)

const (
	DefaultMaxSize             int64 = 20 << 20 // 20 MiB
	DefaultMaxConcurrentStores       = 20
	DefaultRandomHexLen              = internal.DefaultRandomHexLen
)

// DriverLocal is the only storage driver in v0.2.0.
const DriverLocal = "local"

// NameMode selects how Store names files.
type NameMode string

const (
	NamePreserve NameMode = "preserve"
	NameRandom   NameMode = "random"
)

// StorageConfig describes one named storage backend.
type StorageConfig struct {
	Driver string
	Path   string // root for local
}

// Config configures a Median instance.
type Config struct {
	Storages            map[string]StorageConfig
	DefaultKey          string
	MIMEAllow           []string
	MIMEDeny            []string
	MaxSize             int64 // 0 => DefaultMaxSize
	MaxConcurrentStores int   // 0 => DefaultMaxConcurrentStores
	FilenameMode        NameMode
	RandomHexLen        int // 0 => DefaultRandomHexLen
	DirLetterCount      *int
	DirNestDepth        *int
}

// Median is the Store/Delete/Get entry point.
type Median struct {
	cfg      Config
	adapters map[string]storage.Adapter
	slots    chan struct{}
}

// New validates cfg and builds a Median instance.
func New(cfg Config) (*Median, error) {
	if len(cfg.Storages) == 0 {
		return nil, errors.New("median: storages is empty")
	}
	if cfg.DefaultKey == "" {
		return nil, errors.New("median: defaultKey is empty")
	}
	if _, ok := cfg.Storages[cfg.DefaultKey]; !ok {
		return nil, fmt.Errorf("median: defaultKey %q not in storages", cfg.DefaultKey)
	}
	if cfg.DirLetterCount == nil {
		return nil, errors.New("median: DirLetterCount is required")
	}
	if cfg.DirNestDepth == nil {
		return nil, errors.New("median: DirNestDepth is required")
	}
	if *cfg.DirLetterCount <= 0 || *cfg.DirNestDepth <= 0 {
		return nil, errors.New("median: DirLetterCount and DirNestDepth must be positive")
	}

	if cfg.MaxSize == 0 {
		cfg.MaxSize = DefaultMaxSize
	}
	if cfg.MaxSize < 0 {
		return nil, errors.New("median: MaxSize must be >= 0")
	}
	if cfg.MaxConcurrentStores == 0 {
		cfg.MaxConcurrentStores = DefaultMaxConcurrentStores
	}
	if cfg.MaxConcurrentStores < 0 {
		return nil, errors.New("median: MaxConcurrentStores must be >= 0")
	}
	if cfg.FilenameMode == "" {
		cfg.FilenameMode = NameRandom
	}
	if cfg.FilenameMode != NamePreserve && cfg.FilenameMode != NameRandom {
		return nil, fmt.Errorf("median: unknown FilenameMode %q", cfg.FilenameMode)
	}
	if cfg.RandomHexLen == 0 {
		cfg.RandomHexLen = DefaultRandomHexLen
	}

	adapters := make(map[string]storage.Adapter, len(cfg.Storages))
	for key, sc := range cfg.Storages {
		switch sc.Driver {
		case DriverLocal:
			fs, err := local.New(sc.Path)
			if err != nil {
				return nil, fmt.Errorf("median: storage %q: %w", key, err)
			}
			adapters[key] = fs
		case "":
			return nil, fmt.Errorf("median: storage %q: driver is empty", key)
		default:
			return nil, fmt.Errorf("median: storage %q: unknown driver %q", key, sc.Driver)
		}
	}

	return &Median{
		cfg:      cfg,
		adapters: adapters,
		slots:    make(chan struct{}, cfg.MaxConcurrentStores),
	}, nil
}

func toInternalMode(m NameMode) internal.NameMode {
	switch m {
	case NamePreserve:
		return internal.NamePreserve
	default:
		return internal.NameRandom
	}
}
