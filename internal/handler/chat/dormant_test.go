package chat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian"
)

func TestPostDormant_unloadsAndReportsDormant(t *testing.T) {
	t.Cleanup(farmguardian.ClearDormantFlag)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models", "/models":
			w.WriteHeader(http.StatusOK)
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

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/guardian/dormant", strings.NewReader(`{"mode":"quick"}`))
	h.PostDormant(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var resp dormantResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.State != "dormant" {
		t.Fatalf("state=%q want dormant", resp.State)
	}
}

func TestPostDormant_AIDisabled(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: false}, nil, nil, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/guardian/dormant", strings.NewReader(`{"mode":"quick"}`))
	h.PostDormant(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
