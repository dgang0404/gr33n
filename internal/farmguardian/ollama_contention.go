// Phase 130 WS2 — unload embedding model before grounded chat when RAM contended.

package farmguardian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// MaybeUnloadEmbedForChat unloads the embedding model from Ollama when it is loaded
// but the chat model is not (or both are CPU-only), freeing RAM for chat prefill.
// Skipped when chat and embed run on different hosts (Phase 138 split inference).
func MaybeUnloadEmbedForChat(ctx context.Context, llmBaseURL, embedModel, chatModel string) {
	if InferenceHostsSplit() {
		return
	}
	embedModel = strings.TrimSpace(embedModel)
	chatModel = strings.TrimSpace(chatModel)
	if embedModel == "" || chatModel == "" {
		return
	}
	base := OllamaNativeBase(llmBaseURL)
	if base == "" || !IsLocalInferenceURL(strings.TrimSpace(llmBaseURL)) {
		return
	}
	loaded, err := listOllamaPS(ctx, base, http.DefaultClient)
	if err != nil || len(loaded) == 0 {
		return
	}
	embedLoaded, embedCPU := psEntry(loaded, embedModel)
	chatLoaded, _ := psEntry(loaded, chatModel)
	if !embedLoaded {
		return
	}
	// Chat already warm and embed not blocking — skip.
	if chatLoaded && !embedCPU {
		return
	}
	if !chatLoaded || embedCPU {
		if err := unloadOllamaModel(ctx, base, embedModel, http.DefaultClient); err != nil {
			slog.Warn("guardian: embed unload failed", "embed_model", embedModel, "err", err)
			return
		}
		slog.Info("guardian: embed unloaded for chat", "embed_model", embedModel, "chat_model", chatModel)
	}
}

func psEntry(loaded map[string]ollamaPsModel, name string) (found bool, cpuOnly bool) {
	for _, key := range modelLookupKeys(name) {
		if m, ok := loaded[key]; ok {
			return true, m.SizeVRAM == 0
		}
	}
	return false, false
}

func listOllamaPS(ctx context.Context, base string, client *http.Client) (map[string]ollamaPsModel, error) {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/ps", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, err
	}
	var parsed ollamaPsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	out := make(map[string]ollamaPsModel, len(parsed.Models))
	for _, m := range parsed.Models {
		name := strings.TrimSpace(m.Name)
		if name == "" {
			continue
		}
		out[name] = m
		for _, key := range modelLookupKeys(name) {
			if _, exists := out[key]; !exists {
				out[key] = m
			}
		}
	}
	return out, nil
}

func unloadOllamaModel(ctx context.Context, base, model string, client *http.Client) error {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	payload, _ := json.Marshal(map[string]any{
		"model":      model,
		"prompt":     "",
		"keep_alive": 0,
	})
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
		return fmt.Errorf("ollama unload HTTP %s", strconv.Itoa(resp.StatusCode))
	}
	return nil
}

// EmbedModelFromEnv returns EMBEDDING_MODEL for contention checks.
func EmbedModelFromEnv() string {
	return strings.TrimSpace(os.Getenv("EMBEDDING_MODEL"))
}

// VisionModelFromEnv returns LLM_VISION_MODEL when zone photo analysis is configured.
func VisionModelFromEnv() string {
	return strings.TrimSpace(os.Getenv("LLM_VISION_MODEL"))
}
