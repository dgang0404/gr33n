package chat

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/farmguardian/procedures"
)

const llmReachabilityTTL = 15 * time.Second

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

func probeLLMReachable(ctx context.Context) bool {
	base := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	if base == "" {
		return false
	}
	llmReachabilityCache.mu.Lock()
	defer llmReachabilityCache.mu.Unlock()
	if time.Since(llmReachabilityCache.checkedAt) < llmReachabilityTTL {
		return llmReachabilityCache.reachable
	}
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := ai.VerifyChatBackend(probeCtx, base, strings.TrimSpace(os.Getenv("LLM_API_KEY")))
	llmReachabilityCache.checkedAt = time.Now()
	llmReachabilityCache.reachable = err == nil
	if err != nil {
		llmReachabilityCache.errText = err.Error()
	} else {
		llmReachabilityCache.errText = ""
	}
	return llmReachabilityCache.reachable
}

// ResetLLMReachabilityCache clears the probe cache (tests only).
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
