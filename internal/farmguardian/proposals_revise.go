package farmguardian

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian/tools"
)

var (
	reviseVolumePattern  = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(?:l\b|liters?\b|litres?\b)`)
	reviseECPattern      = regexp.MustCompile(`(?i)\bec\s*(?:target|of|to|=|:)?\s*(\d+(?:\.\d+)?)`)
	revisePHRangePattern = regexp.MustCompile(`(?i)\bph\s*(?:of|to|=|:)?\s*(\d(?:\.\d+)?)\s*(?:-|–|to)\s*(\d(?:\.\d+)?)`)
	reviseRHPattern      = regexp.MustCompile(`(?i)(?:rh|humidity)[^\d%]{0,24}?(\d{1,3})\s*%?`)
)

// tryReviseActiveProposal revises the live draft in a session when the turn reads
// as a correction or supplies an unsensed fact (Phase 34 WS2/WS3). It returns
// handled=true when it owns the turn (so the caller must not build a fresh
// proposal), and the revised proposal slice (which may be empty when the chain has
// hit its revision cap and the operator should start over).
func tryReviseActiveProposal(
	ctx context.Context,
	q db.Querier,
	userID uuid.UUID,
	sessionID uuid.UUID,
	question string,
	snap Snapshot,
) ([]ActionProposal, bool, error) {
	prior, err := q.GetLatestPendingProposalBySession(ctx, db.GetLatestPendingProposalBySessionParams{
		UserID:    userID,
		SessionID: pgtype.UUID{Bytes: sessionID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var priorArgs map[string]any
	if len(prior.Args) > 0 {
		_ = json.Unmarshal(prior.Args, &priorArgs)
	}
	if priorArgs == nil {
		priorArgs = map[string]any{}
	}

	newArgs, changed := applyRevisionDeltas(prior.ToolID, priorArgs, question)
	facts := extractOperatorFacts(question)

	// Nothing actionable in this turn — let normal matching / the model answer it.
	if !changed && len(facts) == 0 {
		return nil, false, nil
	}

	// Bounded chains: stop superseding once the chain is deep; the operator should
	// start a fresh request rather than grow this one unbounded.
	if int(prior.Revision) >= MaxProposalRevisions {
		return []ActionProposal{}, true, nil
	}

	mergedFacts := mergeOperatorFacts(priorMetaFacts(prior.Meta), facts)

	if _, err := q.SupersedeProposal(ctx, db.SupersedeProposalParams{
		ProposalID: prior.ProposalID,
		UserID:     userID,
	}); err != nil {
		return nil, false, err
	}

	summary := prior.Summary
	if prior.ToolID == "apply_grow_setup_pack" {
		summary = tools.GrowSetupPackSummary(newArgs)
	}

	row, err := insertProposal(ctx, q, insertProposalInput{
		userID:     userID,
		farmID:     prior.FarmID,
		sessionID:  sessionID,
		toolID:     prior.ToolID,
		args:       newArgs,
		summary:    summary,
		revision:   prior.Revision + 1,
		supersedes: prior.ProposalID,
		facts:      mergedFacts,
	})
	if err != nil {
		return nil, false, err
	}
	return []ActionProposal{ActionProposalFromRow(row)}, true, nil
}

// applyRevisionDeltas returns a copy of priorArgs with the corrections from the
// turn applied, scoped to fields the given tool actually accepts. changed is true
// only when at least one field was rewritten.
func applyRevisionDeltas(toolID string, priorArgs map[string]any, question string) (map[string]any, bool) {
	next := deepCopyArgs(priorArgs)
	changed := false

	switch toolID {
	case "apply_grow_setup_pack":
		program := childMap(next, "program")
		cycle := childMap(next, "cycle")
		if v, ok := parseVolume(question); ok && program != nil {
			program["total_volume_liters"] = v
			changed = true
		}
		if v, ok := parseEC(question); ok && program != nil {
			program["ec_trigger_low"] = v
			changed = true
		}
		if lo, hi, ok := parsePHRange(question); ok && program != nil {
			program["ph_trigger_low"] = lo
			program["ph_trigger_high"] = hi
			changed = true
		}
		if stage := parseStage(question); stage != "" && cycle != nil {
			cycle["current_stage"] = stage
			changed = true
		}
	case "create_fertigation_program":
		if v, ok := parseVolume(question); ok {
			next["total_volume_liters"] = v
			changed = true
		}
		if v, ok := parseEC(question); ok {
			next["ec_trigger_low"] = v
			changed = true
		}
		if lo, hi, ok := parsePHRange(question); ok {
			next["ph_trigger_low"] = lo
			next["ph_trigger_high"] = hi
			changed = true
		}
	case "patch_fertigation_program":
		if v, ok := parseVolume(question); ok {
			next["total_volume_liters"] = v
			changed = true
		}
	case "create_crop_cycle", "update_cycle_stage":
		if stage := parseStage(question); stage != "" {
			next["current_stage"] = stage
			changed = true
		}
	}

	if !changed {
		return priorArgs, false
	}
	return next, true
}

// extractOperatorFacts pulls unsensed ground-truth the operator asserts, labeled
// operator-stated and never merged into args as a measurement (Phase 34 WS3).
func extractOperatorFacts(question string) []OperatorFact {
	lower := strings.ToLower(question)
	facts := []OperatorFact{}

	if hasOperatorAssertionCue(lower) {
		if m := reviseRHPattern.FindStringSubmatch(question); len(m) > 1 {
			if pct, err := strconv.Atoi(m[1]); err == nil && pct >= 0 && pct <= 100 {
				facts = append(facts, OperatorFact{
					Field: "rh_pct",
					Value: pct,
					Basis: "operator_stated",
					Label: "RH " + strconv.Itoa(pct) + "% (operator-stated, not measured)",
				})
			}
		}
		if src := parseWaterSource(lower); src != "" {
			facts = append(facts, OperatorFact{
				Field: "water_source",
				Value: src,
				Basis: "operator_stated",
				Label: "water source " + src + " (operator-stated, not measured)",
			})
		}
	}
	return facts
}

func hasOperatorAssertionCue(lower string) bool {
	cues := []string{"assume", "around", "approx", "no ", "there's no", "theres no",
		"~", "call it", "stated", "water source", "well water", "ro water", "tap water"}
	for _, c := range cues {
		if strings.Contains(lower, c) {
			return true
		}
	}
	return false
}

func parseWaterSource(lower string) string {
	switch {
	case strings.Contains(lower, "well water"):
		return "well"
	case strings.Contains(lower, "ro water"), strings.Contains(lower, "reverse osmosis"):
		return "ro"
	case strings.Contains(lower, "tap water"):
		return "tap"
	case strings.Contains(lower, "rain water"), strings.Contains(lower, "rainwater"):
		return "rain"
	default:
		return ""
	}
}

func parseVolume(question string) (float64, bool) {
	if m := reviseVolumePattern.FindStringSubmatch(question); len(m) > 1 {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil && v > 0 {
			return v, true
		}
	}
	return 0, false
}

func parseEC(question string) (float64, bool) {
	if m := reviseECPattern.FindStringSubmatch(question); len(m) > 1 {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil && v > 0 && v < 10 {
			return v, true
		}
	}
	return 0, false
}

func parsePHRange(question string) (float64, float64, bool) {
	if m := revisePHRangePattern.FindStringSubmatch(question); len(m) > 2 {
		lo, err1 := strconv.ParseFloat(m[1], 64)
		hi, err2 := strconv.ParseFloat(m[2], 64)
		if err1 == nil && err2 == nil && lo > 0 && hi >= lo && hi < 14 {
			return lo, hi, true
		}
	}
	return 0, 0, false
}

func parseStage(question string) string {
	lower := strings.ToLower(question)
	return inferStageKeyword(lower)
}

func priorMetaFacts(meta []byte) []OperatorFact {
	if len(meta) == 0 {
		return nil
	}
	var m proposalMeta
	if err := json.Unmarshal(meta, &m); err != nil {
		return nil
	}
	return m.OperatorProvided
}

// mergeOperatorFacts combines prior facts with new ones, with later assertions for
// the same field overriding earlier ones.
func mergeOperatorFacts(prior, next []OperatorFact) []OperatorFact {
	if len(prior) == 0 {
		return next
	}
	byField := map[string]OperatorFact{}
	order := []string{}
	for _, f := range prior {
		if _, seen := byField[f.Field]; !seen {
			order = append(order, f.Field)
		}
		byField[f.Field] = f
	}
	for _, f := range next {
		if _, seen := byField[f.Field]; !seen {
			order = append(order, f.Field)
		}
		byField[f.Field] = f
	}
	out := make([]OperatorFact, 0, len(order))
	for _, field := range order {
		out = append(out, byField[field])
	}
	return out
}

func deepCopyArgs(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		if child, ok := v.(map[string]any); ok {
			out[k] = deepCopyArgs(child)
		} else {
			out[k] = v
		}
	}
	return out
}

func childMap(parent map[string]any, key string) map[string]any {
	if parent == nil {
		return nil
	}
	if child, ok := parent[key].(map[string]any); ok {
		return child
	}
	return nil
}
