package farmguardian

import (
	"strings"
	"testing"
)

func TestSystemPromptShapesPersona(t *testing.T) {
	p := SystemPrompt()
	if !strings.Contains(p, "Farm Guardian") {
		t.Fatal("system prompt missing persona name")
	}
	for _, want := range []string{"setpoint", "schedule", "rule", "zone"} {
		if !strings.Contains(p, want) {
			t.Fatalf("system prompt missing glossary term %q", want)
		}
	}
	if strings.HasSuffix(p, "\n") {
		t.Fatal("system prompt should be trimmed")
	}
}

func TestBuildUserMessage(t *testing.T) {
	t.Run("rejects empty", func(t *testing.T) {
		if _, err := BuildUserMessage("   "); err == nil {
			t.Fatal("expected error on empty message")
		}
	})
	t.Run("trims whitespace", func(t *testing.T) {
		got, err := BuildUserMessage("  hello\n")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got != "hello" {
			t.Fatalf("got %q want hello", got)
		}
	})
	t.Run("rejects too long", func(t *testing.T) {
		big := strings.Repeat("a", MaxMessageRunes+1)
		if _, err := BuildUserMessage(big); err == nil {
			t.Fatal("expected length error")
		}
	})
}
