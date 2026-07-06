package guardian

import (
	"encoding/json"
	"net/http"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/rag/reingest"
)

type reingestRequest struct {
	Scope string `json:"scope"`
}

// PostReingest handles POST /farms/{id}/guardian/reingest — farm admin async RAG ingest (Phase 135).
func (h *Handler) PostReingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.cfg.Enabled {
		httputil.WriteError(w, http.StatusServiceUnavailable, "AI features are disabled on this installation")
		return
	}
	farmID, err := parseFarmID(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}

	var req reingestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Scope == "" {
		httputil.WriteError(w, http.StatusBadRequest, "scope is required")
		return
	}

	job, err := reingest.Start(r.Context(), h.q, farmID, req.Scope)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	status := http.StatusAccepted
	if job.Status != reingest.StatusRunning {
		status = http.StatusOK
	}
	httputil.WriteJSON(w, status, job)
}

// GetReingestStatus handles GET /farms/{id}/guardian/reingest/status (Phase 135).
func (h *Handler) GetReingestStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	farmID, err := parseFarmID(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	job := reingest.ForFarm(farmID)
	if job == nil {
		httputil.WriteJSON(w, http.StatusOK, map[string]any{"status": "idle"})
		return
	}
	httputil.WriteJSON(w, http.StatusOK, job)
}
