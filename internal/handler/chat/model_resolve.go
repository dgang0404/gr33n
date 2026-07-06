package chat

import (
	"context"
	"log/slog"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/rag/llm"
)

func (h *Handler) resolveChatClient(ctx context.Context, sessionModel string, farmID int64, grounded, vision bool) (llm.ChatCompleter, farmguardian.ResolveOutcome) {
	if vision {
		if h.visionLLM != nil {
			return h.visionLLM, farmguardian.ResolveOutcome{ModelName: h.visionLLM.ModelLabel()}
		}
		return nil, farmguardian.ResolveOutcome{}
	}
	if h.llm == nil {
		return nil, farmguardian.ResolveOutcome{RejectReason: "Farm Guardian chat is not configured (set LLM_BASE_URL and LLM_MODEL)"}
	}

	envDefault := farmguardian.EnvServerDefaultModel()
	if h.baseLLM != nil {
		envDefault = h.baseLLM.ModelLabel()
	} else if h.llm != nil {
		envDefault = h.llm.ModelLabel()
	}

	var farmPref *string
	if farmID > 0 && h.q != nil {
		if farm, err := h.q.GetFarmByID(ctx, farmID); err == nil {
			if grounded {
				farmPref = farmguardian.FarmCounselModel(&farm)
			} else {
				farmPref = farmguardian.FarmQuickModel(&farm)
			}
		}
	}

	outcome := farmguardian.ResolveChatModel(h.modelCache, sessionModel, farmPref, envDefault, grounded)
	if outcome.RejectReason != "" {
		return nil, outcome
	}
	if outcome.Fallback {
		slog.Warn("guardian: model not available, falling back",
			"requested", sessionModel,
			"farm_id", farmID,
			"using", outcome.ModelName,
		)
	}

	client := h.llm
	if h.baseLLM != nil && outcome.ModelName != "" && outcome.ModelName != h.baseLLM.ModelLabel() {
		client = h.baseLLM.WithModel(outcome.ModelName)
	}
	if outcome.ModelName == "" && h.baseLLM != nil {
		outcome.ModelName = h.baseLLM.ModelLabel()
	}
	return client, outcome
}

func applyModelMeta(resp *postResponse, outcome farmguardian.ResolveOutcome) {
	if resp == nil {
		return
	}
	if outcome.ModelName != "" {
		resp.LLMModel = outcome.ModelName
		resp.ModelUsed = outcome.ModelName
	}
	resp.ModelFallback = outcome.Fallback
}
