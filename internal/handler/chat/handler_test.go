package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gr33n-api/internal/ai"
)

type fakeLLM struct {
	reply     string
	failWith  error
	lastUser  string
	lastSys   string
	callCount int
}

func (f *fakeLLM) ChatCompletion(_ context.Context, system, user string) (string, error) {
	f.callCount++
	f.lastSys = system
	f.lastUser = user
	if f.failWith != nil {
		return "", f.failWith
	}
	return f.reply, nil
}
func (f *fakeLLM) ModelLabel() string { return "fake-model" }

func doPost(t *testing.T, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.PostV1(rec, req)
	return rec
}

func TestPostV1_AIDisabled503(t *testing.T) {
	h := NewHandlerWithClient(ai.Config{Enabled: false}, &fakeLLM{reply: "ignored"})
	rec := doPost(t, h, `{"message":"hi"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "disabled") {
		t.Fatalf("expected disabled message, got %s", rec.Body.String())
	}
}

func TestPostV1_NoClient503(t *testing.T) {
	h := NewHandlerWithClient(ai.Config{Enabled: true}, nil)
	rec := doPost(t, h, `{"message":"hi"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "LLM_BASE_URL") {
		t.Fatalf("expected config hint, got %s", rec.Body.String())
	}
}

func TestPostV1_MissingMessage400(t *testing.T) {
	h := NewHandlerWithClient(ai.Config{Enabled: true}, &fakeLLM{reply: "x"})
	rec := doPost(t, h, `{"message":"   "}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestPostV1_EmptyBody400(t *testing.T) {
	h := NewHandlerWithClient(ai.Config{Enabled: true}, &fakeLLM{reply: "x"})
	rec := doPost(t, h, ``)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestPostV1_HappyPath200(t *testing.T) {
	llm := &fakeLLM{reply: "Check the irrigation schedule on the Dashboard."}
	h := NewHandlerWithClient(ai.Config{Enabled: true}, llm)
	rec := doPost(t, h, `{"message":"What should I do this morning?"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp postResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Answer == "" || resp.LLMModel != "fake-model" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if llm.callCount != 1 {
		t.Fatalf("expected one LLM call, got %d", llm.callCount)
	}
	if !strings.Contains(llm.lastSys, "Farm Guardian") {
		t.Fatalf("system prompt not used: %q", llm.lastSys)
	}
	if !strings.Contains(llm.lastUser, "What should I do this morning?") {
		t.Fatalf("user message not threaded: %q", llm.lastUser)
	}
}

func TestPostV1_LLMError502(t *testing.T) {
	h := NewHandlerWithClient(ai.Config{Enabled: true}, &fakeLLM{failWith: errors.New("upstream boom")})
	rec := doPost(t, h, `{"message":"hi"}`)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("want 502, got %d", rec.Code)
	}
}

func TestPostV1_WrongMethod405(t *testing.T) {
	h := NewHandlerWithClient(ai.Config{Enabled: true}, &fakeLLM{reply: "x"})
	req := httptest.NewRequest(http.MethodGet, "/v1/chat", nil)
	rec := httptest.NewRecorder()
	h.PostV1(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}
