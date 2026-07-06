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
	if len(strings.TrimSpace(string(body))) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "request body required")
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	hasCounsel := raw["guardian_counsel_model"] != nil
	hasQuick := raw["guardian_quick_model"] != nil
	hasPreferred := raw["guardian_preferred_model"] != nil
	hasTimeout := raw["guardian_grounded_timeout_seconds"] != nil
	if !hasCounsel && !hasQuick && !hasPreferred && !hasTimeout {
		httputil.WriteError(w, http.StatusBadRequest, "at least one guardian model policy field is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	existing, err := h.q.GetFarmByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "farm not found")
		return
	}

	counsel := farmguardian.FarmCounselModel(&existing)
	quick := farmguardian.FarmQuickModel(&existing)
	timeout := existing.GuardianGroundedTimeoutSeconds

	if hasPreferred {
		c, err := parseOptionalModelField(raw["guardian_preferred_model"], "guardian_preferred_model")
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		counsel = c
	}
	if hasCounsel {
		c, err := parseOptionalModelField(raw["guardian_counsel_model"], "guardian_counsel_model")
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		counsel = c
	}
	if hasQuick {
		q, err := parseOptionalModelField(raw["guardian_quick_model"], "guardian_quick_model")
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		quick = q
	}
	if hasTimeout {
		t, err := parseOptionalTimeoutField(raw["guardian_grounded_timeout_seconds"])
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		timeout = t
	}

	if counsel != nil && strings.TrimSpace(*counsel) != "" {
		if err := h.ensureModelAvailable(ctx, *counsel); err != nil {
			if strings.Contains(err.Error(), "not loaded") || strings.Contains(err.Error(), "too small") {
				httputil.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			httputil.WriteError(w, http.StatusBadGateway, err.Error())
			return
		}
		if err := h.ensureGroundedFarmModel(*counsel); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if quick != nil && strings.TrimSpace(*quick) != "" {
		if err := h.ensureModelAvailable(ctx, *quick); err != nil {
			if strings.Contains(err.Error(), "not loaded") {
				httputil.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			httputil.WriteError(w, http.StatusBadGateway, err.Error())
			return
		}
	}

	updated, err := h.q.UpdateFarmGuardianModelPolicy(ctx, db.UpdateFarmGuardianModelPolicyParams{
		ID:                             id,
		GuardianCounselModel:           counsel,
		GuardianQuickModel:             quick,
		GuardianGroundedTimeoutSeconds: timeout,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update farm settings")
		return
	}

	from := modelLabelPtr(existing.GuardianCounselModel)
	if from == "" {
		from = modelLabelPtr(existing.GuardianPreferredModel)
	}
	to := modelLabelPtr(updated.GuardianCounselModel)
	if to == "" {
		to = modelLabelPtr(updated.GuardianPreferredModel)
	}
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
				"policy":             "counsel",
			},
		})
	}

	httputil.WriteJSON(w, http.StatusOK, guardianModelPolicyResponse(updated))
}

func guardianModelPolicyResponse(f db.Gr33ncoreFarm) map[string]any {
	return map[string]any{
		"id":                                f.ID,
		"guardian_preferred_model":          f.GuardianPreferredModel,
		"guardian_counsel_model":            f.GuardianCounselModel,
		"guardian_quick_model":              f.GuardianQuickModel,
		"guardian_grounded_timeout_seconds": f.GuardianGroundedTimeoutSeconds,
	}
}

func parseOptionalModelField(fieldRaw json.RawMessage, fieldName string) (*string, error) {
	if string(fieldRaw) == "null" {
		return nil, nil
	}
	var modelName string
	if err := json.Unmarshal(fieldRaw, &modelName); err != nil {
		return nil, fmt.Errorf("%s must be a string or null", fieldName)
	}
	trimmed := strings.TrimSpace(modelName)
	if trimmed == "" {
		return nil, nil
	}
	return &trimmed, nil
}

func parseOptionalTimeoutField(fieldRaw json.RawMessage) (*int32, error) {
	if string(fieldRaw) == "null" {
		return nil, nil
	}
	var n int32
	if err := json.Unmarshal(fieldRaw, &n); err != nil {
		return nil, fmt.Errorf("guardian_grounded_timeout_seconds must be a positive integer or null")
	}
	if n <= 0 {
		return nil, fmt.Errorf("guardian_grounded_timeout_seconds must be positive")
	}
	return &n, nil
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

// Farm counsel model must meet the grounded context minimum.
func (h *Handler) ensureGroundedFarmModel(name string) error {
	if h.modelCache == nil {
		return nil
	}
	info, ok := h.modelCache.Get(name)
	if !ok {
		return nil
	}
	if info.ContextWindow > 0 && info.ContextWindow < farmguardian.GuardianMinContextWindow {
		return fmt.Errorf(
			"guardian_counsel_model %q context window (%d) is below the minimum required for grounded Guardian chat (%d)",
			info.Name, info.ContextWindow, farmguardian.GuardianMinContextWindow,
		)
	}
	return nil
}
