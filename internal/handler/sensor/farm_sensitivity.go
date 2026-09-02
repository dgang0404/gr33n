package sensor

// Phase 217 WS5 — farm-wide default for the 5-tier sensor quality scale's
// band width (per-sensor override lives on gr33ncore.sensors.quality_sensitivity
// instead). Stored in farms.meta_data since it's a single small setting, not
// a first-class column (mirrors weather.PatchSettings' opt-in flags).

import (
	"encoding/json"
	"net/http"
	"strconv"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// PATCH /farms/{id}/sensor-sensitivity
//
// Body: { "default_sensor_sensitivity": "narrow"|"normal"|"wide" }
func (h *Handler) PatchSensitivity(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}

	var body struct {
		DefaultSensorSensitivity string `json:"default_sensor_sensitivity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	switch body.DefaultSensorSensitivity {
	case "narrow", "normal", "wide":
	default:
		httputil.WriteError(w, http.StatusBadRequest, "default_sensor_sensitivity must be narrow, normal, or wide")
		return
	}

	patchJSON, _ := json.Marshal(map[string]any{
		"default_sensor_sensitivity": body.DefaultSensorSensitivity,
	})
	farm, err := h.q.MergeFarmMetaData(r.Context(), db.MergeFarmMetaDataParams{
		ID:    farmID,
		Patch: patchJSON,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update sensor sensitivity")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, farm)
}
