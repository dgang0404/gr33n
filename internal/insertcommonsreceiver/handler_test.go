package insertcommonsreceiver

import (
	"net/http/httptest"
	"testing"
)

func TestIngestIdempotencyKey(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/ingest", nil)
	req.Header.Set("Idempotency-Key", "  abc-123  ")
	key, err := ingestIdempotencyKey(req)
	if err != nil || key != "abc-123" {
		t.Fatalf("key=%q err=%v", key, err)
	}
}

func TestIngestIdempotencyKey_MissingReturnsEmpty(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/ingest", nil)
	key, err := ingestIdempotencyKey(req)
	if err != nil || key != "" {
		t.Fatalf("key=%q err=%v", key, err)
	}
}

func TestIngestIdempotencyKey_TooLong(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/ingest", nil)
	long := make([]byte, maxIdempotencyKeyLen+1)
	for i := range long {
		long[i] = 'a'
	}
	req.Header.Set("Idempotency-Key", string(long))
	_, err := ingestIdempotencyKey(req)
	if err == nil {
		t.Fatal("expected error for overlong key")
	}
}

func TestNewHandler_AllowNoAuth(t *testing.T) {
	h := NewHandler(nil, "", true, 30)
	if h == nil || !h.allowNoAuth {
		t.Fatal("expected allowNoAuth handler")
	}
}
