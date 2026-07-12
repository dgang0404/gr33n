package farm

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// PATCH /farms/{id}/site — set lat/long + optional elevation (Phase 66).
func (h *Handler) PatchSite(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var req struct {
		Latitude   float64  `json:"latitude"`
		Longitude  float64  `json:"longitude"`
		ElevationM *float64 `json:"elevation_m"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		httputil.WriteError(w, http.StatusBadRequest, "latitude/longitude out of range")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, id) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	farm, err := h.q.UpdateFarmSiteCoords(ctx, db.UpdateFarmSiteCoordsParams{
		ID:         id,
		Longitude:  req.Longitude,
		Latitude:   req.Latitude,
		ElevationM: req.ElevationM,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update site coordinates")
		return
	}
	// PostGIS geometry JSON-encodes as EWKB/base64 — rewrite as GeoJSON so the UI can parse it.
	httputil.WriteJSON(w, http.StatusOK, farmWithGeoJSONPoint(farm, req.Longitude, req.Latitude))
}

// farmWithGeoJSONPoint returns the farm as a JSON object with location_gis as a GeoJSON Point.
func farmWithGeoJSONPoint(farm db.Gr33ncoreFarm, longitude, latitude float64) map[string]any {
	raw, err := json.Marshal(farm)
	if err != nil {
		return map[string]any{
			"id": farm.ID,
			"location_gis": map[string]any{
				"type":        "Point",
				"coordinates": []float64{longitude, latitude},
			},
		}
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{
			"id": farm.ID,
			"location_gis": map[string]any{
				"type":        "Point",
				"coordinates": []float64{longitude, latitude},
			},
		}
	}
	out["location_gis"] = map[string]any{
		"type":        "Point",
		"coordinates": []float64{longitude, latitude},
	}
	return out
}
