package farmguardian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

type ollamaShowResponse struct {
	ModelInfo    map[string]any `json:"model_info"`
	Capabilities []string       `json:"capabilities"`
}

type modelShowDetails struct {
	ContextWindow          int
	EffectiveContextWindow int
	Capabilities           []string
}

type ollamaPullRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream"`
}

type ollamaPullProgress struct {
	Status string `json:"status"`
}

// parseContextLength scans Ollama model_info for *.context_length keys.
func parseContextLength(modelInfo map[string]any) int {
	if len(modelInfo) == 0 {
		return 0
	}
	max := 0
	for k, v := range modelInfo {
		if !strings.HasSuffix(k, ".context_length") {
			continue
		}
		n := jsonNumberInt(v)
		if n > max {
			max = n
		}
	}
	return max
}

func jsonNumberInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func fetchModelShowDetails(ctx context.Context, nativeBase, name string, client *http.Client) (modelShowDetails, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return modelShowDetails{}, fmt.Errorf("model name required")
	}
	body, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return modelShowDetails{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, nativeBase+"/api/show", bytes.NewReader(body))
	if err != nil {
		return modelShowDetails{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return modelShowDetails{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return modelShowDetails{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return modelShowDetails{}, fmt.Errorf("ollama /api/show: HTTP %d: %s", resp.StatusCode, truncateBytes(raw, 256))
	}
	var parsed ollamaShowResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return modelShowDetails{}, fmt.Errorf("ollama show decode: %w", err)
	}
	return modelShowDetails{
		ContextWindow:          parseContextLength(parsed.ModelInfo),
		EffectiveContextWindow: ResolveEffectiveContextWindow(name, parseContextLength(parsed.ModelInfo), parsed.ModelInfo),
		Capabilities:           append([]string(nil), parsed.Capabilities...),
	}, nil
}

func fetchModelContextWindow(ctx context.Context, nativeBase, name string, client *http.Client) (int, error) {
	details, err := fetchModelShowDetails(ctx, nativeBase, name, client)
	if err != nil {
		return 0, err
	}
	return details.ContextWindow, nil
}

// EnrichModelContextWindows fills ContextWindow via parallel POST /api/show calls.
func EnrichModelContextWindows(ctx context.Context, llmBaseURL string, models []ModelInfo, client *http.Client, concurrency int) []ModelInfo {
	if len(models) == 0 {
		return models
	}
	base := OllamaNativeBase(llmBaseURL)
	if base == "" {
		return models
	}
	if client == nil {
		client = http.DefaultClient
	}
	if concurrency < 1 {
		concurrency = defaultShowConcurrency
	}
	out := make([]ModelInfo, len(models))
	copy(out, models)
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i := range out {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()
			details, err := fetchModelShowDetails(ctx, base, out[i].Name, client)
			if err != nil {
				slog.Warn("guardian: ollama show failed", "model", out[i].Name, "err", err)
				return
			}
			out[i].ContextWindow = details.ContextWindow
			out[i].EffectiveContextWindow = details.EffectiveContextWindow
			if out[i].EffectiveContextWindow <= 0 && out[i].ContextWindow > 0 {
				out[i].EffectiveContextWindow = ResolveEffectiveContextWindow(out[i].Name, out[i].ContextWindow, nil)
			}
			out[i].Capabilities = details.Capabilities
		}(i)
	}
	wg.Wait()
	return out
}

// PullOllamaModel downloads a model via POST /api/pull (stream=false).
func PullOllamaModel(ctx context.Context, llmBaseURL, name string, client *http.Client) error {
	base := OllamaNativeBase(llmBaseURL)
	if base == "" {
		return fmt.Errorf("LLM_BASE_URL not set")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("model name required")
	}
	if client == nil {
		client = http.DefaultClient
	}
	raw, err := json.Marshal(ollamaPullRequest{Name: name, Stream: false})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/pull", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ollama /api/pull: HTTP %d: %s", resp.StatusCode, truncateBytes(body, 256))
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	var progress ollamaPullProgress
	if err := json.Unmarshal(body, &progress); err != nil {
		return nil
	}
	status := strings.ToLower(strings.TrimSpace(progress.Status))
	if status != "" && status != "success" {
		return fmt.Errorf("ollama pull status: %s", progress.Status)
	}
	return nil
}

// PullAndRefresh pulls a model then reloads the cache from Ollama.
func (c *ModelCache) PullAndRefresh(ctx context.Context, modelName string) error {
	base := LLMBaseURLFromEnv()
	if !IsLocalOllamaConfigured() {
		return fmt.Errorf("model pull requires a local Ollama LLM_BASE_URL")
	}
	pullCtx, cancel := context.WithTimeout(ctx, PullTimeoutFromEnv())
	defer cancel()
	client := &http.Client{Timeout: PullTimeoutFromEnv()}
	if err := PullOllamaModel(pullCtx, base, modelName, client); err != nil {
		return err
	}
	return c.RefreshFromEnv(ctx)
}
