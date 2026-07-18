package httputil

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrInvalidLimit  = errors.New("invalid limit")
	ErrInvalidOffset = errors.New("invalid offset")
)

// PathValueInt64 parses r.PathValue(key) as base-10 int64.
func PathValueInt64(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(r.PathValue(key), 10, 64)
}

// ParseLimitOffset reads ?limit and ?offset; invalid values fall back to defaults.
func ParseLimitOffset(r *http.Request, defaultLimit, maxLimit int) (limit, offset int) {
	limit = defaultLimit
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
			if limit > maxLimit {
				limit = maxLimit
			}
		}
	}
	if s := strings.TrimSpace(r.URL.Query().Get("offset")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

// ParseLimitOffsetStrict returns ErrInvalidLimit or ErrInvalidOffset for bad query values.
func ParseLimitOffsetStrict(r *http.Request, defaultLimit, maxLimit int) (limit, offset int, err error) {
	limit = defaultLimit
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		n, parseErr := strconv.ParseInt(s, 10, 32)
		if parseErr != nil || n < 1 {
			return 0, 0, ErrInvalidLimit
		}
		if int(n) > maxLimit {
			n = int64(maxLimit)
		}
		limit = int(n)
	}
	offset = 0
	if s := strings.TrimSpace(r.URL.Query().Get("offset")); s != "" {
		n, parseErr := strconv.ParseInt(s, 10, 32)
		if parseErr != nil || n < 0 {
			return 0, 0, ErrInvalidOffset
		}
		offset = int(n)
	}
	return limit, offset, nil
}
