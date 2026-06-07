package hardware

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
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

// CheckSensorWiringConflict returns a ConflictError when another sensor or actuator on the farm
// uses the same device_id + gpio_pin or device_id + i2c_channel.
func CheckSensorWiringConflict(ctx context.Context, pool *pgxpool.Pool, farmID, sensorID int64, w *Wiring) error {
	if w == nil || w.DeviceID == nil {
		return nil
	}
	if w.GPIOPin != nil {
		if err := findGPIOConflict(ctx, pool, farmID, sensorID, 0, *w.DeviceID, *w.GPIOPin, w.Source); err != nil {
			return err
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
	if w.Source == "derived" {
		if err := ValidateDerivedInputs(ctx, pool, farmID, w); err != nil {
			return err
		}
	}
	return nil
}

// CheckActuatorWiringConflict returns a ConflictError when another actuator or sensor shares gpio_pin on device.
func CheckActuatorWiringConflict(ctx context.Context, pool *pgxpool.Pool, farmID, actuatorID int64, w *Wiring) error {
	if w == nil || w.DeviceID == nil || w.GPIOPin == nil {
		return nil
	}
	return findGPIOConflict(ctx, pool, farmID, 0, actuatorID, *w.DeviceID, *w.GPIOPin, w.Source)
}

func sharedDHT22GPIOAllowed(a, b string) bool {
	return a == "dht22" && b == "dht22"
}

func findGPIOConflict(ctx context.Context, pool *pgxpool.Pool, farmID, excludeSensorID, excludeActuatorID, deviceID int64, gpioPin int, newSource string) error {
	if excludeSensorID > 0 {
		var id int64
		var name string
		var otherSource string
		err := pool.QueryRow(ctx, `
			SELECT id, name, COALESCE(config->'wiring'->>'source', '') FROM gr33ncore.sensors
			WHERE farm_id = $1 AND deleted_at IS NULL AND id != $2
			  AND (config->'wiring'->>'device_id')::bigint = $3
			  AND (config->'wiring'->>'gpio_pin')::int = $4
			LIMIT 1`, farmID, excludeSensorID, deviceID, gpioPin).Scan(&id, &name, &otherSource)
		if err == nil && !sharedDHT22GPIOAllowed(newSource, otherSource) {
			return &ConflictError{EntityType: "sensor", EntityID: id, EntityName: name}
		}
	}
	if excludeActuatorID > 0 {
		var id int64
		var name string
		err := pool.QueryRow(ctx, `
			SELECT id, name FROM gr33ncore.actuators
			WHERE farm_id = $1 AND deleted_at IS NULL AND id != $2
			  AND (config->'wiring'->>'device_id')::bigint = $3
			  AND (config->'wiring'->>'gpio_pin')::int = $4
			LIMIT 1`, farmID, excludeActuatorID, deviceID, gpioPin).Scan(&id, &name)
		if err == nil {
			return &ConflictError{EntityType: "actuator", EntityID: id, EntityName: name}
		}
	}
	// Cross-type: sensor checking actuators, actuator checking sensors.
	if excludeSensorID > 0 {
		var id int64
		var name string
		err := pool.QueryRow(ctx, `
			SELECT id, name FROM gr33ncore.actuators
			WHERE farm_id = $1 AND deleted_at IS NULL
			  AND (config->'wiring'->>'device_id')::bigint = $2
			  AND (config->'wiring'->>'gpio_pin')::int = $3
			LIMIT 1`, farmID, deviceID, gpioPin).Scan(&id, &name)
		if err == nil {
			return &ConflictError{EntityType: "actuator", EntityID: id, EntityName: name}
		}
	}
	if excludeActuatorID > 0 {
		var id int64
		var name string
		var otherSource string
		err := pool.QueryRow(ctx, `
			SELECT id, name, COALESCE(config->'wiring'->>'source', '') FROM gr33ncore.sensors
			WHERE farm_id = $1 AND deleted_at IS NULL
			  AND (config->'wiring'->>'device_id')::bigint = $2
			  AND (config->'wiring'->>'gpio_pin')::int = $3
			LIMIT 1`, farmID, deviceID, gpioPin).Scan(&id, &name, &otherSource)
		if err == nil && !sharedDHT22GPIOAllowed(newSource, otherSource) {
			return &ConflictError{EntityType: "sensor", EntityID: id, EntityName: name}
		}
	}
	return nil
}

// DerivedInputError reports a missing derived-sensor input reference.
type DerivedInputError struct {
	InputKey string
	SensorID int64
}

func (e *DerivedInputError) Error() string {
	return fmt.Sprintf("derived input %q references missing sensor %d", e.InputKey, e.SensorID)
}

// ValidateDerivedInputs ensures derived wiring inputs point at existing farm sensors.
func ValidateDerivedInputs(ctx context.Context, pool *pgxpool.Pool, farmID int64, w *Wiring) error {
	if w == nil || w.Source != "derived" || len(w.Inputs) == 0 {
		return nil
	}
	var inputs map[string]int
	if err := json.Unmarshal(w.Inputs, &inputs); err != nil {
		return fmt.Errorf("wiring.inputs: invalid JSON")
	}
	for key, sensorID := range inputs {
		var one int
		err := pool.QueryRow(ctx, `
			SELECT 1 FROM gr33ncore.sensors
			WHERE farm_id = $1 AND id = $2 AND deleted_at IS NULL
			LIMIT 1`, farmID, int64(sensorID)).Scan(&one)
		if err != nil {
			if err == pgx.ErrNoRows {
				return &DerivedInputError{InputKey: key, SensorID: int64(sensorID)}
			}
			return err
		}
	}
	return nil
}
