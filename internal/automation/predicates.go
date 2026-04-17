package automation

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

// Predicate is the canonical wire shape shared by Phase 19 schedule
// preconditions (`gr33ncore.schedules.preconditions`) and Phase 20
// automation-rule conditions (`gr33ncore.automation_rules.conditions_jsonb.predicates`).
//
// Two variants live in the same struct, discriminated by `Type`:
//
//   - Type "" or "hard" (default, back-compat) → {sensor_id, op, value}.
//     Reads the latest value from `sensor_id` and compares with an op
//     (lt|lte|eq|gte|gt|ne).
//   - Type "setpoint" (Phase 20.6) → {sensor_type, scope, op}. Resolves
//     a `gr33ncore.zone_setpoints` row via the precedence order
//     (cycle+stage > cycle-any-stage > zone+stage > zone-any-stage) and
//     checks the latest reading for that sensor_type in the rule's zone
//     against the resolved row's (min, max, ideal) using a setpoint op
//     (out_of_range | below_ideal | above_ideal | inside_range).
//
// The two variants share JSON tags so the existing write-path validator
// keeps working untouched for hard predicates. Setpoint predicates are
// only meaningful for rules (schedules have no zone context) — when a
// schedule precondition accidentally uses Type "setpoint" the evaluator
// surfaces it as a skip with Reason `no_scope_for_setpoint` rather than
// crashing.
type Predicate struct {
	Type string `json:"type,omitempty"`

	SensorID int64   `json:"sensor_id,omitempty"`
	Op       string  `json:"op"`
	Value    float64 `json:"value,omitempty"`

	SensorType string `json:"sensor_type,omitempty"`
	Scope      string `json:"scope,omitempty"`
}

// FailedPredicate captures why a single predicate didn't pass so the
// operator can understand the skip reason on the runs page. Actual is
// nil when we couldn't read a numeric value at all (reason `no_reading`
// or `reading_lookup_failed`). For setpoint predicates SensorID is 0
// and SensorType / Scope identify the predicate instead.
type FailedPredicate struct {
	Type       string   `json:"type,omitempty"`
	SensorID   int64    `json:"sensor_id,omitempty"`
	SensorType string   `json:"sensor_type,omitempty"`
	Scope      string   `json:"scope,omitempty"`
	Op         string   `json:"op"`
	Expected   float64  `json:"expected,omitempty"`
	Actual     *float64 `json:"actual,omitempty"`
	Reason     string   `json:"reason"`
}

// Logic values for rule conditions_jsonb. Schedule preconditions behave
// like LogicAll implicitly.
const (
	LogicAll = "ALL"
	LogicAny = "ANY"
)

// Setpoint-predicate values.
const (
	PredicateTypeHard     = "hard"
	PredicateTypeSetpoint = "setpoint"

	SetpointScopeCurrentStage = "current_stage"
	SetpointScopeZoneDefault  = "zone_default"

	SetpointOpOutOfRange = "out_of_range"
	SetpointOpBelowIdeal = "below_ideal"
	SetpointOpAboveIdeal = "above_ideal"
	SetpointOpInsideRange = "inside_range"

	// SkipNoSetpointForScope is the skip message surfaced when a
	// setpoint-typed predicate resolves to no matching row. This is a
	// normal state — operators just haven't configured the setpoint yet
	// — not a failure. Mirrors Phase 20.6 plan wording verbatim.
	SkipNoSetpointForScope = "no_setpoint_for_scope"
	SkipNoScopeForSetpoint = "no_scope_for_setpoint"
)

// ScopeContext threads the rule's zone / farm down to the per-predicate
// evaluator. Setpoint predicates need the zone to find the active crop
// cycle and resolve the precedence chain; hard predicates ignore it
// entirely. Passing the zero value is safe — it just means setpoint
// predicates can't resolve and the rule skips.
type ScopeContext struct {
	FarmID int64
	ZoneID *int64
}

// validSetpointOps is the closed set of ops for setpoint predicates.
// Resist growing it — the plan calls this out explicitly ("every new
// op is a new rule-engine code path"). The four here cover the 95%
// case; anything exotic should layer on top of these four with a hard
// predicate on a derived sensor.
var validSetpointOps = map[string]struct{}{
	SetpointOpOutOfRange:  {},
	SetpointOpBelowIdeal:  {},
	SetpointOpAboveIdeal:  {},
	SetpointOpInsideRange: {},
}

