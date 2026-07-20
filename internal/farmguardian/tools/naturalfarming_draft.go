// Phase 210 WS3 — Confirm-gated natural farming write tools.

package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/naturalfarmingcatalog"
)

var (
	nfDraftCatalogOnce sync.Once
	nfDraftMaterial    map[string]any
	nfDraftCanon       map[string]any
	nfDraftCatalogErr  error
)

func nfDraftCatalogs() (material, canon map[string]any, err error) {
	nfDraftCatalogOnce.Do(func() {
		root, err := croplibrary.FindRepoRoot()
		if err != nil {
			nfDraftCatalogErr = err
			return
		}
		nfDraftMaterial, nfDraftCatalogErr = naturalfarmingcatalog.LoadMaterialCatalog(root)
		if nfDraftCatalogErr != nil {
			return
		}
		nfDraftCanon, nfDraftCatalogErr = naturalfarmingcatalog.LoadRecipeCanon(root)
	})
	return nfDraftMaterial, nfDraftCanon, nfDraftCatalogErr
}

func execDraftInputDefinition(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	if raw, ok := args["farm_id"]; ok && raw != nil {
		return nil, errors.New("farm_id is set by proposal scope — omit from args")
	}

	resolved, err := resolveDraftInputFromCatalog(args)
	if err != nil {
		return nil, err
	}
	if resolved.name == "" {
		return nil, errors.New("name or catalog material_id required")
	}
	if resolved.category == "" {
		return nil, errors.New("category required")
	}
	if !isNFInputCategory(resolved.category) {
		return nil, fmt.Errorf("invalid category %q", resolved.category)
	}

	desc, err := optionalStringFromArgs(args, "description")
	if err != nil {
		return nil, err
	}
	if resolved.sourceTier != "" {
		note := "source_tier: " + resolved.sourceTier
		if desc == nil {
			desc = &note
		} else {
			merged := strings.TrimSpace(*desc) + "\n" + note
			desc = &merged
		}
	}

	if deps.Q == nil {
		return nil, errors.New("database unavailable")
	}

	row, err := deps.Q.CreateInputDefinition(ctx, db.CreateInputDefinitionParams{
		FarmID:             deps.FarmID,
		Name:               resolved.name,
		Category:           db.Gr33nnaturalfarmingInputCategoryEnum(resolved.category),
		Description:        desc,
		TypicalIngredients: strPtrFromArgs(args, "typical_ingredients"),
		PreparationSummary: strPtrFromArgs(args, "preparation_summary"),
		StorageGuidelines:  strPtrFromArgs(args, "storage_guidelines"),
		SafetyPrecautions:  strPtrFromArgs(args, "safety_precautions"),
		ReferenceSource:    strPtrFromArgs(args, "reference_source"),
	})
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"input_definition_id": row.ID,
		"name":                row.Name,
		"category":            string(row.Category),
		"natural_farming_url": "/natural-farming?tab=recipes",
	}
	if resolved.materialID != "" {
		out["material_id"] = resolved.materialID
	}
	return out, nil
}

