package chat

import (
	"encoding/json"
	"net/http"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

type warmupRequest struct {
	Mode   string `json:"mode"`
	FarmID int64  `json:"farm_id"`
}

type warmupResponse struct {
	State     string `json:"state"`
	ChatModel string `json:"chat_model,omitempty"`
}

// PostWarmup handles POST /guardian/warmup — async Guardian model preload (Phase 129 WS1).
func (h *Handler) PostWarmup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.cfg.Enabled {
		httputil.WriteError(w, http.StatusServiceUnavailable, "AI features are disabled on this installation")
		return
	}

	var req warmupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.FarmID > 0 && !farmauthz.RequireFarmMember(w, r, h.q, req.FarmID) {
		return
	}

	envDefault := farmguardian.EnvServerDefaultModel()
	if h.baseLLM != nil {
		envDefault = h.baseLLM.ModelLabel()
	}

	var farmPref *string
	if req.FarmID > 0 && h.q != nil {
		if farm, err := h.q.GetFarmByID(r.Context(), req.FarmID); err == nil {
			farmPref = farm.GuardianPreferredModel
		}
	}

	llmBase := ""
	if h.baseLLM != nil {
		llmBase = h.baseLLM.BaseURL
	}
	state, chatModel := farmguardian.StartWarmup(r.Context(), llmBase, req.Mode, farmPref, envDefault, h.modelCache)

	status := http.StatusAccepted
	if state == farmguardian.AwakeningStateReady {
		status = http.StatusOK
	} else if state == farmguardian.AwakeningStateUnavailable {
		status = http.StatusServiceUnavailable
	}
	httputil.WriteJSON(w, status, warmupResponse{State: state, ChatModel: chatModel})
}
