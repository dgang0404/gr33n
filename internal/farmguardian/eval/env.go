package eval

import (
	"os"
	"strconv"
	"strings"
	"time"

	"gr33n-api/internal/rag/llm"
)

const defaultWarmupTimeout = 5 * time.Minute
const cpuLaptopWarmupTimeout = 90 * time.Second
const evalTimeoutBuffer = 30 * time.Minute // was 15m — late grounded turns need headroom on CPU
const defaultSuiteTimeoutHours = 12
const defaultLaptopPromptCooldown = 60 * time.Second

// WarmupTimeoutFromEnv returns how long eval waits on POST /guardian/warmup before continuing.
// GUARDIAN_EVAL_WARMUP_TIMEOUT overrides; cpu-16gb profile defaults to 90s when unset.
func WarmupTimeoutFromEnv() time.Duration {
	if s := strings.TrimSpace(os.Getenv("GUARDIAN_EVAL_WARMUP_TIMEOUT")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	if profile := strings.TrimSpace(os.Getenv("GUARDIAN_TUNE_PROFILE")); profile == "" || profile == "cpu-16gb" {
		return cpuLaptopWarmupTimeout
	}
	return defaultWarmupTimeout
}

// LeavePendingTTLFromEnv is how long bumped proposals stay in the Pending tab after eval.
func LeavePendingTTLFromEnv() time.Duration {
	if s := strings.TrimSpace(os.Getenv("GUARDIAN_LEAVE_PENDING_HOURS")); s != "" {
		if h, err := strconv.Atoi(s); err == nil && h > 0 {
			return time.Duration(h) * time.Hour
		}
	}
	return 24 * time.Hour
}

// SuiteTimeoutFromEnv is the wall-clock budget for one guardian-eval process (all prompts in -suite).
// GUARDIAN_EVAL_SUITE_TIMEOUT_HOURS overrides; default 12h (laptop NF batches need headroom vs old 4h cap).
func SuiteTimeoutFromEnv() time.Duration {
	if s := strings.TrimSpace(os.Getenv("GUARDIAN_EVAL_SUITE_TIMEOUT_HOURS")); s != "" {
		if h, err := strconv.Atoi(s); err == nil && h > 0 {
			return time.Duration(h) * time.Hour
		}
	}
	return defaultSuiteTimeoutHours * time.Hour
}

// ClientTimeoutFromEnv is the HTTP client timeout for each eval chat POST.
func ClientTimeoutFromEnv() time.Duration {
	if s := strings.TrimSpace(os.Getenv("GUARDIAN_EVAL_TIMEOUT_SECONDS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return llm.EvalTimeoutFromEnv() + evalTimeoutBuffer
}

// PromptCooldownFromEnv is a pause between eval prompts (laptop thermal).
// GUARDIAN_EVAL_PROMPT_COOLDOWN_SECONDS overrides; cpu-16gb / unset profile defaults to 60s.
// Set to 0 to disable. ponytail: global sleep — upgrade = cool-down only after slow/timeout turns.
func PromptCooldownFromEnv() time.Duration {
	if s := strings.TrimSpace(os.Getenv("GUARDIAN_EVAL_PROMPT_COOLDOWN_SECONDS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			if n <= 0 {
				return 0
			}
			return time.Duration(n) * time.Second
		}
	}
	if profile := strings.TrimSpace(os.Getenv("GUARDIAN_TUNE_PROFILE")); profile == "" || profile == "cpu-16gb" {
		return defaultLaptopPromptCooldown
	}
	return 0
}

// PromptCooldownAfterLatency returns cool-down, extended after a slow turn.
func PromptCooldownAfterLatency(base, lastLatency time.Duration) time.Duration {
	if lastLatency >= 20*time.Minute {
		extra := 90 * time.Second
		if base > extra {
			return base
		}
		return extra
	}
	return base
}
