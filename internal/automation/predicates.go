package automation

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

// Predicate is the canonical {sensor_id, op, value} shape shared by
// Phase 19 schedule preconditions (`gr33ncore.schedules.preconditions`)
// and Phase 20 automation-rule conditions
// (`gr33ncore.automation_rules.conditions_jsonb.predicates`). Keeping the
// JSON tags identical means the two evaluators, the two write-path
// validators, and the UI component all speak the same wire shape.
type Predicate struct {
	SensorID int64   `json:"sensor_id"`
	Op       string  `json:"op"`
	Value    float64 `json:"value"`
}

// FailedPredicate captures why a single predicate didn't pass so the
// operator can understand the skip reason on the runs page. Actual is
// nil when we couldn't read a numeric value at all (`reason=no_reading`
// or `reading_lookup_failed`).
type FailedPredicate struct {
	SensorID int64    `json:"sensor_id"`
	Op       string   `json:"op"`
	Expected float64  `json:"expected"`
	Actual   *float64 `json:"actual,omitempty"`
	Reason   string   `json:"reason"`
}

// Logic values for rule conditions_jsonb. Schedule preconditions behave
// like LogicAll implicitly.
const (
	LogicAll = "ALL"
	LogicAny = "ANY"
)

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
		return false, FailedPredicate{SensorID: p.SensorID, Op: p.Op, Expected: p.Value, Reason: reason}
	}
	actual, ok := numericToFloat64(reading.ValueRaw)
	if !ok {
		return false, FailedPredicate{SensorID: p.SensorID, Op: p.Op, Expected: p.Value, Reason: "no_reading"}
	}
	if !evalPredicate(actual, p.Op, p.Value) {
		a := actual
		return false, FailedPredicate{SensorID: p.SensorID, Op: p.Op, Expected: p.Value, Actual: &a, Reason: "predicate_failed"}
	}
	return true, FailedPredicate{}
}

// EvaluatePredicates runs all predicates and returns an overall decision
// plus the list of failures. Semantics:
//   - logic == "ALL" (or empty/unknown): every predicate must pass; the
//     returned `failed` slice contains every predicate that didn't.
//   - logic == "ANY": at least one predicate must pass; the returned
//     `failed` slice contains every non-passing predicate (useful for
//     "nothing passed, here's why" messages in the runs page).
//
// Returns (true, nil) for an empty predicate list — an empty ALL is
// vacuously satisfied; the rule evaluator treats "no predicates" as
// "fire on trigger, no gating" in a later phase. For WS2 rules without
// predicates still fire.
func EvaluatePredicates(ctx context.Context, q *db.Queries, logic string, preds []Predicate) (passed bool, failed []FailedPredicate) {
	if len(preds) == 0 {
		return true, nil
	}
	failed = make([]FailedPredicate, 0, len(preds))
	anyPassed := false
	for _, p := range preds {
		ok, fp := evaluatePredicate(ctx, q, p)
		if ok {
			anyPassed = true
			continue
		}
		failed = append(failed, fp)
	}
	switch logic {
	case LogicAny:
		return anyPassed, failed
	default:
		return len(failed) == 0, failed
	}
}
