package chat

import (
	"net/http"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/httputil"
)

// Handler exposes Phase 27 Farm Guardian routes (stub until WS5).
type Handler struct {
	cfg ai.Config
}

func NewHandler(cfg ai.Config) *Handler {
	return &Handler{cfg: cfg}
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
	httputil.WriteError(w, http.StatusNotImplemented, "Farm Guardian chat is not implemented yet")
}
