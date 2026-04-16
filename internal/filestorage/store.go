package filestorage

import (
	"context"
	"errors"
	"io"
	"time"
)

var ErrDownloadURLNotSupported = errors.New("download URL not supported")

// Store is the shared blob storage contract for receipt attachments.
type Store interface {
	Backend() string
	Put(ctx context.Context, key string, r io.Reader, maxBytes int64) (int64, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	DownloadURL(ctx context.Context, key, filename, mime string, ttl time.Duration) (string, error)
}
