package farmguardian

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultShowConcurrency = 4
	defaultPullTimeoutSec  = 600
)

// LLMBaseURLFromEnv returns trimmed LLM_BASE_URL.
func LLMBaseURLFromEnv() string {
	return strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
}

// IsLocalOllamaConfigured reports whether LLM_BASE_URL points at local inference.
func IsLocalOllamaConfigured() bool {
	return IsLocalInferenceURL(LLMBaseURLFromEnv())
}

// AutoPullEnabled reads GUARDIAN_OLLAMA_AUTO_PULL (default false).
func AutoPullEnabled() bool {
	s := strings.ToLower(strings.TrimSpace(os.Getenv("GUARDIAN_OLLAMA_AUTO_PULL")))
	return s == "1" || s == "true" || s == "yes"
}

// ShowConcurrencyFromEnv returns GUARDIAN_OLLAMA_SHOW_CONCURRENCY (default 4).
func ShowConcurrencyFromEnv() int {
	s := strings.TrimSpace(os.Getenv("GUARDIAN_OLLAMA_SHOW_CONCURRENCY"))
	if s == "" {
		return defaultShowConcurrency
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return defaultShowConcurrency
	}
	if n > 32 {
		return 32
	}
	return n
}

// PullTimeoutFromEnv returns GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS (default 600).
func PullTimeoutFromEnv() time.Duration {
	s := strings.TrimSpace(os.Getenv("GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS"))
	if s == "" {
		return defaultPullTimeoutSec * time.Second
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return defaultPullTimeoutSec * time.Second
	}
	return time.Duration(n) * time.Second
}
