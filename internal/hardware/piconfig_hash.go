package hardware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

type wiringHashPayload struct {
	Sensors     []PiRuntimeSensor   `json:"sensors"`
	Actuators   []PiRuntimeActuator `json:"actuators"`
	MixChannels []int64             `json:"mix_channels,omitempty"`
}

// PiRuntimeConfigWiringSHA256 hashes sensors/actuators wiring (ignores api keys and URLs).
func PiRuntimeConfigWiringSHA256(cfg *PiRuntimeConfig) string {
	if cfg == nil {
		return ""
	}
	sensors := append([]PiRuntimeSensor(nil), cfg.Sensors...)
	actuators := append([]PiRuntimeActuator(nil), cfg.Actuators...)
	sort.Slice(sensors, func(i, j int) bool { return sensors[i].SensorID < sensors[j].SensorID })
	sort.Slice(actuators, func(i, j int) bool { return actuators[i].ActuatorID < actuators[j].ActuatorID })

	payload := wiringHashPayload{
		Sensors:   sensors,
		Actuators: actuators,
	}
	if len(cfg.MixChannels) > 0 {
		payload.MixChannels = append([]int64(nil), cfg.MixChannels...)
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
