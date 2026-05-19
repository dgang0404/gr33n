package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestStreamUsage_CapturesTerminalUsageChunk simulates the OpenAI / Ollama
// contract: deltas in non-terminal chunks, a terminal chunk with empty
// choices + populated usage, then `data: [DONE]`. The streaming client
// must (a) forward every delta verbatim, (b) return the usage block, and
// (c) record stream_options.include_usage in the outgoing request body.
func TestStreamUsage_CapturesTerminalUsageChunk(t *testing.T) {
	t.Parallel()
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		// Two delta chunks, then a terminal usage chunk, then [DONE].
		_, _ = io.WriteString(w, `data: {"choices":[{"delta":{"content":"Hello"}}]}`+"\n\n")
		_, _ = io.WriteString(w, `data: {"choices":[{"delta":{"content":" world"}}]}`+"\n\n")
		_, _ = io.WriteString(w, `data: {"choices":[],"usage":{"prompt_tokens":17,"completion_tokens":4,"total_tokens":21}}`+"\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		BaseURL:    srv.URL,
		Model:      "fake",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
	var got strings.Builder
	usage, err := c.ChatCompletionStreamMessagesWithUsage(
		context.Background(),
		[]Message{{Role: "user", Content: "hi"}},
		func(s string) { got.WriteString(s) },
	)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}
	if got.String() != "Hello world" {
		t.Fatalf("delta content: %q", got.String())
	}
	if usage.PromptTokens != 17 || usage.CompletionTokens != 4 || usage.TotalTokens != 21 {
		t.Fatalf("usage: %+v", usage)
	}

	// Confirm the outgoing request asked for stream_options.include_usage.
	var sent struct {
		Stream        bool `json:"stream"`
		StreamOptions struct {
			IncludeUsage bool `json:"include_usage"`
		} `json:"stream_options"`
	}
	if err := json.Unmarshal(capturedBody, &sent); err != nil {
		t.Fatalf("decode request body: %v (body=%s)", err, string(capturedBody))
	}
	if !sent.Stream {
		t.Fatalf("request must set stream=true")
	}
	if !sent.StreamOptions.IncludeUsage {
		t.Fatalf("request must set stream_options.include_usage=true (body=%s)", string(capturedBody))
	}
}

// TestStreamUsage_BackwardsCompatibleWhenNoUsage proves a server that
// ignores stream_options just emits deltas — the new method returns
// zero usage with nil error and the legacy method keeps working.
func TestStreamUsage_BackwardsCompatibleWhenNoUsage(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `data: {"choices":[{"delta":{"content":"plain"}}]}`+"\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	t.Cleanup(srv.Close)
	c := &Client{
		BaseURL:    srv.URL,
		Model:      "fake",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
	var got strings.Builder
	usage, err := c.ChatCompletionStreamMessagesWithUsage(
		context.Background(),
		[]Message{{Role: "user", Content: "hi"}},
		func(s string) { got.WriteString(s) },
	)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got.String() != "plain" {
		t.Fatalf("content: %q", got.String())
	}
	if usage != (Usage{}) {
		t.Fatalf("backend didn't report usage; expected zero, got %+v", usage)
	}
}

// TestStreamUsage_LegacyMethodStillWorks confirms the old signature wraps
// the new one and ignores usage cleanly.
func TestStreamUsage_LegacyMethodStillWorks(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `data: {"choices":[{"delta":{"content":"legacy"}}]}`+"\n\n")
		_, _ = io.WriteString(w, `data: {"choices":[],"usage":{"prompt_tokens":3,"completion_tokens":1,"total_tokens":4}}`+"\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	t.Cleanup(srv.Close)
	c := &Client{
		BaseURL:    srv.URL,
		Model:      "fake",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
	var got strings.Builder
	err := c.ChatCompletionStreamMessages(
		context.Background(),
		[]Message{{Role: "user", Content: "hi"}},
		func(s string) { got.WriteString(s) },
	)
	if err != nil {
		t.Fatalf("legacy stream: %v", err)
	}
	if got.String() != "legacy" {
		t.Fatalf("content: %q", got.String())
	}
}
