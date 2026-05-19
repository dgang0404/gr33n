package llm

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsTransientLLMError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"context canceled", context.Canceled, false},
		{"context deadline exceeded", context.DeadlineExceeded, true},
		{"http 500", &HTTPStatusError{StatusCode: 500}, true},
		{"http 502", &HTTPStatusError{StatusCode: 502}, true},
		{"http 503", &HTTPStatusError{StatusCode: 503}, true},
		{"http 504", &HTTPStatusError{StatusCode: 504}, true},
		{"http 429", &HTTPStatusError{StatusCode: 429}, true},
		{"http 408", &HTTPStatusError{StatusCode: 408}, true},
		{"http 400", &HTTPStatusError{StatusCode: 400}, false},
		{"http 401", &HTTPStatusError{StatusCode: 401}, false},
		{"http 403", &HTTPStatusError{StatusCode: 403}, false},
		{"http 404", &HTTPStatusError{StatusCode: 404}, false},
		{"net op error", &net.OpError{Op: "dial", Err: errors.New("connection refused")}, true},
		{"plain error", errors.New("decode boom"), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsTransientLLMError(tc.err); got != tc.want {
				t.Fatalf("IsTransientLLMError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestRetryOp_TransientRetried(t *testing.T) {
	t.Parallel()
	var attempts int32
	cfg := RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		Sleeper:        func(ctx context.Context, d time.Duration) error { return nil },
	}
	err := retryOp(context.Background(), cfg, func(_ int) error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			return &HTTPStatusError{StatusCode: 503}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestRetryOp_PermanentNotRetried(t *testing.T) {
	t.Parallel()
	var attempts int32
	cfg := RetryConfig{
		MaxAttempts:    5,
		InitialBackoff: 1 * time.Millisecond,
		Sleeper:        func(ctx context.Context, d time.Duration) error { return nil },
	}
	wantErr := &HTTPStatusError{StatusCode: 400, Body: "bad input"}
	err := retryOp(context.Background(), cfg, func(_ int) error {
		atomic.AddInt32(&attempts, 1)
		return wantErr
	})
	if !errors.Is(err, wantErr) && err != wantErr { // *HTTPStatusError is concrete
		t.Fatalf("expected the 400 to surface, got %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("permanent error should not retry, attempts=%d", got)
	}
}

func TestRetryOp_MaxAttemptsHonoured(t *testing.T) {
	t.Parallel()
	var attempts int32
	cfg := RetryConfig{
		MaxAttempts:    4,
		InitialBackoff: 1 * time.Millisecond,
		Sleeper:        func(ctx context.Context, d time.Duration) error { return nil },
	}
	err := retryOp(context.Background(), cfg, func(_ int) error {
		atomic.AddInt32(&attempts, 1)
		return &HTTPStatusError{StatusCode: 502}
	})
	if err == nil {
		t.Fatalf("expected final transient error, got nil")
	}
	if got := atomic.LoadInt32(&attempts); got != 4 {
		t.Fatalf("expected 4 attempts, got %d", got)
	}
}

func TestRetryOp_ContextCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cfg := RetryConfig{
		MaxAttempts:    5,
		InitialBackoff: 1 * time.Millisecond,
		Sleeper: func(ctx context.Context, d time.Duration) error {
			// Caller cancels during the first backoff — confirm we honour it
			// instead of starting another attempt.
			cancel()
			return ctx.Err()
		},
	}
	var attempts int32
	err := retryOp(ctx, cfg, func(_ int) error {
		atomic.AddInt32(&attempts, 1)
		return &HTTPStatusError{StatusCode: 503}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("expected only the first attempt before cancel, got %d", got)
	}
}

// Integration-style: hit a test HTTP server that fails twice with 503 then
// returns a healthy chat response. Verifies the *Client end-to-end retry.
func TestClient_ChatCompletion_RetriesTransient(t *testing.T) {
	t.Parallel()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, `{"error":"overloaded"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"hi there"}}],"usage":{"prompt_tokens":5,"completion_tokens":2,"total_tokens":7}}`)
	}))
	defer srv.Close()

	c := &Client{
		BaseURL:    srv.URL,
		Model:      "test-model",
		HTTPClient: srv.Client(),
		Retry: RetryConfig{
			MaxAttempts:    3,
			InitialBackoff: 1 * time.Millisecond,
			MaxBackoff:     5 * time.Millisecond,
			Sleeper:        func(ctx context.Context, d time.Duration) error { return nil },
		},
	}
	ans, usage, err := c.ChatCompletionMessagesWithUsage(context.Background(), []Message{
		{Role: "user", Content: "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ans != "hi there" {
		t.Fatalf("unexpected answer: %q", ans)
	}
	if usage.PromptTokens != 5 || usage.CompletionTokens != 2 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 server calls, got %d", got)
	}
}

func TestClient_ChatCompletion_DoesNotRetry4xx(t *testing.T) {
	t.Parallel()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `bad model`)
	}))
	defer srv.Close()

	c := &Client{
		BaseURL:    srv.URL,
		Model:      "test-model",
		HTTPClient: srv.Client(),
		Retry: RetryConfig{
			MaxAttempts:    3,
			InitialBackoff: 1 * time.Millisecond,
			Sleeper:        func(ctx context.Context, d time.Duration) error { return nil },
		},
	}
	_, _, err := c.ChatCompletionMessagesWithUsage(context.Background(), []Message{{Role: "user", Content: "x"}})
	if err == nil {
		t.Fatalf("expected error from 400 response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected error to mention 400, got: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("4xx must not retry, got %d calls", got)
	}
}