// evalPredicate applies a comparison op to two floats. Unknown ops
// evaluate to false — validators at the write path are responsible for
// rejecting them before they reach here.
func evalPredicate(actual float64, op string, expected float64) bool {
	switch op {
	case "lt":
		return actual < expected
	case "lte":
		return actual <= expected
	case "eq":
		return actual == expected
	case "gte":
		return actual >= expected
	case "gt":
		return actual > expected
	case "ne":
		return actual != expected
	default:
		return false
	}
}

func numericToFloat64(n pgtype.Numeric) (float64, bool) {
	if !n.Valid {
		return 0, false
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0, false
	}
	return f.Float64, true
}

// predicateType returns the effective type — missing / empty falls
// back to "hard" so existing stored rules keep working without a
// migration.
func (p Predicate) predicateType() string {
	if p.Type == "" {
		return PredicateTypeHard
	}
	return p.Type
}

// evaluatePredicate fetches the latest reading for the predicate's
// sensor and decides pass/fail. Missing readings always count as
// failure — if the operator asked for an interlock, we can't assert
// safety (or "conditions met") without data.
func evaluatePredicate(ctx context.Context, q *db.Queries, p Predicate) (passed bool, fp FailedPredicate) {
	reading, err := q.GetLatestReadingBySensor(ctx, p.SensorID)
	if err != nil {
		reason := "reading_lookup_failed"
		if errors.Is(err, pgx.ErrNoRows) {
			reason = "no_reading"
		}
		return false, FailedPredicate{Type: PredicateTypeHard, SensorID: p.SensorID, Op: p.Op, Expected: p.Value, Reason: reason}
	}
	actual, ok := numericToFloat64(reading.ValueRaw)
	if !ok {
		return false, FailedPredicate{Type: PredicateTypeHard, SensorID: p.SensorID, Op: p.Op, Expected: p.Value, Reason: "no_reading"}
	}
	if !evalPredicate(actual, p.Op, p.Value) {
		a := actual
		return false, FailedPredicate{Type: PredicateTypeHard, SensorID: p.SensorID, Op: p.Op, Expected: p.Value, Actual: &a, Reason: "predicate_failed"}
	}
	return true, FailedPredicate{}
}

// evalSetpointOp applies a setpoint op to (reading, min, max, ideal).
// Any NULL bound short-circuits to `true` on that side of the comparison
// (e.g. out_of_range when max is NULL is only a high-side violation,
// never a low-side one). `okMin`/`okMax`/`okIdeal` flag whether each
// bound was provided. The function returns (passed, wasInconclusive).
// Inconclusive happens when a required bound is missing — e.g. asking
// below_ideal on a setpoint row with no ideal_value.
func evalSetpointOp(op string, reading, min, max, ideal float64, okMin, okMax, okIdeal bool) (passed bool, inconclusive bool) {
	switch op {
	case SetpointOpOutOfRange:
		low := okMin && reading < min
		high := okMax && reading > max
		if !okMin && !okMax {
			return false, true
		}
		return low || high, false
	case SetpointOpInsideRange:
		if !okMin && !okMax {
			return true, true
		}
		low := !okMin || reading >= min
		high := !okMax || reading <= max
		return low && high, false
	case SetpointOpBelowIdeal:
		if !okIdeal {
			return false, true
		}
		return reading < ideal, false
	case SetpointOpAboveIdeal:
		if !okIdeal {
			return false, true
		}
		return reading > ideal, false
	default:
		return false, false
	}
}

