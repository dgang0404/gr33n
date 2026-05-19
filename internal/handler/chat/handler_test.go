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
	h := NewHandlerWithDeps(ai.Config{Enabled: false}, nil, &fakeLLM{reply: "ignored"}, nil)
	rec := doPost(t, h, `{"message":"hi"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "disabled") {
		t.Fatalf("expected disabled message, got %s", rec.Body.String())
	}
}

func TestPostV1_NoClient503(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, nil, nil)
	rec := doPost(t, h, `{"message":"hi"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "LLM_BASE_URL") {
		t.Fatalf("expected config hint, got %s", rec.Body.String())
	}
}

func TestPostV1_MissingMessage400(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, &fakeLLM{reply: "x"}, nil)
	rec := doPost(t, h, `{"message":"   "}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestPostV1_EmptyBody400(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, &fakeLLM{reply: "x"}, nil)
	rec := doPost(t, h, ``)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestPostV1_HappyPath200(t *testing.T) {
	llm := &fakeLLM{reply: "Check the irrigation schedule on the Dashboard."}
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, llm, nil)
	rec := doPost(t, h, `{"message":"What should I do this morning?","session_id":"sess-1"}`)
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
	if resp.Grounded {
		t.Fatalf("expected grounded=false on plain path, got %+v", resp)
	}
	if resp.SessionID != "sess-1" {
		t.Fatalf("expected session_id echo, got %q", resp.SessionID)
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
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, &fakeLLM{failWith: errors.New("upstream boom")}, nil)
	rec := doPost(t, h, `{"message":"hi"}`)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("want 502, got %d", rec.Code)
	}
}

func TestPostV1_WrongMethod405(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, &fakeLLM{reply: "x"}, nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/chat", nil)
	rec := httptest.NewRecorder()
	h.PostV1(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}

func TestPostV1_InvalidFarmID400(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, &fakeLLM{reply: "x"}, nil)
	rec := doPost(t, h, `{"message":"hi","farm_id":0}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

// fakeStreamingLLM implements both ChatCompleter and StreamingChatCompleter
// so the handler picks the streaming branch when stream=true.
type fakeStreamingLLM struct {
	deltas []string
	failOn int  // 1-indexed delta to fail after; 0 disables
	called bool // true if ChatCompletionStream ran
}

func (f *fakeStreamingLLM) ChatCompletion(_ context.Context, _, _ string) (string, error) {
	return strings.Join(f.deltas, ""), nil
}
func (f *fakeStreamingLLM) ModelLabel() string { return "fake-streamer" }
func (f *fakeStreamingLLM) ChatCompletionStream(_ context.Context, _, _ string, onDelta func(string)) error {
	f.called = true
	for i, d := range f.deltas {
		onDelta(d)
		if f.failOn > 0 && i+1 == f.failOn {
			return errors.New("midstream boom")
		}
	}
	return nil
}

func TestPostV1_StreamHappyPath(t *testing.T) {
	llm := &fakeStreamingLLM{deltas: []string{"Hello", ", ", "world"}}
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, llm, nil)
	rec := doPost(t, h, `{"message":"hi","stream":true,"session_id":"abc"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("expected SSE content-type, got %q", ct)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`event: delta`,
		`"text":"Hello"`,
		`"text":", "`,
		`"text":"world"`,
		`event: done`,
		`"answer":"Hello, world"`,
		`"session_id":"abc"`,
		`data: [DONE]`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q\nfull body:\n%s", want, body)
		}
	}
	if !llm.called {
		t.Fatal("expected streaming branch to run")
	}
}

func TestPostV1_StreamUnsupportedLLM501(t *testing.T) {
	// fakeLLM only implements ChatCompleter, not StreamingChatCompleter.
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, &fakeLLM{reply: "x"}, nil)
	rec := doPost(t, h, `{"message":"hi","stream":true}`)
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("want 501, got %d", rec.Code)
	}
}

func TestPostV1_StreamErrorMidstream(t *testing.T) {
	llm := &fakeStreamingLLM{deltas: []string{"part-1", "part-2"}, failOn: 1}
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, llm, nil)
	rec := doPost(t, h, `{"message":"hi","stream":true}`)
	// SSE always returns 200 (headers committed before the error).
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 (SSE), got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: error") || !strings.Contains(body, "LLM request failed") {
		t.Fatalf("expected error event, got %s", body)
	}
	if !strings.Contains(body, "data: [DONE]") {
		t.Fatalf("expected stream termination, got %s", body)
	}
}
