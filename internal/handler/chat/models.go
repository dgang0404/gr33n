package chat

import (
	"net/http"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

type modelsResponse struct {
	AvailableModels []farmguardian.ModelInfo `json:"available_models"`
	ServerDefault   string                   `json:"server_default"`
}

// GetModels handles GET /guardian/models — server-wide Ollama snapshot (not farm-scoped).
// Query ?all=true includes embedding-only models for debugging.
func (h *Handler) GetModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.cfg.Enabled {
		httputil.WriteError(w, http.StatusServiceUnavailable, "AI features are disabled on this installation")
		return
	}
	includeAll := r.URL.Query().Get("all") == "true"
	models, serverDefault := h.modelCache.Snapshot(includeAll)
	if models == nil {
		models = []farmguardian.ModelInfo{}
	}
	if serverDefault == "" && h.baseLLM != nil {
		serverDefault = h.baseLLM.ModelLabel()
	}
	httputil.WriteJSON(w, http.StatusOK, modelsResponse{
		AvailableModels: models,
		ServerDefault:   serverDefault,
	})
}
