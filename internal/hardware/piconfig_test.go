package hardware

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGeneratePiConfigYAML_roundTrip(t *testing.T) {
	dev := int64(1)
	interval := int32(60)
	opts := PiConfigOptions{
		FarmID:     1,
		DeviceID:   1,
		DeviceUID:  "demo-veg-relay-01",
		DeviceName: "Veg Pi",
		BaseURL:    "http://192.168.1.100:8080",
		Sensors: []PiConfigSensor{
			{
				ID: 3, SensorType: "temperature", ReadingIntervalSeconds: &interval,
				Config: mustConfig(t, map[string]any{
					"wiring": map[string]any{"source": "dht22", "gpio_pin": 4, "device_id": 1},
				}),
			},
			{
				ID: 8, SensorType: "ec", ReadingIntervalSeconds: &interval,
				Config: mustConfig(t, map[string]any{
					"wiring": map[string]any{"source": "ads1115", "i2c_channel": 1, "device_id": 1},
				}),
			},
		},
		Actuators: []PiConfigActuator{
			{
				ID: 1, ActuatorType: "light", DeviceID: &dev,
				Config: mustConfig(t, map[string]any{
					"wiring": map[string]any{"source": "gpio_relay", "gpio_pin": 17, "device_id": 1},
				}),
			},
		},
	}
	out, err := GeneratePiConfigYAML(opts)
	if err != nil {
		t.Fatal(err)
	}
	yaml := string(out)
	if !containsAll(yaml, "sensor_id: 3", "pin: 4", "sensor_id: 8", "channel: 1", "gpio_pin: 17", "farm_id: 1") {
		t.Fatalf("yaml missing expected fields:\n%s", yaml)
	}

	sensors, actuators, err := ParsePiConfigYAML(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(sensors) != 2 || len(actuators) != 1 {
		t.Fatalf("parsed %d sensors %d actuators", len(sensors), len(actuators))
	}
	w0 := PiSensorEntryToWiring(sensors[0], 1)
	if w0.Source != "dht22" || w0.GPIOPin == nil || *w0.GPIOPin != 4 {
		t.Fatalf("round-trip sensor0 %+v", w0)
	}
	w1 := PiSensorEntryToWiring(sensors[1], 1)
	if w1.Source != "ads1115" || w1.I2CChannel == nil || *w1.I2CChannel != 1 {
		t.Fatalf("round-trip sensor1 %+v", w1)
	}
	aw := PiActuatorEntryToWiring(actuators[0])
	if aw.GPIOPin == nil || *aw.GPIOPin != 17 {
		t.Fatalf("round-trip actuator %+v", aw)
	}
}

func mustConfig(t *testing.T, v map[string]any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}
