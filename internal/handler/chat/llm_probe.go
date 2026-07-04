package chat

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/farmguardian/procedures"
)

const (
	reachableCacheTTL   = 30 * time.Second
	unreachableCacheTTL = 2 * time.Second
	defaultLocalProbe = 12 * time.Second
	defaultRemoteProbe = 5 * time.Second
)

var llmReachabilityCache struct {
	mu        sync.Mutex
	checkedAt time.Time
	reachable bool
	errText   string
}

// llmConfigured reports whether env suggests a chat backend is intended.
func llmConfigured() bool {
	return strings.TrimSpace(os.Getenv("LLM_BASE_URL")) != "" &&
		strings.TrimSpace(os.Getenv("LLM_MODEL")) != ""
}

func (h *Handler) llmReachable(ctx context.Context) bool {
	if h != nil && h.llm != nil {
		return probeLLMReachable(ctx)
	}
	if !llmConfigured() {
		return false
	}
	return probeLLMReachable(ctx)
}

func probeCacheTTL(reachable bool) time.Duration {
	if reachable {
		return reachableCacheTTL
	}
	return unreachableCacheTTL
}

func llmProbeTimeout(base string) time.Duration {
	if raw := strings.TrimSpace(os.Getenv("GUARDIAN_LLM_PROBE_TIMEOUT_SECONDS")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	if farmguardian.IsLocalInferenceURL(base) {
		return defaultLocalProbe
	}
	return defaultRemoteProbe
}

func probeLLMReachable(ctx context.Context) bool {
	base := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	if base == "" {
		return false
	}
	llmReachabilityCache.mu.Lock()
	defer llmReachabilityCache.mu.Unlock()
	if !llmReachabilityCache.checkedAt.IsZero() {
		if time.Since(llmReachabilityCache.checkedAt) < probeCacheTTL(llmReachabilityCache.reachable) {
			return llmReachabilityCache.reachable
		}
	}
	probeCtx, cancel := context.WithTimeout(ctx, llmProbeTimeout(base))
	defer cancel()
	err := ai.VerifyChatBackend(probeCtx, base, strings.TrimSpace(os.Getenv("LLM_API_KEY")))
	reachable := err == nil
	if !reachable && farmguardian.IsLocalInferenceURL(base) && probeOllamaBusyAlive(ctx, base) {
		slog.Debug("farm guardian: LLM /models probe failed but Ollama is busy-alive — treating as reachable",
			"probe_err", err)
		reachable = true
		err = nil
	}
	llmReachabilityCache.checkedAt = time.Now()
	llmReachabilityCache.reachable = reachable
	if err != nil {
		llmReachabilityCache.errText = err.Error()
	} else {
		llmReachabilityCache.errText = ""
	}
	return llmReachabilityCache.reachable
}

// probeOllamaBusyAlive returns true when the native Ollama daemon responds even if
// GET /v1/models is slow (model load, embedding job on CPU).
func probeOllamaBusyAlive(ctx context.Context, openAIBase string) bool {
	native := farmguardian.OllamaNativeBase(openAIBase)
	if native == "" {
		return false
	}
	probeCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, native+"/api/ps", nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// ResetLLMReachabilityCache clears the probe cache (tests and after chat errors).
func ResetLLMReachabilityCache() {
	llmReachabilityCache.mu.Lock()
	defer llmReachabilityCache.mu.Unlock()
	llmReachabilityCache.checkedAt = time.Time{}
}

func (h *Handler) fieldDegradeEligible() bool {
	base := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	return base != "" && farmguardian.IsLocalInferenceURL(base)
}

// ProceduresAvailable reports whether authored YAML procedures are on disk.
func ProceduresAvailable() bool {
	_, err := procedures.List(procedures.RepoRoot())
	return err == nil
}
