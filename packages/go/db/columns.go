package db

// ColumnMap maps logical media fields to physical column names.
// Empty fields fall back to DefaultColumnMap values.
type ColumnMap struct {
	ID               string
	FilePath         string
	FileName         string
	OriginalFileName string
	MIME             string
	Size             string
	Width            string
	Height           string
	Hash             string
	OriginalID       string
	VariantKey       string
	CreatedBy        string
	OwnedBy          string
	Status           string
	CreatedAt        string
}

// DefaultColumnMap returns the canonical physical column names.
func DefaultColumnMap() ColumnMap {
	return ColumnMap{
		ID:               "id",
		FilePath:         "file_path",
		FileName:         "file_name",
		OriginalFileName: "original_file_name",
		MIME:             "mime",
		Size:             "size",
		Width:            "width",
		Height:           "height",
		Hash:             "hash",
		OriginalID:       "original_id",
		VariantKey:       "variant_key",
		CreatedBy:        "created_by",
		OwnedBy:          "owned_by",
		Status:           "status",
		CreatedAt:        "created_at",
	}
}

// WithDefaults fills empty entries from DefaultColumnMap.
func (c ColumnMap) WithDefaults() ColumnMap {
	d := DefaultColumnMap()
	if c.ID != "" {
		d.ID = c.ID
	}
	if c.FilePath != "" {
		d.FilePath = c.FilePath
	}
	if c.FileName != "" {
		d.FileName = c.FileName
	}
	if c.OriginalFileName != "" {
		d.OriginalFileName = c.OriginalFileName
	}
	if c.MIME != "" {
		d.MIME = c.MIME
	}
	if c.Size != "" {
		d.Size = c.Size
	}
	if c.Width != "" {
		d.Width = c.Width
	}
	if c.Height != "" {
		d.Height = c.Height
	}
	if c.Hash != "" {
		d.Hash = c.Hash
	}
	if c.OriginalID != "" {
		d.OriginalID = c.OriginalID
	}
	if c.VariantKey != "" {
		d.VariantKey = c.VariantKey
	}
	if c.CreatedBy != "" {
		d.CreatedBy = c.CreatedBy
	}
	if c.OwnedBy != "" {
		d.OwnedBy = c.OwnedBy
	}
	if c.Status != "" {
		d.Status = c.Status
	}
	if c.CreatedAt != "" {
		d.CreatedAt = c.CreatedAt
	}
	return d
}
