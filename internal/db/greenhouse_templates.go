package db

import (
	"context"
	"encoding/json"
)

// ApplyGreenhouseRuleTemplates runs gr33ncore.apply_greenhouse_rule_templates
// for Phase 36 WS3. All actuator/sensor IDs are optional (pass nil to skip
// the corresponding rule family).
func (q *Queries) ApplyGreenhouseRuleTemplates(
	ctx context.Context,
	farmID, zoneID int64,
	shadeActuatorID, fanActuatorID *int64,
	luxSensorID, tempSensorID *int64,
) (map[string]any, error) {
	row := q.db.QueryRow(ctx,
		`SELECT gr33ncore.apply_greenhouse_rule_templates($1, $2, $3, $4, $5, $6)`,
		farmID, zoneID,
		shadeActuatorID, fanActuatorID,
		luxSensorID, tempSensorID,
	)
	var raw []byte
	if err := row.Scan(&raw); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
