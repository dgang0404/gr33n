// Phase 39 WS1 — hand-written (mirrors sqlc style).
// Run `make sqlc` after adding queries to db/queries/device_commands.sql to regenerate.

package db

import (
	"context"
	"encoding/json"
)

// ---- params ----------------------------------------------------------------

type EnqueueDeviceCommandParams struct {
	DeviceID    int64           `db:"device_id" json:"device_id"`
	FarmID      int64           `db:"farm_id" json:"farm_id"`
	CommandType string          `db:"command_type" json:"command_type"`
	Payload     json.RawMessage `db:"payload" json:"payload"`
	Source      string          `db:"source" json:"source"`
	// nullable provenance
	ActuatorID *int64 `db:"actuator_id" json:"actuator_id"`
	ScheduleID *int64 `db:"schedule_id" json:"schedule_id"`
	RuleID     *int64 `db:"rule_id" json:"rule_id"`
	ProgramID  *int64 `db:"program_id" json:"program_id"`
}

type AckDeviceCommandParams struct {
	ID     int64           `db:"id" json:"id"`
	Status string          `db:"status" json:"status"`
	Result json.RawMessage `db:"result" json:"result"`
}

// ---- helpers ---------------------------------------------------------------

const deviceCommandColumns = `id, device_id, farm_id, command_type, payload, status, source,
	actuator_id, schedule_id, rule_id, program_id,
	created_at, started_at, completed_at, result`

func scanDeviceCommand(row interface{ Scan(...any) error }) (Gr33ncoreDeviceCommand, error) {
	var c Gr33ncoreDeviceCommand
	err := row.Scan(
		&c.ID, &c.DeviceID, &c.FarmID, &c.CommandType, &c.Payload, &c.Status, &c.Source,
		&c.ActuatorID, &c.ScheduleID, &c.RuleID, &c.ProgramID,
		&c.CreatedAt, &c.StartedAt, &c.CompletedAt, &c.Result,
	)
	return c, err
}

// ---- queries ----------------------------------------------------------------

const enqueueDeviceCommand = `
INSERT INTO gr33ncore.device_commands (
    device_id, farm_id, command_type, payload,
    source, actuator_id, schedule_id, rule_id, program_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING ` + deviceCommandColumns

func (q *Queries) EnqueueDeviceCommand(ctx context.Context, arg EnqueueDeviceCommandParams) (Gr33ncoreDeviceCommand, error) {
	row := q.db.QueryRow(ctx, enqueueDeviceCommand,
		arg.DeviceID, arg.FarmID, arg.CommandType, arg.Payload,
		arg.Source, arg.ActuatorID, arg.ScheduleID, arg.RuleID, arg.ProgramID,
	)
	return scanDeviceCommand(row)
}

const getNextDeviceCommand = `
UPDATE gr33ncore.device_commands
SET status = 'in_progress', started_at = NOW()
WHERE id = (
    SELECT id FROM gr33ncore.device_commands
    WHERE device_id = $1 AND status = 'pending'
    ORDER BY created_at ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
RETURNING ` + deviceCommandColumns

func (q *Queries) GetNextDeviceCommand(ctx context.Context, deviceID int64) (Gr33ncoreDeviceCommand, error) {
	row := q.db.QueryRow(ctx, getNextDeviceCommand, deviceID)
	return scanDeviceCommand(row)
}

const ackDeviceCommand = `
UPDATE gr33ncore.device_commands
SET status = $2, completed_at = NOW(), result = $3
WHERE id = $1
RETURNING ` + deviceCommandColumns

func (q *Queries) AckDeviceCommand(ctx context.Context, arg AckDeviceCommandParams) (Gr33ncoreDeviceCommand, error) {
	row := q.db.QueryRow(ctx, ackDeviceCommand, arg.ID, arg.Status, arg.Result)
	return scanDeviceCommand(row)
}

const listDeviceCommands = `
SELECT ` + deviceCommandColumns + `
FROM gr33ncore.device_commands
WHERE device_id = $1
  AND ($2::text IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT 100`

func (q *Queries) ListDeviceCommands(ctx context.Context, deviceID int64, status *string) ([]Gr33ncoreDeviceCommand, error) {
	rows, err := q.db.Query(ctx, listDeviceCommands, deviceID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Gr33ncoreDeviceCommand
	for rows.Next() {
		c, err := scanDeviceCommand(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

const cancelDeviceCommand = `
UPDATE gr33ncore.device_commands
SET status = 'cancelled', completed_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING ` + deviceCommandColumns

func (q *Queries) CancelDeviceCommand(ctx context.Context, id int64) (Gr33ncoreDeviceCommand, error) {
	row := q.db.QueryRow(ctx, cancelDeviceCommand, id)
	return scanDeviceCommand(row)
}

const countPendingCommandsByDevice = `
SELECT COUNT(*)::bigint AS cnt
FROM gr33ncore.device_commands
WHERE device_id = $1 AND status IN ('pending', 'in_progress')`

func (q *Queries) CountPendingCommandsByDevice(ctx context.Context, deviceID int64) (int64, error) {
	var cnt int64
	err := q.db.QueryRow(ctx, countPendingCommandsByDevice, deviceID).Scan(&cnt)
	return cnt, err
}

const countPendingCommandsByFarm = `
SELECT device_id, COUNT(*)::bigint AS cnt
FROM gr33ncore.device_commands
WHERE farm_id = $1 AND status IN ('pending', 'in_progress')
GROUP BY device_id
ORDER BY device_id`

func (q *Queries) CountPendingCommandsByFarm(ctx context.Context, farmID int64) ([]CountPendingCommandsByFarmRow, error) {
	rows, err := q.db.Query(ctx, countPendingCommandsByFarm, farmID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CountPendingCommandsByFarmRow
	for rows.Next() {
		var r CountPendingCommandsByFarmRow
		if err := rows.Scan(&r.DeviceID, &r.Cnt); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, rows.Err()
}
