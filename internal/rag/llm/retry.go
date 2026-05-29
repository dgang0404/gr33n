package llm

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// RetryConfig controls the Phase 27 WS3 follow-up retry policy.
//
//   - MaxAttempts is the total number of tries including the first one — so
//     MaxAttempts = 1 disables retry entirely, MaxAttempts = 3 means
//     1 attempt + up to 2 retries.
//   - InitialBackoff is the wait before retry #1; subsequent retries double
//     the backoff up to MaxBackoff. We also apply ±25% jitter so a fleet of
//     restarted Pis doesn't thunder the LLM at the same millisecond.
//   - Sleeper is the function used to wait between attempts (test seam).
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Sleeper        func(ctx context.Context, d time.Duration) error
}

// Defaults are intentionally conservative — three attempts at 500ms / 1s / 2s
// (with jitter) keeps the worst-case latency below the default 666s LLM
// timeout while still papering over the most common transient failures.
const (
	DefaultRetryMaxAttempts    = 3
	DefaultRetryInitialBackoff = 500 * time.Millisecond
	DefaultRetryMaxBackoff     = 10 * time.Second
)

// HTTPStatusError carries the HTTP status code from a non-2xx LLM response.
// We attach it so the retry classifier can decide on transient (5xx/429) vs
// permanent (4xx other than 429) without scraping error strings.
type HTTPStatusError struct {
	StatusCode int
	Body       string // truncated preview
}

func (e *HTTPStatusError) Error() string {
	if e.Body == "" {
		return "chat HTTP " + strconv.Itoa(e.StatusCode)
	}
	return "chat HTTP " + strconv.Itoa(e.StatusCode) + ": " + e.Body
}

// retryAttemptsFromEnv reads LLM_RETRY_MAX_ATTEMPTS (clamped to [1, 8]).
func retryAttemptsFromEnv() int {
	raw := strings.TrimSpace(os.Getenv("LLM_RETRY_MAX_ATTEMPTS"))
	if raw == "" {
		return DefaultRetryMaxAttempts
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return DefaultRetryMaxAttempts
	}
	if n > 8 {
		return 8
	}
	return n
}

// retryBackoffFromEnv reads LLM_RETRY_BACKOFF_MS as the initial backoff
// (clamped to [10ms, 30s]).
func retryBackoffFromEnv() time.Duration {
	raw := strings.TrimSpace(os.Getenv("LLM_RETRY_BACKOFF_MS"))
	if raw == "" {
		return DefaultRetryInitialBackoff
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 10 {
		return DefaultRetryInitialBackoff
	}
	d := time.Duration(n) * time.Millisecond
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}

// retryConfigFromEnv builds the *Client default retry config.
func retryConfigFromEnv() RetryConfig {
	return RetryConfig{
		MaxAttempts:    retryAttemptsFromEnv(),
		InitialBackoff: retryBackoffFromEnv(),
		MaxBackoff:     DefaultRetryMaxBackoff,
		Sleeper:        ctxSleep,
	}
}

// IsTransientLLMError returns true when the error is worth retrying.
// Network failures, context-deadline-exceeded (the request-level timeout —
// NOT the caller's ctx.Done()), and the standard 408/429/5xx HTTP statuses
// are transient. Caller cancellation (ctx.Err()) is treated as permanent so
// we never resurrect a request the operator already gave up on.
func IsTransientLLMError(err error) bool {
	if err == nil {
		return false
	}

	// HTTP status — transient if 408, 429, 5xx.
	var httpErr *HTTPStatusError
	if errors.As(err, &httpErr) {
		return isRetryableStatus(httpErr.StatusCode)
	}

	// Caller-cancelled — don't paper over it.
	if errors.Is(err, context.Canceled) {
		return false
	}

	// Per-attempt timeout (from c.HTTPClient.Timeout) surfaces as
	// context.DeadlineExceeded wrapped in a url.Error. That's transient —
	// the next attempt may succeed.
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Anything net-related (DNS, connect, reset, dropped conn) is worth
	// retrying. *net.OpError covers most cases; url.Error wraps the
	// http.Client errors.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		// url.Error wraps temporary connection failures.
		return true
	}

	return false
}

func isRetryableStatus(code int) bool {
	switch code {
	case 408, 425, 429, 500, 502, 503, 504:
		return true
	}
	return false
}

// retryOp runs fn with the configured retry/backoff policy. It returns
// the first nil error (success) or the last error after MaxAttempts.
// ctx cancellation short-circuits between attempts.
func retryOp(ctx context.Context, cfg RetryConfig, fn func(attempt int) error) error {
	attempts := cfg.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := fn(attempt)
		if err == nil {
			return nil
		}
		lastErr = err
		if !IsTransientLLMError(err) || attempt == attempts {
			return err
		}
		wait := backoffFor(cfg, attempt)
		sleep := cfg.Sleeper
		if sleep == nil {
			sleep = ctxSleep
		}
		if serr := sleep(ctx, wait); serr != nil {
			return serr
		}
	}
	return lastErr
}

// backoffFor returns the post-attempt-N wait: initial * 2^(N-1) capped at
// MaxBackoff, with ±25% jitter to avoid thundering-herd.
func backoffFor(cfg RetryConfig, attempt int) time.Duration {
	if cfg.InitialBackoff <= 0 {
		return 0
	}
	exp := cfg.InitialBackoff << (attempt - 1)
	if exp <= 0 || exp > cfg.MaxBackoff {
		exp = cfg.MaxBackoff
	}
	if exp <= 0 {
		return 0
	}
	// ±25% jitter. rand.Float64 is good enough for backoff spread; no need for crypto/rand.
	jitterFraction := (rand.Float64()*0.5 - 0.25) // -0.25 .. +0.25
	jittered := time.Duration(float64(exp) * (1.0 + jitterFraction))
	if jittered <= 0 {
		return 0
	}
	return jittered
}

// ctxSleep is the production sleeper — honours ctx cancellation.
func ctxSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