func execDraftApplicationRecipe(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	if raw, ok := args["farm_id"]; ok && raw != nil {
		return nil, errors.New("farm_id is set by proposal scope — omit from args")
	}

	name, err := stringFromArgs(args, "name")
	if err != nil {
		return nil, err
	}
	target, err := stringFromArgs(args, "target_application_type")
	if err != nil {
		return nil, err
	}
	if !isNFApplicationTarget(target) {
		return nil, fmt.Errorf("invalid target_application_type %q", target)
	}
	dilution, err := stringFromArgs(args, "dilution_ratio")
	if err != nil {
		return nil, err
	}
	components, err := recipeComponentsFromArgs(args)
	if err != nil {
		return nil, err
	}

	if deps.Q == nil {
		return nil, errors.New("database unavailable")
	}
	var inputDefID *int64
	if id, err := optionalInt64FromArgs(args, "input_definition_id"); err != nil {
		return nil, err
	} else if id != nil {
		if err := ensureInputDefinitionOnFarm(ctx, deps.Q, deps.FarmID, *id); err != nil {
			return nil, err
		}
		inputDefID = id
	}

	row, err := deps.Q.CreateRecipe(ctx, db.CreateRecipeParams{
		FarmID:                deps.FarmID,
		Name:                  name,
		InputDefinitionID:     inputDefID,
		Description:           strPtrFromArgs(args, "description"),
		TargetApplicationType: db.Gr33nnaturalfarmingApplicationTargetEnum(strings.TrimSpace(target)),
		DilutionRatio:         &dilution,
		Instructions:          strPtrFromArgs(args, "instructions"),
		FrequencyGuidelines:   strPtrFromArgs(args, "frequency_guidelines"),
		Notes:                 strPtrFromArgs(args, "notes"),
	})
	if err != nil {
		return nil, err
	}

	added := 0
	for _, c := range components {
		if err := ensureInputDefinitionOnFarm(ctx, deps.Q, deps.FarmID, c.inputDefinitionID); err != nil {
			return nil, err
		}
		pv, err := httputil.NumericFromFloat64(c.partValue)
		if err != nil {
			return nil, fmt.Errorf("invalid part_value for input_definition_id %d", c.inputDefinitionID)
		}
		if err := deps.Q.AddRecipeComponent(ctx, db.AddRecipeComponentParams{
			ApplicationRecipeID: row.ID,
			InputDefinitionID:   c.inputDefinitionID,
			PartValue:           pv,
			PartUnitID:          c.partUnitID,
			Notes:               c.notes,
		}); err != nil {
			return nil, err
		}
		added++
	}

	out := map[string]any{
		"recipe_id":           row.ID,
		"name":                row.Name,
		"dilution_ratio":      dilution,
		"target_application_type": string(row.TargetApplicationType),
		"natural_farming_url": "/natural-farming?tab=recipes",
	}
	if inputDefID != nil {
		out["input_definition_id"] = *inputDefID
	}
	if added > 0 {
		out["components_added"] = added
	}
	if tier := strPtrFromArgs(args, "source_tier"); tier != nil {
		out["source_tier"] = *tier
	}
	return out, nil
}

func execDraftInputBatch(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	if raw, ok := args["farm_id"]; ok && raw != nil {
		return nil, errors.New("farm_id is set by proposal scope — omit from args")
	}

	inputDefID, err := int64FromArgs(args, "input_definition_id")
	if err != nil {
		return nil, err
	}

	status := db.Gr33nnaturalfarmingInputBatchStatusEnumFermentingBrewing
	if s, err := optionalStringFromArgs(args, "status"); err != nil {
		return nil, err
	} else if s != nil && strings.TrimSpace(*s) != "" {
		if !isNFBatchStatus(*s) {
			return nil, fmt.Errorf("invalid status %q", *s)
		}
		status = db.Gr33nnaturalfarmingInputBatchStatusEnum(strings.TrimSpace(*s))
	}

	if deps.Q == nil {
		return nil, errors.New("database unavailable")
	}
	if err := ensureInputDefinitionOnFarm(ctx, deps.Q, deps.FarmID, inputDefID); err != nil {
		return nil, err
	}

	params := db.CreateInputBatchParams{
		FarmID:            deps.FarmID,
		InputDefinitionID: inputDefID,
		Status:            status,
		BatchIdentifier:   strPtrFromArgs(args, "batch_identifier"),
		IngredientsUsed:   strPtrFromArgs(args, "ingredients_used"),
		ProcedureFollowed: strPtrFromArgs(args, "procedure_followed"),
		ObservationsNotes: strPtrFromArgs(args, "observations_notes"),
		StorageLocation:   strPtrFromArgs(args, "storage_location"),
	}

	if d, ok, err := optionalDateFromArgs(args, "creation_start_date"); err != nil {
		return nil, err
	} else if ok {
		params.CreationStartDate = d
	}
	if d, ok, err := optionalDateFromArgs(args, "creation_end_date"); err != nil {
		return nil, err
	} else if ok {
		params.CreationEndDate = d
	}
	if d, ok, err := optionalDateFromArgs(args, "expected_ready_date"); err != nil {
		return nil, err
	} else if ok {
		params.ExpectedReadyDate = d
	}
	if f, err := optionalFloat64FromArgs(args, "quantity_produced"); err != nil {
		return nil, err
	} else if f != nil {
		n, err := httputil.NumericFromFloat64(*f)
		if err != nil {
			return nil, errors.New("invalid quantity_produced")
		}
		params.QuantityProduced = n
	}
	if f, err := optionalFloat64FromArgs(args, "current_quantity_remaining"); err != nil {
		return nil, err
	} else if f != nil {
		n, err := httputil.NumericFromFloat64(*f)
		if err != nil {
			return nil, errors.New("invalid current_quantity_remaining")
		}
		params.CurrentQuantityRemaining = n
	}
	if uid, err := optionalInt64FromArgs(args, "quantity_unit_id"); err != nil {
		return nil, err
	} else if uid != nil {
		params.QuantityUnitID = uid
	}

	row, err := deps.Q.CreateInputBatch(ctx, params)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"batch_id":            row.ID,
		"input_definition_id": row.InputDefinitionID,
		"status":              string(row.Status),
		"natural_farming_url": "/natural-farming?tab=stock",
	}, nil
}

