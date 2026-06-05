package farmguardian

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	db "gr33n-api/internal/db"
)

var (
	ruleIDPattern = regexp.MustCompile(`(?i)rule\s*#\s*(\d+)`)

	disableRuleIntent = regexp.MustCompile(`(?i)\b(?:disable|turn\s+off|pause|stop)\b.*\b(?:rule|automation|shade|vent|greenhouse)\b|\b(?:shade|vent)\b.*\b(?:rule|automation)\b`)
	enableRuleIntent  = regexp.MustCompile(`(?i)\b(?:enable|turn\s+on|resume|start)\b.*\b(?:rule|automation|shade|vent)\b`)

	pauseComfortScheduleIntent = regexp.MustCompile(`(?i)\b(?:pause|disable|stop|turn\s+off)\b.*\b(?:lights?|lighting)\b`)
	enableComfortScheduleIntent = regexp.MustCompile(`(?i)\b(?:enable|resume|turn\s+on|start)\b.*\b(?:lights?|lighting)\b.*\bschedule\b`)

	ecTargetIntent = regexp.MustCompile(`(?i)(?:set|change|update).*(?:ec|conductivity).*(?:to\s+)(\d+(?:\.\d+)?)`)
)

// matchComfortAutomationIntent proposes patch_rule, patch_schedule, or
// patch_fertigation_program for plain-language comfort/automation edits (Phase 42 WS8).
// Feeding-specific volume/irrigation matchers live in matchFeedingProgramIntent; this
// pass covers automation rules, lighting schedules, and EC target tweaks.
func matchComfortAutomationIntent(
	ctx context.Context,
	querier comfortMatchQuerier,
	farmID int64,
	question string,
	snap Snapshot,
) (toolID string, args map[string]any, summary string, ok bool) {
	if querier == nil || farmID <= 0 {
		return "", nil, "", false
	}
	q := strings.TrimSpace(question)
	if q == "" {
		return "", nil, "", false
	}

	if disableRuleIntent.MatchString(q) {
		if rule, okRule := resolveRuleForIntent(ctx, querier, farmID, q, snap); okRule {
			return "patch_rule", map[string]any{
				"rule_id":   rule.ID,
				"is_active": false,
				"rule_name": rule.Name,
			}, "Disable automation rule \"" + rule.Name + "\"", true
		}
	}

	if enableRuleIntent.MatchString(q) {
		if rule, okRule := resolveRuleForIntent(ctx, querier, farmID, q, snap); okRule {
			return "patch_rule", map[string]any{
				"rule_id":   rule.ID,
				"is_active": true,
				"rule_name": rule.Name,
			}, "Enable automation rule \"" + rule.Name + "\"", true
		}
	}

	if pauseComfortScheduleIntent.MatchString(q) {
		if sch, okSch := resolveScheduleForIntent(ctx, querier, farmID, q, snap); okSch {
			return "patch_schedule", map[string]any{
				"schedule_id":   sch.ID,
				"is_active":     false,
				"schedule_name": sch.Name,
			}, "Pause schedule \"" + sch.Name + "\" — no automatic runs until re-enabled", true
		}
	}

	if enableComfortScheduleIntent.MatchString(q) {
		if sch, okSch := resolveScheduleForIntent(ctx, querier, farmID, q, snap); okSch {
			return "patch_schedule", map[string]any{
				"schedule_id":   sch.ID,
				"is_active":     true,
				"schedule_name": sch.Name,
			}, "Enable schedule \"" + sch.Name + "\"", true
		}
	}

	if m := ecTargetIntent.FindStringSubmatch(q); len(m) > 1 {
		if ec, err := strconv.ParseFloat(m[1], 64); err == nil {
			if prog, okProg := resolveActiveProgramForIntent(ctx, querier, q, farmID, snap); okProg {
				return "patch_fertigation_program", map[string]any{
					"program_id":      prog.ID,
					"ec_trigger_low":  ec,
					"program_name":    prog.Name,
				}, "Set feeding plan \"" + prog.Name + "\" EC target to " + formatEC(ec), true
			}
		}
	}

	return "", nil, "", false
}

