package commonscatalog

import (
	"fmt"
	"strings"

	db "gr33n-api/internal/db"
)

// ValidatePublishBody checks a pack before POST /commons/catalog.
func ValidatePublishBody(b PackBody) error {
	switch b.Kind {
	case KindFertigationRecipePack:
		if len(b.Programs) == 0 {
			return fmt.Errorf("fertigation_recipe_pack requires at least one program")
		}
		for i, p := range b.Programs {
			if strings.TrimSpace(p.Name) == "" {
				return fmt.Errorf("programs[%d].name is required", i)
			}
			if p.TotalVolumeLiters <= 0 {
				return fmt.Errorf("programs[%d].total_volume_liters must be positive", i)
			}
		}
	case KindNaturalFarmingRecipePack:
		return validateNaturalFarmingRecipePack(b)
	case KindAgronomySeedPack:
		if b.PlatformCatalogVersion < 1 {
			return fmt.Errorf("agronomy_seed_pack requires platform_catalog_version >= 1")
		}
	case KindDocumentationPack:
		if strings.TrimSpace(b.ReadmeMD) == "" {
			return fmt.Errorf("documentation_pack requires readme_md")
		}
	default:
		return fmt.Errorf("unsupported pack kind %q", b.Kind)
	}
	return nil
}

func validateNaturalFarmingRecipePack(b PackBody) error {
	if len(b.InputDefinitions) == 0 {
		return fmt.Errorf("natural_farming_recipe_pack requires at least one input_definition")
	}
	inputNames := map[string]struct{}{}
	for i, inp := range b.InputDefinitions {
		name := strings.TrimSpace(inp.Name)
		if name == "" {
			return fmt.Errorf("input_definitions[%d].name is required", i)
		}
		if _, dup := inputNames[name]; dup {
			return fmt.Errorf("duplicate input_definitions name %q", name)
		}
		inputNames[name] = struct{}{}
		cat := strings.TrimSpace(inp.Category)
		if cat == "" {
			return fmt.Errorf("input_definitions[%d].category is required", i)
		}
		if !isNFInputCategory(cat) {
			return fmt.Errorf("input_definitions[%d].invalid category %q", i, cat)
		}
	}

	recipeNames := map[string]struct{}{}
	for i, rec := range b.ApplicationRecipes {
		name := strings.TrimSpace(rec.Name)
		if name == "" {
			return fmt.Errorf("application_recipes[%d].name is required", i)
		}
		if _, dup := recipeNames[name]; dup {
			return fmt.Errorf("duplicate application_recipes name %q", name)
		}
		recipeNames[name] = struct{}{}
		target := strings.TrimSpace(rec.TargetApplicationType)
		if target == "" {
			return fmt.Errorf("application_recipes[%d].target_application_type is required", i)
		}
		if !isNFApplicationTarget(target) {
			return fmt.Errorf("application_recipes[%d].invalid target_application_type %q", i, target)
		}
		if rec.PrimaryInputName != nil {
			pn := strings.TrimSpace(*rec.PrimaryInputName)
			if pn != "" {
				if _, ok := inputNames[pn]; !ok {
					return fmt.Errorf("application_recipes[%d].primary_input_name %q not in input_definitions", i, pn)
				}
			}
		}
	}

	for i, comp := range b.RecipeInputComponents {
		rn := strings.TrimSpace(comp.RecipeName)
		in := strings.TrimSpace(comp.InputName)
		if rn == "" {
			return fmt.Errorf("recipe_input_components[%d].recipe_name is required", i)
		}
		if in == "" {
			return fmt.Errorf("recipe_input_components[%d].input_name is required", i)
		}
		if _, ok := recipeNames[rn]; !ok {
			return fmt.Errorf("recipe_input_components[%d].recipe_name %q not in application_recipes", i, rn)
		}
		if _, ok := inputNames[in]; !ok {
			return fmt.Errorf("recipe_input_components[%d].input_name %q not in input_definitions", i, in)
		}
		if comp.PartValue <= 0 {
			return fmt.Errorf("recipe_input_components[%d].part_value must be positive", i)
		}
	}
	return nil
}

func isNFInputCategory(cat string) bool {
	switch db.Gr33nnaturalfarmingInputCategoryEnum(strings.TrimSpace(cat)) {
	case db.Gr33nnaturalfarmingInputCategoryEnumMicrobialInoculant,
		db.Gr33nnaturalfarmingInputCategoryEnumFermentedPlantJuice,
		db.Gr33nnaturalfarmingInputCategoryEnumWaterSolubleNutrient,
		db.Gr33nnaturalfarmingInputCategoryEnumOrientalHerbalNutrient,
		db.Gr33nnaturalfarmingInputCategoryEnumFishAminoAcid,
		db.Gr33nnaturalfarmingInputCategoryEnumInsectAttractantRepellent,
		db.Gr33nnaturalfarmingInputCategoryEnumSoilConditioner,
		db.Gr33nnaturalfarmingInputCategoryEnumCompostTeaExtract,
		db.Gr33nnaturalfarmingInputCategoryEnumBiocharPreparation,
		db.Gr33nnaturalfarmingInputCategoryEnumOtherFerment,
		db.Gr33nnaturalfarmingInputCategoryEnumOtherExtract,
		db.Gr33nnaturalfarmingInputCategoryEnumAnimalFeed,
		db.Gr33nnaturalfarmingInputCategoryEnumBedding,
		db.Gr33nnaturalfarmingInputCategoryEnumVeterinarySupply:
		return true
	default:
		return false
	}
}

func isNFApplicationTarget(target string) bool {
	switch db.Gr33nnaturalfarmingApplicationTargetEnum(strings.TrimSpace(target)) {
	case db.Gr33nnaturalfarmingApplicationTargetEnumSoilDrench,
		db.Gr33nnaturalfarmingApplicationTargetEnumFoliarSpray,
		db.Gr33nnaturalfarmingApplicationTargetEnumSeedTreatment,
		db.Gr33nnaturalfarmingApplicationTargetEnumCompostPileInoculant,
		db.Gr33nnaturalfarmingApplicationTargetEnumLivestockWaterSupplement,
		db.Gr33nnaturalfarmingApplicationTargetEnumOther:
		return true
	default:
		return false
	}
}

// ValidateRecipeCropKeys ensures recommended_crop_keys exist in platform catalog.
func ValidateRecipeCropKeys(programs []RecipeProgram, entries []db.Gr33ncropsCropCatalogEntry) error {
	valid := map[string]struct{}{}
	for _, e := range entries {
		k := strings.ToLower(strings.TrimSpace(e.CropKey))
		if k != "" {
			valid[k] = struct{}{}
		}
	}
	if len(valid) == 0 {
		return fmt.Errorf("platform crop catalog is empty")
	}
	for _, p := range programs {
		for _, ck := range p.RecommendedCropKeys {
			norm := strings.ToLower(strings.TrimSpace(ck))
			if norm == "" {
				continue
			}
			if _, ok := valid[norm]; !ok {
				return fmt.Errorf("unknown crop_key %q in program %q", ck, p.Name)
			}
		}
		if p.ProfileECSource != nil {
			norm := strings.ToLower(strings.TrimSpace(p.ProfileECSource.CropKey))
			if norm != "" {
				if _, ok := valid[norm]; !ok {
					return fmt.Errorf("unknown profile_ec_source.crop_key %q in program %q", p.ProfileECSource.CropKey, p.Name)
				}
			}
		}
	}
	return nil
}
