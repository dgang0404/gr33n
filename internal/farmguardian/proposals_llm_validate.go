package farmguardian

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian/tools"
	"gr33n-api/internal/naturalfarmingcatalog"
)

// llmGrowthStageAliases mirrors tools/cycles.go for WS2 stage validation.
var llmGrowthStageAliases = map[string]struct{}{
	"clone": {}, "seedling": {}, "early_veg": {}, "late_veg": {}, "transition": {},
	"early_flower": {}, "mid_flower": {}, "late_flower": {}, "flush": {}, "harvest": {}, "dry_cure": {},
	"veg": {}, "vegetative": {}, "vegetation": {}, "flower": {}, "flowering": {}, "bloom": {}, "blooming": {},
	"dry": {}, "drying": {}, "cure": {}, "curing": {}, "flushing": {}, "harvesting": {}, "cutting": {}, "sprout": {}, "germination": {},
}

// ValidateLLMProposalDraft runs WS1 gates, per-tool schema (WS2), and farm ID binding.
func ValidateLLMProposalDraft(
	ctx context.Context,
	q db.Querier,
	farmID int64,
	draft LLMProposalDraft,
	hasAdmin bool,
) (rejectReason string) {
	if reason := validateLLMProposalCore(draft, hasAdmin); reason != "" {
		return reason
	}
	if reason := validateLLMProposalSchema(draft.Tool, draft.Args); reason != "" {
		return reason
	}
	if q != nil && farmID > 0 {
		if reason := bindLLMProposalFarmIDs(ctx, q, farmID, draft.Tool, draft.Args); reason != "" {
			return reason
		}
	}
	return ""
}

func validateLLMProposalCore(draft LLMProposalDraft, hasAdmin bool) string {
	tool := strings.TrimSpace(draft.Tool)
	if tool == "" {
		return "missing tool"
	}
	if !IsLLMToolAllowed(tool) {
		return "tool not on LLM allowlist"
	}
	t, err := tools.Lookup(tool)
	if err != nil {
		return "unknown tool"
	}
	if t.RequiresAdmin && !hasAdmin {
		return "tool requires admin"
	}
	if draft.Summary == "" {
		return "missing summary"
	}
	if len(draft.Args) == 0 {
		return "missing args"
	}
	if strings.EqualFold(draft.Confidence, "low") && tools.RiskTierForTool(tool, draft.Args) == "high" {
		return "low confidence on high-tier tool"
	}
	return ""
}

