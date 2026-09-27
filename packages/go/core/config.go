package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/b4moss/median/packages/go/db"
	"github.com/b4moss/median/packages/go/internal"
	"github.com/b4moss/median/packages/go/pipeline"
	"github.com/b4moss/median/packages/go/storage"
	"github.com/b4moss/median/packages/go/storage/local"
	s3store "github.com/b4moss/median/packages/go/storage/s3"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	DefaultMaxSize             int64 = 20 << 20
	DefaultMaxConcurrentStores       = 20
	DefaultRandomHexLen              = internal.DefaultRandomHexLen
	DefaultPresignTTL                = time.Hour
)

const (
	DriverLocal = "local"
	DriverS3    = s3store.Driver
)

type NameMode string

const (
	NamePreserve NameMode = "preserve"
	NameRandom   NameMode = "random"
)

type StorageConfig struct {
	Driver         string
	Path           string
	Bucket         string
	Region         string
	Endpoint       string
	ForcePathStyle bool
	Prefix         string
	AccessKey      string
	SecretKey      string
}

type DBConfig struct {
	Gorm       *gorm.DB
	DSN        string
	IDStrategy string
	TableName  string
	Columns    db.ColumnMap
}

type Config struct {
	Storages            map[string]StorageConfig
	DefaultKey          string
	MIMEAllow           []string
	MIMEDeny            []string
	MaxSize             int64
	MaxConcurrentStores int
	FilenameMode        NameMode
	RandomHexLen        int
	DirLetterCount      *int
	DirNestDepth        *int
	DB                  *DBConfig
	DefaultActor        string
	Thumbnails          *pipeline.ThumbnailsConfig
	PresignTTL          time.Duration
	// PDFRenderer rasterizes PDF pages for PDFThumbnail. Required when generating PDF thumbs.
	PDFRenderer pipeline.PDFRenderer

	// AdapterOverrides injects prebuilt adapters (tests).
	AdapterOverrides map[string]storage.Adapter
	S3Overrides      map[string]*s3store.FS
}

type Median struct {
	cfg       Config
	adapters  map[string]storage.Adapter
	s3        map[string]*s3store.FS
	slots     chan struct{}
	repo      *db.MediaRepo
	dbEnabled bool
}

func New(cfg Config) (*Median, error) {
	if len(cfg.Storages) == 0 && len(cfg.AdapterOverrides) == 0 {
		return nil, errors.New("median: storages is empty")
	}
	if cfg.DefaultKey == "" {
		return nil, errors.New("median: defaultKey is empty")
	}
	if cfg.Storages != nil {
		if _, ok := cfg.Storages[cfg.DefaultKey]; !ok {
			if _, ok2 := cfg.AdapterOverrides[cfg.DefaultKey]; !ok2 {
				return nil, fmt.Errorf("median: defaultKey %q not in storages", cfg.DefaultKey)
			}
		}
	} else if _, ok := cfg.AdapterOverrides[cfg.DefaultKey]; !ok {
		return nil, fmt.Errorf("median: defaultKey %q not in storages", cfg.DefaultKey)
	}
	if cfg.DirLetterCount == nil || cfg.DirNestDepth == nil {
		return nil, errors.New("median: DirLetterCount and DirNestDepth are required")
	}
	if *cfg.DirLetterCount <= 0 || *cfg.DirNestDepth <= 0 {
		return nil, errors.New("median: DirLetterCount and DirNestDepth must be positive")
	}
	if cfg.MaxSize == 0 {
		cfg.MaxSize = DefaultMaxSize
	}
	if cfg.MaxConcurrentStores == 0 {
		cfg.MaxConcurrentStores = DefaultMaxConcurrentStores
	}
	if cfg.FilenameMode == "" {
		cfg.FilenameMode = NameRandom
	}
	if cfg.RandomHexLen == 0 {
		cfg.RandomHexLen = DefaultRandomHexLen
	}
	if cfg.PresignTTL == 0 {
		cfg.PresignTTL = DefaultPresignTTL
	}
	if cfg.Thumbnails != nil {
		for _, k := range cfg.Thumbnails.DefaultKeys {
			if _, ok := cfg.Thumbnails.Presets[k]; !ok {
				return nil, fmt.Errorf("median: thumbnail defaultKey %q missing from presets", k)
			}
		}
	}

	adapters := make(map[string]storage.Adapter)
	s3map := make(map[string]*s3store.FS)
	for key, sc := range cfg.Storages {
		if ov, ok := cfg.AdapterOverrides[key]; ok {
			adapters[key] = ov
			if s3fs, ok := cfg.S3Overrides[key]; ok {
				s3map[key] = s3fs
			}
			continue
		}
		switch sc.Driver {
		case DriverLocal:
			fs, err := local.New(sc.Path)
			if err != nil {
				return nil, fmt.Errorf("median: storage %q: %w", key, err)
			}
			adapters[key] = fs
		case DriverS3:
			if sc.Bucket == "" {
				return nil, fmt.Errorf("median: storage %q: s3 bucket required", key)
			}
			fs, err := s3store.New(context.Background(), s3store.Config{
				Bucket: sc.Bucket, Region: sc.Region, Endpoint: sc.Endpoint,
				ForcePathStyle: sc.ForcePathStyle, Prefix: sc.Prefix,
				AccessKey: sc.AccessKey, SecretKey: sc.SecretKey,
				PresignTTL: cfg.PresignTTL,
			})
			if err != nil {
				return nil, fmt.Errorf("median: storage %q: %w", key, err)
			}
			adapters[key] = fs
			s3map[key] = fs
		case "":
			return nil, fmt.Errorf("median: storage %q: driver is empty", key)
		default:
			return nil, fmt.Errorf("median: storage %q: unknown driver %q", key, sc.Driver)
		}
	}
	for key, ov := range cfg.AdapterOverrides {
		adapters[key] = ov
		if s3fs, ok := cfg.S3Overrides[key]; ok {
			s3map[key] = s3fs
		}
	}

	m := &Median{
		cfg:      cfg,
		adapters: adapters,
		s3:       s3map,
		slots:    make(chan struct{}, cfg.MaxConcurrentStores),
	}

	if cfg.DB != nil {
		gdb := cfg.DB.Gorm
		var err error
		if gdb == nil {
			dsn := cfg.DB.DSN
			if dsn == "" {
				return nil, errors.New("median: DB.Gorm or DB.DSN required")
			}
			gdb, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
			if err != nil {
				return nil, fmt.Errorf("median: open db: %w", err)
			}
		}
		strategy := cfg.DB.IDStrategy
		if strategy == "" {
			strategy = db.IDAutoIncrement
		}
		repo, err := db.NewMediaRepo(gdb, db.RepoOptions{
			IDStrategy: strategy,
			TableName:  cfg.DB.TableName,
			Columns:    cfg.DB.Columns,
		})
		if err != nil {
			return nil, err
		}
		m.repo = repo
		m.dbEnabled = true
	}
	return m, nil
}

func toInternalMode(m NameMode) internal.NameMode {
	if m == NamePreserve {
		return internal.NamePreserve
	}
	return internal.NameRandom
}
