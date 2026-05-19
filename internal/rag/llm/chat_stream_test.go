package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newSSETestServer(t *testing.T, payload string, status int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(payload))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func makeClient(srv *httptest.Server) *Client {
	return &Client{
		BaseURL:    srv.URL,
		Model:      "fake",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func TestChatCompletionStream_HappyPath(t *testing.T) {
	body := strings.Join([]string{
		`data: {"choices":[{"delta":{"content":"Hello"}}]}`,
		``,
		`data: {"choices":[{"delta":{"content":", world"}}]}`,
		``,
		`data: {"choices":[{"delta":{"content":"!"}}]}`,
		``,
		`data: [DONE]`,
		``,
	}, "\n")
	srv := newSSETestServer(t, body, http.StatusOK)
	c := makeClient(srv)

	var got strings.Builder
	err := c.ChatCompletionStream(context.Background(), "sys", "user", func(s string) {
		got.WriteString(s)
	})
	if err != nil {
		t.Fatalf("stream err: %v", err)
	}
	if got.String() != "Hello, world!" {
		t.Fatalf("got %q want %q", got.String(), "Hello, world!")
	}
}

func TestChatCompletionStream_RejectsNilCallback(t *testing.T) {
	srv := newSSETestServer(t, "", http.StatusOK)
	c := makeClient(srv)
	if err := c.ChatCompletionStream(context.Background(), "s", "u", nil); err == nil {
		t.Fatal("expected error for nil onDelta")
	}
}

func TestChatCompletionStream_NonOKStatus(t *testing.T) {
	srv := newSSETestServer(t, "boom", http.StatusInternalServerError)
	c := makeClient(srv)
	err := c.ChatCompletionStream(context.Background(), "s", "u", func(string) {})
	if err == nil || !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("expected HTTP 500 error, got %v", err)
	}
}

func TestChatCompletionStream_ContextCancel(t *testing.T) {
	// Server holds the connection open without sending [DONE]; we cancel ctx.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"slow\"}}]}\n\n"))
		if flusher != nil {
			flusher.Flush()
		}
		time.Sleep(500 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)
	c := makeClient(srv)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	err := c.ChatCompletionStream(ctx, "s", "u", func(string) {})
	if err == nil {
		t.Fatal("expected error after context cancel")
	}
}

func TestChatCompletionStream_ErrorFromUpstream(t *testing.T) {
	body := `data: {"error":{"message":"upstream blew up"}}` + "\n\n"
	srv := newSSETestServer(t, body, http.StatusOK)
	c := makeClient(srv)
	err := c.ChatCompletionStream(context.Background(), "s", "u", func(string) {})
	if err == nil || !strings.Contains(err.Error(), "upstream blew up") {
		t.Fatalf("expected upstream error, got %v", err)
	}
}

func TestChatCompletionStream_TolerantToOddLines(t *testing.T) {
	// Some servers emit comments (':keepalive') or blank padding — neither should fail the stream.
	body := strings.Join([]string{
		`:keepalive`,
		``,
		`data: not-json`,
		``,
		`data: {"choices":[{"delta":{"content":"ok"}}]}`,
		``,
		`data: [DONE]`,
		``,
	}, "\n")
	srv := newSSETestServer(t, body, http.StatusOK)
	c := makeClient(srv)

	var got strings.Builder
	if err := c.ChatCompletionStream(context.Background(), "s", "u", func(s string) { got.WriteString(s) }); err != nil {
		t.Fatalf("stream err: %v", err)
	}
	if got.String() != "ok" {
		t.Fatalf("got %q want %q", got.String(), "ok")
	}
}
