package chat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

type pullModelRequest struct {
	Name   string `json:"name"`
	FarmID int64  `json:"farm_id"`
}

type pullModelResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// PostPullModel handles POST /guardian/models/pull — farm admin, local Ollama only.
func (h *Handler) PostPullModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.cfg.Enabled {
		httputil.WriteError(w, http.StatusServiceUnavailable, "AI features are disabled on this installation")
		return
	}
	if !farmguardian.IsLocalOllamaConfigured() {
		httputil.WriteError(w, http.StatusBadRequest, "model pull is only available when LLM_BASE_URL points at local Ollama")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var req pullModelRequest
	if err := json.Unmarshal(body, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.FarmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "farm_id is required")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, req.FarmID) {
		return
	}
	if h.modelCache == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "model cache not configured")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), farmguardian.PullTimeoutFromEnv()+5*time.Second)
	defer cancel()

	if err := h.modelCache.PullAndRefresh(ctx, name); err != nil {
		if ctx.Err() != nil {
			httputil.WriteError(w, http.StatusGatewayTimeout, "model pull timed out")
			return
		}
		httputil.WriteError(w, http.StatusBadGateway, "model pull failed: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, pullModelResponse{
		Name:   name,
		Status: "success",
	})
}
