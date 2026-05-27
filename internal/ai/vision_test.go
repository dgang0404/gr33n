package ai

import (
	"os"
	"testing"
)

func TestLoadVisionConfigFromEnv(t *testing.T) {
	t.Setenv("LLM_VISION_MODEL", "llava")
	t.Setenv("LLM_VISION_BASE_URL", "")
	t.Setenv("LLM_BASE_URL", "http://127.0.0.1:11434/v1")
	cfg := LoadVisionConfigFromEnv()
	if !cfg.Enabled || cfg.Model != "llava" {
		t.Fatalf("cfg %#v", cfg)
	}
}

func TestVisionConfiguredFalseWithoutModel(t *testing.T) {
	os.Unsetenv("LLM_VISION_MODEL")
	if VisionConfigured() {
		t.Fatal("expected false")
	}
}
