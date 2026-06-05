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
	pauseScheduleIntent  = regexp.MustCompile(`(?i)(?:pause|disable|stop|turn\s+off)\s+(?:the\s+)?schedule`)
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
	for _, s := range schedules {
		if strings.Contains(lower, strings.ToLower(s.Name)) {
			return s, true
		}
	}
	if len(schedules) == 1 {
		return schedules[0], true
	}
	return db.Gr33ncoreSchedule{}, false
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
	}
	for _, name := range snap.ZoneNames {
		if strings.Contains(lower, strings.ToLower(name)) {
			for _, z := range zones {
				if z.Name == name {
					return z.ID
				}
			}
		}
	}
	return 0
}
