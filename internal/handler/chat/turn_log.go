package chat

import (
	"context"
	"log/slog"
	"time"

	"gr33n-api/internal/authctx"
	"gr33n-api/internal/rag/llm"
)

type chatTurnMeta struct {
	farmID          int64
	grounded        bool
	stream          bool
	model           string
	question        string
	historyTurns    int
	contextChunks   int
	effectiveWindow int
	sessionID       string
}

func (h *Handler) logChatTurnStarted(ctx context.Context, meta chatTurnMeta) {
	attrs := []any{
		"request_id", authctx.RequestID(ctx),
		"farm_id", meta.farmID,
		"grounded", meta.grounded,
		"stream", meta.stream,
		"model", meta.model,
		"question_chars", len(meta.question),
		"history_turns", meta.historyTurns,
		"context_chunks", meta.contextChunks,
		"effective_context_window", meta.effectiveWindow,
		"llm_timeout_seconds", llmTimeoutSecondsForLog(meta.grounded),
	}
	if meta.sessionID != "" {
		attrs = append(attrs, "session_id", meta.sessionID)
	}
	slog.Info("guardian: chat turn started", attrs...)
}

func (h *Handler) logChatTurnFailed(ctx context.Context, meta chatTurnMeta, started time.Time, err error) {
	payload := classifyLLMError(err)
	attrs := []any{
		"request_id", authctx.RequestID(ctx),
		"farm_id", meta.farmID,
		"grounded", meta.grounded,
		"stream", meta.stream,
		"model", meta.model,
		"elapsed_ms", time.Since(started).Milliseconds(),
		"error_code", payload.ErrorCode,
		"err", err,
	}
	if meta.sessionID != "" {
		attrs = append(attrs, "session_id", meta.sessionID)
	}
	slog.Warn("guardian: chat turn failed", attrs...)
}

func llmTimeoutSecondsForLog(grounded bool) int {
	var d time.Duration
	if grounded {
		d = llm.GroundedChatTimeoutFromEnv()
	} else {
		d = llm.ChatTimeoutFromEnv()
	}
	return int(d / time.Second)
}