// Streaming connect retry: first attempt 503, second succeeds and emits a
// single delta + [DONE]. Verifies pre-first-delta retries happen.
func TestClient_Stream_RetriesConnect(t *testing.T) {
	t.Parallel()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, `overloaded`)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	c := &Client{
		BaseURL:    srv.URL,
		Model:      "test-model",
		HTTPClient: srv.Client(),
		Retry: RetryConfig{
			MaxAttempts:    2,
			InitialBackoff: 1 * time.Millisecond,
			Sleeper:        func(ctx context.Context, d time.Duration) error { return nil },
		},
	}
	var got strings.Builder
	err := c.ChatCompletionStreamMessages(context.Background(), []Message{{Role: "user", Content: "hi"}}, func(s string) {
		got.WriteString(s)
	})
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}
	if got.String() != "hello" {
		t.Fatalf("unexpected stream content: %q", got.String())
	}
	if c := atomic.LoadInt32(&calls); c != 2 {
		t.Fatalf("expected 2 connect calls, got %d", c)
	}
}

func TestRetryConfigFromEnv(t *testing.T) {
	t.Setenv("LLM_RETRY_MAX_ATTEMPTS", "5")
	t.Setenv("LLM_RETRY_BACKOFF_MS", "250")
	cfg := retryConfigFromEnv()
	if cfg.MaxAttempts != 5 {
		t.Fatalf("MaxAttempts = %d, want 5", cfg.MaxAttempts)
	}
	if cfg.InitialBackoff != 250*time.Millisecond {
		t.Fatalf("InitialBackoff = %v, want 250ms", cfg.InitialBackoff)
	}

	// Out-of-range and garbage inputs fall back to defaults.
	t.Setenv("LLM_RETRY_MAX_ATTEMPTS", "abc")
	t.Setenv("LLM_RETRY_BACKOFF_MS", "1")
	cfg = retryConfigFromEnv()
	if cfg.MaxAttempts != DefaultRetryMaxAttempts {
		t.Fatalf("bad attempts -> default; got %d", cfg.MaxAttempts)
	}
	if cfg.InitialBackoff != DefaultRetryInitialBackoff {
		t.Fatalf("bad backoff -> default; got %v", cfg.InitialBackoff)
	}

	// Upper-bound clamps.
	t.Setenv("LLM_RETRY_MAX_ATTEMPTS", "999")
	t.Setenv("LLM_RETRY_BACKOFF_MS", "999999")
	cfg = retryConfigFromEnv()
	if cfg.MaxAttempts != 8 {
		t.Fatalf("attempts clamp = %d, want 8", cfg.MaxAttempts)
	}
	if cfg.InitialBackoff != 30*time.Second {
		t.Fatalf("backoff clamp = %v, want 30s", cfg.InitialBackoff)
	}
}
