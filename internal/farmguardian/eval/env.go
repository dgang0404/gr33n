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
const evalTimeoutBuffer = 15 * time.Minute

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

// ClientTimeoutFromEnv is the HTTP client timeout for each eval chat POST.
func ClientTimeoutFromEnv() time.Duration {
	if s := strings.TrimSpace(os.Getenv("GUARDIAN_EVAL_TIMEOUT_SECONDS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return llm.EvalTimeoutFromEnv() + evalTimeoutBuffer
}
