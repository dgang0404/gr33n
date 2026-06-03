package actuator

import (
	"encoding/json"
	"fmt"
)

// GreenhouseActuatorTypes are the Phase 36 first-class greenhouse types.
// Legacy types (shade_cloth_motor, etc.) are accepted without validation
// to preserve backward compatibility.
var GreenhouseActuatorTypes = map[string]struct{}{
	"shade_screen":    {},
	"ridge_vent":      {},
	"exhaust_fan":     {},
	"circulation_fan": {},
	"glazing_panel":   {},
}

// AllKnownActuatorTypes is the full validated set across all bootstrap types.
// Anything not in this set is accepted as a legacy free-text type with a warning.
var AllKnownActuatorTypes = map[string]struct{}{
	// Greenhouse (Phase 36)
	"shade_screen":    {},
	"ridge_vent":      {},
	"exhaust_fan":     {},
	"circulation_fan": {},
	"glazing_panel":   {},
	// Lighting (Phase 35)
	"light":       {},
	"grow_light":  {},
	// Climate / legacy greenhouse
	"shade_cloth_motor": {},
	"humidifier":        {},
	"dehumidifier":      {},
	"co2_injector":      {},
	// Animal / farm utility
	"feeder_hopper": {},
	"water_valve":   {},
	"heat_lamp":     {},
	// Aquaponics
	"return_pump": {},
	"air_pump":    {},
	// Generic
	"relay": {},
	"pump":  {},
}

// ActuatorConfig is the typed config shape for greenhouse actuators.
// All fields are optional; only semantically relevant fields per type
// are expected by the Pi client.
type ActuatorConfig struct {
	// Channel is the Pi GPIO/relay channel number (0-indexed).
	Channel *int `json:"channel,omitempty"`
	// NormallyOpen indicates the relay is normally-open (true) or normally-closed (false).
	NormallyOpen *bool `json:"normally_open,omitempty"`
	// MaxRunSeconds caps motor run time to prevent over-travel (shade_screen, ridge_vent).
	MaxRunSeconds *int `json:"max_run_seconds,omitempty"`
	// PercentOpenSupported signals the Pi can control ridge_vent to a percentage.
	PercentOpenSupported *bool `json:"percent_open_supported,omitempty"`
}

// ValidCommands returns the valid command strings for the given actuator_type.
func ValidCommands(actuatorType string) []string {
	switch actuatorType {
	case "shade_screen":
		return []string{"deploy", "retract", "stop", "on", "off"}
	case "ridge_vent":
		return []string{"open", "close", "stop", "on", "off"}
	case "glazing_panel":
		return []string{"open", "close", "on", "off"}
	case "exhaust_fan", "circulation_fan",
		"humidifier", "dehumidifier", "co2_injector",
		"heat_lamp", "return_pump", "air_pump",
		"light", "grow_light", "relay", "pump":
		return []string{"on", "off"}
	case "water_valve":
		return []string{"open", "close", "on", "off"}
	case "feeder_hopper":
		return []string{"dispense", "on", "off"}
	case "shade_cloth_motor":
		return []string{"on", "off", "deploy", "retract"}
	default:
		return []string{"on", "off"}
	}
}

// ValidateActuatorConfig parses and validates the config JSON for the
// given actuator_type.  Returns a hard error for malformed JSON or
// type-specific constraint violations.
func ValidateActuatorConfig(actuatorType string, cfg json.RawMessage) error {
	if len(cfg) == 0 {
		return nil
	}
	var ac ActuatorConfig
	if err := json.Unmarshal(cfg, &ac); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	switch actuatorType {
	case "shade_screen", "ridge_vent", "glazing_panel":
		if ac.MaxRunSeconds != nil && *ac.MaxRunSeconds <= 0 {
			return fmt.Errorf("config.max_run_seconds must be positive for %s", actuatorType)
		}
	}
	return nil
}
