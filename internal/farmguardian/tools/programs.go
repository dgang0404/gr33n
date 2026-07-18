package tools

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

func execCreateFertigationProgram(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	name, err := stringFromArgs(args, "name")
	if err != nil {
		return nil, err
	}
	targetZoneID, err := int64FromArgs(args, "target_zone_id")
	if err != nil {
		return nil, err
	}
	totalVolF, err := float64FromArgs(args, "total_volume_liters")
	if err != nil {
		return nil, err
	}
	ecLowF, err := float64FromArgs(args, "ec_trigger_low")
	if err != nil {
		return nil, err
	}
	phLowF, err := float64FromArgs(args, "ph_trigger_low")
	if err != nil {
		return nil, err
	}
	phHighF, err := float64FromArgs(args, "ph_trigger_high")
	if err != nil {
		return nil, err
	}

	z, err := deps.Q.GetZoneByID(ctx, targetZoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("zone %d not found", targetZoneID)
		}
		return nil, err
	}
	if err := ensureFarmScope(z.FarmID, deps.FarmID); err != nil {
		return nil, err
	}

	totalVol, err := httputil.NumericFromFloat64(totalVolF)
	if err != nil {
		return nil, fmt.Errorf("invalid total_volume_liters")
	}
	ecLow, err := httputil.NumericFromFloat64(ecLowF)
	if err != nil {
		return nil, fmt.Errorf("invalid ec_trigger_low")
	}
	phLow, err := httputil.NumericFromFloat64(phLowF)
	if err != nil {
		return nil, fmt.Errorf("invalid ph_trigger_low")
	}
	phHigh, err := httputil.NumericFromFloat64(phHighF)
	if err != nil {
		return nil, fmt.Errorf("invalid ph_trigger_high")
	}

	isActive := true
	if v, err := optionalBoolFromArgs(args, "is_active"); err != nil {
		return nil, err
	} else if v != nil {
		isActive = *v
	}
	desc, err := optionalStringFromArgs(args, "description")
	if err != nil {
		return nil, err
	}

	zoneID := targetZoneID
	row, err := deps.Q.CreateProgram(ctx, db.CreateProgramParams{
		FarmID:            deps.FarmID,
		Name:              name,
		Description:       desc,
		TargetZoneID:      &zoneID,
		TotalVolumeLiters: totalVol,
		EcTriggerLow:      ecLow,
		PhTriggerLow:      phLow,
		PhTriggerHigh:     phHigh,
		IsActive:          isActive,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"program_id":          row.ID,
		"name":                row.Name,
		"target_zone_id":      targetZoneID,
		"total_volume_liters": totalVolF,
		"is_active":           row.IsActive,
	}, nil
}

func execPatchFertigationProgram(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	programID, err := int64FromArgs(args, "program_id")
	if err != nil {
		return nil, err
	}
	if len(args) <= 1 {
		return nil, errors.New("at least one patch field required")
	}
	prog, err := deps.Q.GetFertigationProgramByID(ctx, programID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("program %d not found", programID)
		}
		return nil, err
	}
	if err := ensureFarmScope(prog.FarmID, deps.FarmID); err != nil {
		return nil, err
	}

	ecTarget := prog.EcTargetID
	if v, err := optionalInt64FromArgs(args, "ec_target_id"); err != nil {
		return nil, err
	} else if v != nil {
		ecTarget = v
	}
	isActive := prog.IsActive
	if v, err := optionalBoolFromArgs(args, "is_active"); err != nil {
		return nil, err
	} else if v != nil {
		isActive = *v
	}
	irrigationOnly := prog.IrrigationOnly
	if v, err := optionalBoolFromArgs(args, "irrigation_only"); err != nil {
		return nil, err
	} else if v != nil {
		irrigationOnly = *v
	}
	recipeID := prog.ApplicationRecipeID
	if irrigationOnly {
		recipeID = nil
	}
	totalVol := prog.TotalVolumeLiters
	if v, err := optionalFloat64FromArgs(args, "total_volume_liters"); err != nil {
		return nil, err
	} else if v != nil {
		var n pgtype.Numeric
		if err := n.Scan(strconv.FormatFloat(*v, 'f', -1, 64)); err != nil {
			return nil, fmt.Errorf("invalid total_volume_liters")
		}
		totalVol = n
	}
	row, err := deps.Q.UpdateProgram(ctx, db.UpdateProgramParams{
		ID:                  programID,
		Name:                prog.Name,
		Description:         prog.Description,
		ReservoirID:         prog.ReservoirID,
		TargetZoneID:        prog.TargetZoneID,
		EcTargetID:          ecTarget,
		TotalVolumeLiters:   totalVol,
		IsActive:            isActive,
		IrrigationOnly:      irrigationOnly,
		ApplicationRecipeID: recipeID,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"program_id": row.ID,
		"is_active":  row.IsActive,
	}, nil
}
