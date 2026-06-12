package farmguardian

import (
	"context"
	"fmt"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/fertigation/programfit"
)

// ProgramFitHintLine returns Guardian-facing text when the linked feeding program
// metadata mismatches the active grow crop_key or stage (Phase 96).
func ProgramFitHintLine(ctx context.Context, q db.Querier, cycle db.Gr33nfertigationCropCycle) string {
	if q == nil || cycle.PrimaryProgramID == nil || *cycle.PrimaryProgramID <= 0 {
		return ""
	}
	stage := ""
	if cycle.CurrentStage != nil {
		stage = string(*cycle.CurrentStage)
	}
	cropKey := ""
	if cycle.PlantID != nil && *cycle.PlantID > 0 {
		if p, err := q.GetPlant(ctx, *cycle.PlantID); err == nil && p.CropKey != nil {
			cropKey = strings.TrimSpace(*p.CropKey)
		}
	}
	warnings, err := programfit.ValidateProgramForGrow(ctx, q, *cycle.PrimaryProgramID, cropKey, stage)
	if err != nil || len(warnings) == 0 {
		return ""
	}
	msg := warnings[0]
	if stage != "" {
		msg = fmt.Sprintf("Your grow is in %s but %s", stage, msg)
	}
	return msg + ". EC on the zone strip comes from the crop profile; the pump recipe may differ — see Water tab or switch program."
}
