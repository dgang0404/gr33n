package chat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gr33n-api/internal/ai"
)

func TestGetHealth_AIEnabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" || r.URL.Path == "/v1/models" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	t.Setenv("LLM_BASE_URL", srv.URL+"/v1")

	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, nil, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/health", nil)
	h.GetHealth(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	fa, _ := body["field_assistant"].(map[string]any)
	if fa["field_mode"] != true {
		t.Fatalf("field_mode: %v", fa["field_mode"])
	}
	if fa["llm_reachable"] != true {
		t.Fatalf("llm_reachable: %v", fa["llm_reachable"])
	}
	aw, ok := body["awakening"].(map[string]any)
	if !ok {
		t.Fatalf("missing awakening block: %v", body)
	}
	if aw["state"] == nil {
		t.Fatalf("awakening.state missing: %v", aw)
	}
}
