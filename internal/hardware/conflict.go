package hardware

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConflictError is returned when two entities share the same pin/channel on one device.
type ConflictError struct {
	EntityType string
	EntityID   int64
	EntityName string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s %d (%s) already uses this pin/channel on the device", e.EntityType, e.EntityID, e.EntityName)
}

// CheckSensorWiringConflict returns a ConflictError when another sensor on the farm
// uses the same device_id + gpio_pin or device_id + i2c_channel.
func CheckSensorWiringConflict(ctx context.Context, pool *pgxpool.Pool, farmID, sensorID int64, w *Wiring) error {
	if w == nil || w.DeviceID == nil {
		return nil
	}
	if w.GPIOPin != nil {
		var id int64
		var name string
		err := pool.QueryRow(ctx, `
			SELECT id, name FROM gr33ncore.sensors
			WHERE farm_id = $1 AND deleted_at IS NULL AND id != $2
			  AND (config->'wiring'->>'device_id')::bigint = $3
			  AND (config->'wiring'->>'gpio_pin')::int = $4
			LIMIT 1`, farmID, sensorID, *w.DeviceID, *w.GPIOPin).Scan(&id, &name)
		if err == nil {
			return &ConflictError{EntityType: "sensor", EntityID: id, EntityName: name}
		}
	}
	if w.I2CChannel != nil {
		var id int64
		var name string
		err := pool.QueryRow(ctx, `
			SELECT id, name FROM gr33ncore.sensors
			WHERE farm_id = $1 AND deleted_at IS NULL AND id != $2
			  AND (config->'wiring'->>'device_id')::bigint = $3
			  AND (config->'wiring'->>'i2c_channel')::int = $4
			LIMIT 1`, farmID, sensorID, *w.DeviceID, *w.I2CChannel).Scan(&id, &name)
		if err == nil {
			return &ConflictError{EntityType: "sensor", EntityID: id, EntityName: name}
		}
	}
	return nil
}

// CheckActuatorWiringConflict returns a ConflictError when another actuator shares gpio_pin on device.
func CheckActuatorWiringConflict(ctx context.Context, pool *pgxpool.Pool, farmID, actuatorID int64, w *Wiring) error {
	if w == nil || w.DeviceID == nil || w.GPIOPin == nil {
		return nil
	}
	var id int64
	var name string
	err := pool.QueryRow(ctx, `
		SELECT id, name FROM gr33ncore.actuators
		WHERE farm_id = $1 AND deleted_at IS NULL AND id != $2
		  AND (config->'wiring'->>'device_id')::bigint = $3
		  AND (config->'wiring'->>'gpio_pin')::int = $4
		LIMIT 1`, farmID, actuatorID, *w.DeviceID, *w.GPIOPin).Scan(&id, &name)
	if err == nil {
		return &ConflictError{EntityType: "actuator", EntityID: id, EntityName: name}
	}
	return nil
}
