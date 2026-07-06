package llm

import (
	"net/http"
	"testing"
	"time"
)

func TestGroundedChatTimeoutFromEnv_defaultMinimum(t *testing.T) {
	t.Setenv("LLM_TIMEOUT_SECONDS", "")
	t.Setenv("GUARDIAN_GROUNDED_TIMEOUT_SECONDS", "")
	if got := GroundedChatTimeoutFromEnv(); got != DefaultGroundedTimeoutMinimum {
		t.Fatalf("got %v want %v", got, DefaultGroundedTimeoutMinimum)
	}
}

func TestGroundedChatTimeoutFromEnv_respectsHigherLLM(t *testing.T) {
	t.Setenv("LLM_TIMEOUT_SECONDS", "2000")
	t.Setenv("GUARDIAN_GROUNDED_TIMEOUT_SECONDS", "")
	if got := GroundedChatTimeoutFromEnv(); got != 2000*time.Second {
		t.Fatalf("got %v", got)
	}
}

func TestGroundedChatTimeoutFromEnv_explicit(t *testing.T) {
	t.Setenv("LLM_TIMEOUT_SECONDS", "777")
	t.Setenv("GUARDIAN_GROUNDED_TIMEOUT_SECONDS", "1800")
	if got := GroundedChatTimeoutFromEnv(); got != 1800*time.Second {
		t.Fatalf("got %v", got)
	}
}

func TestEvalTimeoutFromEnv_usesGroundedWhenHigher(t *testing.T) {
	t.Setenv("GUARDIAN_EVAL_TIMEOUT_SECONDS", "")
	t.Setenv("LLM_TIMEOUT_SECONDS", "777")
	t.Setenv("GUARDIAN_GROUNDED_TIMEOUT_SECONDS", "1800")
	if got := EvalTimeoutFromEnv(); got != 1800*time.Second {
		t.Fatalf("got %v", got)
	}
}

func TestWithHTTPTimeout(t *testing.T) {
	c := &Client{Model: "phi3:mini", HTTPClient: &http.Client{Timeout: 10 * time.Second}}
	c2 := c.WithHTTPTimeout(30 * time.Second)
	if c2.HTTPClient.Timeout != 30*time.Second {
		t.Fatalf("timeout=%v", c2.HTTPClient.Timeout)
	}
	if c.HTTPClient.Timeout != 10*time.Second {
		t.Fatal("original mutated")
	}
}
