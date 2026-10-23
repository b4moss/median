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
	"sort"
	"time"

	"github.com/b4moss/median/packages/go/db"
	"github.com/b4moss/median/packages/go/internal"
	"github.com/b4moss/median/packages/go/pipeline"
	"github.com/b4moss/median/packages/go/storage"
	s3store "github.com/b4moss/median/packages/go/storage/s3"
)

type VariantResult struct {
	Key        string
	Path       string
	MIME       string
	Size       int64
	Hash       string
	Width      int
	Height     int
	ID         any
	StorageKey string
}

type StoreResult struct {
	ID         any
	Path       string
	MIME       string
	Size       int64
	Hash       string
	StorageKey string
	Width      *int
	Height     *int
	Variants   []VariantResult
}

type StoreOptions struct {
	MIME             string
	Filename         string
	StorageKey       string
	FilenameMode     *NameMode
	RandomHexLen     int
	Actor            string
	RejectDuplicate  bool
	Compress         bool
	KeepOriginal     bool
	Quality          float64
	Resize           *pipeline.ResizeConstraint
	ThumbnailKeys    []string
	PDFThumbnail     bool
}

type GetOptions struct {
	StorageKey string
	WithBody   bool
	ID         any // required when DB enabled
}

// GetWithAllThumbnailsOptions configures GetWithAllThumbnails.
// StorageKey is optional (same as Get); empty uses defaultKey.
type GetWithAllThumbnailsOptions struct {
	StorageKey string
}

type GetResult struct {
	ID         any
	Path       string
	MIME       string
	Size       int64
	Hash       string
	StorageKey string
	Width      *int
	Height     *int
	Body       []byte
}

type DeleteOptions struct {
	StorageKey string
	ID         any // required when DB enabled
}

