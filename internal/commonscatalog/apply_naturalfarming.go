package commonscatalog

import (
	"context"
	"fmt"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

func applyNaturalFarmingRecipePack(ctx context.Context, q db.Querier, farmID int64, body PackBody) (ApplyResult, error) {
	existingInputs, err := q.ListInputDefinitionsByFarm(ctx, farmID)
	if err != nil {
		return ApplyResult{Kind: KindNaturalFarmingRecipePack, Status: "failed"}, err
	}
	existingRecipes, err := q.ListRecipesByFarm(ctx, farmID)
	if err != nil {
		return ApplyResult{Kind: KindNaturalFarmingRecipePack, Status: "failed"}, err
	}

	inputByName := map[string]db.Gr33nnaturalfarmingInputDefinition{}
	for _, row := range existingInputs {
		inputByName[row.Name] = row
	}
	recipeByName := map[string]db.Gr33nnaturalfarmingApplicationRecipe{}
	for _, row := range existingRecipes {
		recipeByName[row.Name] = row
	}

	res := ApplyResult{
		Kind:    KindNaturalFarmingRecipePack,
		Status:  "applied",
		Message: "Natural farming inputs and recipes imported. Review in Natural Farming → Recipes.",
		NextSteps: []string{
			"Open Natural Farming → Recipes and confirm definitions match your operation.",
			"Make first batches for inputs you plan to use this season.",
		},
	}

	for _, spec := range body.InputDefinitions {
		name := strings.TrimSpace(spec.Name)
		if _, ok := inputByName[name]; ok {
			res.InputsSkipped++
			res.Details = append(res.Details, fmt.Sprintf("Skipped existing input %q", name))
			continue
		}
		row, err := q.CreateInputDefinition(ctx, db.CreateInputDefinitionParams{
			FarmID:             farmID,
			Name:               name,
			Category:           db.Gr33nnaturalfarmingInputCategoryEnum(strings.TrimSpace(spec.Category)),
			Description:        spec.Description,
			TypicalIngredients: spec.TypicalIngredients,
			PreparationSummary: spec.PreparationSummary,
			StorageGuidelines:  spec.StorageGuidelines,
			SafetyPrecautions:  spec.SafetyPrecautions,
			ReferenceSource:    spec.ReferenceSource,
		})
		if err != nil {
			return res, fmt.Errorf("input %q: %w", name, err)
		}
		inputByName[name] = row
		res.InputsCreated++
		res.Details = append(res.Details, fmt.Sprintf("Created input %q (id %d)", name, row.ID))
	}

	for _, spec := range body.ApplicationRecipes {
		name := strings.TrimSpace(spec.Name)
		if _, ok := recipeByName[name]; ok {
			res.RecipesSkipped++
			res.Details = append(res.Details, fmt.Sprintf("Skipped existing recipe %q", name))
			continue
		}
		var inputDefID *int64
		if spec.PrimaryInputName != nil {
			pn := strings.TrimSpace(*spec.PrimaryInputName)
			if pn != "" {
				if inp, ok := inputByName[pn]; ok {
					id := inp.ID
					inputDefID = &id
				}
			}
		}
		row, err := q.CreateRecipe(ctx, db.CreateRecipeParams{
			FarmID:                farmID,
			Name:                  name,
			InputDefinitionID:     inputDefID,
			Description:           spec.Description,
			TargetApplicationType: db.Gr33nnaturalfarmingApplicationTargetEnum(strings.TrimSpace(spec.TargetApplicationType)),
			DilutionRatio:         spec.DilutionRatio,
			Instructions:          spec.Instructions,
			FrequencyGuidelines:   spec.FrequencyGuidelines,
			Notes:                 spec.Notes,
		})
		if err != nil {
			return res, fmt.Errorf("recipe %q: %w", name, err)
		}
		recipeByName[name] = row
		res.RecipesCreated++
		res.Details = append(res.Details, fmt.Sprintf("Created recipe %q (id %d)", name, row.ID))
	}

	unitByName := map[string]int64{}
	for _, comp := range body.RecipeInputComponents {
		recipeName := strings.TrimSpace(comp.RecipeName)
		inputName := strings.TrimSpace(comp.InputName)
		recipe, ok := recipeByName[recipeName]
		if !ok {
			return ApplyResult{Kind: KindNaturalFarmingRecipePack, Status: "failed",
				Message: fmt.Sprintf("unknown recipe %q in components", recipeName)},
				fmt.Errorf("recipe_input_components: unknown recipe %q", recipeName)
		}
		inp, ok := inputByName[inputName]
		if !ok {
			return ApplyResult{Kind: KindNaturalFarmingRecipePack, Status: "failed",
				Message: fmt.Sprintf("unknown input %q in components", inputName)},
				fmt.Errorf("recipe_input_components: unknown input %q", inputName)
		}
		unitName := strings.TrimSpace(comp.PartUnitName)
		if unitName == "" {
			unitName = "decimal_fraction"
		}
		unitID, err := resolveUnitID(ctx, q, unitByName, unitName)
		if err != nil {
			return res, fmt.Errorf("component %q + %q: %w", recipeName, inputName, err)
		}
		pv, err := httputil.NumericFromFloat64(comp.PartValue)
		if err != nil {
			return res, fmt.Errorf("component %q + %q: invalid part_value", recipeName, inputName)
		}
		if err := q.AddRecipeComponent(ctx, db.AddRecipeComponentParams{
			ApplicationRecipeID: recipe.ID,
			InputDefinitionID:   inp.ID,
			PartValue:           pv,
			PartUnitID:          &unitID,
			Notes:               comp.Notes,
		}); err != nil {
			return res, fmt.Errorf("component %q + %q: %w", recipeName, inputName, err)
		}
		res.ComponentsUpserted++
	}

	if res.InputsCreated == 0 && res.RecipesCreated == 0 && res.ComponentsUpserted == 0 &&
		res.InputsSkipped > 0 && res.RecipesSkipped > 0 {
		res.Status = "noop"
		res.Message = "All pack inputs and recipes already exist on this farm — components refreshed where listed."
	}
	return res, nil
}

func resolveUnitID(ctx context.Context, q db.Querier, cache map[string]int64, name string) (int64, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}
	row, err := q.GetUnitByName(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("unknown part_unit_name %q", name)
	}
	cache[name] = row.ID
	return row.ID, nil
}
