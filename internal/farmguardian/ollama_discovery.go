package farmguardian

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ModelInfo describes one model reported by the Ollama runtime.
type ModelInfo struct {
	Name           string `json:"name"`
	ContextWindow  int    `json:"context_window"`
	ParameterCount int64  `json:"parameter_count,omitempty"`
	SpeedClass     string `json:"speed_class"`
}

type ollamaTagsResponse struct {
	Models []ollamaTagModel `json:"models"`
}

type ollamaTagModel struct {
	Name    string `json:"name"`
	Model   string `json:"model"`
	Size    int64  `json:"size"`
	Details struct {
		ParameterSize string `json:"parameter_size"`
		Family        string `json:"family"`
	} `json:"details"`
}

var paramSizeRE = regexp.MustCompile(`(?i)([\d.]+)\s*([bmk])`)

// OllamaNativeBase strips a trailing /v1 from an OpenAI-compatible LLM base URL.
func OllamaNativeBase(openAICompatBase string) string {
	base := strings.TrimSuffix(strings.TrimSpace(openAICompatBase), "/")
	if strings.HasSuffix(strings.ToLower(base), "/v1") {
		return strings.TrimSuffix(base, "/v1")
	}
	return base
}

// DiscoverOllamaModels queries GET /api/tags on the Ollama native API.
func DiscoverOllamaModels(ctx context.Context, llmBaseURL string, client *http.Client) ([]ModelInfo, error) {
	base := OllamaNativeBase(llmBaseURL)
	if base == "" {
		return nil, fmt.Errorf("LLM_BASE_URL not set")
	}
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/tags", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama /api/tags: HTTP %d: %s", resp.StatusCode, truncateBytes(body, 256))
	}
	var parsed ollamaTagsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("ollama tags decode: %w", err)
	}
	out := make([]ModelInfo, 0, len(parsed.Models))
	for _, m := range parsed.Models {
		name := strings.TrimSpace(m.Name)
		if name == "" {
			name = strings.TrimSpace(m.Model)
		}
		if name == "" {
			continue
		}
		params := parseParameterCount(m.Details.ParameterSize)
		out = append(out, ModelInfo{
			Name:           name,
			ContextWindow:  0,
			ParameterCount: params,
			SpeedClass:     classifySpeedClass(name, params),
		})
	}
	return out, nil
}

func parseParameterCount(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	m := paramSizeRE.FindStringSubmatch(raw)
	if len(m) < 3 {
		return 0
	}
	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0
	}
	switch strings.ToLower(m[2]) {
	case "b":
		return int64(f)
	case "m":
		return int64(f / 1000)
	case "k":
		return int64(f / 1_000_000)
	default:
		return 0
	}
}

func classifySpeedClass(name string, paramB int64) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "r1") || strings.Contains(lower, "reason") || strings.Contains(lower, "deepseek-r") {
		return "reasoning"
	}
	if paramB > 0 && paramB <= 7 {
		return "fast"
	}
	if paramB > 30 {
		return "general"
	}
	return "general"
}

func truncateBytes(b []byte, n int) string {
	s := string(b)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// EnvServerDefaultModel returns the configured LLM_MODEL env default.
func EnvServerDefaultModel() string {
	return strings.TrimSpace(os.Getenv("LLM_MODEL"))
}
