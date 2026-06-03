package chat

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gr33n-api/internal/ai"
)

func TestPostV1_FieldDegradeWithoutLLMClient(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, nil, nil)
	body, _ := json.Marshal(map[string]any{
		"message": "start procedure wire-pi-relay-light",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bytes.NewReader(body))
	h.PostV1(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var resp postResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Procedure == nil || resp.Procedure.ProcedureID != "wire-pi-relay-light" {
		t.Fatalf("procedure: %+v", resp.Procedure)
	}
	if !resp.FieldDegraded && resp.LLMModel != procedureModelLabel {
		// procedure path uses field-procedure; degrade uses field-degrade
	}
	if resp.LLMModel != procedureModelLabel {
		t.Fatalf("model: %s", resp.LLMModel)
	}
}

func TestPostV1_FieldDegradeSuggestWithoutLLM(t *testing.T) {
	t.Setenv("LLM_BASE_URL", "http://127.0.0.1:11434/v1")
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, nil, nil)
	body, _ := json.Marshal(map[string]any{
		"message": "help me wire the pi to a light",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bytes.NewReader(body))
	h.PostV1(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var resp postResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Procedure == nil || resp.Procedure.StepN != 1 {
		t.Fatalf("expected auto-started procedure step 1: %+v", resp.Procedure)
	}
	// May be field-procedure (direct start) or field-degrade (TryFieldDegrade auto-start).
	if resp.LLMModel != fieldDegradeModelLabel && resp.LLMModel != procedureModelLabel {
		t.Fatalf("unexpected model %s", resp.LLMModel)
	}
}
