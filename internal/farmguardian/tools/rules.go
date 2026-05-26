package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

type ruleConditionsPatch struct {
	Logic      string          `json:"logic"`
	Predicates []rulePredicate `json:"predicates"`
}

type rulePredicate struct {
	Type       string  `json:"type,omitempty"`
	SensorID   int64   `json:"sensor_id,omitempty"`
	Op         string  `json:"op"`
	Value      float64 `json:"value,omitempty"`
	SensorType string  `json:"sensor_type,omitempty"`
	Scope      string  `json:"scope,omitempty"`
}

func execPatchRule(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	ruleID, err := int64FromArgs(args, "rule_id")
	if err != nil {
		return nil, err
	}
	rule, err := deps.Q.GetAutomationRuleByID(ctx, ruleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("rule %d not found", ruleID)
		}
		return nil, err
	}
	if err := ensureFarmScope(rule.FarmID, deps.FarmID); err != nil {
		return nil, err
	}

	active, err := optionalBoolFromArgs(args, "is_active")
	if err != nil {
		return nil, err
	}
	threshold, err := optionalFloat64FromArgs(args, "threshold")
	if err != nil {
		return nil, err
	}
	if active == nil && threshold == nil {
		return nil, errors.New("is_active or threshold required")
	}

	out := map[string]any{"rule_id": ruleID}

	if active != nil && threshold == nil {
		row, err := deps.Q.UpdateAutomationRuleActive(ctx, db.UpdateAutomationRuleActiveParams{
			ID:       ruleID,
			IsActive: *active,
		})
		if err != nil {
			return nil, err
		}
		out["is_active"] = row.IsActive
		return out, nil
	}

	isActive := rule.IsActive
	if active != nil {
		isActive = *active
	}
	conds := rule.ConditionsJsonb
	if threshold != nil {
		patched, err := patchRuleThreshold(conds, *threshold)
		if err != nil {
			return nil, err
		}
		conds = patched
	}
	logic := rule.ConditionLogic
	row, err := deps.Q.UpdateAutomationRule(ctx, db.UpdateAutomationRuleParams{
		ID:                    ruleID,
		Name:                  rule.Name,
		Description:           rule.Description,
		IsActive:              isActive,
		TriggerSource:         rule.TriggerSource,
		TriggerConfiguration:  rule.TriggerConfiguration,
		ConditionLogic:        logic,
		ConditionsJsonb:       conds,
		CooldownPeriodSeconds: rule.CooldownPeriodSeconds,
	})
	if err != nil {
		return nil, err
	}
	out["is_active"] = row.IsActive
	if threshold != nil {
		out["threshold"] = *threshold
	}
	return out, nil
}

func patchRuleThreshold(conds []byte, threshold float64) ([]byte, error) {
	if len(conds) == 0 {
		return nil, errors.New("rule has no conditions to patch")
	}
	var rc ruleConditionsPatch
	if err := json.Unmarshal(conds, &rc); err != nil {
		return nil, errors.New("invalid conditions_jsonb on rule")
	}
	if len(rc.Predicates) == 0 {
		return nil, errors.New("rule has no predicates to patch")
	}
	updated := false
	for i := range rc.Predicates {
		pt := rc.Predicates[i].Type
		if pt == "" || pt == "hard" {
			rc.Predicates[i].Value = threshold
			updated = true
			break
		}
	}
	if !updated {
		return nil, errors.New("no hard threshold predicate found on rule")
	}
	return json.Marshal(rc)
}
