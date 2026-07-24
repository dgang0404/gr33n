package reciperevision

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
)

// Snapshot is the immutable audit payload stored on each revision row.
type Snapshot struct {
	Recipe     RecipeSnapshot      `json:"recipe"`
	Components []ComponentSnapshot `json:"components"`
}

type RecipeSnapshot struct {
	ID                    int64   `json:"id"`
	FarmID                int64   `json:"farm_id"`
	Name                  string  `json:"name"`
	InputDefinitionID     *int64  `json:"input_definition_id"`
	Description           *string `json:"description"`
	TargetApplicationType string  `json:"target_application_type"`
	DilutionRatio         *string `json:"dilution_ratio"`
	Instructions          *string `json:"instructions"`
	FrequencyGuidelines   *string `json:"frequency_guidelines"`
	Notes                 *string `json:"notes"`
}

type ComponentSnapshot struct {
	InputDefinitionID int64   `json:"input_definition_id"`
	InputName         string  `json:"input_name"`
	PartValue         float64 `json:"part_value"`
	PartUnitID        *int64  `json:"part_unit_id"`
	Notes             *string `json:"notes"`
}

func userIDFromContext(ctx context.Context) pgtype.UUID {
	uid, ok := authctx.UserID(ctx)
	if !ok {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: uid, Valid: true}
}

func partValueFloat(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

// BuildSnapshot copies the live recipe row and components into JSON.
func BuildSnapshot(recipe db.Gr33nnaturalfarmingApplicationRecipe, components []db.ListRecipeComponentsRow) (json.RawMessage, error) {
	snap := Snapshot{
		Recipe: RecipeSnapshot{
			ID:                    recipe.ID,
			FarmID:                recipe.FarmID,
			Name:                  recipe.Name,
			InputDefinitionID:     recipe.InputDefinitionID,
			Description:           recipe.Description,
			TargetApplicationType: string(recipe.TargetApplicationType),
			DilutionRatio:         recipe.DilutionRatio,
			Instructions:          recipe.Instructions,
			FrequencyGuidelines:   recipe.FrequencyGuidelines,
			Notes:                 recipe.Notes,
		},
		Components: make([]ComponentSnapshot, 0, len(components)),
	}
	for _, c := range components {
		snap.Components = append(snap.Components, ComponentSnapshot{
			InputDefinitionID: c.InputDefinitionID,
			InputName:         c.InputName,
			PartValue:         partValueFloat(c.PartValue),
			PartUnitID:        c.PartUnitID,
			Notes:             c.Notes,
		})
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

// SnapshotFromRecipe loads the current recipe + components and builds JSON.
func SnapshotFromRecipe(ctx context.Context, q db.Querier, recipeID int64) (json.RawMessage, error) {
	recipe, err := q.GetRecipeByID(ctx, recipeID)
	if err != nil {
		return nil, err
	}
	components, err := q.ListRecipeComponents(ctx, recipeID)
	if err != nil {
		return nil, err
	}
	return BuildSnapshot(recipe, components)
}

// Record appends a new revision for the recipe's current live state.
func Record(ctx context.Context, q db.Querier, recipeID int64, summary string) (db.Gr33nnaturalfarmingApplicationRecipeRevision, error) {
	snapshot, err := SnapshotFromRecipe(ctx, q, recipeID)
	if err != nil {
		return db.Gr33nnaturalfarmingApplicationRecipeRevision{}, err
	}
	var summaryPtr *string
	if summary != "" {
		summaryPtr = &summary
	}
	return q.CreateRecipeRevision(ctx, db.CreateRecipeRevisionParams{
		ApplicationRecipeID: recipeID,
		Snapshot:            snapshot,
		ChangeSummary:       summaryPtr,
		CreatedByUserID:     userIDFromContext(ctx),
	})
}

// LatestRevisionID returns the newest revision for a recipe, or nil when none exist.
func LatestRevisionID(ctx context.Context, q db.Querier, recipeID int64) (*int64, error) {
	rev, err := q.GetLatestRecipeRevision(ctx, recipeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	id := rev.ID
	return &id, nil
}

// PinProgramRecipeRevision picks the revision id stored on a fertigation program link.
// New links pin the latest revision; unchanged recipe keeps the existing pin.
func PinProgramRecipeRevision(
	ctx context.Context,
	q db.Querier,
	previousRecipeID, newRecipeID *int64,
	previousRevisionID *int64,
	irrigationOnly bool,
) (*int64, error) {
	if irrigationOnly || newRecipeID == nil {
		return nil, nil
	}
	if previousRecipeID != nil && *previousRecipeID == *newRecipeID {
		return previousRevisionID, nil
	}
	return LatestRevisionID(ctx, q, *newRecipeID)
}
