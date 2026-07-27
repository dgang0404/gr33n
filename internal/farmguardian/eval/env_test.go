package eval

import (
	"testing"
	"time"
)

func TestWarmupTimeoutFromEnv_cpuProfileDefault(t *testing.T) {
	t.Setenv("GUARDIAN_EVAL_WARMUP_TIMEOUT", "")
	t.Setenv("GUARDIAN_TUNE_PROFILE", "cpu-16gb")
	if got := WarmupTimeoutFromEnv(); got != cpuLaptopWarmupTimeout {
		t.Fatalf("got %v want %v", got, cpuLaptopWarmupTimeout)
	}
}

func TestWarmupTimeoutFromEnv_override(t *testing.T) {
	t.Setenv("GUARDIAN_EVAL_WARMUP_TIMEOUT", "120")
	if got := WarmupTimeoutFromEnv(); got != 120*time.Second {
		t.Fatalf("got %v", got)
	}
}

func TestSuiteTimeoutFromEnv_default(t *testing.T) {
	t.Setenv("GUARDIAN_EVAL_SUITE_TIMEOUT_HOURS", "")
	if got := SuiteTimeoutFromEnv(); got != defaultSuiteTimeoutHours*time.Hour {
		t.Fatalf("got %v want %v", got, defaultSuiteTimeoutHours*time.Hour)
	}
}

func TestSuiteTimeoutFromEnv_override(t *testing.T) {
	t.Setenv("GUARDIAN_EVAL_SUITE_TIMEOUT_HOURS", "6")
	if got := SuiteTimeoutFromEnv(); got != 6*time.Hour {
		t.Fatalf("got %v", got)
	}
}

func TestClientTimeoutFromEnv_addsBuffer(t *testing.T) {
	t.Setenv("GUARDIAN_EVAL_TIMEOUT_SECONDS", "")
	t.Setenv("LLM_TIMEOUT_SECONDS", "1500")
	t.Setenv("GUARDIAN_GROUNDED_TIMEOUT_SECONDS", "1800")
	got := ClientTimeoutFromEnv()
	if got < 1800*time.Second {
		t.Fatalf("expected buffer above grounded timeout, got %v", got)
	}
}
