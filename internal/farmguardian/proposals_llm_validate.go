package farmguardian

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian/tools"
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
		return ""
	case "create_task_from_alert":
		if _, err := llmArgInt64(args, "alert_id"); err != nil {
			return "alert_id required"
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
	}
	return ""
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
