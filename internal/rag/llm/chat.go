// Package llm provides OpenAI-compatible chat completions for RAG answer synthesis (Phase 24 WS5).
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultTimeout = 120 * time.Second

// Client calls POST /v1/chat/completions (OpenAI-compatible; LM Studio, local gateways).
type Client struct {
	BaseURL     string
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
	HTTPClient  *http.Client
	// Retry controls transient-failure backoff. Zero value disables retry.
	// Defaults populated by NewChatClientFromEnv (LLM_RETRY_MAX_ATTEMPTS /
	// LLM_RETRY_BACKOFF_MS). Phase 27 WS3 follow-up.
	Retry RetryConfig
}

func temperatureFromEnv() float64 {
	s := strings.TrimSpace(os.Getenv("LLM_TEMPERATURE"))
	if s == "" {
		return 0.2
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.2
	}
	if f < 0 {
		return 0
	}
	if f > 2 {
		return 2
	}
	return f
}

func maxTokensFromEnv() int {
	s := strings.TrimSpace(os.Getenv("LLM_MAX_TOKENS"))
	if s == "" {
		return 1024
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 1024
	}
	if n > 8192 {
		return 8192
	}
	return n
}

// NewChatClientFromEnv requires LLM_MODEL and LLM_BASE_URL (no implicit default URL so operators
// opt in to a specific endpoint — local LM Studio vs cloud).
func timeoutFromEnv() time.Duration {
	s := strings.TrimSpace(os.Getenv("LLM_TIMEOUT_SECONDS"))
	if s == "" {
		return DefaultTimeout
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return DefaultTimeout
	}
	return time.Duration(n) * time.Second
}

// NewChatClientFromEnv requires LLM_MODEL and LLM_BASE_URL (no implicit default URL so operators
// opt in to a specific endpoint — local LM Studio vs cloud).
func NewChatClientFromEnv() (*Client, error) {
	model := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	base := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	if model == "" || base == "" {
		return nil, errors.New("LLM synthesis requires LLM_MODEL and LLM_BASE_URL")
	}
	key := strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	return &Client{
		BaseURL:     strings.TrimSuffix(base, "/"),
		APIKey:      key,
		Model:       model,
		Temperature: temperatureFromEnv(),
		MaxTokens:   maxTokensFromEnv(),
		HTTPClient:  &http.Client{Timeout: timeoutFromEnv()},
		Retry:       retryConfigFromEnv(),
	}, nil
}

// Message is a single role + content entry in an OpenAI-style chat history.
// Use "system", "user", or "assistant" for Role. Phase 27 WS5 follow-up uses
// this for multi-turn requests (history + current user message).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatMessage = Message

// Usage captures the token accounting OpenAI-compatible servers return on
// each completion. Zero values are valid — backends that don't report usage
// (some Ollama builds, mocks) just leave the counters at 0.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage *Usage `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatCompletion runs a single turn (system + user). Equivalent to
// ChatCompletionMessages with a two-element slice; preserved for callers that
// don't need multi-turn history.
func (c *Client) ChatCompletion(ctx context.Context, system, user string) (string, error) {
	answer, _, err := c.ChatCompletionMessagesWithUsage(ctx, []Message{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	})
	return answer, err
}

// ChatCompletionMessages runs a multi-turn completion. Messages are passed
// through verbatim — callers are responsible for the system / history /
// current-user ordering. Phase 27 WS5 uses this for session history replay.
func (c *Client) ChatCompletionMessages(ctx context.Context, messages []Message) (string, error) {
	answer, _, err := c.ChatCompletionMessagesWithUsage(ctx, messages)
	return answer, err
}

// ChatCompletionMessagesWithUsage is the canonical non-streaming path. It
// returns both the answer text and the OpenAI-style token usage block when
// the backend reports it (Usage{} on backends that don't).
//
// Transient failures (HTTP 408/425/429/5xx, net.OpError, per-attempt
// DeadlineExceeded) are retried per c.Retry with exponential backoff +
// jitter. Caller cancellation (ctx.Done()) short-circuits between attempts
// and is never retried. Phase 27 WS3 follow-up.
func (c *Client) ChatCompletionMessagesWithUsage(ctx context.Context, messages []Message) (string, Usage, error) {
	if len(messages) == 0 {
		return "", Usage{}, errors.New("messages required")
	}
	body := chatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: c.Temperature,
		MaxTokens:   c.MaxTokens,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", Usage{}, err
	}

	var (
		answer string
		usage  Usage
	)
	rerr := retryOp(ctx, c.Retry, func(_ int) error {
		a, u, err := c.doChatOnce(ctx, raw)
		if err != nil {
			return err
		}
		answer, usage = a, u
		return nil
	})
	if rerr != nil {
		return "", Usage{}, rerr
	}
	return answer, usage, nil
}

