package programrules

import "errors"

// ErrIrrigationOnlyNoRecipe is returned when a client tries to attach a recipe
// to an irrigation-only program.
var ErrIrrigationOnlyNoRecipe = errors.New(
	"irrigation-only programs cannot use a nutrient recipe — uncheck irrigation only or remove the recipe")

// ErrRecipeRequiresFertigation is returned when a fertigation (mix) program
// is missing a recipe but is not marked irrigation-only.
var ErrRecipeOnNonIrrigation = errors.New(
	"application_recipe_id is only allowed on fertigation programs; for RO/well water use irrigation_only")

// ValidateCreateUpdate checks irrigation_only vs application_recipe_id invariants.
func ValidateCreateUpdate(irrigationOnly bool, recipeID *int64) error {
	if irrigationOnly {
		if recipeID != nil {
			return ErrIrrigationOnlyNoRecipe
		}
		return nil
	}
	return nil
}

// NeedsMixBatch reports whether the program tick should enqueue mix_batch before pulse.
func NeedsMixBatch(irrigationOnly bool, recipeID *int64) bool {
	return !irrigationOnly && recipeID != nil
}
