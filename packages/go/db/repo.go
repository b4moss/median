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

var (
	ErrNotFound     = errors.New("db: media not found")
	ErrInvalidInput = errors.New("db: invalid input")
)

// Media is a media metadata row.
type Media struct {
	ID               any
	FilePath         string
	FileName         string
	OriginalFileName string
	MIME             string
	Size             int64
	Width            *int
	Height           *int
	Hash             string
	OriginalID       any
	VariantKey       *string
	CreatedBy        string
	OwnedBy          string
	Status           string
	CreatedAt        time.Time
}

// RepoOptions configures MediaRepo.
type RepoOptions struct {
	IDStrategy string
	TableName  string
	Columns    ColumnMap
}

// MediaRepo persists media rows via crudian.
type MediaRepo struct {
	crud       *gormcrud.Crud
	idStrategy string
	table      string
	cols       ColumnMap
}

// NewMediaRepo wraps an existing gorm DB with crudian.
func NewMediaRepo(gdb *gorm.DB, opt RepoOptions) (*MediaRepo, error) {
	if gdb == nil {
		return nil, errors.New("db: gorm DB is required")
	}
	if opt.IDStrategy == "" {
		opt.IDStrategy = IDAutoIncrement
	}
	table, err := NormalizeTableName(opt.TableName)
	if err != nil {
		return nil, err
	}
	crud, err := gormcrud.CreateCrud(gdb, crudian.Options{Driver: "sqlite"})
	if err != nil {
		return nil, err
	}
	return &MediaRepo{
		crud:       crud,
		idStrategy: opt.IDStrategy,
		table:      table,
		cols:       opt.Columns.WithDefaults(),
	}, nil
}

// Create inserts a media row. Actor is written to created_by and owned_by.
func (r *MediaRepo) Create(ctx context.Context, m *Media) (*Media, error) {
	if m == nil {
		return nil, ErrInvalidInput
	}
	if m.FilePath == "" || m.FileName == "" || m.MIME == "" || m.Hash == "" {
		return nil, fmt.Errorf("%w: file_path/file_name/mime/hash required", ErrInvalidInput)
	}
	if m.CreatedBy == "" || m.OwnedBy == "" {
		return nil, fmt.Errorf("%w: actor required", ErrInvalidInput)
	}
	if m.Status == "" {
		m.Status = "active"
	}
	c := r.cols
	cols := crudian.Row{
		c.FilePath:         m.FilePath,
		c.FileName:         m.FileName,
		c.OriginalFileName: m.OriginalFileName,
		c.MIME:             m.MIME,
		c.Size:             m.Size,
		c.Hash:             m.Hash,
		c.CreatedBy:        m.CreatedBy,
		c.OwnedBy:          m.OwnedBy,
		c.Status:           m.Status,
	}
	if m.Width != nil {
		cols[c.Width] = *m.Width
	}
	if m.Height != nil {
		cols[c.Height] = *m.Height
	}
	if m.OriginalID != nil {
		cols[c.OriginalID] = m.OriginalID
	}
	if m.VariantKey != nil {
		cols[c.VariantKey] = *m.VariantKey
	}
	switch r.idStrategy {
	case IDUUIDv4:
		cols[c.ID] = uuid.NewString()
	case IDUUIDv7:
		id, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		cols[c.ID] = id.String()
	case IDULID:
		id, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		cols[c.ID] = id.String()
	}
	row, err := r.crud.Create(ctx, r.table, cols)
	if err != nil {
		return nil, err
	}
	return r.rowToMedia(row), nil
}

// FindByID returns a media row by primary key.
func (r *MediaRepo) FindByID(ctx context.Context, id any) (*Media, error) {
	if id == nil || id == "" {
		return nil, fmt.Errorf("%w: empty id", ErrInvalidInput)
	}
	row, err := r.crud.Read(ctx, r.table, crudian.ReadQuery{
		Where: crudian.Where().Eq(r.cols.ID, id),
	})
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrNotFound
	}
	return r.rowToMedia(row), nil
}

// FindByHash returns the first media with hash (original preferred: variant_key IS NULL).
func (r *MediaRepo) FindByHash(ctx context.Context, hash string) (*Media, error) {
	if hash == "" {
		return nil, fmt.Errorf("%w: empty hash", ErrInvalidInput)
	}
	res, err := r.crud.Search(ctx, r.table, crudian.SearchQuery{
		Where: crudian.Where().Eq(r.cols.Hash, hash).IsNull(r.cols.VariantKey),
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		res, err = r.crud.Search(ctx, r.table, crudian.SearchQuery{
			Where: crudian.Where().Eq(r.cols.Hash, hash),
			Limit: 1,
		})
		if err != nil {
			return nil, err
		}
		if len(res.Items) == 0 {
			return nil, ErrNotFound
		}
	}
	return r.rowToMedia(res.Items[0]), nil
}

// ListChildren returns rows whose original_id matches parentID.
func (r *MediaRepo) ListChildren(ctx context.Context, parentID any) ([]*Media, error) {
	res, err := r.crud.Search(ctx, r.table, crudian.SearchQuery{
		Where: crudian.Where().Eq(r.cols.OriginalID, parentID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*Media, 0, len(res.Items))
	for _, row := range res.Items {
		out = append(out, r.rowToMedia(row))
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
		if _, err := r.crud.Delete(ctx, r.table, crudian.DeleteQuery{
			Where: crudian.Where().Eq(r.cols.ID, ch.ID),
		}); err != nil {
			return err
		}
	}
	n, err := r.crud.Delete(ctx, r.table, crudian.DeleteQuery{
		Where: crudian.Where().Eq(r.cols.ID, id),
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MediaRepo) rowToMedia(row crudian.Row) *Media {
	c := r.cols
	m := &Media{
		ID:               row[c.ID],
		FilePath:         asString(row[c.FilePath]),
		FileName:         asString(row[c.FileName]),
		OriginalFileName: asString(row[c.OriginalFileName]),
		MIME:             asString(row[c.MIME]),
		Size:             asInt64(row[c.Size]),
		Hash:             asString(row[c.Hash]),
		CreatedBy:        asString(row[c.CreatedBy]),
		OwnedBy:          asString(row[c.OwnedBy]),
		Status:           asString(row[c.Status]),
	}
	if v, ok := row[c.Width]; ok && v != nil {
		w := int(asInt64(v))
		m.Width = &w
	}
	if v, ok := row[c.Height]; ok && v != nil {
		h := int(asInt64(v))
		m.Height = &h
	}
	if v, ok := row[c.OriginalID]; ok && v != nil {
		m.OriginalID = v
	}
	if v, ok := row[c.VariantKey]; ok && v != nil {
		s := asString(v)
		m.VariantKey = &s
	}
	if v, ok := row[c.CreatedAt]; ok && v != nil {
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
