package farmguardian

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type ollamaPsResponse struct {
	Models []ollamaPsModel `json:"models"`
}

type ollamaPsModel struct {
	Name     string `json:"name"`
	SizeVRAM int64  `json:"size_vram"`
}

// EnrichModelRuntimeHints marks loaded models and sets advisory runtime_hint text
// from GET /api/ps. Failures are ignored — hints are optional.
func EnrichModelRuntimeHints(ctx context.Context, llmBaseURL string, models []ModelInfo, client *http.Client) []ModelInfo {
	if len(models) == 0 {
		return models
	}
	base := OllamaNativeBase(llmBaseURL)
	if base == "" {
		return models
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/ps", nil)
	if err != nil {
		return models
	}
	resp, err := client.Do(req)
	if err != nil {
		return models
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return models
	}
	var parsed ollamaPsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return models
	}
	loaded := make(map[string]ollamaPsModel, len(parsed.Models))
	for _, m := range parsed.Models {
		name := strings.TrimSpace(m.Name)
		if name == "" {
			continue
		}
		loaded[name] = m
		for _, key := range modelLookupKeys(name) {
			if _, exists := loaded[key]; !exists {
				loaded[key] = m
			}
		}
	}
	out := make([]ModelInfo, len(models))
	copy(out, models)
	for i := range out {
		var ps ollamaPsModel
		found := false
		for _, key := range modelLookupKeys(out[i].Name) {
			if m, ok := loaded[key]; ok {
				ps = m
				found = true
				break
			}
		}
		if !found {
			out[i].RuntimeHint = "cold — first message loads the model from local disk (no internet needed); may take a while"
			continue
		}
		out[i].Loaded = true
		if ps.SizeVRAM > 0 {
			out[i].Processor = "gpu"
			out[i].RuntimeHint = "loaded on GPU"
		} else {
			out[i].Processor = "cpu"
			out[i].RuntimeHint = "loaded, CPU-only — expect slow replies"
		}
	}
	return out
}
