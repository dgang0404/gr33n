package hardware

import (
	"encoding/json"
	"fmt"
)

// PiRuntimeConfig is the JSON shape returned by GET /devices/by-uid/{uid}/config (Phase 51).
type PiRuntimeConfig struct {
	DeviceUID                   string              `json:"device_uid"`
	DeviceID                    int64               `json:"device_id"`
	FarmID                      int64               `json:"farm_id"`
	ConfigVersion               int32               `json:"config_version"`
	Sensors                     []PiRuntimeSensor   `json:"sensors"`
	Actuators                   []PiRuntimeActuator `json:"actuators"`
	MixChannels                 []int64             `json:"mix_channels,omitempty"`
	SchedulePollIntervalSeconds int                 `json:"schedule_poll_interval_seconds"`
	OfflineQueuePath            string              `json:"offline_queue_path"`
	OfflineFlushIntervalSeconds int                 `json:"offline_flush_interval_seconds"`
}

// PiRuntimeSensor matches pi_client sensor dict keys (pin/channel, not gpio_pin).
type PiRuntimeSensor struct {
	SensorID           int64          `json:"sensor_id"`
	SensorType         string         `json:"sensor_type"`
	Source             string         `json:"source"`
	Pin                *int           `json:"pin,omitempty"`
	Channel            *int           `json:"channel,omitempty"`
	Port               string         `json:"port,omitempty"`
	Inputs             map[string]int `json:"inputs,omitempty"`
	InputMaxAgeSeconds *int           `json:"input_max_age_seconds,omitempty"`
	IntervalSeconds    int            `json:"interval_seconds"`
}

// PiRuntimeActuator matches pi_client actuator dict keys.
type PiRuntimeActuator struct {
	ActuatorID int64  `json:"actuator_id"`
	DeviceID   int64  `json:"device_id"`
	DeviceType string `json:"device_type"`
	GPIOPin    int    `json:"gpio_pin"`
}

// BuildPiRuntimeConfig assembles the Pi pull-sync payload from DB wiring rows.
func BuildPiRuntimeConfig(opts PiConfigOptions, configVersion int32, deviceConfig json.RawMessage) (*PiRuntimeConfig, error) {
	if opts.FarmID <= 0 || opts.DeviceID <= 0 {
		return nil, fmt.Errorf("farm_id and device_id are required")
	}
	sensors, err := buildPiSensorEntries(opts.Sensors, opts.DeviceID)
	if err != nil {
		return nil, err
	}
	actuators, err := buildPiActuatorEntries(opts.Actuators, opts.DeviceID)
	if err != nil {
		return nil, err
	}
	out := &PiRuntimeConfig{
		DeviceUID:                   opts.DeviceUID,
		DeviceID:                    opts.DeviceID,
		FarmID:                      opts.FarmID,
		ConfigVersion:               configVersion,
		Sensors:                     toRuntimeSensors(sensors),
		Actuators:                   toRuntimeActuators(actuators),
		MixChannels:                 extractMixChannels(deviceConfig),
		SchedulePollIntervalSeconds: 30,
		OfflineQueuePath:            "/var/lib/gr33n/queue.db",
		OfflineFlushIntervalSeconds: 60,
	}
	return out, nil
}

func toRuntimeSensors(entries []piSensorEntry) []PiRuntimeSensor {
	if len(entries) == 0 {
		return []PiRuntimeSensor{}
	}
	out := make([]PiRuntimeSensor, len(entries))
	for i, e := range entries {
		out[i] = PiRuntimeSensor{
			SensorID:           e.SensorID,
			SensorType:         e.SensorType,
			Source:             e.Source,
			Pin:                e.Pin,
			Channel:            e.Channel,
			Port:               e.Port,
			Inputs:             e.Inputs,
			InputMaxAgeSeconds: e.InputMaxAgeSeconds,
			IntervalSeconds:    e.IntervalSeconds,
		}
	}
	return out
}

func toRuntimeActuators(entries []piActuatorEntry) []PiRuntimeActuator {
	if len(entries) == 0 {
		return []PiRuntimeActuator{}
	}
	out := make([]PiRuntimeActuator, len(entries))
	for i, e := range entries {
		out[i] = PiRuntimeActuator{
			ActuatorID: e.ActuatorID,
			DeviceID:   e.DeviceID,
			DeviceType: e.DeviceType,
			GPIOPin:    e.GPIOPin,
		}
	}
	return out
}

func extractMixChannels(deviceConfig json.RawMessage) []int64 {
	if len(deviceConfig) == 0 {
		return nil
	}
	var root struct {
		MixChannels []int64 `json:"mix_channels"`
	}
	if err := json.Unmarshal(deviceConfig, &root); err != nil {
		return nil
	}
	return root.MixChannels
}
