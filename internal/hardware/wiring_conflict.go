package hardware

import "fmt"

// WiringEntity is a sensor or actuator with resolved wiring for conflict checks.
type WiringEntity struct {
	ID     int64
	Name   string
	Type   string // "sensor" or "actuator"
	Wiring map[string]any
}

// WiringConflict describes an overlapping pin/channel on one edge device.
type WiringConflict struct {
	EntityType string
	EntityID   int64
	EntityName string
	Message    string
}

// WiringConflictQuery checks a draft wiring payload against farm entities (UI preview).
type WiringConflictQuery struct {
	Wiring     map[string]any
	EntityType string
	EntityID   int64
	Sensors    []WiringEntity
	Actuators  []WiringEntity
}

// FindWiringConflict returns a conflict when another entity shares gpio_pin or i2c_channel on the same device.
func FindWiringConflict(q WiringConflictQuery) *WiringConflict {
	if q.Wiring == nil {
		return nil
	}
	dev, ok := int64Field(q.Wiring["device_id"])
	if !ok || dev <= 0 {
		return nil
	}
	newSource, _ := q.Wiring["source"].(string)
	if pin, ok := intField(q.Wiring["gpio_pin"]); ok {
		for _, s := range q.Sensors {
			if q.EntityType == "sensor" && s.ID == q.EntityID {
				continue
			}
			if wiringSharesGPIO(s.Wiring, dev, pin) {
				otherSource, _ := s.Wiring["source"].(string)
				if sharedDHT22GPIOAllowed(newSource, otherSource) {
					continue
				}
				return conflictFrom(s)
			}
		}
		for _, a := range q.Actuators {
			if q.EntityType == "actuator" && a.ID == q.EntityID {
				continue
			}
			if wiringSharesGPIO(a.Wiring, dev, pin) {
				return conflictFrom(a)
			}
		}
	}
	if ch, ok := intField(q.Wiring["i2c_channel"]); ok && q.EntityType == "sensor" {
		for _, s := range q.Sensors {
			if s.ID == q.EntityID {
				continue
			}
			if wiringSharesI2C(s.Wiring, dev, ch) {
				return conflictFrom(s)
			}
		}
	}
	return nil
}

func conflictFrom(e WiringEntity) *WiringConflict {
	return &WiringConflict{
		EntityType: e.Type,
		EntityID:   e.ID,
		EntityName: e.Name,
		Message:    fmt.Sprintf("%s %d (%s) already uses this pin/channel on the device", e.Type, e.ID, e.Name),
	}
}

func wiringSharesGPIO(w map[string]any, deviceID int64, pin int) bool {
	if w == nil {
		return false
	}
	d, ok := int64Field(w["device_id"])
	if !ok || d != deviceID {
		return false
	}
	p, ok := intField(w["gpio_pin"])
	return ok && p == pin
}

func wiringSharesI2C(w map[string]any, deviceID int64, channel int) bool {
	if w == nil {
		return false
	}
	d, ok := int64Field(w["device_id"])
	if !ok || d != deviceID {
		return false
	}
	ch, ok := intField(w["i2c_channel"])
	return ok && ch == channel
}

func intField(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func int64Field(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}
