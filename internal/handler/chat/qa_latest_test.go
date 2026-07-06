package chat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/authctx"
	"gr33n-api/internal/farmguardian"
)

func TestGetLatestQARun_ok(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GUARDIAN_QA_RUNS_DIR", dir)
	path := filepath.Join(dir, "20260102T120000_smoke_phi3-mini.json")
	if err := farmguardian.SaveQARunArchive(path, "smoke", "phi3:mini", []farmguardian.EvalQuestionScore{
		{ID: "smoke-cherry-forest", Passed: true},
	}); err != nil {
		t.Fatal(err)
	}

	h := &Handler{cfg: ai.Config{Enabled: true}}
	req := httptest.NewRequest(http.MethodGet, "/v1/guardian/qa/latest", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()
	h.GetLatestQARun(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var out qaLatestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Summary.Suite != "smoke" || out.Summary.Passed != 1 {
		t.Fatalf("summary: %+v", out.Summary)
	}
}

func TestGetLatestQARun_notFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GUARDIAN_QA_RUNS_DIR", dir)

	h := &Handler{cfg: ai.Config{Enabled: true}}
	req := httptest.NewRequest(http.MethodGet, "/v1/guardian/qa/latest", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()
	h.GetLatestQARun(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestGetLatestQARun_aiDisabled(t *testing.T) {
	h := &Handler{cfg: ai.Config{Enabled: false}}
	req := httptest.NewRequest(http.MethodGet, "/v1/guardian/qa/latest", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()
	h.GetLatestQARun(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}
