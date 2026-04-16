package filestorage

import (
	"context"
	"testing"
)

func TestConfigFromEnvDefaultsToLocal(t *testing.T) {
	t.Parallel()

	cfg := ConfigFromEnv(func(string) string { return "" })
	if cfg.Backend != "local" {
		t.Fatalf("Backend = %q, want local", cfg.Backend)
	}
	if cfg.DownloadURLTTL <= 0 {
		t.Fatalf("DownloadURLTTL = %v, want > 0", cfg.DownloadURLTTL)
	}
}

func TestNewStoreRejectsMissingS3Bucket(t *testing.T) {
	t.Parallel()

	_, err := NewStore(context.Background(), Config{Backend: "s3"})
	if err == nil {
		t.Fatal("NewStore with missing bucket succeeded, want error")
	}
}

func TestNormalizeEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		raw          string
		disableHTTPS bool
		want         string
	}{
		{name: "https default", raw: "objects.example.com", want: "https://objects.example.com"},
		{name: "http override", raw: "objects.example.com:9000", disableHTTPS: true, want: "http://objects.example.com:9000"},
		{name: "preserve scheme", raw: "http://minio.internal:9000", want: "http://minio.internal:9000"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeEndpoint(tc.raw, tc.disableHTTPS)
			if err != nil {
				t.Fatalf("normalizeEndpoint: %v", err)
			}
			if got != tc.want {
				t.Fatalf("normalizeEndpoint(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}
