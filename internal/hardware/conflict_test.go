package hardware

import "testing"

func TestFindWiringConflictAllowsSharedDHT22(t *testing.T) {
	sensors := []WiringEntity{
		{ID: 3, Name: "Air Temp", Type: "sensor", Wiring: map[string]any{"source": "dht22", "gpio_pin": 4, "device_id": 1}},
	}
	conflict := FindWiringConflict(WiringConflictQuery{
		Wiring:     map[string]any{"source": "dht22", "gpio_pin": 4, "device_id": 1},
		EntityType: "sensor",
		EntityID:   99,
		Sensors:    sensors,
	})
	if conflict != nil {
		t.Fatalf("shared dht22 should be allowed, got %+v", conflict)
	}
}

func TestFindWiringConflictBlocksMixedGPIO(t *testing.T) {
	sensors := []WiringEntity{
		{ID: 3, Name: "Digital", Type: "sensor", Wiring: map[string]any{"source": "gpio_digital", "gpio_pin": 4, "device_id": 1}},
	}
	conflict := FindWiringConflict(WiringConflictQuery{
		Wiring:     map[string]any{"source": "dht22", "gpio_pin": 4, "device_id": 1},
		EntityType: "sensor",
		EntityID:   99,
		Sensors:    sensors,
	})
	if conflict == nil || conflict.EntityID != 3 {
		t.Fatalf("expected conflict, got %+v", conflict)
	}
}

func TestFindWiringConflictCrossType(t *testing.T) {
	actuators := []WiringEntity{
		{ID: 10, Name: "Pump", Type: "actuator", Wiring: map[string]any{"gpio_pin": 17, "device_id": 1}},
	}
	conflict := FindWiringConflict(WiringConflictQuery{
		Wiring:     map[string]any{"source": "gpio_relay", "gpio_pin": 17, "device_id": 1},
		EntityType: "sensor",
		EntityID:   5,
		Actuators:  actuators,
	})
	if conflict == nil || conflict.EntityType != "actuator" {
		t.Fatalf("expected actuator conflict, got %+v", conflict)
	}
}

func TestFindWiringConflictI2CChannel(t *testing.T) {
	sensors := []WiringEntity{
		{ID: 8, Name: "EC", Type: "sensor", Wiring: map[string]any{"source": "ads1115", "i2c_channel": 1, "device_id": 1}},
	}
	conflict := FindWiringConflict(WiringConflictQuery{
		Wiring:     map[string]any{"source": "ads1115", "i2c_channel": 1, "device_id": 1},
		EntityType: "sensor",
		EntityID:   9,
		Sensors:    sensors,
	})
	if conflict == nil || conflict.EntityID != 8 {
		t.Fatalf("got %+v", conflict)
	}
}