func validateLLMProposalSchema(tool string, args map[string]any) string {
	switch strings.TrimSpace(tool) {
	case "patch_fertigation_program":
		if _, err := llmArgInt64(args, "program_id"); err != nil {
			return "program_id required"
		}
		if !llmHasPatchField(args, "total_volume_liters", "is_active", "irrigation_only", "ec_target_id") {
			return "at least one patch field required"
		}
		if err := llmOptionalFloat64(args, "total_volume_liters"); err != nil {
			return err.Error()
		}
		if _, err := llmOptionalBool(args, "is_active"); err != nil {
			return err.Error()
		}
		if _, err := llmOptionalBool(args, "irrigation_only"); err != nil {
			return err.Error()
		}
		return ""
	case "patch_schedule":
		if _, err := llmArgInt64(args, "schedule_id"); err != nil {
			return "schedule_id required"
		}
		if !llmHasPatchField(args, "is_active", "name", "cron_expression") {
			return "at least one patch field required"
		}
		return ""
	case "patch_rule":
		if _, err := llmArgInt64(args, "rule_id"); err != nil {
			return "rule_id required"
		}
		active, err := llmOptionalBool(args, "is_active")
		if err != nil {
			return err.Error()
		}
		if _, hasThreshold := args["threshold"]; hasThreshold {
			return "LLM patch_rule may not set threshold v1"
		}
		if active == nil {
			return "is_active required for patch_rule"
		}
		if *active {
			return "LLM patch_rule only allows is_active false v1"
		}
		return ""
	case "ack_alert":
		if _, err := llmArgInt64(args, "alert_id"); err != nil {
			return "alert_id required"
		}
		return ""
	case "create_task":
		if _, err := llmArgString(args, "title"); err != nil {
			return "title required"
		}
		if _, err := llmOptionalInt64(args, "zone_id"); err != nil {
			return err.Error()
		}
		if raw, ok := args["due_date"]; ok && raw != nil {
			s, err := llmArgString(args, "due_date")
			if err != nil || !isISODate(s) {
				return "due_date must be YYYY-MM-DD"
			}
		}
		return ""
	case "create_task_from_alert":
		if _, err := llmArgInt64(args, "alert_id"); err != nil {
			return "alert_id required"
		}
		if raw, ok := args["due_date"]; ok && raw != nil {
			s, err := llmArgString(args, "due_date")
			if err != nil || !isISODate(s) {
				return "due_date must be YYYY-MM-DD"
			}
		}
		return ""
	case "update_cycle_stage":
		if _, err := llmArgInt64(args, "crop_cycle_id", "cycle_id"); err != nil {
			return "crop_cycle_id required"
		}
		stage, err := llmArgString(args, "current_stage")
		if err != nil {
			return "current_stage required"
		}
		if !isKnownGrowthStage(stage) {
			return "invalid current_stage"
		}
		return ""
	case "draft_input_definition":
		if reason := llmRejectFarmIDArg(args); reason != "" {
			return reason
		}
		hasName := llmHasNonEmptyString(args, "name")
		matID, err := llmOptionalString(args, "material_id")
		if err != nil {
			return err.Error()
		}
		if !hasName && matID == "" {
			return "name or material_id required"
		}
		if matID != "" && !llmMaterialIDKnown(matID) {
			return "unknown material_id"
		}
		if hasName && matID == "" {
			cat, err := llmArgString(args, "category")
			if err != nil {
				return "category required with name"
			}
			if !isLLMNFInputCategory(cat) {
				return "invalid category"
			}
		}
		if raw, ok := args["category"]; ok && raw != nil {
			cat, err := llmArgString(args, "category")
			if err != nil {
				return err.Error()
			}
			if !isLLMNFInputCategory(cat) {
				return "invalid category"
			}
		}
		return ""
	case "draft_application_recipe":
		if reason := llmRejectFarmIDArg(args); reason != "" {
			return reason
		}
		if _, err := llmArgString(args, "name"); err != nil {
			return "name required"
		}
		target, err := llmArgString(args, "target_application_type")
		if err != nil {
			return "target_application_type required"
		}
		if !isLLMNFApplicationTarget(target) {
			return "invalid target_application_type"
		}
		if _, err := llmArgString(args, "dilution_ratio"); err != nil {
			return "dilution_ratio required"
		}
		if _, err := llmOptionalInt64(args, "input_definition_id"); err != nil {
			return err.Error()
		}
		return llmValidateRecipeComponentsSchema(args)
	case "draft_input_batch":
		if reason := llmRejectFarmIDArg(args); reason != "" {
			return reason
		}
		if _, err := llmArgInt64(args, "input_definition_id"); err != nil {
			return "input_definition_id required"
		}
		if raw, ok := args["status"]; ok && raw != nil {
			st, err := llmArgString(args, "status")
			if err != nil {
				return err.Error()
			}
			if !isLLMNFBatchStatus(st) {
				return "invalid status"
			}
		}
		return ""
	default:
		return "unsupported tool schema"
	}
}

