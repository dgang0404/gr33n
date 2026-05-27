package ai

import (
	"os"
	"strings"
)

// VisionConfig describes optional multimodal chat (Phase 30 WS6).
type VisionConfig struct {
	Enabled bool
	Model   string
	BaseURL string
	APIKey  string
}

// LoadVisionConfigFromEnv returns vision settings when LLM_VISION_MODEL is set.
// LLM_VISION_BASE_URL and LLM_VISION_API_KEY fall back to LLM_BASE_URL / LLM_API_KEY.
func LoadVisionConfigFromEnv() VisionConfig {
	model := strings.TrimSpace(os.Getenv("LLM_VISION_MODEL"))
	if model == "" {
		return VisionConfig{}
	}
	base := strings.TrimSpace(os.Getenv("LLM_VISION_BASE_URL"))
	if base == "" {
		base = strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	}
	key := strings.TrimSpace(os.Getenv("LLM_VISION_API_KEY"))
	if key == "" {
		key = strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	}
	if base == "" {
		return VisionConfig{}
	}
	return VisionConfig{
		Enabled: true,
		Model:   model,
		BaseURL: base,
		APIKey:  key,
	}
}

// VisionConfigured reports whether vision chat can be offered to the UI.
func VisionConfigured() bool {
	return LoadVisionConfigFromEnv().Enabled
}
