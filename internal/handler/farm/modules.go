package farm

import (
	"encoding/json"
	"net/http"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmmodules"
	"gr33n-api/internal/httputil"
)

// GET /farms/{id}/modules
func (h *Handler) ListModules(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListFarmActiveModules(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list modules")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreFarmActiveModule{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// PATCH /farms/{id}/modules/{schema}
func (h *Handler) PatchModule(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	schema := r.PathValue("schema")
	if schema == "" {
		httputil.WriteError(w, http.StatusBadRequest, "module schema required")
		return
	}
	var body struct {
		IsEnabled     bool            `json:"is_enabled"`
		Configuration json.RawMessage `json:"configuration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	cfg := body.Configuration
	if len(cfg) == 0 {
		cfg = json.RawMessage(`{}`)
	}
	row, err := h.q.UpsertFarmActiveModule(r.Context(), db.UpsertFarmActiveModuleParams{
		FarmID:           farmID,
		ModuleSchemaName: schema,
		IsEnabled:        body.IsEnabled,
		Column4:          cfg,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update module")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// KnownModules returns the canonical module list for UI seed hints.
func KnownModules() []map[string]any {
	return []map[string]any{
		{"schema": farmmodules.SchemaCrops, "label": "Crops & grow cycles", "default_enabled": true},
		{"schema": farmmodules.SchemaNaturalFarming, "label": "Natural farming inputs", "default_enabled": true},
		{"schema": farmmodules.SchemaAnimals, "label": "Animal husbandry", "default_enabled": false},
		{"schema": farmmodules.SchemaAquaponics, "label": "Aquaponics", "default_enabled": false},
	}
}

// GET /farm-modules/catalog — static module metadata.
func (h *Handler) ModuleCatalog(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, KnownModules())
}

// GET /farms/{id}/system-logs
func (h *Handler) ListSystemLogs(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	level := r.URL.Query().Get("level")
	limit := int32(100)
	fid := farmID
	rows, err := h.q.ListSystemLogsByFarm(r.Context(), db.ListSystemLogsByFarmParams{
		FarmID:  &fid,
		Column2: level,
		Limit:   limit,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list system logs")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreSystemLog{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"logs": rows})
}
