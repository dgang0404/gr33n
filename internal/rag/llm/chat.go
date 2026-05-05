// Package llm provides OpenAI-compatible chat completions for RAG answer synthesis (Phase 24 WS5).
package llm

import (
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
		HTTPClient:  &http.Client{Timeout: DefaultTimeout},
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
