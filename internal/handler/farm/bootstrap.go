package farm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmbootstrap"
	"gr33n-api/internal/httputil"
)

func (h *Handler) runFarmBootstrap(ctx context.Context, farmID int64, template string) (map[string]any, error) {
	if h.pool == nil {
		return nil, errors.New("database pool not configured for bootstrap")
	}
	if farmbootstrap.IsBlankChoice(template) {
		return map[string]any{"skipped": true}, nil
	}
	row := h.pool.QueryRow(ctx,
		`SELECT gr33ncore.apply_farm_bootstrap_template($1, $2)`,
		farmID, template,
	)
	var raw []byte
	if err := row.Scan(&raw); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ApplyFarmBootstrapTemplate — POST /farms/{id}/bootstrap-template
func (h *Handler) ApplyFarmBootstrapTemplate(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	var body struct {
		Template string `json:"template"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if farmbootstrap.IsBlankChoice(body.Template) {
		httputil.WriteError(w, http.StatusBadRequest, "template is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	out, err := h.runFarmBootstrap(ctx, farmID, body.Template)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "bootstrap failed: "+err.Error())
		return
	}
	if errObj, _ := out["error"].(string); errObj == "farm_not_found" {
		httputil.WriteError(w, http.StatusNotFound, "farm not found")
		return
	}
	if errObj, _ := out["error"].(string); errObj == "unknown_template" {
		httputil.WriteError(w, http.StatusBadRequest, "unknown template")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"farm_id":   farmID,
		"bootstrap": out,
	})
}
