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
	"gr33n-api/internal/authctx"
	"gr33n-api/internal/rag/llm"
)

type fakeLLM struct {
	reply        string
	usage        llm.Usage
	failWith     error
	lastUser     string
	lastSys      string
	lastMessages []llm.Message
	callCount    int
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
func (f *fakeLLM) ChatCompletionMessages(_ context.Context, messages []llm.Message) (string, error) {
	f.callCount++
	f.lastMessages = messages
	if len(messages) > 0 {
		f.lastSys = messages[0].TextContent()
		f.lastUser = messages[len(messages)-1].TextContent()
	}
	if f.failWith != nil {
		return "", f.failWith
	}
	return f.reply, nil
}
func (f *fakeLLM) ChatCompletionMessagesWithUsage(_ context.Context, messages []llm.Message) (string, llm.Usage, error) {
	f.callCount++
	f.lastMessages = messages
	if len(messages) > 0 {
		f.lastSys = messages[0].TextContent()
		f.lastUser = messages[len(messages)-1].TextContent()
	}
	if f.failWith != nil {
		return "", llm.Usage{}, f.failWith
	}
	return f.reply, f.usage, nil
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

func TestPostV1_VisionAttachmentsWithoutVisionClient503(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, &fakeLLM{reply: "ok"}, nil)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bytes.NewBufferString(`{"message":"leaves?","farm_id":1,"attachment_ids":[42]}`))
	req = req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
	rec := httptest.NewRecorder()
	h.PostV1(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "vision") {
		t.Fatalf("expected vision config hint, got %s", rec.Body.String())
	}
}

func TestPostV1_VisionAttachmentsRequireFarmID400(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, &fakeLLM{reply: "ok"}, nil).
		WithVisionLLM(&fakeLLM{reply: "vision"})
	rec := doPost(t, h, `{"message":"leaves?","attachment_ids":[42]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d: %s", rec.Code, rec.Body.String())
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
	const sid = "11111111-1111-4111-8111-111111111111"
	rec := doPost(t, h, `{"message":"What should I do this morning?","session_id":"`+sid+`"}`)
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
	if resp.SessionID != sid {
		t.Fatalf("expected session_id echo %q, got %q", sid, resp.SessionID)
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

func TestPostV1_GeneratesSessionWhenMissing(t *testing.T) {
	llm := &fakeLLM{reply: "Hi."}
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, llm, nil)
	rec := doPost(t, h, `{"message":"hi"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp postResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	// Should be a UUID — at minimum, non-empty and 36 chars with dashes.
	if len(resp.SessionID) != 36 || strings.Count(resp.SessionID, "-") != 4 {
		t.Fatalf("expected generated UUID session_id, got %q", resp.SessionID)
	}
}

func TestPostV1_InvalidSessionID400(t *testing.T) {
	llm := &fakeLLM{reply: "x"}
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, llm, nil)
	rec := doPost(t, h, `{"message":"hi","session_id":"not-a-uuid"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestPostV1_ReturnsTokenUsage(t *testing.T) {
	fl := &fakeLLM{
		reply: "ok",
		usage: llm.Usage{PromptTokens: 42, CompletionTokens: 17, TotalTokens: 59},
	}
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, fl, nil)
	rec := doPost(t, h, `{"message":"hi"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp postResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.PromptTokens != 42 || resp.CompletionTokens != 17 {
		t.Fatalf("expected token usage to flow through, got %+v", resp)
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

// fakeStreamingLLM implements ChatCompleter + MessagesChatCompleter +
// MessagesStreamingChatCompleter so the handler picks the streaming branch
// when stream=true.
type fakeStreamingLLM struct {
	deltas       []string
	failOn       int  // 1-indexed delta to fail after; 0 disables
	called       bool // true if streaming variant ran
	lastMessages []llm.Message
}

func (f *fakeStreamingLLM) ChatCompletion(_ context.Context, _, _ string) (string, error) {
	return strings.Join(f.deltas, ""), nil
}
func (f *fakeStreamingLLM) ChatCompletionMessages(_ context.Context, messages []llm.Message) (string, error) {
	f.lastMessages = messages
	return strings.Join(f.deltas, ""), nil
}
func (f *fakeStreamingLLM) ModelLabel() string { return "fake-streamer" }
func (f *fakeStreamingLLM) ChatCompletionStream(_ context.Context, _, _ string, onDelta func(string)) error {
	return f.ChatCompletionStreamMessages(nil, nil, onDelta)
}
func (f *fakeStreamingLLM) ChatCompletionStreamMessages(_ context.Context, messages []llm.Message, onDelta func(string)) error {
	f.called = true
	f.lastMessages = messages
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
	const sid = "22222222-2222-4222-8222-222222222222"
	rec := doPost(t, h, `{"message":"hi","stream":true,"session_id":"`+sid+`"}`)
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
		`"session_id":"` + sid + `"`,
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

// fakeUsageStreamingLLM implements UsageAwareStreamingChatCompleter so the
// handler picks the usage-aware streaming branch (Phase 27 WS5 follow-up).
type fakeUsageStreamingLLM struct {
	deltas []string
	usage  llm.Usage
	called bool
}

func (f *fakeUsageStreamingLLM) ChatCompletion(_ context.Context, _, _ string) (string, error) {
	return strings.Join(f.deltas, ""), nil
}
func (f *fakeUsageStreamingLLM) ChatCompletionMessages(_ context.Context, _ []llm.Message) (string, error) {
	return strings.Join(f.deltas, ""), nil
}
func (f *fakeUsageStreamingLLM) ModelLabel() string { return "fake-usage-streamer" }
func (f *fakeUsageStreamingLLM) ChatCompletionStream(_ context.Context, _, _ string, onDelta func(string)) error {
	_, err := f.ChatCompletionStreamMessagesWithUsage(nil, nil, onDelta)
	return err
}
func (f *fakeUsageStreamingLLM) ChatCompletionStreamMessages(_ context.Context, _ []llm.Message, onDelta func(string)) error {
	_, err := f.ChatCompletionStreamMessagesWithUsage(nil, nil, onDelta)
	return err
}
func (f *fakeUsageStreamingLLM) ChatCompletionStreamMessagesWithUsage(_ context.Context, _ []llm.Message, onDelta func(string)) (llm.Usage, error) {
	f.called = true
	for _, d := range f.deltas {
		onDelta(d)
	}
	return f.usage, nil
}

func TestPostV1_StreamUsageFlowsToDoneEvent(t *testing.T) {
	fl := &fakeUsageStreamingLLM{
		deltas: []string{"Hello", " ", "world"},
		usage:  llm.Usage{PromptTokens: 42, CompletionTokens: 8, TotalTokens: 50},
	}
	h := NewHandlerWithDeps(ai.Config{Enabled: true}, nil, fl, nil)
	const sid = "33333333-3333-4333-8333-333333333333"
	rec := doPost(t, h, `{"message":"hi","stream":true,"session_id":"`+sid+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !fl.called {
		t.Fatal("usage-aware streaming branch was not selected")
	}
	body := rec.Body.String()
	for _, want := range []string{
		`event: done`,
		`"prompt_tokens":42`,
		`"completion_tokens":8`,
		`"answer":"Hello world"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q\nfull body:\n%s", want, body)
		}
	}
}
