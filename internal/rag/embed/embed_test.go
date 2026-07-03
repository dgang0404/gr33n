package embed

import (
	"strings"
	"testing"
)

func TestTruncateForErr(t *testing.T) {
	if got := truncateForErr([]byte("short"), 10); got != "short" {
		t.Fatalf("got %q", got)
	}
	long := make([]byte, 20)
	for i := range long {
		long[i] = 'x'
	}
	got := truncateForErr(long, 8)
	if !strings.HasPrefix(got, "xxxxxxxx") || !strings.HasSuffix(got, "…") {
		t.Fatalf("got %q", got)
	}
}

func TestClient_ModelID(t *testing.T) {
	c := &Client{Model: "text-embedding-3-small"}
	if c.ModelID() != "text-embedding-3-small" {
		t.Fatalf("ModelID = %q", c.ModelID())
	}
}

func TestNewOpenAICompatibleFromEnv_MissingKey(t *testing.T) {
	t.Setenv("EMBEDDING_API_KEY", "")
	_, err := NewOpenAICompatibleFromEnv()
	if err == nil {
		t.Fatal("expected error without EMBEDDING_API_KEY")
	}
}

func TestNewOpenAICompatibleFromEnv_InvalidDimension(t *testing.T) {
	t.Setenv("EMBEDDING_API_KEY", "test-key")
	t.Setenv("EMBEDDING_DIMENSION", "nope")
	_, err := NewOpenAICompatibleFromEnv()
	if err == nil {
		t.Fatal("expected invalid dimension error")
	}
}