// evaluateSetpointPredicate resolves a setpoint row for the predicate's
// (zone, crop_cycle, stage, sensor_type), fetches the latest reading for
// that sensor_type in the zone, and applies the setpoint op. Any of the
// usual failure modes (no scope, no setpoint row, no reading, missing
// bound) short-circuits to a `FailedPredicate` with a specific Reason
// so operators can tell them apart on the runs page.
func evaluateSetpointPredicate(ctx context.Context, q *db.Queries, scope ScopeContext, p Predicate) (passed bool, fp FailedPredicate) {
	base := FailedPredicate{Type: PredicateTypeSetpoint, SensorType: p.SensorType, Scope: p.Scope, Op: p.Op}
	if scope.ZoneID == nil {
		base.Reason = SkipNoScopeForSetpoint
		return false, base
	}
	if _, ok := validSetpointOps[p.Op]; !ok {
		base.Reason = "invalid_op"
		return false, base
	}
	if p.SensorType == "" {
		base.Reason = "sensor_type_required"
		return false, base
	}

	var cycleID *int64
	var stage *string
	if p.Scope != SetpointScopeZoneDefault {
		cycle, err := q.GetActiveCropCycleForZone(ctx, *scope.ZoneID)
		if err == nil {
			id := cycle.ID
			cycleID = &id
			if cycle.CurrentStage.Valid {
				s := string(cycle.CurrentStage.Gr33nfertigationGrowthStageEnum)
				stage = &s
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			base.Reason = "cycle_lookup_failed"
			return false, base
		}
	}

	row, err := q.GetActiveSetpointForScope(ctx, db.GetActiveSetpointForScopeParams{
		SensorType:  p.SensorType,
		CropCycleID: cycleID,
		Stage:       stage,
		ZoneID:      scope.ZoneID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			base.Reason = SkipNoSetpointForScope
			return false, base
		}
		base.Reason = "setpoint_lookup_failed"
		return false, base
	}

	reading, err := q.GetLatestReadingForZoneSensorType(ctx, db.GetLatestReadingForZoneSensorTypeParams{
		ZoneID:     scope.ZoneID,
		SensorType: p.SensorType,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			base.Reason = "no_reading"
			return false, base
		}
		base.Reason = "reading_lookup_failed"
		return false, base
	}
	actualF, ok := numericToFloat64(reading.ValueRaw)
	if !ok {
		base.Reason = "no_reading"
		return false, base
	}

	minV, okMin := numericToFloat64(row.MinValue)
	maxV, okMax := numericToFloat64(row.MaxValue)
	idealV, okIdeal := numericToFloat64(row.IdealValue)

	passed, inconclusive := evalSetpointOp(p.Op, actualF, minV, maxV, idealV, okMin, okMax, okIdeal)
	if inconclusive {
		base.Reason = "setpoint_missing_bound"
		base.Actual = &actualF
		return false, base
	}
	if !passed {
		a := actualF
		base.Reason = "predicate_failed"
		base.Actual = &a
		return false, base
	}
	return true, FailedPredicate{}
}

// EvaluatePredicates is the back-compat entry point used by schedule
// preconditions, which have no zone context. Equivalent to
// EvaluatePredicatesInScope with a zero ScopeContext; setpoint-typed
// predicates passed to this variant will always skip with
// `no_scope_for_setpoint`.
func EvaluatePredicates(ctx context.Context, q *db.Queries, logic string, preds []Predicate) (passed bool, failed []FailedPredicate) {
	passed, failed, _ = EvaluatePredicatesInScope(ctx, q, logic, preds, ScopeContext{})
	return
}

// EvaluatePredicatesInScope runs all predicates and returns an overall
// decision plus the list of failures AND an optional `skipMessage` the
// caller should use when recording the run. Semantics:
//
//   - logic == "ALL" (or empty/unknown): every predicate must pass.
//   - logic == "ANY": at least one predicate must pass.
//   - Empty predicate list → (true, nil, "") — empty ALL is vacuously
//     satisfied and the rule fires on trigger with no gating.
//   - `skipMessage` is set to the first setpoint-specific skip reason
//     encountered (`no_setpoint_for_scope` / `no_scope_for_setpoint`)
//     so the caller can record a rule skip with that exact message
//     instead of the generic `conditions_not_met`. It's only set when
//     the overall decision is "not passed".
func EvaluatePredicatesInScope(ctx context.Context, q *db.Queries, logic string, preds []Predicate, scope ScopeContext) (passed bool, failed []FailedPredicate, skipMessage string) {
	if len(preds) == 0 {
		return true, nil, ""
	}
	failed = make([]FailedPredicate, 0, len(preds))
	anyPassed := false
	for _, p := range preds {
		var ok bool
		var fp FailedPredicate
		if p.predicateType() == PredicateTypeSetpoint {
			ok, fp = evaluateSetpointPredicate(ctx, q, scope, p)
		} else {
			ok, fp = evaluatePredicate(ctx, q, p)
		}
		if ok {
			anyPassed = true
			continue
		}
		failed = append(failed, fp)
		if skipMessage == "" && (fp.Reason == SkipNoSetpointForScope || fp.Reason == SkipNoScopeForSetpoint) {
			skipMessage = fp.Reason
		}
	}
	switch logic {
	case LogicAny:
		if anyPassed {
			return true, failed, ""
		}
		return false, failed, skipMessage
	default:
		if len(failed) == 0 {
			return true, nil, ""
		}
		return false, failed, skipMessage
	}
}
