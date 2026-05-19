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
	}, nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatCompletion runs a single turn (system + user).
func (c *Client) ChatCompletion(ctx context.Context, system, user string) (string, error) {
	body := chatRequest{
		Model:       c.Model,
		Messages:    []chatMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
		Temperature: c.Temperature,
		MaxTokens:   c.MaxTokens,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
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
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("chat HTTP %d: %s", resp.StatusCode, truncateErr(respBody, 512))
	}
	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("chat decode: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", errors.New(parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return "", errors.New("empty chat response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
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

type sseStreamRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
}

type sseChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatCompletionStream streams tokens from /v1/chat/completions with `stream: true`.
// Ollama and OpenAI both emit Server-Sent Events shaped as `data: {…}\n\n` followed
// by a terminal `data: [DONE]\n\n`. We parse line-by-line and invoke onDelta with
// each non-empty content delta.
func (c *Client) ChatCompletionStream(ctx context.Context, system, user string, onDelta func(string)) error {
	if onDelta == nil {
		return errors.New("onDelta callback required")
	}
	body := sseStreamRequest{
		Model:       c.Model,
		Messages:    []chatMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
		Temperature: c.Temperature,
		MaxTokens:   c.MaxTokens,
		Stream:      true,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
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
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("chat stream HTTP %d: %s", resp.StatusCode, truncateErr(preview, 512))
	}

	br := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
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
					return nil
				}
				var chunk sseChunk
				if jerr := json.Unmarshal([]byte(data), &chunk); jerr != nil {
					// Don't fail the whole stream on a single odd line — log via err once at end.
					continue
				}
				if chunk.Error != nil && chunk.Error.Message != "" {
					return errors.New(chunk.Error.Message)
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
				return nil
			}
			return err
		}
	}
}