func bindLLMProposalFarmIDs(ctx context.Context, q db.Querier, farmID int64, tool string, args map[string]any) string {
	switch strings.TrimSpace(tool) {
	case "patch_fertigation_program":
		id, _ := llmArgInt64(args, "program_id")
		row, err := q.GetFertigationProgramByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "program_id not on farm"
			}
			return "program lookup failed"
		}
		if row.FarmID != farmID {
			return "program_id not on farm"
		}
	case "patch_schedule":
		id, _ := llmArgInt64(args, "schedule_id")
		row, err := q.GetScheduleByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "schedule_id not on farm"
			}
			return "schedule lookup failed"
		}
		if row.FarmID != farmID {
			return "schedule_id not on farm"
		}
	case "patch_rule":
		id, _ := llmArgInt64(args, "rule_id")
		row, err := q.GetAutomationRuleByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "rule_id not on farm"
			}
			return "rule lookup failed"
		}
		if row.FarmID != farmID {
			return "rule_id not on farm"
		}
	case "ack_alert", "create_task_from_alert":
		id, _ := llmArgInt64(args, "alert_id")
		row, err := q.GetAlertNotificationByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "alert_id not on farm"
			}
			return "alert lookup failed"
		}
		if row.FarmID != farmID {
			return "alert_id not on farm"
		}
	case "create_task":
		if zid, err := llmOptionalInt64(args, "zone_id"); err != nil {
			return err.Error()
		} else if zid != nil {
			z, err := q.GetZoneByID(ctx, *zid)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return "zone_id not on farm"
				}
				return "zone lookup failed"
			}
			if z.FarmID != farmID {
				return "zone_id not on farm"
			}
		}
	case "update_cycle_stage":
		id, _ := llmArgInt64(args, "crop_cycle_id", "cycle_id")
		row, err := q.GetCropCycleByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "crop_cycle_id not on farm"
			}
			return "crop cycle lookup failed"
		}
		if row.FarmID != farmID {
			return "crop_cycle_id not on farm"
		}
	case "draft_input_definition":
		if reason := llmRejectFarmIDArg(args); reason != "" {
			return reason
		}
	case "draft_application_recipe":
		if reason := llmRejectFarmIDArg(args); reason != "" {
			return reason
		}
		if id, err := llmOptionalInt64(args, "input_definition_id"); err != nil {
			return err.Error()
		} else if id != nil {
			if reason := bindLLMInputDefinitionOnFarm(ctx, q, farmID, *id); reason != "" {
				return reason
			}
		}
		return bindLLMRecipeComponentsOnFarm(ctx, q, farmID, args)
	case "draft_input_batch":
		if reason := llmRejectFarmIDArg(args); reason != "" {
			return reason
		}
		id, _ := llmArgInt64(args, "input_definition_id")
		return bindLLMInputDefinitionOnFarm(ctx, q, farmID, id)
	}
	return ""
}

func bindLLMInputDefinitionOnFarm(ctx context.Context, q db.Querier, farmID, inputDefID int64) string {
	row, err := q.GetInputDefinitionByID(ctx, inputDefID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "input_definition_id not on farm"
		}
		return "input definition lookup failed"
	}
	if row.FarmID != farmID {
		return "input_definition_id not on farm"
	}
	return ""
}

func bindLLMRecipeComponentsOnFarm(ctx context.Context, q db.Querier, farmID int64, args map[string]any) string {
	raw, ok := args["components"]
	if !ok || raw == nil {
		return ""
	}
	items, ok := raw.([]any)
	if !ok {
		return "components must be an array"
	}
	for i, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			return "components must be an array of objects"
		}
		id, err := llmArgInt64(m, "input_definition_id")
		if err != nil {
			return "components[" + strconv.Itoa(i) + "]: input_definition_id required"
		}
		if reason := bindLLMInputDefinitionOnFarm(ctx, q, farmID, id); reason != "" {
			return "components[" + strconv.Itoa(i) + "]: " + reason
		}
	}
	return ""
}

func llmValidateRecipeComponentsSchema(args map[string]any) string {
	raw, ok := args["components"]
	if !ok || raw == nil {
		return ""
	}
	items, ok := raw.([]any)
	if !ok {
		return "components must be an array"
	}
	for i, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			return "components must be an array of objects"
		}
		if _, err := llmArgInt64(m, "input_definition_id"); err != nil {
			return "components[" + strconv.Itoa(i) + "]: input_definition_id required"
		}
		if err := llmRequireFloat64(m, "part_value"); err != nil {
			return "components[" + strconv.Itoa(i) + "]: " + err.Error()
		}
	}
	return ""
}

