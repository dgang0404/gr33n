package chat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gr33n-api/internal/ai"
)

func TestPostWarmup_AIEnabled_IdempotentStirring(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models", "/models":
			w.WriteHeader(http.StatusOK)
		case "/api/ps":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"models": []any{}})
		case "/api/generate":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"done":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	t.Setenv("LLM_BASE_URL", srv.URL+"/v1")
	t.Setenv("LLM_MODEL", "tinyllama:latest")

	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, nil, nil)

	doWarmup := func() warmupResponse {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/guardian/warmup", strings.NewReader(`{"mode":"quick"}`))
		h.PostWarmup(rec, req)
		if rec.Code != http.StatusAccepted && rec.Code != http.StatusOK {
			t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
		}
		var resp warmupResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		return resp
	}

	first := doWarmup()
	if first.State != "stirring" {
		t.Fatalf("expected stirring, got %+v", first)
	}
	second := doWarmup()
	if second.State != "stirring" {
		t.Fatalf("expected idempotent stirring, got %+v", second)
	}
}

func TestPostWarmup_AIDisabled(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: false}, nil, nil, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/guardian/warmup", strings.NewReader(`{"mode":"quick"}`))
	h.PostWarmup(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