type recipeComponentSpec struct {
	inputDefinitionID int64
	partValue         float64
	partUnitID        *int64
	notes             *string
}

func recipeComponentsFromArgs(args map[string]any) ([]recipeComponentSpec, error) {
	raw, ok := args["components"]
	if !ok || raw == nil {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, errors.New("components must be an array")
	}
	out := make([]recipeComponentSpec, 0, len(items))
	for i, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("components[%d] must be an object", i)
		}
		id, err := int64FromArgs(m, "input_definition_id")
		if err != nil {
			return nil, fmt.Errorf("components[%d]: %w", i, err)
		}
		pv, err := float64FromArgs(m, "part_value")
		if err != nil {
			return nil, fmt.Errorf("components[%d]: part_value required", i)
		}
		uid, err := optionalInt64FromArgs(m, "part_unit_id")
		if err != nil {
			return nil, fmt.Errorf("components[%d]: %w", i, err)
		}
		notes, err := optionalStringFromArgs(m, "notes")
		if err != nil {
			return nil, fmt.Errorf("components[%d]: %w", i, err)
		}
		out = append(out, recipeComponentSpec{
			inputDefinitionID: id,
			partValue:         pv,
			partUnitID:        uid,
			notes:             notes,
		})
	}
	return out, nil
}

type resolvedDraftInput struct {
	name       string
	category   string
	sourceTier string
	materialID string
}

