package tools

import (
	"context"
	"errors"

	db "gr33n-api/internal/db"
)

func execCreatePlant(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	displayName, err := stringFromArgs(args, "display_name")
	if err != nil {
		return nil, err
	}
	variety, err := optionalStringFromArgs(args, "variety_or_cultivar")
	if err != nil {
		return nil, err
	}
	meta, err := optionalMetaJSONFromArgs(args, "meta")
	if err != nil {
		return nil, err
	}
	row, err := deps.Q.CreatePlant(ctx, db.CreatePlantParams{
		FarmID:            deps.FarmID,
		DisplayName:       displayName,
		VarietyOrCultivar: variety,
		Meta:              meta,
	})
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"plant_id":     row.ID,
		"display_name": row.DisplayName,
	}
	if row.VarietyOrCultivar != nil {
		out["variety_or_cultivar"] = *row.VarietyOrCultivar
	}
	return out, nil
}
