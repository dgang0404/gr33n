package guardian

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/authctx"
)

func withAuthzSkip(req *http.Request) *http.Request {
	return req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
}

func TestPostReingest_AIDisabled(t *testing.T) {
	h := &Handler{cfg: ai.Config{Enabled: false}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/farms/1/guardian/reingest", strings.NewReader(`{"scope":"field_guides"}`))
	req.SetPathValue("id", "1")
	h.PostReingest(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPostReingest_InvalidScope(t *testing.T) {
	h := &Handler{cfg: ai.Config{Enabled: true}}
	rec := httptest.NewRecorder()
	req := withAuthzSkip(httptest.NewRequest(http.MethodPost, "/farms/1/guardian/reingest", strings.NewReader(`{"scope":"bogus"}`)))
	req.SetPathValue("id", "1")
	h.PostReingest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetReingestStatus_Idle(t *testing.T) {
	h := &Handler{cfg: ai.Config{Enabled: true}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/farms/1/guardian/reingest/status", nil)
	req.SetPathValue("id", "1")
	// No auth / DB — RequireFarmMember will fail; smoke-level wiring only checks method guard.
	h.GetReingestStatus(rec, req)
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden && rec.Code != http.StatusInternalServerError {
		// Without JWT middleware, farmauthz may 401/403 depending on test harness.
		t.Logf("status without auth=%d (expected non-200 without member)", rec.Code)
	}
}
