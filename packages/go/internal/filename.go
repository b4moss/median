package internal

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"path"
	"strings"
	"unicode"
)

const (
	DefaultRandomHexLen = 32
	MaxRandomHexLen     = 128
)

var (
	ErrEmptyFilename   = errors.New("filename: empty after sanitize")
	ErrInvalidNameLen  = errors.New("filename: invalid random length")
	ErrUnknownNameMode = errors.New("filename: unknown mode")
)

// NameMode selects how Store names files.
type NameMode string

const (
	NamePreserve NameMode = "preserve"
	NameRandom   NameMode = "random"
)

// ResolveFilename returns the final base filename for storage.
func ResolveFilename(original string, mode NameMode, randomLen int) (string, error) {
	switch mode {
	case NamePreserve, "":
		return SanitizePreserve(original)
	case NameRandom:
		if randomLen == 0 {
			randomLen = DefaultRandomHexLen
		}
		return RandomFilename(original, randomLen)
	default:
		return "", ErrUnknownNameMode
	}
}

// SanitizePreserve keeps Unicode but strips dangerous path characters and edge dots/spaces.
func SanitizePreserve(original string) (string, error) {
	name := original
	name = strings.ReplaceAll(name, "\x00", "")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.Map(func(r rune) rune {
		if r == 0 || r == '/' || r == '\\' {
			return -1
		}
		return r
	}, name)
	// Collapse ".." segments as literal dots only after removing separators.
	for strings.Contains(name, "..") {
		name = strings.ReplaceAll(name, "..", ".")
	}
	name = strings.TrimFunc(name, func(r rune) bool {
		return r == '.' || unicode.IsSpace(r)
	})
	if name == "" || name == "." || name == ".." {
		return "", ErrEmptyFilename
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "\x00") {
		return "", ErrEmptyFilename
	}
	return name, nil
}

// RandomFilename returns cryptographically random hex + extension from original.
func RandomFilename(original string, hexLen int) (string, error) {
	if hexLen <= 0 || hexLen > MaxRandomHexLen || hexLen%2 != 0 {
		return "", ErrInvalidNameLen
	}
	buf := make([]byte, hexLen/2)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	base := hex.EncodeToString(buf)
	ext := path.Ext(original)
	if ext == "." {
		ext = ""
	}
	return base + ext, nil
}
