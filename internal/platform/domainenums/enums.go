// Package domainenums is the single source for platform dropdown enums (Phase 88).
package domainenums

import "strings"

// Option is one enum value with a farmer-facing label.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Payload is returned by GET /platform/domain-enums.
type Payload struct {
	GrowthStages              []Option `json:"growth_stages"`
	ReservoirStatuses         []Option `json:"reservoir_statuses"`
	CostCategories            []Option `json:"cost_categories"`
	ApplicationTargets        []Option `json:"application_targets"`
	InputDefinitionCategories []Option `json:"input_definition_categories"`
	BatchStatuses             []Option `json:"batch_statuses"`
}

// Ordered slices — aligned with Postgres enums and croplibrary.ValidGrowthStages.
var (
	growthStageOrder = []string{
		"clone", "seedling", "early_veg", "late_veg", "transition",
		"early_flower", "mid_flower", "late_flower", "flush", "harvest", "dry_cure",
	}
	reservoirStatusOrder = []string{
		"ready", "mixing", "needs_top_up", "needs_flush", "flushing", "offline", "empty",
	}
	costCategoryOrder = []string{
		"seeds_plants", "fertilizers_soil_amendments", "pest_disease_control", "water_irrigation",
		"labor_wages", "equipment_purchase_rental", "equipment_maintenance_fuel",
		"utilities_electricity_gas", "land_rent_mortgage", "insurance", "licenses_permits",
		"feed_livestock", "veterinary_services", "packaging_supplies", "transportation_logistics",
		"marketing_sales", "training_consultancy", "miscellaneous",
	}
	applicationTargetOrder = []string{
		"soil_drench", "foliar_spray", "seed_treatment", "compost_pile_inoculant",
		"livestock_water_supplement", "other",
	}
	inputDefinitionCategoryOrder = []string{
		"microbial_inoculant", "fermented_plant_juice", "water_soluble_nutrient",
		"oriental_herbal_nutrient", "fish_amino_acid", "insect_attractant_repellent",
		"soil_conditioner", "compost_tea_extract", "biochar_preparation",
		"other_ferment", "other_extract", "animal_feed", "bedding", "veterinary_supply",
	}
	batchStatusOrder = []string{
		"planning", "ingredients_gathered", "mixing_in_progress", "fermenting_brewing",
		"maturing_aging", "ready_for_use", "partially_used", "fully_used",
		"expired_discarded", "failed_production",
	}
)

func humanize(value string) string {
	return strings.ReplaceAll(value, "_", " ")
}

func options(values []string) []Option {
	out := make([]Option, len(values))
	for i, v := range values {
		out[i] = Option{Value: v, Label: humanize(v)}
	}
	return out
}

// All returns every platform enum for UI dropdowns.
func All() Payload {
	return Payload{
		GrowthStages:              options(growthStageOrder),
		ReservoirStatuses:         options(reservoirStatusOrder),
		CostCategories:            options(costCategoryOrder),
		ApplicationTargets:        options(applicationTargetOrder),
		InputDefinitionCategories: options(inputDefinitionCategoryOrder),
		BatchStatuses:             options(batchStatusOrder),
	}
}