// doChatOnce performs a single non-streaming HTTP attempt. The retry loop
// decides what to do with the error.
func (c *Client) doChatOnce(ctx context.Context, raw []byte) (string, Usage, error) {
	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", Usage{}, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", Usage{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", Usage{}, &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Body:       truncateErr(respBody, 512),
		}
	}
	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", Usage{}, fmt.Errorf("chat decode: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", Usage{}, errors.New(parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return "", Usage{}, errors.New("empty chat response")
	}
	var u Usage
	if parsed.Usage != nil {
		u = *parsed.Usage
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), u, nil
}

// ModelLabel returns the configured chat model id for API responses.
func (c *Client) ModelLabel() string {
	return c.Model
}

func truncateErr(b []byte, n int) string {
	s := string(b)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ChatCompleter is implemented by *Client for answer synthesis; tests may inject mocks.
type ChatCompleter interface {
	ChatCompletion(ctx context.Context, system, user string) (string, error)
	// ModelLabel identifies the chat model in JSON responses (e.g. OpenAI model id).
	ModelLabel() string
}

// MessagesChatCompleter is the multi-turn non-streaming surface (Phase 27 WS5 follow-up).
type MessagesChatCompleter interface {
	ChatCompletionMessages(ctx context.Context, messages []Message) (string, error)
}

// UsageAwareChatCompleter is the optional non-streaming surface that returns
// token accounting alongside the answer. Phase 27 WS5 token-usage slice.
type UsageAwareChatCompleter interface {
	ChatCompletionMessagesWithUsage(ctx context.Context, messages []Message) (string, Usage, error)
}

// StreamingChatCompleter is the optional Phase 27 WS5 v3 streaming surface.
// ChatCompletionStream runs a single turn (system + user) and invokes onDelta
// for each incremental text token returned by the OpenAI-compatible SSE stream.
// Implementations must:
//   - honour ctx cancellation (return immediately when the caller disconnects);
//   - never call onDelta after returning;
//   - return a non-nil error on non-2xx HTTP, malformed SSE, or transport failure.
//
// On success the full text was streamed via onDelta and the call returns nil.
type StreamingChatCompleter interface {
	ChatCompletionStream(ctx context.Context, system, user string, onDelta func(string)) error
}

// MessagesStreamingChatCompleter is the multi-turn streaming surface.
type MessagesStreamingChatCompleter interface {
	ChatCompletionStreamMessages(ctx context.Context, messages []Message, onDelta func(string)) error
}

// UsageAwareStreamingChatCompleter is the multi-turn streaming surface that
// also returns the OpenAI-style token-usage block from the terminal SSE
// chunk (Phase 27 WS5 follow-up — stream_options.include_usage).
// Implementations that don't get usage from the upstream return Usage{}
// with nil error — callers must treat zero usage as "not reported" rather
// than "zero tokens used".
type UsageAwareStreamingChatCompleter interface {
	ChatCompletionStreamMessagesWithUsage(ctx context.Context, messages []Message, onDelta func(string)) (Usage, error)
}

type sseStreamRequest struct {
	Model         string            `json:"model"`
	Messages      []Message         `json:"messages"`
	Temperature   float64           `json:"temperature,omitempty"`
	MaxTokens     int               `json:"max_tokens,omitempty"`
	Stream        bool              `json:"stream"`
	StreamOptions *sseStreamOptions `json:"stream_options,omitempty"`
}

// sseStreamOptions opts in to the terminal-usage chunk that OpenAI-compatible
// servers emit when stream_options.include_usage is true. Ollama (>= 0.3.x)
// honours the same field; servers that don't recognise it ignore extra JSON.
type sseStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// sseChunk is one parsed `data: {...}` event. The terminal usage chunk has
// `choices: []` and a populated `usage` block — i.e. no delta to forward but
// authoritative token counts. Non-terminal chunks carry deltas and (on most
// backends) leave usage at zero / null.
type sseChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *Usage `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatCompletionStream streams tokens from /v1/chat/completions with `stream: true`.
// Convenience wrapper for the single-turn (system + user) case; delegates to
// ChatCompletionStreamMessages.
func (c *Client) ChatCompletionStream(ctx context.Context, system, user string, onDelta func(string)) error {
	return c.ChatCompletionStreamMessages(ctx, []Message{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}, onDelta)
}

// ChatCompletionStreamMessages streams tokens from /v1/chat/completions with the
// given multi-turn messages slice. Thin wrapper that discards token usage —
// callers that need usage should use ChatCompletionStreamMessagesWithUsage.
func (c *Client) ChatCompletionStreamMessages(ctx context.Context, messages []Message, onDelta func(string)) error {
	_, err := c.ChatCompletionStreamMessagesWithUsage(ctx, messages, onDelta)
	return err
}

// ChatCompletionStreamMessagesWithUsage streams tokens from
// /v1/chat/completions and returns the OpenAI-style token-usage block when
// the server emits one in the terminal SSE chunk. Phase 27 WS5 follow-up:
// the request body now sets `stream_options.include_usage: true`, which
// causes the upstream to send a final `data: {...usage: {...}}` chunk just
// before `data: [DONE]`. Servers that don't recognise the field ignore it
// and the returned Usage stays zero — backwards compatible.
//
// Retry semantics (Phase 27 WS3 follow-up): the **connect + status-check**
// phase is retried per c.Retry, since failing there hasn't emitted any
// content to the caller yet. Once the body has streamed at least one delta
// to onDelta, retrying would duplicate visible text — so any later error
// surfaces directly.
func (c *Client) ChatCompletionStreamMessagesWithUsage(ctx context.Context, messages []Message, onDelta func(string)) (Usage, error) {
	if onDelta == nil {
		return Usage{}, errors.New("onDelta callback required")
	}
	if len(messages) == 0 {
		return Usage{}, errors.New("messages required")
	}
	body := sseStreamRequest{
		Model:         c.Model,
		Messages:      messages,
		Temperature:   c.Temperature,
		MaxTokens:     c.MaxTokens,
		Stream:        true,
		StreamOptions: &sseStreamOptions{IncludeUsage: true},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return Usage{}, err
	}

	// openStream is fully replayable — it doesn't read or forward any SSE
	// chunks before returning. Once we leave the retry loop with a healthy
	// response, mid-stream errors fall through directly to the caller
	// because retrying after deltas have been forwarded would duplicate
	// visible text.
	var resp *http.Response
	connectErr := retryOp(ctx, c.Retry, func(_ int) error {
		r, err := c.openStream(ctx, raw)
		if err != nil {
			return err
		}
		resp = r
		return nil
	})
	if connectErr != nil {
		return Usage{}, connectErr
	}
	defer resp.Body.Close()

	var usage Usage
	br := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return usage, ctx.Err()
		default:
		}
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			payload := strings.TrimSpace(line)
			if payload == "" || !strings.HasPrefix(payload, "data:") {
				// SSE comments / blank separators — ignore.
			} else {
				data := strings.TrimSpace(strings.TrimPrefix(payload, "data:"))
				if data == "[DONE]" {
					return usage, nil
				}
				var chunk sseChunk
				if jerr := json.Unmarshal([]byte(data), &chunk); jerr != nil {
					// Don't fail the whole stream on a single odd line — log via err once at end.
					continue
				}
				if chunk.Error != nil && chunk.Error.Message != "" {
					return usage, errors.New(chunk.Error.Message)
				}
				// Most backends only populate usage on the terminal chunk
				// (choices: [] + usage: {...}). We refresh on every chunk
				// that carries non-zero usage so partial usage updates work
				// too — last-write-wins matches OpenAI's contract.
				if chunk.Usage != nil && (chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0 || chunk.Usage.TotalTokens > 0) {
					usage = *chunk.Usage
				}
				for _, ch := range chunk.Choices {
					if ch.Delta.Content != "" {
						onDelta(ch.Delta.Content)
					}
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return usage, nil
			}
			return usage, err
		}
	}
}

// openStream performs one streaming connect + status check. It returns the
// response with the body open on success so the caller can read SSE chunks.
func (c *Client) openStream(ctx context.Context, raw []byte) (*http.Response, error) {
	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()
		return nil, &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Body:       truncateErr(preview, 512),
		}
	}
	return resp, nil
}
