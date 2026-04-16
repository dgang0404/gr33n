package filestorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Local stores blobs under a single root directory (dev / single-node prod).
type Local struct {
	root string
}

func NewLocal(root string) (*Local, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, err
	}
	return &Local{root: abs}, nil
}

func (l *Local) absKey(key string) (string, error) {
	if key == "" || strings.Contains(key, "..") {
		return "", fmt.Errorf("invalid storage key")
	}
	full := filepath.Join(l.root, filepath.FromSlash(key))
	prefix := l.root + string(os.PathSeparator)
	if !strings.HasPrefix(full, prefix) && full != l.root {
		return "", fmt.Errorf("invalid storage key")
	}
	return full, nil
}

func (l *Local) Backend() string { return "local" }

// Put writes the stream to key, refusing to write more than maxBytes.
func (l *Local) Put(_ context.Context, key string, r io.Reader, maxBytes int64) (written int64, err error) {
	path, err := l.absKey(key)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return 0, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	lr := &io.LimitedReader{R: r, N: maxBytes + 1}
	n, err := io.Copy(f, lr)
	if n > maxBytes {
		_ = os.Remove(path)
		return 0, fmt.Errorf("file too large")
	}
	if err != nil {
		_ = os.Remove(path)
		return 0, err
	}
	return n, nil
}

// Open returns a reader for key.
func (l *Local) Open(_ context.Context, key string) (io.ReadCloser, error) {
	path, err := l.absKey(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

// Delete removes a blob if it exists.
func (l *Local) Delete(_ context.Context, key string) error {
	path, err := l.absKey(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (l *Local) DownloadURL(_ context.Context, key, filename, mime string, ttl time.Duration) (string, error) {
	return "", ErrDownloadURLNotSupported
}

// ExtForMime returns a safe file extension for receipt uploads.
func ExtForMime(mime string) string {
	switch strings.ToLower(strings.TrimSpace(mime)) {
	case "application/pdf":
		return ".pdf"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}
