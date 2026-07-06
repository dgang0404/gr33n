package chat

import (
	"net/http"
	"strconv"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/rag/ingest"
)

// GetHealth handles GET /v1/chat/health — offline field assistant readiness (Phase 37 WS1).
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.cfg.Enabled {
		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"ai_enabled": false,
			"field_assistant": farmguardian.FieldAssistantHealth{},
		})
		return
	}

	var farmID int64
	if s := r.URL.Query().Get("farm_id"); s != "" {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil || id <= 0 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid farm_id")
			return
		}
		farmID = id
	}

	ctx := r.Context()
	var fieldChunks, platformChunks int64
	var corpus *farmguardian.CorpusHealth
	if h.q != nil && farmID > 0 {
		if stats, err := h.q.GetRagCorpusStatsByFarm(ctx, farmID); err == nil {
			fieldChunks = stats.FieldGuideChunks
			platformChunks = stats.PlatformDocChunks
			c := farmguardian.BuildCorpusHealth(farmguardian.CorpusStatsFromRow(stats), time.Now().UTC())
			corpus = &c
		} else {
			if n, err := h.q.CountRagChunksByFarmSourceType(ctx, db.CountRagChunksByFarmSourceTypeParams{
				FarmID: farmID, SourceType: ingest.SourceTypeFieldGuide,
			}); err == nil {
				fieldChunks = n
			}
			if n, err := h.q.CountRagChunksByFarmSourceType(ctx, db.CountRagChunksByFarmSourceTypeParams{
				FarmID: farmID, SourceType: ingest.SourceTypePlatformDoc,
			}); err == nil {
				platformChunks = n
			}
		}
	}

	health := farmguardian.BuildFieldAssistantHealth(ctx, nil, fieldChunks, platformChunks)
	mode := r.URL.Query().Get("mode")

	envDefault := farmguardian.EnvServerDefaultModel()
	if h.baseLLM != nil {
		envDefault = h.baseLLM.ModelLabel()
	}
	var farmCounsel, farmQuick *string
	if farmID > 0 && h.q != nil {
		if farm, err := h.q.GetFarmByID(ctx, farmID); err == nil {
			farmCounsel = farmguardian.FarmCounselModel(&farm)
			farmQuick = farmguardian.FarmQuickModel(&farm)
		}
	}
	awakening := farmguardian.BuildAwakeningHealth(ctx, farmguardian.AwakeningBuildInput{
		AIEnabled:          true,
		Field:              health,
		Mode:               mode,
		FarmID:             farmID,
		FieldGuideChunks:   fieldChunks,
		PlatformDocChunks:  platformChunks,
		Corpus:             corpus,
		Cache:              h.modelCache,
		FarmCounselModel:   farmCounsel,
		FarmQuickModel:     farmQuick,
		EnvDefault:         envDefault,
	})

	procsOK := ProceduresAvailable()
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"ai_enabled":           true,
		"chat_configured":      h.llm != nil || llmConfigured(),
		"procedures_available": procsOK,
		"field_degrade_ready":  procsOK && health.FieldMode,
		"field_assistant":      health,
		"awakening":            awakening,
	})
}
