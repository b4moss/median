package core

import (
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"strings"
)

var (
	ErrInvalidBase64 = errors.New("median: invalid base64")
	ErrInvalidPart   = errors.New("median: invalid multipart part")
)

// DecodeBase64 decodes raw base64 or a data URL into bytes and optional MIME.
func DecodeBase64(s string) (data []byte, declaredMIME string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, "", ErrInvalidBase64
	}
	if strings.HasPrefix(s, "data:") {
		comma := strings.IndexByte(s, ',')
		if comma < 0 {
			return nil, "", ErrInvalidBase64
		}
		meta := s[len("data:"):comma]
		payload := s[comma+1:]
		if !strings.Contains(meta, ";base64") {
			return nil, "", ErrInvalidBase64
		}
		media := strings.TrimSuffix(meta, ";base64")
		media = strings.TrimSuffix(media, ";")
		if media != "" {
			declaredMIME = media
		}
		s = payload
	}
	// Try StdEncoding then RawStdEncoding.
	data, err = base64.StdEncoding.DecodeString(s)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(s)
		if err != nil {
			return nil, "", ErrInvalidBase64
		}
	}
	return data, declaredMIME, nil
}

// MultipartPart is an already-extracted multipart file part.
type MultipartPart struct {
	Filename string
	MIME     string
	Reader   io.Reader
	Size     int64
}

// PartToStoreOptions maps a part into Store options and body reader.
// MIME is required; filename may be empty (random mode fills it).
func PartToStoreOptions(part MultipartPart) (io.Reader, int64, StoreOptions, error) {
	if part.Reader == nil {
		return nil, 0, StoreOptions{}, ErrInvalidPart
	}
	if part.Size < 0 {
		return nil, 0, StoreOptions{}, ErrInvalidPart
	}
	mimeType := strings.TrimSpace(part.MIME)
	if mimeType == "" {
		return nil, 0, StoreOptions{}, ErrInvalidPart
	}
	// Normalize if parameters present.
	if mt, _, err := mime.ParseMediaType(mimeType); err == nil {
		mimeType = mt
	}
	opt := StoreOptions{
		MIME:     mimeType,
		Filename: part.Filename,
	}
	return part.Reader, part.Size, opt, nil
}