var (
	llmNFCatalogOnce sync.Once
	llmNFMaterialCat map[string]any
	llmNFCatalogErr  error
)

func llmNFMaterialCatalog() (map[string]any, error) {
	llmNFCatalogOnce.Do(func() {
		root, err := croplibrary.FindRepoRoot()
		if err != nil {
			llmNFCatalogErr = err
			return
		}
		llmNFMaterialCat, llmNFCatalogErr = naturalfarmingcatalog.LoadMaterialCatalog(root)
	})
	return llmNFMaterialCat, llmNFCatalogErr
}

func llmMaterialIDKnown(materialID string) bool {
	materialID = strings.TrimSpace(materialID)
	if materialID == "" {
		return false
	}
	cat, err := llmNFMaterialCatalog()
	if err != nil {
		return false
	}
	_, ok := naturalfarmingcatalog.MaterialByID(cat, materialID)
	return ok
}

func llmRejectFarmIDArg(args map[string]any) string {
	if raw, ok := args["farm_id"]; ok && raw != nil {
		return "farm_id is proposal scope — omit from args"
	}
	return ""
}

func llmHasNonEmptyString(args map[string]any, key string) bool {
	raw, ok := args[key]
	if !ok || raw == nil {
		return false
	}
	s, ok := raw.(string)
	return ok && strings.TrimSpace(s) != ""
}

func llmOptionalString(args map[string]any, key string) (string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return "", nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", errors.New(key + " must be string")
	}
	return strings.TrimSpace(s), nil
}

func isLLMNFInputCategory(cat string) bool {
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

func isLLMNFApplicationTarget(target string) bool {
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

func isLLMNFBatchStatus(status string) bool {
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

func isKnownGrowthStage(stage string) bool {
	_, ok := llmGrowthStageAliases[strings.ToLower(strings.TrimSpace(stage))]
	return ok
}

func llmHasPatchField(args map[string]any, keys ...string) bool {
	for _, k := range keys {
		if v, ok := args[k]; ok && v != nil {
			return true
		}
	}
	return false
}

func llmArgInt64(args map[string]any, keys ...string) (int64, error) {
	for _, key := range keys {
		raw, ok := args[key]
		if !ok {
			continue
		}
		switch v := raw.(type) {
		case float64:
			if v <= 0 || math.IsNaN(v) {
				return 0, errors.New("invalid " + key)
			}
			return int64(v), nil
		case int64:
			if v <= 0 {
				return 0, errors.New("invalid " + key)
			}
			return v, nil
		case int:
			if v <= 0 {
				return 0, errors.New("invalid " + key)
			}
			return int64(v), nil
		default:
			return 0, errors.New("invalid " + key + " type")
		}
	}
	return 0, errors.New(keys[0] + " required")
}

func llmOptionalInt64(args map[string]any, key string) (*int64, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}
	n, err := llmArgInt64(map[string]any{key: raw}, key)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func llmArgString(args map[string]any, key string) (string, error) {
	raw, ok := args[key]
	if !ok {
		return "", errors.New(key + " required")
	}
	s, ok := raw.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "", errors.New(key + " required")
	}
	return strings.TrimSpace(s), nil
}

func llmOptionalBool(args map[string]any, key string) (*bool, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}
	b, ok := raw.(bool)
	if !ok {
		return nil, errors.New(key + " must be boolean")
	}
	return &b, nil
}

func llmOptionalFloat64(args map[string]any, key string) error {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case float64:
		if math.IsNaN(v) || v < 0 {
			return errors.New("invalid " + key)
		}
		return nil
	case int:
		if v < 0 {
			return errors.New("invalid " + key)
		}
		return nil
	default:
		return errors.New(key + " must be number")
	}
}

func llmRequireFloat64(args map[string]any, key string) error {
	raw, ok := args[key]
	if !ok || raw == nil {
		return errors.New(key + " required")
	}
	return llmOptionalFloat64(map[string]any{key: raw}, key)
}

func isISODate(s string) bool {
	_, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	return err == nil
}
