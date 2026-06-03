package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	db "gr33n-api/internal/db"
)

// Greenhouse template rule families (Phase 36 WS6).
const (
	ghRuleFamilyHighLux = "high_lux"
	ghRuleFamilyHighTemp = "high_temp"
	ghRuleFamilyNightRetract = "night_retract"
)

// IsLuxPARSensorType reports whether a sensor can drive high-lux shade rules.
func IsLuxPARSensorType(sensorType string) bool {
	t := strings.ToLower(strings.TrimSpace(sensorType))
	return t == "lux" || strings.Contains(t, "lux") || t == "par" || strings.Contains(t, "par")
}

// IsTempSensorType reports whether a sensor can drive high-temp / night-retract rules.
func IsTempSensorType(sensorType string) bool {
	t := strings.ToLower(strings.TrimSpace(sensorType))
	return strings.Contains(t, "temp")
}

// IsHumiditySensorType reports whether a sensor can drive humidity interlocks (future rules).
func IsHumiditySensorType(sensorType string) bool {
	t := strings.ToLower(strings.TrimSpace(sensorType))
	return strings.Contains(t, "humid") || t == "rh"
}

// GreenhouseRuleFamily classifies a GH automation rule by name prefix.
func GreenhouseRuleFamily(ruleName string) string {
	n := strings.ToLower(ruleName)
	if strings.Contains(n, "high lux") {
		return ghRuleFamilyHighLux
	}
	if strings.Contains(n, "high temp") {
		return ghRuleFamilyHighTemp
	}
	if strings.Contains(n, "night retract") {
		return ghRuleFamilyNightRetract
	}
	return ""
}

// ZoneSensorInterlockStatus summarizes which climate sensors exist in a zone.
type ZoneSensorInterlockStatus struct {
	HasLux      bool `json:"has_lux_or_par"`
	HasTemp     bool `json:"has_temperature"`
	HasHumidity bool `json:"has_humidity"`
}

// ZoneSensorInterlocks scans non-deleted sensors assigned to zoneID.
func ZoneSensorInterlocks(ctx context.Context, q *db.Queries, zoneID int64) (ZoneSensorInterlockStatus, error) {
	zid := zoneID
	sensors, err := q.ListSensorsByZone(ctx, &zid)
	if err != nil {
		return ZoneSensorInterlockStatus{}, err
	}
	var st ZoneSensorInterlockStatus
	for _, s := range sensors {
		if s.DeletedAt.Valid {
			continue
		}
		if IsLuxPARSensorType(s.SensorType) {
			st.HasLux = true
		}
		if IsTempSensorType(s.SensorType) {
			st.HasTemp = true
		}
		if IsHumiditySensorType(s.SensorType) {
			st.HasHumidity = true
		}
	}
	return st, nil
}

func triggerSensorID(rule db.Gr33ncoreAutomationRule) (int64, bool) {
	if len(rule.TriggerConfiguration) == 0 {
		return 0, false
	}
	var cfg map[string]any
	if err := json.Unmarshal(rule.TriggerConfiguration, &cfg); err != nil {
		return 0, false
	}
	return jsonInt64(cfg["sensor_id"])
}

func triggerInterlockOverride(rule db.Gr33ncoreAutomationRule) bool {
	if len(rule.TriggerConfiguration) == 0 {
		return false
	}
	var cfg map[string]any
	if err := json.Unmarshal(rule.TriggerConfiguration, &cfg); err != nil {
		return false
	}
	v, ok := cfg["sensor_interlock_override"].(bool)
	return ok && v
}

func jsonInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case float64:
		return int64(x), true
	case int64:
		return x, true
	case int:
		return int64(x), true
	}
	return 0, false
}

