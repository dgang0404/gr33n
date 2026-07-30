package eval

import (
	"testing"
	"time"
)

func TestPromptCooldownAfterLatency(t *testing.T) {
	base := 60 * time.Second
	if got := PromptCooldownAfterLatency(base, 5*time.Minute); got != base {
		t.Fatalf("short turn: got %v want %v", got, base)
	}
	if got := PromptCooldownAfterLatency(base, 25*time.Minute); got != 90*time.Second {
		t.Fatalf("slow turn: got %v want 90s", got)
	}
	if got := PromptCooldownAfterLatency(120*time.Second, 25*time.Minute); got != 120*time.Second {
		t.Fatalf("base already longer: got %v", got)
	}
}

func TestPromptCooldownFromEnv_laptopDefault(t *testing.T) {
	t.Setenv("GUARDIAN_EVAL_PROMPT_COOLDOWN_SECONDS", "")
	t.Setenv("GUARDIAN_TUNE_PROFILE", "")
	if got := PromptCooldownFromEnv(); got != defaultLaptopPromptCooldown {
		t.Fatalf("got %v want %v", got, defaultLaptopPromptCooldown)
	}
	t.Setenv("GUARDIAN_EVAL_PROMPT_COOLDOWN_SECONDS", "0")
	if got := PromptCooldownFromEnv(); got != 0 {
		t.Fatalf("explicit 0 should disable, got %v", got)
	}
}
