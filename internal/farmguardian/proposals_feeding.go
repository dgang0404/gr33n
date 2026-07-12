package farmguardian

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	db "gr33n-api/internal/db"
)

var (
	feedVolumeIntent = regexp.MustCompile(`(?i)(?:set|change|update|adjust)\s+(?:the\s+)?(?:feed(?:ing)?\s+)?volume\s+(?:to\s+)?(\d+(?:\.\d+)?)\s*l`)
	feedVolumeAlt    = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*l\s+(?:per\s+)?(?:feed|run|watering)`)
	irrigationOnlyIntent = regexp.MustCompile(`(?i)(?:switch|change|set).*(?:plain\s+)?(?:water[\s-]*only|irrigation[\s-]*only|plain\s+water(?:[\s-]*only)?)`)
	pauseFeedingIntent   = regexp.MustCompile(`(?i)(?:pause|disable|stop|turn\s+off)\s+(?:the\s+)?(?:feed(?:ing)?|watering|irrigation)`)
	pauseScheduleIntent  = regexp.MustCompile(`(?i)(?:pause|disable|stop|turn\s+off).*\bschedule\b`)
	enableScheduleIntent = regexp.MustCompile(`(?i)(?:enable|resume|turn\s+on|start)\s+(?:the\s+)?schedule`)
)

// matchFeedingProgramIntent proposes patch_fertigation_program or patch_schedule for
// plain-language feeding edits (Phase 47 WS6; extends Phase 42 §3.4).
func matchFeedingProgramIntent(
	ctx context.Context,
	querier db.Querier,
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
	lowerQ := strings.ToLower(q)

	if pauseScheduleIntent.MatchString(q) || (pauseFeedingIntent.MatchString(q) && strings.Contains(lowerQ, "schedule")) {
		if sch, okSch := resolveScheduleForIntent(ctx, querier, farmID, q, snap); okSch {
			return "patch_schedule", map[string]any{
				"schedule_id": sch.ID,
				"is_active":   false,
			}, "Pause schedule \"" + sch.Name + "\" — no automatic runs until re-enabled", true
		}
	}

	if enableScheduleIntent.MatchString(q) {
		if sch, okSch := resolveScheduleForIntent(ctx, querier, farmID, q, snap); okSch {
			return "patch_schedule", map[string]any{
				"schedule_id": sch.ID,
				"is_active":   true,
			}, "Enable schedule \"" + sch.Name + "\"", true
		}
	}

	if irrigationOnlyIntent.MatchString(q) {
		if prog, okProg := resolveActiveProgramForIntent(ctx, querier, q, farmID, snap); okProg {
			return "patch_fertigation_program", map[string]any{
				"program_id":      prog.ID,
				"irrigation_only": true,
			}, "Switch feeding plan \"" + prog.Name + "\" to water-only irrigation (no nutrient mix)", true
		}
	}

	if m := feedVolumeIntent.FindStringSubmatch(q); len(m) > 1 {
		if vol, err := strconv.ParseFloat(m[1], 64); err == nil {
			if prog, okProg := resolveActiveProgramForIntent(ctx, querier, q, farmID, snap); okProg {
				return "patch_fertigation_program", map[string]any{
					"program_id":          prog.ID,
					"total_volume_liters": vol,
				}, "Set feeding plan \"" + prog.Name + "\" volume to " + formatLiters(vol), true
			}
		}
	}
	if m := feedVolumeAlt.FindStringSubmatch(q); len(m) > 1 {
		if vol, err := strconv.ParseFloat(m[1], 64); err == nil {
			if prog, okProg := resolveActiveProgramForIntent(ctx, querier, q, farmID, snap); okProg {
				return "patch_fertigation_program", map[string]any{
					"program_id":          prog.ID,
					"total_volume_liters": vol,
				}, "Set feeding plan \"" + prog.Name + "\" volume to " + formatLiters(vol), true
			}
		}
	}

	if pauseFeedingIntent.MatchString(q) {
		if prog, okProg := resolveActiveProgramForIntent(ctx, querier, q, farmID, snap); okProg {
			return "patch_fertigation_program", map[string]any{
				"program_id": prog.ID,
				"is_active":  false,
			}, "Pause feeding plan \"" + prog.Name + "\" — does not run the program now", true
		}
	}

	return "", nil, "", false
}

func resolveActiveProgramForIntent(ctx context.Context, querier farmMatchQuerier, question string, farmID int64, snap Snapshot) (db.Gr33nfertigationProgram, bool) {
	zoneID := resolveZoneIDForIntent(ctx, querier, question, farmID, snap)
	programs, err := querier.ListProgramsByFarm(ctx, farmID)
	if err != nil || len(programs) == 0 {
		return db.Gr33nfertigationProgram{}, false
	}
	var active []db.Gr33nfertigationProgram
	for _, p := range programs {
		if zoneID > 0 && p.TargetZoneID != nil && *p.TargetZoneID != zoneID {
			continue
		}
		if p.IsActive {
			active = append(active, p)
		}
	}
	if len(active) == 0 {
		for _, p := range programs {
			if zoneID > 0 && p.TargetZoneID != nil && *p.TargetZoneID != zoneID {
				continue
			}
			active = append(active, p)
		}
	}
	if len(active) == 0 {
		return db.Gr33nfertigationProgram{}, false
	}
	return active[0], true
}

func resolveScheduleForIntent(ctx context.Context, querier farmMatchQuerier, farmID int64, question string, snap Snapshot) (db.Gr33ncoreSchedule, bool) {
	schedules, err := querier.ListSchedulesByFarm(ctx, farmID)
	if err != nil || len(schedules) == 0 {
		return db.Gr33ncoreSchedule{}, false
	}
	lower := strings.ToLower(question)
	zoneID := resolveZoneIDForIntent(ctx, querier, question, farmID, snap)
	zoneName := resolveZoneNameForIntent(ctx, querier, question, farmID, snap)
	lightingIntent := strings.Contains(lower, "light")

	for _, s := range schedules {
		if s.Name != "" && strings.Contains(lower, strings.ToLower(s.Name)) {
			return s, true
		}
	}

	if lightingIntent && zoneID > 0 {
		if sch, ok := resolveLightingScheduleForZone(ctx, querier, schedules, farmID, zoneID); ok {
			return sch, true
		}
	}

	if zoneName != "" || zoneID > 0 {
		var candidates []db.Gr33ncoreSchedule
		for _, s := range schedules {
			if lightingIntent && s.ScheduleType != "lighting" {
				continue
			}
			if zoneName != "" && !scheduleDescribesZone(s, zoneName) {
				continue
			}
			candidates = append(candidates, s)
		}
		if sch, ok := pickScheduleForIntent(candidates, lightingIntent, lower); ok {
			return sch, true
		}
	}

	if lightingIntent {
		var lighting []db.Gr33ncoreSchedule
		for _, s := range schedules {
			if s.ScheduleType == "lighting" {
				lighting = append(lighting, s)
			}
		}
		if sch, ok := pickScheduleForIntent(lighting, true, lower); ok {
			return sch, true
		}
	}

	if len(schedules) == 1 {
		return schedules[0], true
	}
	return db.Gr33ncoreSchedule{}, false
}

type lightingProgramQuerier interface {
	ListLightingProgramsByFarm(ctx context.Context, farmID int64) ([]db.Gr33ncoreLightingProgram, error)
}

func resolveLightingScheduleForZone(
	ctx context.Context,
	querier farmMatchQuerier,
	schedules []db.Gr33ncoreSchedule,
	farmID, zoneID int64,
) (db.Gr33ncoreSchedule, bool) {
	lpq, ok := querier.(lightingProgramQuerier)
	if !ok {
		return db.Gr33ncoreSchedule{}, false
	}
	programs, err := lpq.ListLightingProgramsByFarm(ctx, farmID)
	if err != nil || len(programs) == 0 {
		return db.Gr33ncoreSchedule{}, false
	}
	for _, lp := range programs {
		if lp.ZoneID != zoneID || !lp.IsActive {
			continue
		}
		if lp.ScheduleOnID != nil {
			if sch, ok := scheduleByID(schedules, *lp.ScheduleOnID); ok {
				return sch, true
			}
		}
	}
	return db.Gr33ncoreSchedule{}, false
}

func scheduleByID(schedules []db.Gr33ncoreSchedule, id int64) (db.Gr33ncoreSchedule, bool) {
	for _, s := range schedules {
		if s.ID == id {
			return s, true
		}
	}
	return db.Gr33ncoreSchedule{}, false
}

func scheduleDescribesZone(s db.Gr33ncoreSchedule, zoneName string) bool {
	if zoneName == "" {
		return true
	}
	if s.Description != nil {
		desc := strings.ToLower(*s.Description)
		zoneLower := strings.ToLower(zoneName)
		if strings.Contains(desc, "zone: "+zoneLower) || strings.Contains(desc, zoneLower) {
			return true
		}
	}
	nameLower := strings.ToLower(s.Name)
	zoneLower := strings.ToLower(zoneName)
	return strings.Contains(nameLower, zoneLower)
}

func pickScheduleForIntent(candidates []db.Gr33ncoreSchedule, lightingIntent bool, lowerQuestion string) (db.Gr33ncoreSchedule, bool) {
	if len(candidates) == 0 {
		return db.Gr33ncoreSchedule{}, false
	}
	if len(candidates) == 1 {
		return candidates[0], true
	}
	if lightingIntent {
		for _, s := range candidates {
			nameLower := strings.ToLower(s.Name)
			if !s.IsActive {
				continue
			}
			if strings.Contains(nameLower, "light on") || strings.Contains(nameLower, "lights on") {
				return s, true
			}
		}
		for _, s := range candidates {
			if s.IsActive {
				return s, true
			}
		}
	}
	for _, s := range candidates {
		if s.IsActive {
			return s, true
		}
	}
	return candidates[0], true
}

func resolveZoneNameForIntent(ctx context.Context, querier farmMatchQuerier, question string, farmID int64, snap Snapshot) string {
	zoneID := resolveZoneIDForIntent(ctx, querier, question, farmID, snap)
	if zoneID <= 0 {
		return ""
	}
	zones, err := querier.ListZonesByFarm(ctx, farmID)
	if err != nil {
		return ""
	}
	for _, z := range zones {
		if z.ID == zoneID {
			return z.Name
		}
	}
	return ""
}

func resolveZoneIDForIntent(ctx context.Context, querier farmMatchQuerier, question string, farmID int64, snap Snapshot) int64 {
	lower := strings.ToLower(strings.TrimSpace(question))
	zones, err := querier.ListZonesByFarm(ctx, farmID)
	if err == nil {
		for _, z := range zones {
			if z.Name != "" && strings.Contains(lower, strings.ToLower(z.Name)) {
				return z.ID
			}
		}
		for _, z := range zones {
			if zoneIntentMatchesNickname(lower, z.Name) {
				return z.ID
			}
		}
	}
	for _, name := range snap.ZoneNames {
		if strings.Contains(lower, strings.ToLower(name)) {
			for _, z := range zones {
				if z.Name == name {
					return z.ID
				}
			}
		}
		if zoneIntentMatchesNickname(lower, name) {
			for _, z := range zones {
				if z.Name == name {
					return z.ID
				}
			}
		}
	}
	return 0
}

func zoneIntentMatchesNickname(questionLower, zoneName string) bool {
	zoneLower := strings.ToLower(strings.TrimSpace(zoneName))
	for _, nick := range zoneNicknamesFor(zoneLower) {
		if strings.Contains(questionLower, nick) {
			return true
		}
	}
	return false
}

func zoneNicknamesFor(zoneNameLower string) []string {
	switch {
	case strings.Contains(zoneNameLower, "veg"):
		return []string{"veg tent", "veg stage", "vegetative tent"}
	case strings.Contains(zoneNameLower, "flower"):
		return []string{"flower tent", "bloom room", "flowering tent"}
	case strings.Contains(zoneNameLower, "propagation"):
		return []string{"prop room", "clone room", "propagation tent"}
	default:
		return nil
	}
}
