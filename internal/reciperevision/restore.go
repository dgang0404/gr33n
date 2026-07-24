package reciperevision

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

// RevisionSummary extracts operator-facing fields from a revision row.
func RevisionSummary(rev db.Gr33nnaturalfarmingApplicationRecipeRevision) (dilution string, componentCount int, err error) {
	var snap Snapshot
	if uerr := json.Unmarshal(rev.Snapshot, &snap); uerr != nil {
		return "", 0, uerr
	}
	if snap.Recipe.DilutionRatio != nil {
		dilution = *snap.Recipe.DilutionRatio
	}
	return dilution, len(snap.Components), nil
}

// RestoreFromRevision copies a revision snapshot onto the live recipe and appends a new revision row.
func RestoreFromRevision(ctx context.Context, q db.Querier, recipeID, revisionID int64) (db.Gr33nnaturalfarmingApplicationRecipeRevision, error) {
	rev, err := q.GetRecipeRevisionByID(ctx, revisionID)
	if err != nil {
		return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, err
	}
	if rev.ApplicationRecipeID != recipeID {
		return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, pgx.ErrNoRows
	}

	var snap Snapshot
	if err := json.Unmarshal(rev.Snapshot, &snap); err != nil {
		return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, fmt.Errorf("invalid revision snapshot: %w", err)
	}

	rec, err := q.GetRecipeByID(ctx, recipeID)
	if err != nil {
		return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, err
	}
	_ = rec // live row loaded for existence; caller enforces farm auth

	targetType := db.Gr33nnaturalfarmingApplicationTargetEnum(snap.Recipe.TargetApplicationType)
	if snap.Recipe.TargetApplicationType == "" {
		targetType = rec.TargetApplicationType
	}
	name := snap.Recipe.Name
	if name == "" {
		name = rec.Name
	}

	if _, err := q.UpdateRecipe(ctx, db.UpdateRecipeParams{
		ID:                    recipeID,
		Name:                  name,
		InputDefinitionID:     snap.Recipe.InputDefinitionID,
		Description:           snap.Recipe.Description,
		TargetApplicationType: targetType,
		DilutionRatio:         snap.Recipe.DilutionRatio,
		Instructions:          snap.Recipe.Instructions,
		FrequencyGuidelines:   snap.Recipe.FrequencyGuidelines,
		Notes:                 snap.Recipe.Notes,
	}); err != nil {
		return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, err
	}

	current, err := q.ListRecipeComponents(ctx, recipeID)
	if err != nil {
		return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, err
	}
	want := make(map[int64]ComponentSnapshot, len(snap.Components))
	for _, c := range snap.Components {
		want[c.InputDefinitionID] = c
	}
	for _, c := range current {
		if _, ok := want[c.InputDefinitionID]; !ok {
			if err := q.RemoveRecipeComponent(ctx, db.RemoveRecipeComponentParams{
				ApplicationRecipeID: recipeID,
				InputDefinitionID:   c.InputDefinitionID,
			}); err != nil {
				return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, err
			}
		}
	}
	for _, c := range snap.Components {
		pv, err := httputil.NumericFromFloat64(c.PartValue)
		if err != nil {
			return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, fmt.Errorf("component %d: %w", c.InputDefinitionID, err)
		}
		if err := q.AddRecipeComponent(ctx, db.AddRecipeComponentParams{
			ApplicationRecipeID: recipeID,
			InputDefinitionID:   c.InputDefinitionID,
			PartValue:           pv,
			PartUnitID:          c.PartUnitID,
			Notes:               c.Notes,
		}); err != nil {
			return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, err
		}
	}

	summary := fmt.Sprintf("restored from revision %d", rev.RevisionNumber)
	return Record(ctx, q, recipeID, summary)
}
