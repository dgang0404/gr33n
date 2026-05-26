package tools

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

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
		ID:                programID,
		Name:              prog.Name,
		Description:       prog.Description,
		ReservoirID:       prog.ReservoirID,
		TargetZoneID:      prog.TargetZoneID,
		EcTargetID:        ecTarget,
		TotalVolumeLiters: totalVol,
		IsActive:          isActive,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"program_id": row.ID,
		"is_active":  row.IsActive,
	}, nil
}