var (
	ErrNegativeSize        = errors.New("median: negative size")
	ErrSizeExceeded        = errors.New("median: size exceeds MaxSize")
	ErrEmptyPath           = errors.New("median: path is empty")
	ErrUnknownKey          = errors.New("median: unknown storage key")
	ErrActorRequired       = errors.New("median: actor required")
	ErrIDRequired          = errors.New("median: id required")
	ErrDBRequired          = errors.New("median: db required")
	ErrDuplicate           = errors.New("median: duplicate hash")
	ErrNotS3               = errors.New("median: storage is not s3")
	ErrPDFRendererRequired = errors.New("median: PDFRenderer required")
)

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
	actor := opt.Actor
	if actor == "" {
		actor = m.cfg.DefaultActor
	}
	if m.dbEnabled && actor == "" {
		return nil, ErrActorRequired
	}

	if err := m.acquire(ctx); err != nil {
		return nil, err
	}
	defer m.release()

	data, err := readExact(r, size)
	if err != nil {
		return nil, err
	}
	if isSVGMIME(opt.MIME) {
		data, err = pipeline.SanitizeSVG(data)
		if err != nil {
			return nil, err
		}
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])

	if m.dbEnabled {
		existing, err := m.repo.FindByHash(ctx, hash)
		if err == nil {
			if opt.RejectDuplicate {
				return nil, ErrDuplicate
			}
			return mediaToStoreResult(existing, opt.StorageKey, m.cfg.DefaultKey), nil
		}
		if !errors.Is(err, db.ErrNotFound) {
			return nil, err
		}
	}

	key, ad, err := m.resolveKey(opt.StorageKey)
	if err != nil {
		return nil, err
	}

	procOpts := pipeline.ProcessOptions{
		MIME: opt.MIME, Compress: opt.Compress, KeepOriginal: opt.KeepOriginal,
		Quality: opt.Quality, Resize: opt.Resize, Thumbnails: m.cfg.Thumbnails,
		ThumbnailKeys: opt.ThumbnailKeys,
	}
	runPipe := isImageMIME(opt.MIME) && (opt.Compress || opt.Resize != nil || len(opt.ThumbnailKeys) > 0 ||
		(m.cfg.Thumbnails != nil && len(m.cfg.Thumbnails.DefaultKeys) > 0))
	runPDFThumbs := isPDFMIME(opt.MIME) && opt.PDFThumbnail

	var (
		mainData = data
		width    *int
		height   *int
		variants []pipeline.Variant
	)
	if runPipe {
		pres, err := pipeline.Process(data, procOpts)
		if err != nil {
			return nil, err
		}
		mainData = pres.Bytes
		w, h := pres.Width, pres.Height
		width, height = &w, &h
		variants = pres.Variants
	} else if runPDFThumbs {
		keys := opt.ThumbnailKeys
		if len(keys) == 0 && m.cfg.Thumbnails != nil {
			keys = append([]string{}, m.cfg.Thumbnails.DefaultKeys...)
		}
		if len(keys) > 0 {
			if m.cfg.PDFRenderer == nil {
				return nil, ErrPDFRendererRequired
			}
			img, err := m.cfg.PDFRenderer.RenderPage(data, 0)
			if err != nil {
				return nil, err
			}
			variants, err = pipeline.ThumbnailsFromImage(img, m.cfg.Thumbnails, keys)
			if err != nil {
				return nil, err
			}
		}
	} else if !isImageMIME(opt.MIME) && (opt.Resize != nil || len(opt.ThumbnailKeys) > 0) {
		return nil, pipeline.ErrInvalidImage
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

	written := []string{relPath}
	if err := ad.Put(ctx, relPath, bytes.NewReader(mainData), int64(len(mainData))); err != nil {
		return nil, err
	}
	mainHash := sha256.Sum256(mainData)
	mainHashHex := hex.EncodeToString(mainHash[:])

	var varResults []VariantResult
	for _, v := range variants {
		ext := path.Ext(baseName)
		vName, err := internal.ResolveFilename(v.Key+ext, internal.NameRandom, randLen)
		if err != nil {
			_ = m.rollbackPaths(ctx, ad, written)
			return nil, err
		}
		// keep key in name for readability when preserve — use random + key prefix via preserve of sanitized
		vName = v.Key + "-" + vName
		vPath, err := internal.BuildStoragePath(vName, *m.cfg.DirLetterCount, *m.cfg.DirNestDepth)
		if err != nil {
			_ = m.rollbackPaths(ctx, ad, written)
			return nil, err
		}
		if err := ad.Put(ctx, vPath, bytes.NewReader(v.Bytes), int64(len(v.Bytes))); err != nil {
			_ = m.rollbackPaths(ctx, ad, written)
			return nil, err
		}
		written = append(written, vPath)
		vh := sha256.Sum256(v.Bytes)
		varResults = append(varResults, VariantResult{
			Key: v.Key, Path: vPath, MIME: v.MIME, Size: int64(len(v.Bytes)),
			Hash: hex.EncodeToString(vh[:]), Width: v.Width, Height: v.Height, StorageKey: key,
		})
	}

	out := &StoreResult{
		Path: relPath, MIME: opt.MIME, Size: int64(len(mainData)),
		Hash: mainHashHex, StorageKey: key, Width: width, Height: height,
		Variants: varResults,
	}

	if m.dbEnabled {
		row := &db.Media{
			FilePath: relPath, FileName: baseName, OriginalFileName: opt.Filename,
			MIME: opt.MIME, Size: out.Size, Hash: hash,
			Width: width, Height: height, CreatedBy: actor, OwnedBy: actor,
		}
		created, err := m.repo.Create(ctx, row)
		if err != nil {
			_ = m.rollbackPaths(ctx, ad, written)
			return nil, err
		}
		out.ID = created.ID
		for i := range varResults {
			vk := varResults[i].Key
			w, h := varResults[i].Width, varResults[i].Height
			childName := path.Base(varResults[i].Path)
			child, err := m.repo.Create(ctx, &db.Media{
				FilePath: varResults[i].Path, FileName: childName, OriginalFileName: "",
				MIME: varResults[i].MIME, Size: varResults[i].Size,
				Hash: varResults[i].Hash, Width: &w, Height: &h,
				OriginalID: created.ID, VariantKey: &vk,
				CreatedBy: actor, OwnedBy: actor,
			})
			if err != nil {
				_ = m.repo.DeleteByID(ctx, created.ID)
				_ = m.rollbackPaths(ctx, ad, written)
				return nil, err
			}
			varResults[i].ID = child.ID
		}
		out.Variants = varResults
	}
	return out, nil
}

func (m *Median) Delete(ctx context.Context, objectPath string, opt DeleteOptions) error {
	if m.dbEnabled {
		if opt.ID == nil || opt.ID == "" {
			return ErrIDRequired
		}
		parent, err := m.repo.FindByID(ctx, opt.ID)
		if err != nil {
			return err
		}
		children, err := m.repo.ListChildren(ctx, opt.ID)
		if err != nil {
			return err
		}
		key, ad, err := m.resolveKey(opt.StorageKey)
		_ = key
		if err != nil {
			return err
		}
		var firstErr error
		for _, ch := range children {
			if err := ad.Delete(ctx, ch.FilePath); err != nil && !errors.Is(err, storage.ErrNotFound) && firstErr == nil {
				firstErr = err
			}
		}
		if err := ad.Delete(ctx, parent.FilePath); err != nil && !errors.Is(err, storage.ErrNotFound) && firstErr == nil {
			firstErr = err
		}
		if err := m.repo.DeleteByID(ctx, opt.ID); err != nil {
			return err
		}
		return firstErr
	}
	if objectPath == "" {
		return ErrEmptyPath
	}
	_, ad, err := m.resolveKey(opt.StorageKey)
	if err != nil {
		return err
	}
	return ad.Delete(ctx, objectPath)
}

