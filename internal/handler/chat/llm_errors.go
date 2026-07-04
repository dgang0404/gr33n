package chat

import (
	"context"
	"errors"
	"net"
	"strings"
	"syscall"

	"gr33n-api/internal/rag/llm"
)

// LLMErrorPayload is returned to clients when chat streaming or completion fails.
type LLMErrorPayload struct {
	ErrorCode string `json:"error_code"`
	Error     string `json:"error"`
}

func classifyLLMError(err error) LLMErrorPayload {
	if err == nil {
		return LLMErrorPayload{ErrorCode: "llm_failed", Error: "LLM request failed"}
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		// Canceled is usually Stop button — still a timeout class for operators.
		if errors.Is(err, context.Canceled) {
			return LLMErrorPayload{
				ErrorCode: "llm_canceled",
				Error:     "Request canceled.",
			}
		}
		return LLMErrorPayload{
			ErrorCode: "llm_timeout",
			Error:     "Local model is still loading or CPU is slow. Wait and retry, or switch to tinyllama for faster replies.",
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return LLMErrorPayload{
			ErrorCode: "llm_timeout",
			Error:     "Local model is still loading or CPU is slow. Wait and retry, or switch to tinyllama for faster replies.",
		}
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if errors.Is(opErr.Err, syscall.ECONNREFUSED) || errors.Is(opErr.Err, syscall.ECONNRESET) {
			return LLMErrorPayload{
				ErrorCode: "llm_unreachable",
				Error:     "Ollama is not reachable at LLM_BASE_URL. Check that the service is running.",
			}
		}
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "context length") || strings.Contains(msg, "context window") || strings.Contains(msg, "too many tokens"):
		return LLMErrorPayload{
			ErrorCode: "llm_context",
			Error:     "Prompt too large for this model. Turn off farm context, use a larger model, or wait for prompt trimming (Phase 126).",
		}
	case strings.Contains(msg, "status 503") || strings.Contains(msg, "server busy") || strings.Contains(msg, "runner"):
		return LLMErrorPayload{
			ErrorCode: "llm_busy",
			Error:     "Ollama is busy (often the embedding model). Run: ollama stop <embed-model> then retry.",
		}
	case strings.Contains(msg, "connection refused"):
		return LLMErrorPayload{
			ErrorCode: "llm_unreachable",
			Error:     "Ollama is not reachable at LLM_BASE_URL. Check that the service is running.",
		}
	}
	if llm.IsTransientLLMError(err) {
		return LLMErrorPayload{
			ErrorCode: "llm_busy",
			Error:     "Ollama is busy or temporarily overloaded. Retry in a moment.",
		}
	}
	return LLMErrorPayload{
		ErrorCode: "llm_failed",
		Error:     "LLM request failed",
	}
}
