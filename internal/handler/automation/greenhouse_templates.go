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
//	  "temp_sensor_id": 3,         // optional — enables high-temp + night retract
//	  "allow_missing_lux_sensor": false,  // WS6: must be true to skip high-lux when no lux_sensor_id
//	  "allow_missing_temp_sensor": false  // WS6: must be true to skip temp-driven rules when no temp_sensor_id
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
		ZoneID                int64  `json:"zone_id"`
		ShadeActuatorID       *int64 `json:"shade_actuator_id"`
		FanActuatorID         *int64 `json:"fan_actuator_id"`
		LuxSensorID           *int64 `json:"lux_sensor_id"`
		TempSensorID          *int64 `json:"temp_sensor_id"`
		AllowMissingLuxSensor bool   `json:"allow_missing_lux_sensor"`
		AllowMissingTempSensor bool  `json:"allow_missing_temp_sensor"`
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
	skipped, err := planGreenhouseTemplateSkips(
		body.ShadeActuatorID, body.FanActuatorID,
		body.LuxSensorID, body.TempSensorID,
		body.AllowMissingLuxSensor, body.AllowMissingTempSensor,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate sensors belong to this farm and match expected types when provided.
	if body.LuxSensorID != nil {
		if msg, ok := checkSensorFarm(ctx, h.q, farmID, *body.LuxSensorID, "lux_sensor_id"); !ok {
			httputil.WriteError(w, http.StatusBadRequest, msg)
			return
		}
		if err := validateTemplateLuxSensor(ctx, h.q, farmID, *body.LuxSensorID); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if body.TempSensorID != nil {
		if msg, ok := checkSensorFarm(ctx, h.q, farmID, *body.TempSensorID, "temp_sensor_id"); !ok {
			httputil.WriteError(w, http.StatusBadRequest, msg)
			return
		}
		if err := validateTemplateTempSensor(ctx, h.q, farmID, *body.TempSensorID); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	interlocks, err := ZoneSensorInterlocks(ctx, h.q, body.ZoneID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to evaluate zone sensors")
		return
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

	result["skipped_rule_families"] = skipped
	result["zone_sensor_interlocks"] = interlocks
	result["required_sensors"] = map[string]any{
		ghRuleFamilyHighLux: map[string]any{
			"sensor_types": []string{"lux", "par"},
			"field":        "lux_sensor_id",
		},
		ghRuleFamilyHighTemp: map[string]any{
			"sensor_types": []string{"temperature", "air_temperature"},
			"field":        "temp_sensor_id",
		},
		ghRuleFamilyNightRetract: map[string]any{
			"sensor_types": []string{"temperature"},
			"field":        "temp_sensor_id",
			"note":         "Night retract uses temp proxy; cron schedules may add preconditions on the same temp sensor",
		},
	}

	httputil.WriteJSON(w, http.StatusCreated, result)
}
