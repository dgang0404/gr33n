package filestorage

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Backend        string
	LocalRoot      string
	S3Bucket       string
	S3Region       string
	S3Endpoint     string
	S3Prefix       string
	S3AccessKeyID  string
	S3SecretKey    string
	S3UsePathStyle bool
	S3DisableHTTPS bool
	DownloadURLTTL time.Duration
}

func ConfigFromEnv(getenv func(string) string) Config {
	backend := strings.ToLower(strings.TrimSpace(getenv("FILE_STORAGE_BACKEND")))
	if backend == "" {
		backend = "local"
	}
	return Config{
		Backend:        backend,
		LocalRoot:      strings.TrimSpace(getenv("FILE_STORAGE_DIR")),
		S3Bucket:       strings.TrimSpace(getenv("S3_BUCKET")),
		S3Region:       strings.TrimSpace(getenv("S3_REGION")),
		S3Endpoint:     strings.TrimSpace(getenv("S3_ENDPOINT")),
		S3Prefix:       strings.Trim(strings.TrimSpace(getenv("S3_PREFIX")), "/"),
		S3AccessKeyID:  strings.TrimSpace(getenv("S3_ACCESS_KEY_ID")),
		S3SecretKey:    strings.TrimSpace(getenv("S3_SECRET_ACCESS_KEY")),
		S3UsePathStyle: envBool(getenv("S3_USE_PATH_STYLE")),
		S3DisableHTTPS: envBool(getenv("S3_DISABLE_HTTPS")),
		DownloadURLTTL: envDurationSeconds(getenv("FILE_STORAGE_SIGNED_URL_TTL_SECONDS"), 300),
	}
}

func envBool(v string) bool {
	b, _ := strconv.ParseBool(strings.TrimSpace(v))
	return b
}

func envDurationSeconds(v string, fallback int) time.Duration {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n <= 0 {
		n = fallback
	}
	return time.Duration(n) * time.Second
}

func NewFromEnv(ctx context.Context) (Store, Config, error) {
	cfg := ConfigFromEnv(os.Getenv)
	if cfg.LocalRoot == "" {
		cfg.LocalRoot = "./data/files"
	}
	store, err := NewStore(ctx, cfg)
	return store, cfg, err
}

func NewStore(ctx context.Context, cfg Config) (Store, error) {
	switch cfg.Backend {
	case "local":
		return NewLocal(cfg.LocalRoot)
	case "s3":
		return NewS3(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported FILE_STORAGE_BACKEND %q", cfg.Backend)
	}
}
