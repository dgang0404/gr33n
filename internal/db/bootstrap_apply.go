package db

import (
	"context"
	"encoding/json"
)

// ApplyFarmBootstrapTemplate runs gr33ncore.apply_farm_bootstrap_template (Phase 30 WS3).
func (q *Queries) ApplyFarmBootstrapTemplate(ctx context.Context, farmID int64, template string) (map[string]any, error) {
	row := q.db.QueryRow(ctx,
		`SELECT gr33ncore.apply_farm_bootstrap_template($1, $2)`,
		farmID, template,
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
