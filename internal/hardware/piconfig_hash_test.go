package hardware

import "testing"

func TestPiRuntimeConfigWiringSHA256_stable(t *testing.T) {
	cfg := &PiRuntimeConfig{
		Sensors: []PiRuntimeSensor{
			{SensorID: 3, SensorType: "temperature", Source: "dht22", Pin: intPtr(4), IntervalSeconds: 60},
			{SensorID: 8, SensorType: "ec", Source: "ads1115", Channel: intPtr(1), IntervalSeconds: 60},
		},
		Actuators: []PiRuntimeActuator{
			{ActuatorID: 1, DeviceID: 1, DeviceType: "relay", GPIOPin: intPtr(17)},
		},
	}
	h1 := PiRuntimeConfigWiringSHA256(cfg)
	h2 := PiRuntimeConfigWiringSHA256(cfg)
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash not stable: %q vs %q", h1, h2)
	}
	// Order of input slices should not matter.
	cfg2 := &PiRuntimeConfig{
		Sensors: []PiRuntimeSensor{
			{SensorID: 8, SensorType: "ec", Source: "ads1115", Channel: intPtr(1), IntervalSeconds: 60},
			{SensorID: 3, SensorType: "temperature", Source: "dht22", Pin: intPtr(4), IntervalSeconds: 60},
		},
		Actuators: []PiRuntimeActuator{
			{ActuatorID: 1, DeviceID: 1, DeviceType: "relay", GPIOPin: intPtr(17)},
		},
	}
	if PiRuntimeConfigWiringSHA256(cfg2) != h1 {
		t.Fatalf("hash order-sensitive")
	}
}

func intPtr(n int) *int { return &n }
