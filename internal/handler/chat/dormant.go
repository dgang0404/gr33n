package chat

import (
	"encoding/json"
	"net/http"
	"strings"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

type dormantRequest struct {
	Mode          string `json:"mode"`
	FarmID        int64  `json:"farm_id"`
	IncludeVision bool   `json:"include_vision"`
	ChatModel     string `json:"chat_model"`
}

type dormantResponse struct {
	State string `json:"state"`
}

// PostDormant handles POST /guardian/dormant — deliberate Guardian rest
// (Phase 163 WS1). Unloads the chat (and optional vision) model from Ollama
// and marks awakening health as "dormant" until the next warmup.
func (h *Handler) PostDormant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.cfg.Enabled {
		httputil.WriteError(w, http.StatusServiceUnavailable, "AI features are disabled on this installation")
		return
	}

	var req dormantRequest
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

	var farmCounsel, farmQuick *string
	if req.FarmID > 0 && h.q != nil {
		if farm, err := h.q.GetFarmByID(r.Context(), req.FarmID); err == nil {
			farmCounsel = farmguardian.FarmCounselModel(&farm)
			farmQuick = farmguardian.FarmQuickModel(&farm)
		}
	}

	chatModel, _, reject := farmguardian.ResolveWarmupModel(h.modelCache, req.Mode, strings.TrimSpace(req.ChatModel), farmCounsel, farmQuick, envDefault)
	if reject != "" || chatModel == "" {
		httputil.WriteError(w, http.StatusServiceUnavailable, "no chat model to rest")
		return
	}

	visionModel := ""
	if req.IncludeVision {
		visionModel = farmguardian.VisionModelFromEnv()
	}

	llmBase := farmguardian.LLMBaseURLFromEnv()
	if h.baseLLM != nil && strings.TrimSpace(h.baseLLM.BaseURL) != "" {
		llmBase = h.baseLLM.BaseURL
	}
	if err := farmguardian.RequestDormant(r.Context(), llmBase, chatModel, visionModel, false); err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "failed to rest Guardian: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, dormantResponse{State: farmguardian.AwakeningStateDormant})
}
