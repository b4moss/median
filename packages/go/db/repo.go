package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/b4moss/crudian/go/crudian"
	gormcrud "github.com/b4moss/crudian/go/gorm"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const tableMedia = "media"

var (
	ErrNotFound     = errors.New("db: media not found")
	ErrInvalidInput = errors.New("db: invalid input")
)

// Media is a media metadata row.
type Media struct {
	ID         any
	Path       string
	MIME       string
	Size       int64
	Width      *int
	Height     *int
	Hash       string
	OriginalID any
	VariantKey *string
	CreatedBy  string
	OwnedBy    string
	Status     string
	CreatedAt  time.Time
}

// MediaRepo persists media rows via crudian.
type MediaRepo struct {
	crud       *gormcrud.Crud
	idStrategy string
}

// NewMediaRepo wraps an existing gorm DB with crudian.
func NewMediaRepo(gdb *gorm.DB, idStrategy string) (*MediaRepo, error) {
	if gdb == nil {
		return nil, errors.New("db: gorm DB is required")
	}
	if idStrategy == "" {
		idStrategy = IDAutoIncrement
	}
	crud, err := gormcrud.CreateCrud(gdb, crudian.Options{Driver: "sqlite"})
	if err != nil {
		return nil, err
	}
	return &MediaRepo{crud: crud, idStrategy: idStrategy}, nil
}

// Create inserts a media row. Actor is written to created_by and owned_by.
func (r *MediaRepo) Create(ctx context.Context, m *Media) (*Media, error) {
	if m == nil {
		return nil, ErrInvalidInput
	}
	if m.Path == "" || m.MIME == "" || m.Hash == "" {
		return nil, fmt.Errorf("%w: path/mime/hash required", ErrInvalidInput)
	}
	if m.CreatedBy == "" || m.OwnedBy == "" {
		return nil, fmt.Errorf("%w: actor required", ErrInvalidInput)
	}
	if m.Status == "" {
		m.Status = "active"
	}
	cols := crudian.Row{
		"path":       m.Path,
		"mime":       m.MIME,
		"size":       m.Size,
		"hash":       m.Hash,
		"created_by": m.CreatedBy,
		"owned_by":   m.OwnedBy,
		"status":     m.Status,
	}
	if m.Width != nil {
		cols["width"] = *m.Width
	}
	if m.Height != nil {
		cols["height"] = *m.Height
	}
	if m.OriginalID != nil {
		cols["original_id"] = m.OriginalID
	}
	if m.VariantKey != nil {
		cols["variant_key"] = *m.VariantKey
	}
	switch r.idStrategy {
	case IDUUIDv4:
		cols["id"] = uuid.NewString()
	case IDUUIDv7:
		// uuid v7 not in stdlib; use v4 placeholder with time-ordered note — use google uuid v7 if available.
		id, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		cols["id"] = id.String()
	case IDULID:
		// ULID-like: use uuid v7 string as sortable text id for MVP.
		id, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		cols["id"] = id.String()
	}
	row, err := r.crud.Create(ctx, tableMedia, cols)
	if err != nil {
		return nil, err
	}
	return rowToMedia(row), nil
}

// FindByID returns a media row by primary key.
func (r *MediaRepo) FindByID(ctx context.Context, id any) (*Media, error) {
	if id == nil || id == "" {
		return nil, fmt.Errorf("%w: empty id", ErrInvalidInput)
	}
	row, err := r.crud.Read(ctx, tableMedia, crudian.ReadQuery{
		Where: crudian.Where().Eq("id", id),
	})
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	return rowToMedia(row), nil
}

// FindByHash returns the first media with hash (original preferred: variant_key IS NULL).
func (r *MediaRepo) FindByHash(ctx context.Context, hash string) (*Media, error) {
	if hash == "" {
		return nil, fmt.Errorf("%w: empty hash", ErrInvalidInput)
	}
	res, err := r.crud.Search(ctx, tableMedia, crudian.SearchQuery{
		Where: crudian.Where().Eq("hash", hash).IsNull("variant_key"),
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		// fallback any row with hash
		res, err = r.crud.Search(ctx, tableMedia, crudian.SearchQuery{
			Where: crudian.Where().Eq("hash", hash),
			Limit: 1,
		})
		if err != nil {
			return nil, err
		}
		if len(res.Items) == 0 {
			return nil, ErrNotFound
		}
	}
	return rowToMedia(res.Items[0]), nil
}

// ListChildren returns rows whose original_id matches parentID.
func (r *MediaRepo) ListChildren(ctx context.Context, parentID any) ([]*Media, error) {
	res, err := r.crud.Search(ctx, tableMedia, crudian.SearchQuery{
		Where: crudian.Where().Eq("original_id", parentID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*Media, 0, len(res.Items))
	for _, row := range res.Items {
		out = append(out, rowToMedia(row))
	}
	return out, nil
}

// DeleteByID deletes the row and all children (DB only).
func (r *MediaRepo) DeleteByID(ctx context.Context, id any) error {
	if id == nil || id == "" {
		return fmt.Errorf("%w: empty id", ErrInvalidInput)
	}
	children, err := r.ListChildren(ctx, id)
	if err != nil {
		return err
	}
	for _, ch := range children {
		if _, err := r.crud.Delete(ctx, tableMedia, crudian.DeleteQuery{
			Where: crudian.Where().Eq("id", ch.ID),
		}); err != nil {
			return err
		}
	}
	n, err := r.crud.Delete(ctx, tableMedia, crudian.DeleteQuery{
		Where: crudian.Where().Eq("id", id),
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func rowToMedia(row crudian.Row) *Media {
	m := &Media{
		ID:        row["id"],
		Path:      asString(row["path"]),
		MIME:      asString(row["mime"]),
		Size:      asInt64(row["size"]),
		Hash:      asString(row["hash"]),
		CreatedBy: asString(row["created_by"]),
		OwnedBy:   asString(row["owned_by"]),
		Status:    asString(row["status"]),
	}
	if v, ok := row["width"]; ok && v != nil {
		w := int(asInt64(v))
		m.Width = &w
	}
	if v, ok := row["height"]; ok && v != nil {
		h := int(asInt64(v))
		m.Height = &h
	}
	if v, ok := row["original_id"]; ok && v != nil {
		m.OriginalID = v
	}
	if v, ok := row["variant_key"]; ok && v != nil {
		s := asString(v)
		m.VariantKey = &s
	}
	if v, ok := row["created_at"]; ok && v != nil {
		switch t := v.(type) {
		case time.Time:
			m.CreatedAt = t
		case string:
			if parsed, err := time.Parse("2006-01-02 15:04:05", t); err == nil {
				m.CreatedAt = parsed
			}
		}
	}
	return m
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprint(t)
	}
}

func asInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case int32:
		return int64(t)
	case float64:
		return int64(t)
	case []byte:
		var n int64
		fmt.Sscan(string(t), &n)
		return n
	case string:
		var n int64
		fmt.Sscan(t, &n)
		return n
	default:
		return 0
	}
}
