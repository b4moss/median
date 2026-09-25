package internal

import (
	"errors"
	"strings"
)

var (
	ErrMIMEDenied = errors.New("mime: denied")
	ErrMIMEEmpty  = errors.New("mime: empty")
)

// CheckMIME validates a declared MIME against allow/deny lists.
// Empty allow means all non-denied types are accepted. Deny always wins.
func CheckMIME(declared string, allow, deny []string) error {
	mime := strings.TrimSpace(declared)
	if mime == "" {
		return ErrMIMEEmpty
	}
	for _, d := range deny {
		if strings.EqualFold(strings.TrimSpace(d), mime) {
			return ErrMIMEDenied
		}
	}
	if len(allow) == 0 {
		return nil
	}
	for _, a := range allow {
		if strings.EqualFold(strings.TrimSpace(a), mime) {
			return nil
		}
	}
	return ErrMIMEDenied
}