func resolveDraftInputFromCatalog(args map[string]any) (resolvedDraftInput, error) {
	out := resolvedDraftInput{}
	name, err := optionalStringFromArgs(args, "name")
	if err != nil {
		return out, err
	}
	if name != nil {
		out.name = strings.TrimSpace(*name)
	}
	materialID, err := optionalStringFromArgs(args, "material_id")
	if err != nil {
		return out, err
	}
	if materialID != nil {
		out.materialID = strings.TrimSpace(*materialID)
	}
	processType, err := optionalStringFromArgs(args, "process_type")
	if err != nil {
		return out, err
	}
	sourceTier, err := optionalStringFromArgs(args, "source_tier")
	if err != nil {
		return out, err
	}
	if sourceTier != nil {
		out.sourceTier = strings.TrimSpace(*sourceTier)
	}
	category, err := optionalStringFromArgs(args, "category")
	if err != nil {
		return out, err
	}
	if category != nil {
		out.category = strings.TrimSpace(*category)
	}

	matCat, canon, catErr := nfDraftCatalogs()
	if catErr != nil {
		return out, nil
	}
	pt := ""
	if processType != nil {
		pt = strings.TrimSpace(*processType)
	}
	if out.materialID != "" {
		mat, ok := naturalfarmingcatalog.MaterialByID(matCat, out.materialID)
		if !ok {
			return out, fmt.Errorf("unknown material_id %q", out.materialID)
		}
		if out.name == "" {
			out.name = materialDisplayName(mat)
		}
		if out.sourceTier == "" {
			if tier, _ := mat["source_tier"].(string); tier != "" {
				out.sourceTier = strings.TrimSpace(tier)
			}
		}
		if pt == "" {
			if procs, _ := mat["processes"].([]any); len(procs) > 0 {
				if proc, ok := procs[0].(map[string]any); ok {
					pt, _ = proc["type"].(string)
					pt = strings.TrimSpace(pt)
				}
			}
		}
	}
	if out.category == "" && pt != "" {
		if inp, ok := naturalfarmingcatalog.CanonInputByProcessType(canon, pt); ok {
			if sc, _ := inp["schema_category"].(string); sc != "" {
				out.category = strings.TrimSpace(sc)
			}
			if out.name == "" && out.materialID != "" {
				if seed, _ := inp["seed_name"].(string); seed != "" {
					out.name = materialDisplayNameFromID(out.materialID) + " (" + strings.TrimSpace(seed) + ")"
				}
			}
		}
	}
	return out, nil
}

func materialDisplayName(mat map[string]any) string {
	if names, _ := mat["common_names"].([]any); len(names) > 0 {
		if n, _ := names[0].(string); strings.TrimSpace(n) != "" {
			return strings.TrimSpace(n)
		}
	}
	if id, _ := mat["id"].(string); id != "" {
		return strings.ReplaceAll(id, "_", " ")
	}
	return ""
}

func materialDisplayNameFromID(id string) string {
	matCat, _, err := nfDraftCatalogs()
	if err != nil {
		return id
	}
	if mat, ok := naturalfarmingcatalog.MaterialByID(matCat, id); ok {
		if n := materialDisplayName(mat); n != "" {
			return n
		}
	}
	return id
}

func ensureInputDefinitionOnFarm(ctx context.Context, q db.Querier, farmID, inputDefID int64) error {
	row, err := q.GetInputDefinitionByID(ctx, inputDefID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("input_definition_id not on farm")
		}
		return err
	}
	return ensureFarmScope(row.FarmID, farmID)
}

func strPtrFromArgs(args map[string]any, key string) *string {
	s, err := optionalStringFromArgs(args, key)
	if err != nil || s == nil {
		return nil
	}
	return s
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

func isNFBatchStatus(status string) bool {
	switch db.Gr33nnaturalfarmingInputBatchStatusEnum(strings.TrimSpace(status)) {
	case db.Gr33nnaturalfarmingInputBatchStatusEnumPlanning,
		db.Gr33nnaturalfarmingInputBatchStatusEnumIngredientsGathered,
		db.Gr33nnaturalfarmingInputBatchStatusEnumMixingInProgress,
		db.Gr33nnaturalfarmingInputBatchStatusEnumFermentingBrewing,
		db.Gr33nnaturalfarmingInputBatchStatusEnumMaturingAging,
		db.Gr33nnaturalfarmingInputBatchStatusEnumReadyForUse,
		db.Gr33nnaturalfarmingInputBatchStatusEnumPartiallyUsed,
		db.Gr33nnaturalfarmingInputBatchStatusEnumFullyUsed,
		db.Gr33nnaturalfarmingInputBatchStatusEnumExpiredDiscarded,
		db.Gr33nnaturalfarmingInputBatchStatusEnumFailedProduction:
		return true
	default:
		return false
	}
}
