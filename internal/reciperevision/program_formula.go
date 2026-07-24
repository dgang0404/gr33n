package reciperevision

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

// ProgramFormulaContext is stamped onto automation_runs.details and mixing_events.metadata.
type ProgramFormulaContext struct {
	ApplicationRecipeID         int64
	ApplicationRecipeRevisionID *int64
	FormulaSnapshot             map[string]any
	RevisionUnpinned            bool
}

// FormulaSummaryFromSnapshot builds the compact run-time snapshot from revision JSON.
func FormulaSummaryFromSnapshot(raw json.RawMessage) (map[string]any, error) {
	var snap Snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return nil, err
	}
	components := make([]map[string]any, 0, len(snap.Components))
	for _, c := range snap.Components {
		row := map[string]any{
			"input_definition_id": c.InputDefinitionID,
			"input_name":          c.InputName,
			"part_value":          c.PartValue,
		}
		if c.PartUnitID != nil {
			row["part_unit_id"] = *c.PartUnitID
		}
		if c.Notes != nil {
			row["notes"] = *c.Notes
		}
		components = append(components, row)
	}
	out := map[string]any{
		"recipe_name": snap.Recipe.Name,
		"components":  components,
	}
	if snap.Recipe.DilutionRatio != nil {
		out["dilution_ratio"] = *snap.Recipe.DilutionRatio
	}
	return out, nil
}

// ResolveProgramFormula loads the pinned revision or latest for the program's recipe.
func ResolveProgramFormula(ctx context.Context, q db.Querier, p db.Gr33nfertigationProgram) (ProgramFormulaContext, error) {
	if p.IrrigationOnly || p.ApplicationRecipeID == nil {
		return ProgramFormulaContext{}, nil
	}
	recipeID := *p.ApplicationRecipeID
	var rev db.Gr33nnaturalfarmingApplicationRecipeRevision
	var err error
	unpinned := false
	if p.ApplicationRecipeRevisionID != nil {
		rev, err = q.GetRecipeRevisionByID(ctx, *p.ApplicationRecipeRevisionID)
	} else {
		unpinned = true
		rev, err = q.GetLatestRecipeRevision(ctx, recipeID)
		if err == nil {
			log.Printf("program %d (%s): application_recipe_revision_id unset — using latest revision %d", p.ID, p.Name, rev.ID)
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ProgramFormulaContext{ApplicationRecipeID: recipeID, RevisionUnpinned: unpinned}, nil
	}
	if err != nil {
		return ProgramFormulaContext{}, err
	}
	summary, err := FormulaSummaryFromSnapshot(rev.Snapshot)
	if err != nil {
		return ProgramFormulaContext{}, err
	}
	revID := rev.ID
	return ProgramFormulaContext{
		ApplicationRecipeID:         recipeID,
		ApplicationRecipeRevisionID: &revID,
		FormulaSnapshot:             summary,
		RevisionUnpinned:            unpinned,
	}, nil
}

// MergeProgramFormulaDetails adds recipe revision fields to an automation_runs.details map.
func MergeProgramFormulaDetails(ctx context.Context, q db.Querier, p db.Gr33nfertigationProgram, details map[string]any) {
	if details == nil {
		return
	}
	fc, err := ResolveProgramFormula(ctx, q, p)
	if err != nil {
		details["formula_resolve_error"] = err.Error()
		return
	}
	if fc.ApplicationRecipeID == 0 {
		return
	}
	details["application_recipe_id"] = fc.ApplicationRecipeID
	if fc.ApplicationRecipeRevisionID != nil {
		details["application_recipe_revision_id"] = *fc.ApplicationRecipeRevisionID
	}
	if fc.FormulaSnapshot != nil {
		details["formula_snapshot"] = fc.FormulaSnapshot
	}
	if fc.RevisionUnpinned {
		details["formula_revision_unpinned"] = true
	}
}

// MixingEventMetadata builds metadata JSON for a mixing event tied to a program.
func MixingEventMetadata(ctx context.Context, q db.Querier, p db.Gr33nfertigationProgram) (json.RawMessage, error) {
	fc, err := ResolveProgramFormula(ctx, q, p)
	if err != nil {
		return nil, err
	}
	if fc.ApplicationRecipeID == 0 {
		return json.RawMessage(`{}`), nil
	}
	meta := map[string]any{
		"application_recipe_id": fc.ApplicationRecipeID,
	}
	if fc.ApplicationRecipeRevisionID != nil {
		meta["application_recipe_revision_id"] = *fc.ApplicationRecipeRevisionID
	}
	if fc.FormulaSnapshot != nil {
		meta["formula_snapshot"] = fc.FormulaSnapshot
	}
	if fc.RevisionUnpinned {
		meta["formula_revision_unpinned"] = true
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}