func resolveRuleForIntent(
	ctx context.Context,
	querier comfortMatchQuerier,
	farmID int64,
	question string,
	snap Snapshot,
) (db.Gr33ncoreAutomationRule, bool) {
	rules, err := querier.ListAutomationRulesByFarm(ctx, farmID)
	if err != nil || len(rules) == 0 {
		return db.Gr33ncoreAutomationRule{}, false
	}

	zoneID := resolveZoneIDForIntent(ctx, querier, question, farmID, snap)
	sensors, _ := querier.ListSensorsByFarm(ctx, farmID)
	return pickRuleForIntent(rules, question, zoneID, sensors)
}

func pickRuleForIntent(
	rules []db.Gr33ncoreAutomationRule,
	question string,
	zoneID int64,
	sensors []db.Gr33ncoreSensor,
) (db.Gr33ncoreAutomationRule, bool) {
	if len(rules) == 0 {
		return db.Gr33ncoreAutomationRule{}, false
	}

	if m := ruleIDPattern.FindStringSubmatch(question); len(m) > 1 {
		if id, err := strconv.ParseInt(m[1], 10, 64); err == nil && id > 0 {
			for _, r := range rules {
				if r.ID == id {
					return r, true
				}
			}
		}
	}

	lower := strings.ToLower(question)
	var zoneScoped []db.Gr33ncoreAutomationRule
	for _, r := range rules {
		if zoneID > 0 && !ruleAppliesToZoneForIntent(r, zoneID, sensors) {
			continue
		}
		zoneScoped = append(zoneScoped, r)
	}
	candidates := zoneScoped
	if len(candidates) == 0 {
		candidates = rules
	}

	// Prefer name substring matches (shade, vent, greenhouse keywords).
	keywords := ruleKeywordsFromQuestion(lower)
	for _, r := range candidates {
		nameLower := strings.ToLower(r.Name)
		for _, kw := range keywords {
			if kw != "" && strings.Contains(nameLower, kw) {
				return r, true
			}
		}
		if strings.Contains(lower, nameLower) && nameLower != "" {
			return r, true
		}
	}

	// Keyword in question but not rule name — pick first active GH-ish rule in zone.
	if len(keywords) > 0 {
		for _, r := range candidates {
			if !r.IsActive {
				continue
			}
			nameLower := strings.ToLower(r.Name)
			if strings.Contains(nameLower, "gh") || strings.Contains(nameLower, "shade") || strings.Contains(nameLower, "vent") {
				return r, true
			}
		}
	}

	if len(candidates) == 1 {
		return candidates[0], true
	}

	// Last resort: first active rule in scope.
	for _, r := range candidates {
		if r.IsActive {
			return r, true
		}
	}
	return db.Gr33ncoreAutomationRule{}, false
}

func ruleKeywordsFromQuestion(lower string) []string {
	var out []string
	for _, kw := range []string{"shade", "vent", "greenhouse", "gh", "humidity", "temperature", "cooling"} {
		if strings.Contains(lower, kw) {
			out = append(out, kw)
		}
	}
	return out
}

func ruleAppliesToZoneForIntent(rule db.Gr33ncoreAutomationRule, zoneID int64, sensors []db.Gr33ncoreSensor) bool {
	if zid := ruleZoneIDFromConfig(rule.TriggerConfiguration); zid != nil && *zid == zoneID {
		return true
	}
	zoneSensorIDs := make(map[int64]struct{})
	for _, s := range sensors {
		if s.ZoneID != nil && *s.ZoneID == zoneID {
			zoneSensorIDs[s.ID] = struct{}{}
		}
	}
	preds := rulePredicatesFromJSON(rule.ConditionsJsonb)
	for _, p := range preds {
		if sid, ok := p["sensor_id"].(float64); ok {
			if _, ok := zoneSensorIDs[int64(sid)]; ok {
				return true
			}
		}
	}
	return zoneID == 0
}

func ruleZoneIDFromConfig(raw json.RawMessage) *int64 {
	if len(raw) == 0 {
		return nil
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil
	}
	rawID, ok := cfg["zone_id"]
	if !ok {
		return nil
	}
	switch v := rawID.(type) {
	case float64:
		id := int64(v)
		if id > 0 {
			return &id
		}
	case string:
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return &n
		}
	}
	return nil
}

func rulePredicatesFromJSON(raw json.RawMessage) []map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var wrapper struct {
		Predicates []map[string]any `json:"predicates"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil
	}
	return wrapper.Predicates
}