func (m *Median) Get(ctx context.Context, objectPath string, opt GetOptions) (*GetResult, error) {
	if m.dbEnabled {
		if opt.ID == nil || opt.ID == "" {
			return nil, ErrIDRequired
		}
		row, err := m.repo.FindByID(ctx, opt.ID)
		if err != nil {
			return nil, err
		}
		key, ad, err := m.resolveKey(opt.StorageKey)
		if err != nil {
			return nil, err
		}
		out := &GetResult{
			ID: row.ID, Path: row.FilePath, MIME: row.MIME, Size: row.Size, Hash: row.Hash,
			StorageKey: key, Width: row.Width, Height: row.Height,
		}
		if opt.WithBody {
			rc, _, err := ad.Get(ctx, row.FilePath)
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			body, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			out.Body = body
		}
		return out, nil
	}
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
	out := &GetResult{Path: objectPath, Size: size, StorageKey: key}
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

// GetWithAllThumbnails returns the active original row and its active variants
// as a StoreResult (metadata only). DB is required. Parent is resolved via
// original_id only. Variants are ordered by width ascending (0 included),
// then variant_key ascending; NULL width last.
func (m *Median) GetWithAllThumbnails(ctx context.Context, id any, opt GetWithAllThumbnailsOptions) (*StoreResult, error) {
	if !m.dbEnabled {
		return nil, ErrDBRequired
	}
	if id == nil || id == "" {
		return nil, ErrIDRequired
	}
	key, _, err := m.resolveKey(opt.StorageKey)
	if err != nil {
		return nil, err
	}

	row, err := m.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	parentID := row.ID
	parent := row
	if row.OriginalID != nil {
		parentID = row.OriginalID
		parent, err = m.repo.FindByID(ctx, parentID)
		if err != nil {
			return nil, err
		}
	}
	if parent.Status != "active" {
		return nil, db.ErrNotFound
	}

	children, err := m.repo.ListChildren(ctx, parentID)
	if err != nil {
		return nil, err
	}

	active := make([]*db.Media, 0, len(children))
	for _, ch := range children {
		if ch.Status == "active" {
			active = append(active, ch)
		}
	}
	sort.SliceStable(active, func(i, j int) bool {
		wi, wj := active[i].Width, active[j].Width
		if wi == nil && wj == nil {
			return variantKeyOf(active[i]) < variantKeyOf(active[j])
		}
		if wi == nil {
			return false
		}
		if wj == nil {
			return true
		}
		if *wi != *wj {
			return *wi < *wj
		}
		return variantKeyOf(active[i]) < variantKeyOf(active[j])
	})

	variants := make([]VariantResult, 0, len(active))
	for _, ch := range active {
		variants = append(variants, mediaToVariantResult(ch, key))
	}

	out := mediaToStoreResult(parent, key, m.cfg.DefaultKey)
	out.StorageKey = key
	out.Variants = variants
	return out, nil
}

func variantKeyOf(m *db.Media) string {
	if m.VariantKey == nil {
		return ""
	}
	return *m.VariantKey
}

func mediaToVariantResult(row *db.Media, storageKey string) VariantResult {
	vr := VariantResult{
		Path:       row.FilePath,
		MIME:       row.MIME,
		Size:       row.Size,
		Hash:       row.Hash,
		ID:         row.ID,
		StorageKey: storageKey,
	}
	if row.VariantKey != nil {
		vr.Key = *row.VariantKey
	}
	if row.Width != nil {
		vr.Width = *row.Width
	}
	if row.Height != nil {
		vr.Height = *row.Height
	}
	return vr
}

// PresignGet returns a signed GET URL for an S3 storage key.
func (m *Median) PresignGet(ctx context.Context, objectPath, storageKey string, ttl time.Duration) (string, error) {
	key := storageKey
	if key == "" {
		key = m.cfg.DefaultKey
	}
	fs, ok := m.s3[key]
	if !ok {
		return "", ErrNotS3
	}
	if ttl <= 0 {
		ttl = m.cfg.PresignTTL
	}
	return fs.PresignGet(ctx, objectPath, ttl)
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

func (m *Median) rollbackPaths(ctx context.Context, ad storage.Adapter, paths []string) error {
	var first error
	for i := len(paths) - 1; i >= 0; i-- {
		if err := ad.Delete(ctx, paths[i]); err != nil && !errors.Is(err, storage.ErrNotFound) && first == nil {
			first = err
		}
	}
	return first
}

func readExact(r io.Reader, size int64) ([]byte, error) {
	if size == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, size)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

func isImageMIME(mime string) bool {
	switch mime {
	case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

func isSVGMIME(mime string) bool {
	return mime == "image/svg+xml"
}

func isPDFMIME(mime string) bool {
	return mime == "application/pdf"
}

func mediaToStoreResult(row *db.Media, storageKey, defaultKey string) *StoreResult {
	key := storageKey
	if key == "" {
		key = defaultKey
	}
	return &StoreResult{
		ID: row.ID, Path: row.FilePath, MIME: row.MIME, Size: row.Size, Hash: row.Hash,
		StorageKey: key, Width: row.Width, Height: row.Height,
	}
}

// Ensure s3store import used when only Presign path references maps.
var _ = s3store.Driver

func BaseName(p string) string { return path.Base(p) }
