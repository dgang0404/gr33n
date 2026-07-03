package sensor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type calPoint struct {
	Raw       float64 `json:"raw"`
	Reference float64 `json:"reference"`
}

// PATCH /sensors/{id}/calibration
//
// Two-point calibration for EC/pH (buffer solutions) or single-offset for temperature.
// Stores computed slope/offset in sensors.calibration_data JSONB.
func (h *Handler) PatchCalibration(w http.ResponseWriter, r *http.Request) {
	sensorID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid sensor id")
		return
	}

	var body struct {
		PointA *calPoint `json:"point_a"`
		PointB *calPoint `json:"point_b"`
		Offset *float64  `json:"offset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := r.Context()
	s0, err := h.q.GetSensorByID(ctx, sensorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load sensor")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, s0.FarmID) {
		return
	}

	var cal map[string]any
	switch {
	case body.PointA != nil && body.PointB != nil:
		denom := body.PointB.Raw - body.PointA.Raw
		if denom == 0 {
			httputil.WriteError(w, http.StatusBadRequest, "calibration points must have distinct raw readings")
			return
		}
		slope := (body.PointB.Reference - body.PointA.Reference) / denom
		offset := body.PointA.Reference - slope*body.PointA.Raw
		cal = map[string]any{
			"method": "two_point",
			"slope":  slope,
			"offset": offset,
			"points": []calPoint{*body.PointA, *body.PointB},
		}
	case body.Offset != nil:
		cal = map[string]any{
			"method": "offset",
			"slope":  1.0,
			"offset": *body.Offset,
		}
	default:
		httputil.WriteError(w, http.StatusBadRequest, "provide point_a+point_b or offset")
		return
	}

	calBytes, err := json.Marshal(cal)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to encode calibration")
		return
	}

	row, err := h.q.UpdateSensorCalibration(ctx, db.UpdateSensorCalibrationParams{
		ID:              sensorID,
		CalibrationData: calBytes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "sensor not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save calibration: %v", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}
