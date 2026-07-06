// Phase 129 WS2 — Ollama chat model preload for Guardian warmup.

package farmguardian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func preloadOllamaChatModel(ctx context.Context, llmBaseURL, model, keepAlive string) error {
	model = strings.TrimSpace(model)
	if model == "" {
		return fmt.Errorf("empty chat model")
	}
	base := OllamaNativeBase(llmBaseURL)
	if base == "" {
		return fmt.Errorf("not an Ollama base URL")
	}
	if keepAlive == "" {
		keepAlive = "30m"
	}
	payload, _ := json.Marshal(map[string]any{
		"model":      model,
		"prompt":     "ok",
		"stream":     false,
		"keep_alive": keepAlive,
	})
	client := &http.Client{Timeout: 10 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/generate", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ollama preload HTTP %s", strconv.Itoa(resp.StatusCode))
	}
	return nil
}
