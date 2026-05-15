// Package ai holds Phase 27 deployment flags for LLM-backed features (Farm Guardian, RAG synthesis).
package ai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Config is loaded once at API startup from environment variables.
type Config struct {
	// Enabled is the master switch (AI_ENABLED). When false, synthesis and /v1/chat are off.
	Enabled bool
}

// LoadConfigFromEnv parses AI_ENABLED.
//
// If AI_ENABLED is unset, Enabled defaults to true so existing deployments that set LLM_*
// keep RAG answer synthesis without a config change. Explicit false/0/off disables AI.
func LoadConfigFromEnv() Config {
	raw, set := os.LookupEnv("AI_ENABLED")
	if !set {
		return Config{Enabled: true}
	}
	return Config{Enabled: parseTruthy(raw)}
}

func parseTruthy(s string) bool {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// VerifyChatBackend performs GET {LLM_BASE_URL}/models (OpenAI-compatible discovery).
// Pass the same base URL and optional API key used for chat completions (e.g. Ollama /v1 or cloud).
func VerifyChatBackend(ctx context.Context, baseURL, apiKey string) error {
	base := strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return fmt.Errorf("empty LLM_BASE_URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/models", nil)
	if err != nil {
		return err
	}
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("LLM backend GET /models: HTTP %s", resp.Status)
	}
	return nil
}
