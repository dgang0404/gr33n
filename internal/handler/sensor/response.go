package sensor

import (
	db "gr33n-api/internal/db"
	"gr33n-api/internal/hardware"
)

// sensorResponse is the list/get shape with top-level wiring extracted from config.
type sensorResponse struct {
	db.Gr33ncoreSensor
	Wiring *hardware.Wiring `json:"wiring"`
}

func wrapSensor(row db.Gr33ncoreSensor) sensorResponse {
	w, _ := hardware.ExtractWiring(row.Config)
	return sensorResponse{Gr33ncoreSensor: row, Wiring: w}
}
