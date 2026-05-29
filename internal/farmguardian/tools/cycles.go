package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	db "gr33n-api/internal/db"
)

func execCreateCropCycle(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	zoneID, err := int64FromArgs(args, "zone_id")
	if err != nil {
		return nil, err
	}
	name, err := stringFromArgs(args, "name")
	if err != nil {
		return nil, err
	}
	strain, err := stringFromArgs(args, "strain_or_variety")
	if err != nil {
		return nil, err
	}
	stage, err := stringFromArgs(args, "current_stage")
	if err != nil {
		return nil, err
	}
	startedAt, err := dateFromArgs(args, "started_at")
	if err != nil {
		return nil, err
	}
	notes, err := optionalStringFromArgs(args, "cycle_notes")
	if err != nil {
		return nil, err
	}

	z, err := deps.Q.GetZoneByID(ctx, zoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("zone %d not found", zoneID)
		}
		return nil, err
	}
	if err := ensureFarmScope(z.FarmID, deps.FarmID); err != nil {
		return nil, err
	}

	active := true
	if v, err := optionalBoolFromArgs(args, "is_active"); err != nil {
		return nil, err
	} else if v != nil {
		active = *v
	}
	if active {
		if err := ensureZoneHasNoActiveCycle(ctx, deps.Q, deps.FarmID, zoneID); err != nil {
			return nil, err
		}
	}

	strainPtr := &strain
	row, err := deps.Q.CreateCropCycle(ctx, db.CreateCropCycleParams{
		FarmID:          deps.FarmID,
		ZoneID:          zoneID,
		Name:            name,
		StrainOrVariety: strainPtr,
		CurrentStage:    parseGrowthStage(stage),
		IsActive:        active,
		StartedAt:       startedAt,
		CycleNotes:      notes,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("only one active crop cycle per zone is allowed")
		}
		return nil, err
	}
	return map[string]any{
		"crop_cycle_id":     row.ID,
		"name":              row.Name,
		"zone_id":           row.ZoneID,
		"strain_or_variety": strain,
		"current_stage":     strings.TrimSpace(stage),
	}, nil
}

func ensureZoneHasNoActiveCycle(ctx context.Context, q db.Querier, farmID, zoneID int64) error {
	cycles, err := q.ListCropCyclesByFarm(ctx, farmID)
	if err != nil {
		return err
	}
	for _, c := range cycles {
		if c.IsActive && c.ZoneID == zoneID {
			return fmt.Errorf("zone %d already has active crop cycle %q (#%d)", zoneID, c.Name, c.ID)
		}
	}
	return nil
}

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
