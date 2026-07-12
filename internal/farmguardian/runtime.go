// Phase 130 — runtime feature flags and inline warmup on send.

package farmguardian

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
)

// EarlySSEEnabled reports whether grounded streams emit phase status before prompt build.
// Disabled when GUARDIAN_EARLY_SSE=0|false.
func EarlySSEEnabled() bool {
	v := strings.TrimSpace(os.Getenv("GUARDIAN_EARLY_SSE"))
	if v == "0" || strings.EqualFold(v, "false") {
		return false
	}
	return true
}

// InlineWarmupOnSendEnabled reports whether cold chat models are preloaded on grounded send.
func InlineWarmupOnSendEnabled() bool {
	v := strings.TrimSpace(os.Getenv("GUARDIAN_INLINE_WARMUP_ON_SEND"))
	if v == "0" || strings.EqualFold(v, "false") {
		return false
	}
	return true
}

// MaybeInlineWarmupOnSend preloads chat when not loaded, capped by maxWait (Phase 130 WS7).
func MaybeInlineWarmupOnSend(ctx context.Context, llmBaseURL, chatModel string, maxWait time.Duration) {
	if !InlineWarmupOnSendEnabled() || strings.TrimSpace(chatModel) == "" {
		return
	}
	_, loadedMap, _ := probeOllamaRuntime(ctx, llmBaseURL)
	if ok, _ := psEntry(loadedMap, chatModel); ok {
		return
	}
	runCtx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()
	embedModel := EmbedModelFromEnv()
	MaybeUnloadEmbedForChat(runCtx, llmBaseURL, embedModel, chatModel)
	if err := preloadOllamaChatModel(runCtx, llmBaseURL, chatModel, warmupKeepAlive); err != nil {
		slog.Warn("guardian: inline warmup on send failed", "chat_model", chatModel, "err", err)
		return
	}
	slog.Info("guardian: inline warmup on send", "chat_model", chatModel)
	ClearDormantFlag()
}
