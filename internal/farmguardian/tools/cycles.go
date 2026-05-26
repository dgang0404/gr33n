package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

func execUpdateCycleStage(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	cycleID, err := int64FromArgs(args, "crop_cycle_id")
	if err != nil {
		cycleID, err = int64FromArgs(args, "cycle_id")
		if err != nil {
			return nil, errors.New("crop_cycle_id required")
		}
	}
	stage, err := stringFromArgs(args, "current_stage")
	if err != nil {
		return nil, err
	}
	cc, err := deps.Q.GetCropCycleByID(ctx, cycleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("crop cycle %d not found", cycleID)
		}
		return nil, err
	}
	if err := ensureFarmScope(cc.FarmID, deps.FarmID); err != nil {
		return nil, err
	}
	row, err := deps.Q.UpdateCropCycleStage(ctx, db.UpdateCropCycleStageParams{
		ID:           cycleID,
		CurrentStage: parseGrowthStage(stage),
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"crop_cycle_id":  row.ID,
		"current_stage":  strings.TrimSpace(stage),
		"cycle_name":     row.Name,
	}, nil
}

func parseGrowthStage(s string) db.NullGr33nfertigationGrowthStageEnum {
	s = strings.TrimSpace(s)
	if s == "" {
		return db.NullGr33nfertigationGrowthStageEnum{
			Gr33nfertigationGrowthStageEnum: db.Gr33nfertigationGrowthStageEnumSeedling,
			Valid:                           true,
		}
	}
	return db.NullGr33nfertigationGrowthStageEnum{
		Gr33nfertigationGrowthStageEnum: db.Gr33nfertigationGrowthStageEnum(s),
		Valid:                           true,
	}
}
