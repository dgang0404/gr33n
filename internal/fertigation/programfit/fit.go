// Package programfit validates primary_program_id against grow context (Phases 96/102).
package programfit

import (
	"context"
	"os"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/fertigation/programmeta"
)

// ValidateProgramForGrow returns fit warnings when program metadata mismatches crop/stage.
func ValidateProgramForGrow(ctx context.Context, q db.Querier, programID int64, cropKey, stage string) ([]string, error) {
	if programID <= 0 {
		return nil, nil
	}
	prog, err := q.GetFertigationProgramByID(ctx, programID)
	if err != nil {
		return nil, err
	}
	meta := programmeta.Parse(prog.Metadata)
	fit := meta.CheckFit(cropKey, stage)
	return fit.Warnings, nil
}

// StrictMode is true when attach-time mismatches should block (422).
func StrictMode() bool {
	return strings.TrimSpace(os.Getenv("STRICT_PROGRAM_STAGE_MATCH")) == "1"
}
