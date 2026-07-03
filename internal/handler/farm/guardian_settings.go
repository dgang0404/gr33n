package farm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

// PatchGuardianSettings handles PATCH /farms/{id}/settings — farm admin only.
func (h *Handler) PatchGuardianSettings(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, id) {
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var raw map[string]json.RawMessage
	if len(strings.TrimSpace(string(body))) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "request body required")
		return
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	fieldRaw, hasField := raw["guardian_preferred_model"]
	if !hasField {
		httputil.WriteError(w, http.StatusBadRequest, "guardian_preferred_model is required")
		return
	}

	candidateModel, parseErr := parsePreferredModelCandidate(fieldRaw)
	if parseErr != nil {
		httputil.WriteError(w, http.StatusBadRequest, parseErr.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), settingsRequestTimeout(candidateModel))
	defer cancel()

	existing, err := h.q.GetFarmByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "farm not found")
		return
	}

	var newModel *string
	if candidateModel == "" {
		newModel = nil
	} else if err := h.ensureModelAvailable(ctx, candidateModel); err != nil {
		if strings.Contains(err.Error(), "not loaded") {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.WriteError(w, http.StatusBadGateway, err.Error())
		return
	} else {
		trimmed := candidateModel
		newModel = &trimmed
	}

	updated, err := h.q.UpdateFarmGuardianPreferredModel(ctx, db.UpdateFarmGuardianPreferredModelParams{
		ID:                     id,
		GuardianPreferredModel: newModel,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update farm settings")
		return
	}

	from := modelLabelPtr(existing.GuardianPreferredModel)
	to := modelLabelPtr(updated.GuardianPreferredModel)
	if from != to {
		var uid uuid.UUID
		if u, ok := authctx.UserID(r.Context()); ok {
			uid = u
		}
		auditlog.Submit(r.Context(), h.q, r, auditlog.Event{
			FarmID: auditlog.FarmIDPtr(id),
			Action: db.Gr33ncoreUserActionTypeEnumGuardianModelChanged,
			Details: map[string]any{
				"from":               from,
				"to":                 to,
				"changed_by_user_id": uid.String(),
			},
		})
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"id":                       updated.ID,
		"guardian_preferred_model": updated.GuardianPreferredModel,
	})
}

func modelLabelPtr(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func settingsRequestTimeout(candidate string) time.Duration {
	if strings.TrimSpace(candidate) == "" {
		return 5 * time.Second
	}
	if farmguardian.AutoPullEnabled() && farmguardian.IsLocalOllamaConfigured() {
		return farmguardian.PullTimeoutFromEnv() + 10*time.Second
	}
	return 5 * time.Second
}

func parsePreferredModelCandidate(fieldRaw json.RawMessage) (string, error) {
	if string(fieldRaw) == "null" {
		return "", nil
	}
	var modelName string
	if err := json.Unmarshal(fieldRaw, &modelName); err != nil {
		return "", fmt.Errorf("guardian_preferred_model must be a string or null")
	}
	return strings.TrimSpace(modelName), nil
}

func (h *Handler) ensureModelAvailable(ctx context.Context, name string) error {
	if h.modelCache != nil && h.modelCache.Contains(name) {
		return nil
	}
	if farmguardian.AutoPullEnabled() && farmguardian.IsLocalOllamaConfigured() && h.modelCache != nil {
		if err := h.modelCache.PullAndRefresh(ctx, name); err != nil {
			return fmt.Errorf("model pull failed: %w", err)
		}
		if h.modelCache.Contains(name) {
			return nil
		}
	}
	return fmt.Errorf("model is not loaded in Ollama — use POST /guardian/models/pull or enable GUARDIAN_OLLAMA_AUTO_PULL")
}
