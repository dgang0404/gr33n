package zone

import (
	"encoding/json"
	"fmt"

	"gr33n-api/internal/platform/domainenums"
)

// GreenhouseClimate is the typed JSON schema stored at
// meta_data.greenhouse_climate on zones with zone_type="greenhouse".
// All fields are optional to allow incremental population.
type GreenhouseClimate struct {
	// CoverType describes the glazing material: glass, polycarbonate, or film.
	CoverType string `json:"cover_type,omitempty"`
	// ShadeActuatorID is the actuator responsible for shade-screen deployment.
	ShadeActuatorID *int64 `json:"shade_actuator_id,omitempty"`
	// VentActuatorID is the ridge vent or motorised opening actuator.
	VentActuatorID *int64 `json:"vent_actuator_id,omitempty"`
	// FanActuatorIDs lists exhaust or circulation fan actuator IDs.
	FanActuatorIDs []int64 `json:"fan_actuator_ids,omitempty"`
	// AutomationPolicy controls how rules interact with this zone:
	//   auto          — rules fire automatically (sensors required)
	//   manual        — operator-only commands; rules disabled
	//   schedule_only — cron retract/deploy only; no sensor predicates
	AutomationPolicy string `json:"automation_policy,omitempty"`
	// Notes is free-text for operator observations (end-wall materials, etc.).
	Notes string `json:"notes,omitempty"`
}

// ValidateGreenhouseClimate parses and validates a greenhouse_climate value
// extracted from zone meta_data. It returns the parsed struct, a slice of
// non-fatal warnings (e.g. "auto policy but no shade actuator"), and any
// hard error that should block the request.
func ValidateGreenhouseClimate(raw json.RawMessage) (*GreenhouseClimate, []string, error) {
	if len(raw) == 0 {
		return nil, nil, nil
	}
	var gc GreenhouseClimate
	if err := json.Unmarshal(raw, &gc); err != nil {
		return nil, nil, fmt.Errorf("greenhouse_climate: %w", err)
	}

	if gc.CoverType != "" {
		if !domainenums.IsValidGreenhouseCoverType(gc.CoverType) {
			return nil, nil, fmt.Errorf(
				"greenhouse_climate.cover_type %q is not valid; must be one of: glass, polycarbonate, film",
				gc.CoverType,
			)
		}
	}

	if gc.AutomationPolicy != "" {
		if !domainenums.IsValidGreenhouseAutomationPolicy(gc.AutomationPolicy) {
			return nil, nil, fmt.Errorf(
				"greenhouse_climate.automation_policy %q is not valid; must be one of: auto, manual, schedule_only",
				gc.AutomationPolicy,
			)
		}
	}

	var warnings []string
	if gc.AutomationPolicy == "auto" {
		if gc.ShadeActuatorID == nil && len(gc.FanActuatorIDs) == 0 {
			warnings = append(warnings,
				"automation_policy=auto but no shade_actuator_id or fan_actuator_ids linked; rules will not control any actuators")
		}
	}

	return &gc, warnings, nil
}

// ExtractGreenhouseClimate reads the greenhouse_climate key from zone meta_data.
// Returns nil, nil if the key is absent.
func ExtractGreenhouseClimate(meta json.RawMessage) (json.RawMessage, error) {
	if len(meta) == 0 {
		return nil, nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(meta, &m); err != nil {
		return nil, err
	}
	return m["greenhouse_climate"], nil
}
