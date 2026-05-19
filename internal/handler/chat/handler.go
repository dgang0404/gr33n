// Package chat exposes Phase 27 Farm Guardian endpoints. WS5 v1 is a
// non-streaming single-turn completion behind AI_ENABLED. RAG context
// injection, session history, streaming, and source attribution are
// scoped to follow-up slices in the Phase 27 plan.
package chat

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/rag/llm"
)

// Handler exposes Phase 27 Farm Guardian routes.
type Handler struct {
	cfg ai.Config
	llm llm.ChatCompleter
}

// NewHandler wires the configured chat client when AI is enabled. When AI is
// off the LLM is left nil and POST /v1/chat answers 503 — the same contract
// as POST /farms/{id}/rag/answer in Lite mode.
func NewHandler(cfg ai.Config) *Handler {
	h := &Handler{cfg: cfg}
	if cfg.Enabled {
		if c, err := llm.NewChatClientFromEnv(); err == nil {
			h.llm = c
		}
	}
	return h
}

// NewHandlerWithClient is the test seam — inject any ChatCompleter (mock or
// real) without relying on env vars.
func NewHandlerWithClient(cfg ai.Config, client llm.ChatCompleter) *Handler {
	return &Handler{cfg: cfg, llm: client}
}

type postBody struct {
	Message string `json:"message"`
	// Stream is accepted for forward compatibility; v1 ignores it (always non-streaming).
	Stream bool `json:"stream"`
}

type postResponse struct {
	Answer   string `json:"answer"`
	LLMModel string `json:"llm_model"`
}

// PostV1 handles POST /v1/chat — JWT required by route wiring.
func (h *Handler) PostV1(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.cfg.Enabled {
		httputil.WriteError(w, http.StatusServiceUnavailable, "AI features are disabled on this installation")
		return
	}
	if h.llm == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "Farm Guardian chat is not configured (set LLM_BASE_URL and LLM_MODEL)")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "request body required")
		return
	}
	var pb postBody
	if err := json.Unmarshal(body, &pb); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	user, err := farmguardian.BuildUserMessage(pb.Message)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	answer, err := h.llm.ChatCompletion(r.Context(), farmguardian.SystemPrompt(), user)
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			return
		}
		slog.Warn("farm guardian chat failed", "err", err)
		httputil.WriteError(w, http.StatusBadGateway, "LLM request failed")
		return
	}

	slog.Info("farm guardian chat completed", "model", h.llm.ModelLabel(), "message_runes", len([]rune(user)))
	httputil.WriteJSON(w, http.StatusOK, postResponse{
		Answer:   answer,
		LLMModel: h.llm.ModelLabel(),
	})
}
