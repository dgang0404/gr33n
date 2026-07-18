package llm

import "testing"

// Phase 190 — completion token budget default was bumped 1024 -> 1536.
// A live grounded turn on the phi3:mini CPU profile hit exactly 1024/1024
// completion tokens mid-list ("...while refilling calcium nitrate:" with
// nothing after) — a real budget cutoff, not the model choosing to stop.
func TestMaxTokensFromEnv_defaultBumped(t *testing.T) {
	if got := maxTokensFromEnv(); got != 1536 {
		t.Fatalf("default max tokens = %d, want 1536", got)
	}
}

func TestMaxTokensFromEnv_overrideRespected(t *testing.T) {
	t.Setenv("LLM_MAX_TOKENS", "2048")
	if got := maxTokensFromEnv(); got != 2048 {
		t.Fatalf("max tokens = %d, want 2048", got)
	}
}

func TestMaxTokensFromEnv_overrideCappedAt8192(t *testing.T) {
	t.Setenv("LLM_MAX_TOKENS", "99999")
	if got := maxTokensFromEnv(); got != 8192 {
		t.Fatalf("max tokens = %d, want 8192 cap", got)
	}
}

func TestMaxTokensFromEnv_invalidFallsBackToDefault(t *testing.T) {
	t.Setenv("LLM_MAX_TOKENS", "not-a-number")
	if got := maxTokensFromEnv(); got != 1536 {
		t.Fatalf("max tokens = %d, want 1536 default fallback", got)
	}
}
