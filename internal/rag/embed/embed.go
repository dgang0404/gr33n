// Package embed abstracts remote embedding providers (OpenAI-compatible HTTP API).
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	DefaultBaseURL      = "https://api.openai.com/v1"
	DefaultModel        = "text-embedding-3-small"
	DefaultExpectedDims = 1536
)

// Embedder produces float32 vectors for indexing (dimension must match DB schema).
type Embedder interface {
	ModelID() string
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// Client calls OpenAI-compatible POST /embeddings (LM Studio and others).
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
	Dim     int
}

// OpenAI embeddings response JSON (subset).
type embeddingsResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func NewOpenAICompatibleFromEnv() (*Client, error) {
	apiKey := strings.TrimSpace(os.Getenv("EMBEDDING_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("EMBEDDING_API_KEY is required for embeddings")
	}
	base := strings.TrimSpace(os.Getenv("EMBEDDING_BASE_URL"))
	if base == "" {
		base = DefaultBaseURL
	}
	model := strings.TrimSpace(os.Getenv("EMBEDDING_MODEL"))
	if model == "" {
		model = DefaultModel
	}
	dim := DefaultExpectedDims
	// Optional override for non-1536 models — must align with migrations.
	if s := strings.TrimSpace(os.Getenv("EMBEDDING_DIMENSION")); s != "" {
		var parsed int
		if _, err := fmt.Sscanf(s, "%d", &parsed); err != nil || parsed <= 0 {
			return nil, fmt.Errorf("invalid EMBEDDING_DIMENSION %q", s)
		}
		dim = parsed
	}
	c := &Client{
		BaseURL: strings.TrimSuffix(base, "/"),
		APIKey:  apiKey,
		Model:   model,
		Client:  &http.Client{Timeout: 120 * time.Second},
		Dim:     dim,
	}
	return c, nil
}

func (c *Client) ModelID() string {
	return c.Model
}

func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body := map[string]any{
		"model": c.Model,
		"input": texts,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := c.BaseURL + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embeddings HTTP %d: %s", resp.StatusCode, truncateForErr(respBody, 512))
	}
	var parsed embeddingsResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("embeddings decode: %w", err)
	}
	if len(parsed.Data) != len(texts) {
		return nil, fmt.Errorf("embeddings count mismatch: got %d want %d", len(parsed.Data), len(texts))
	}
	out := make([][]float32, len(parsed.Data))
	for i := range parsed.Data {
		el := parsed.Data[i].Embedding
		if len(el) != c.Dim {
			return nil, fmt.Errorf("embedding dim %d != expected %d (model %s)", len(el), c.Dim, c.ModelID())
		}
		row := make([]float32, len(el))
		for j, v := range el {
			row[j] = float32(v)
		}
		out[i] = row
	}
	return out, nil
}

func truncateForErr(b []byte, n int) string {
	s := string(b)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