// ValidateGreenhouseRuleActivation blocks activating GH sensor rules without a
// valid sensor (unless trigger_configuration.sensor_interlock_override is true).
func ValidateGreenhouseRuleActivation(ctx context.Context, q *db.Queries, rule db.Gr33ncoreAutomationRule) error {
	family := GreenhouseRuleFamily(rule.Name)
	if family == "" {
		return nil
	}
	if triggerInterlockOverride(rule) {
		return nil
	}
	sensorID, ok := triggerSensorID(rule)
	if !ok || sensorID <= 0 {
		return fmt.Errorf(
			"cannot activate %q: missing sensor_id in trigger_configuration (set sensor_interlock_override only when operator waived the interlock)",
			rule.Name,
		)
	}
	sensor, err := q.GetSensorByID(ctx, sensorID)
	if err != nil {
		return fmt.Errorf("cannot activate %q: sensor %d not found", rule.Name, sensorID)
	}
	if sensor.FarmID != rule.FarmID {
		return fmt.Errorf("cannot activate %q: sensor %d belongs to another farm", rule.Name, sensorID)
	}
	switch family {
	case ghRuleFamilyHighLux:
		if !IsLuxPARSensorType(sensor.SensorType) {
			return fmt.Errorf(
				"cannot activate %q: sensor %d type %q is not lux/par — add a lux or PAR sensor or waive with sensor_interlock_override",
				rule.Name, sensorID, sensor.SensorType,
			)
		}
	case ghRuleFamilyHighTemp, ghRuleFamilyNightRetract:
		if !IsTempSensorType(sensor.SensorType) {
			return fmt.Errorf(
				"cannot activate %q: sensor %d type %q is not a temperature sensor",
				rule.Name, sensorID, sensor.SensorType,
			)
		}
	}
	return nil
}

// validateTemplateLuxSensor ensures lux_sensor_id references a lux/par sensor on the farm.
func validateTemplateLuxSensor(ctx context.Context, q *db.Queries, farmID, sensorID int64) error {
	sensor, err := q.GetSensorByID(ctx, sensorID)
	if err != nil {
		return fmt.Errorf("lux_sensor_id: sensor %d not found", sensorID)
	}
	if sensor.FarmID != farmID {
		return fmt.Errorf("lux_sensor_id: sensor %d does not belong to this farm", sensorID)
	}
	if !IsLuxPARSensorType(sensor.SensorType) {
		return fmt.Errorf("lux_sensor_id: sensor %d must be type lux or par (got %q)", sensorID, sensor.SensorType)
	}
	return nil
}

// validateTemplateTempSensor ensures temp_sensor_id references a temperature sensor.
func validateTemplateTempSensor(ctx context.Context, q *db.Queries, farmID, sensorID int64) error {
	sensor, err := q.GetSensorByID(ctx, sensorID)
	if err != nil {
		return fmt.Errorf("temp_sensor_id: sensor %d not found", sensorID)
	}
	if sensor.FarmID != farmID {
		return fmt.Errorf("temp_sensor_id: sensor %d does not belong to this farm", sensorID)
	}
	if !IsTempSensorType(sensor.SensorType) {
		return fmt.Errorf("temp_sensor_id: sensor %d must be a temperature sensor (got %q)", sensorID, sensor.SensorType)
	}
	return nil
}

// planGreenhouseTemplateSkips returns rule families that will not be created
// given the request (before calling SQL).
func planGreenhouseTemplateSkips(
	shadeID, fanID, luxID, tempID *int64,
	allowMissingLux, allowMissingTemp bool,
) (skipped []string, err error) {
	if shadeID != nil && luxID == nil && !allowMissingLux {
		return nil, fmt.Errorf(
			"lux_sensor_id is required when shade_actuator_id is set (high-lux deploy template); pass allow_missing_lux_sensor=true to skip that rule family",
		)
	}
	if fanID != nil && tempID == nil && !allowMissingTemp {
		return nil, fmt.Errorf(
			"temp_sensor_id is required when fan_actuator_id is set (high-temp fan template); pass allow_missing_temp_sensor=true to skip that rule family",
		)
	}
	if shadeID != nil && luxID == nil && allowMissingLux {
		skipped = append(skipped, ghRuleFamilyHighLux)
	}
	if fanID != nil && tempID == nil && allowMissingTemp {
		skipped = append(skipped, ghRuleFamilyHighTemp)
	}
	if shadeID != nil && tempID == nil && allowMissingTemp {
		skipped = append(skipped, ghRuleFamilyNightRetract)
	}
	return skipped, nil
}
