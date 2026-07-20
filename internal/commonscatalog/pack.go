// Package commonscatalog validates and applies commons catalog pack bodies on import.
package commonscatalog

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	CatalogVersion = "gr33n.commons_catalog.v1"

	KindFertigationRecipePack     = "fertigation_recipe_pack"
	KindNaturalFarmingRecipePack  = "natural_farming_recipe_pack"
	KindAgronomySeedPack          = "agronomy_seed_pack"
	KindDocumentationPack         = "documentation_pack"
)

// PackBody is the JSON shape stored in commons_catalog_entries.body.
type PackBody struct {
	CatalogVersion string `json:"catalog_version"`
	Kind           string `json:"kind"`
	ReadmeMD       string `json:"readme_md"`
	// fertigation_recipe_pack
	Programs []RecipeProgram `json:"programs"`
	// natural_farming_recipe_pack
	PackKey              string                      `json:"pack_key"`
	ReferenceSource      string                      `json:"reference_source"`
	InputDefinitions     []NFInputDefinitionSpec     `json:"input_definitions"`
	ApplicationRecipes   []NFApplicationRecipeSpec   `json:"application_recipes"`
	RecipeInputComponents []NFRecipeComponentSpec    `json:"recipe_input_components"`
	// agronomy_seed_pack
	PlatformCatalogVersion int            `json:"platform_catalog_version"`
	ExpectedCounts         map[string]int `json:"expected_counts"`
}

// RecipeProgram is one fertigation program in a recipe pack.
type RecipeProgram struct {
	Name                string             `json:"name"`
	Description         *string            `json:"description"`
	TotalVolumeLiters   float64            `json:"total_volume_liters"`
	EcTriggerLow        float64            `json:"ec_trigger_low"`
	PhTriggerLow        float64            `json:"ph_trigger_low"`
	PhTriggerHigh       float64            `json:"ph_trigger_high"`
	IsActive            bool               `json:"is_active"`
	RecommendedCropKeys []string           `json:"recommended_crop_keys"`
	RecommendedStages   []string           `json:"recommended_stages"`
	ProfileECSource     *profileECSource   `json:"profile_ec_source"`
	ECBandMSCM          *ecBand            `json:"ec_band_mscm"`
}

type profileECSource struct {
	CropKey string `json:"crop_key"`
	Stage   string `json:"stage"`
}

type ecBand struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// NFInputDefinitionSpec is one natural-farming input in a recipe pack.
type NFInputDefinitionSpec struct {
	Name               string  `json:"name"`
	Category           string  `json:"category"`
	Description        *string `json:"description"`
	TypicalIngredients *string `json:"typical_ingredients"`
	PreparationSummary *string `json:"preparation_summary"`
	StorageGuidelines  *string `json:"storage_guidelines"`
	SafetyPrecautions  *string `json:"safety_precautions"`
	ReferenceSource    *string `json:"reference_source"`
}

// NFApplicationRecipeSpec is one application recipe in a natural-farming pack.
type NFApplicationRecipeSpec struct {
	Name                  string   `json:"name"`
	Description           *string  `json:"description"`
	TargetApplicationType string   `json:"target_application_type"`
	DilutionRatio         *string  `json:"dilution_ratio"`
	Instructions          *string  `json:"instructions"`
	FrequencyGuidelines   *string  `json:"frequency_guidelines"`
	TargetCropCategories  []string `json:"target_crop_categories"`
	TargetGrowthStages    []string `json:"target_growth_stages"`
	PrimaryInputName      *string  `json:"primary_input_name"`
	Notes                 *string  `json:"notes"`
}

// NFRecipeComponentSpec links a recipe to an input by name (resolved at apply time).
type NFRecipeComponentSpec struct {
	RecipeName    string  `json:"recipe_name"`
	InputName     string  `json:"input_name"`
	PartValue     float64 `json:"part_value"`
	PartUnitName  string  `json:"part_unit_name"`
	Notes         *string `json:"notes"`
}

// ApplyResult is returned after import auto-apply.
type ApplyResult struct {
	Kind            string   `json:"kind"`
	Status          string   `json:"status"` // applied | verified | noop | skipped | failed
	Message         string   `json:"message"`
	ProgramsCreated int      `json:"programs_created,omitempty"`
	ProgramsUpdated int      `json:"programs_updated,omitempty"`
	ProgramsSkipped int      `json:"programs_skipped,omitempty"`
	InputsCreated   int      `json:"inputs_created,omitempty"`
	InputsSkipped   int      `json:"inputs_skipped,omitempty"`
	RecipesCreated  int      `json:"recipes_created,omitempty"`
	RecipesSkipped  int      `json:"recipes_skipped,omitempty"`
	ComponentsUpserted int   `json:"components_upserted,omitempty"`
	NextSteps       []string `json:"next_steps,omitempty"`
	Details         []string `json:"details,omitempty"`
}

func ParsePackBody(raw json.RawMessage) (PackBody, error) {
	if len(raw) == 0 {
		return PackBody{}, fmt.Errorf("pack body is empty")
	}
	var b PackBody
	if err := json.Unmarshal(raw, &b); err != nil {
		return PackBody{}, fmt.Errorf("invalid pack body JSON: %w", err)
	}
	b.Kind = strings.TrimSpace(b.Kind)
	if b.Kind == "" {
		return b, fmt.Errorf("pack body missing kind")
	}
	if b.CatalogVersion != "" && b.CatalogVersion != CatalogVersion {
		return b, fmt.Errorf("unsupported catalog_version %q (want %s)", b.CatalogVersion, CatalogVersion)
	}
	return b, nil
}

func NormalizeSlug(slug string) (string, error) {
	s := strings.TrimSpace(strings.ToLower(slug))
	if s == "" {
		return "", fmt.Errorf("slug is required")
	}
	if len(s) > 120 {
		return "", fmt.Errorf("slug too long (max 120)")
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return "", fmt.Errorf("slug must use lowercase letters, digits, and hyphens only")
	}
	return s, nil
}
