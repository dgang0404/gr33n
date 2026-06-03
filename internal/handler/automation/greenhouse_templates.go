package automation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// POST /farms/{id}/automation/rule-templates/greenhouse
//
// Applies Phase 36 greenhouse climate rule templates (high-lux shade deploy,
// high-temp fan, night retract) for a specified zone. All actuator/sensor
// references are optional; pass null to skip that rule family.
//
// Request body:
//
//	{
//	  "zone_id": 12,
//	  "shade_actuator_id": 20,    // optional — enables deploy/retract rules
//	  "fan_actuator_id": 21,       // optional — enables high-temp fan rule
//	  "lux_sensor_id": 5,          // optional — enables high-lux shade rule
//	  "temp_sensor_id": 3          // optional — enables high-temp + night retract
//	}
func (h *Handler) ApplyGreenhouseRuleTemplates(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}

	var body struct {
		ZoneID          int64  `json:"zone_id"`
		ShadeActuatorID *int64 `json:"shade_actuator_id"`
		FanActuatorID   *int64 `json:"fan_actuator_id"`
		LuxSensorID     *int64 `json:"lux_sensor_id"`
		TempSensorID    *int64 `json:"temp_sensor_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.ZoneID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "zone_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Verify the zone exists and belongs to this farm.
	zone, err := h.q.GetZoneByID(ctx, body.ZoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "zone not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load zone")
		return
	}
	if zone.FarmID != farmID {
		httputil.WriteError(w, http.StatusForbidden, "zone does not belong to this farm")
		return
	}

	// Validate actuators belong to this farm when provided.
	if body.ShadeActuatorID != nil {
		if msg, ok := checkActuatorFarm(ctx, h.q, farmID, *body.ShadeActuatorID, "shade_actuator_id"); !ok {
			httputil.WriteError(w, http.StatusBadRequest, msg)
			return
		}
	}
	if body.FanActuatorID != nil {
		if msg, ok := checkActuatorFarm(ctx, h.q, farmID, *body.FanActuatorID, "fan_actuator_id"); !ok {
			httputil.WriteError(w, http.StatusBadRequest, msg)
			return
		}
	}
	// Validate sensors belong to this farm when provided.
	if body.LuxSensorID != nil {
		if msg, ok := checkSensorFarm(ctx, h.q, farmID, *body.LuxSensorID, "lux_sensor_id"); !ok {
			httputil.WriteError(w, http.StatusBadRequest, msg)
			return
		}
	}
	if body.TempSensorID != nil {
		if msg, ok := checkSensorFarm(ctx, h.q, farmID, *body.TempSensorID, "temp_sensor_id"); !ok {
			httputil.WriteError(w, http.StatusBadRequest, msg)
			return
		}
	}

	result, err := h.q.ApplyGreenhouseRuleTemplates(ctx,
		farmID, body.ZoneID,
		body.ShadeActuatorID, body.FanActuatorID,
		body.LuxSensorID, body.TempSensorID,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to apply greenhouse rule templates")
		return
	}
	if errMsg, ok := result["error"]; ok {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Sprintf("%v", errMsg))
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, result)
}
