package hardware

import (
	"encoding/json"
	"testing"
)

func TestBuildPiRuntimeConfig_sensorAndActuatorShape(t *testing.T) {
	dev := int64(1)
	interval := int32(60)
	opts := PiConfigOptions{
		FarmID:     1,
		DeviceID:   1,
		DeviceUID:  "demo-veg-relay-01",
		DeviceName: "Veg Pi",
		Sensors: []PiConfigSensor{
			{
				ID: 3, SensorType: "temperature", ReadingIntervalSeconds: &interval,
				Config: mustConfig(t, map[string]any{
					"wiring": map[string]any{"source": "dht22", "gpio_pin": 4, "device_id": 1},
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
	deviceCfg := mustConfig(t, map[string]any{"mix_channels": []int64{1, 2}})
	cfg, err := BuildPiRuntimeConfig(opts, 7, deviceCfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConfigVersion != 7 || cfg.DeviceUID != "demo-veg-relay-01" || cfg.FarmID != 1 {
		t.Fatalf("header fields: %+v", cfg)
	}
	if len(cfg.Sensors) != 1 || cfg.Sensors[0].Pin == nil || *cfg.Sensors[0].Pin != 4 {
		t.Fatalf("sensor pin: %+v", cfg.Sensors)
	}
	if cfg.Sensors[0].SensorID != 3 || cfg.Sensors[0].IntervalSeconds != 60 {
		t.Fatalf("sensor meta: %+v", cfg.Sensors[0])
	}
	if len(cfg.Actuators) != 1 || cfg.Actuators[0].GPIOPin != 17 {
		t.Fatalf("actuator: %+v", cfg.Actuators)
	}
	if len(cfg.MixChannels) != 2 || cfg.MixChannels[0] != 1 {
		t.Fatalf("mix_channels: %v", cfg.MixChannels)
	}
	if cfg.SchedulePollIntervalSeconds != 30 || cfg.OfflineFlushIntervalSeconds != 60 {
		t.Fatalf("poll defaults: %+v", cfg)
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["gpio_pin"]; ok {
		t.Fatalf("top-level gpio_pin leaked: %v", raw)
	}
	sensors := raw["sensors"].([]any)
	s0 := sensors[0].(map[string]any)
	if _, hasPin := s0["pin"]; !hasPin {
		t.Fatalf("sensor missing pin key: %v", s0)
	}
	if _, hasGpio := s0["gpio_pin"]; hasGpio {
		t.Fatalf("sensor must use pin not gpio_pin: %v", s0)
	}
}

func TestBuildPiRuntimeConfig_emptySlicesNotNull(t *testing.T) {
	cfg, err := BuildPiRuntimeConfig(PiConfigOptions{
		FarmID: 1, DeviceID: 99, DeviceUID: "empty-pi",
	}, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Sensors == nil || cfg.Actuators == nil {
		t.Fatalf("expected empty slices, got sensors=%v actuators=%v", cfg.Sensors, cfg.Actuators)
	}
}
