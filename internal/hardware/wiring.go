// Package hardware — structured Pi wiring metadata stored in sensors/actuators config JSONB.
package hardware

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Wiring describes how a sensor or actuator is connected on an edge device (Pi).
// Stored at config.wiring on gr33ncore.sensors and gr33ncore.actuators.
type Wiring struct {
	Source      string          `json:"source,omitempty"`
	GPIOPin     *int            `json:"gpio_pin,omitempty"`
	I2CChannel  *int            `json:"i2c_channel,omitempty"`
	I2CAddress  string          `json:"i2c_address,omitempty"`
	SerialPort  string          `json:"serial_port,omitempty"`
	DeviceID    *int64          `json:"device_id,omitempty"`
	Notes       string          `json:"notes,omitempty"`
	Inputs      json.RawMessage `json:"inputs,omitempty"` // derived sensors only
}

var knownSensorSources = map[string]struct{}{
	"dht22": {}, "ads1115": {}, "mhz19": {}, "bh1750": {}, "derived": {}, "gpio_digital": {},
}

var knownActuatorSources = map[string]struct{}{
	"gpio_relay": {},
}

// ExtractWiring reads config.wiring from a JSONB config blob.
func ExtractWiring(config json.RawMessage) (*Wiring, error) {
	if len(config) == 0 {
		return nil, nil
	}
	var root struct {
		Wiring *Wiring `json:"wiring"`
	}
	if err := json.Unmarshal(config, &root); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if root.Wiring == nil {
		return nil, nil
	}
	w := *root.Wiring
	if w.Source == "" && w.GPIOPin == nil && w.I2CChannel == nil && w.SerialPort == "" && w.DeviceID == nil {
		return nil, nil
	}
	return &w, nil
}

// MergeWiring returns updated config JSON with wiring merged (or removed when nil).
func MergeWiring(config json.RawMessage, wiring *Wiring) (json.RawMessage, error) {
	root := map[string]json.RawMessage{}
	if len(config) > 0 {
		if err := json.Unmarshal(config, &root); err != nil {
			return nil, fmt.Errorf("config: %w", err)
		}
	}
	if wiring == nil {
		delete(root, "wiring")
	} else {
		b, err := json.Marshal(wiring)
		if err != nil {
			return nil, err
		}
		root["wiring"] = b
	}
	if len(root) == 0 {
		return json.RawMessage(`{}`), nil
	}
	out, err := json.Marshal(root)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ValidateSensorWiring checks a sensor wiring payload.
func ValidateSensorWiring(w *Wiring) error {
	if w == nil {
		return nil
	}
	if w.Source == "" {
		return fmt.Errorf("wiring.source is required")
	}
	if _, ok := knownSensorSources[w.Source]; !ok {
		return fmt.Errorf("wiring.source %q is not supported", w.Source)
	}
	switch w.Source {
	case "dht22", "gpio_digital":
		if w.GPIOPin == nil {
			return fmt.Errorf("wiring.gpio_pin is required for %s", w.Source)
		}
	case "ads1115":
		if w.I2CChannel == nil {
			return fmt.Errorf("wiring.i2c_channel is required for ads1115")
		}
	case "mhz19":
		if strings.TrimSpace(w.SerialPort) == "" {
			return fmt.Errorf("wiring.serial_port is required for mhz19")
		}
	}
	if w.GPIOPin != nil && (*w.GPIOPin < 0 || *w.GPIOPin > 27) {
		return fmt.Errorf("wiring.gpio_pin must be 0–27 (BCM)")
	}
	if w.I2CChannel != nil && (*w.I2CChannel < 0 || *w.I2CChannel > 3) {
		return fmt.Errorf("wiring.i2c_channel must be 0–3")
	}
	return nil
}

// ValidateActuatorWiring checks an actuator wiring payload.
func ValidateActuatorWiring(w *Wiring) error {
	if w == nil {
		return nil
	}
	src := w.Source
	if src == "" {
		src = "gpio_relay"
		w.Source = src
	}
	if _, ok := knownActuatorSources[src]; !ok {
		return fmt.Errorf("wiring.source %q is not supported for actuators", src)
	}
	if w.GPIOPin == nil {
		return fmt.Errorf("wiring.gpio_pin is required for actuators")
	}
	if *w.GPIOPin < 0 || *w.GPIOPin > 27 {
		return fmt.Errorf("wiring.gpio_pin must be 0–27 (BCM)")
	}
	return nil
}

// FormatLabel returns farmer-facing wiring summary text.
func FormatLabel(w *Wiring) string {
	if w == nil {
		return ""
	}
	parts := []string{}
	if w.Source != "" {
		parts = append(parts, strings.ToUpper(w.Source))
	}
	if w.GPIOPin != nil {
		parts = append(parts, fmt.Sprintf("BCM GPIO %d", *w.GPIOPin))
	}
	if w.I2CChannel != nil {
		parts = append(parts, fmt.Sprintf("I2C ch %d", *w.I2CChannel))
	}
	if w.SerialPort != "" {
		parts = append(parts, w.SerialPort)
	}
	return strings.Join(parts, " · ")
}
