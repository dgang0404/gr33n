package farm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	existing, err := h.q.GetFarmByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "farm not found")
		return
	}

	var newModel *string
	if string(fieldRaw) == "null" {
		newModel = nil
	} else {
		var modelName string
		if err := json.Unmarshal(fieldRaw, &modelName); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "guardian_preferred_model must be a string or null")
			return
		}
		trimmed := strings.TrimSpace(modelName)
		if trimmed == "" {
			newModel = nil
		} else {
			if h.modelCache != nil && !h.modelCache.Contains(trimmed) {
				httputil.WriteError(w, http.StatusBadRequest, "model is not loaded in Ollama — pull it first or pick an available model from GET /guardian/models")
				return
			}
			newModel = &trimmed
		}
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
