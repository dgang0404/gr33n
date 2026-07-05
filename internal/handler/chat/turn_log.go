package chat

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"gr33n-api/internal/authctx"
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
		"llm_timeout_seconds", llmTimeoutSecondsForLog(),
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

func llmTimeoutSecondsForLog() int {
	s := strings.TrimSpace(os.Getenv("LLM_TIMEOUT_SECONDS"))
	if s == "" {
		return 120
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 120
	}
	return n
}
