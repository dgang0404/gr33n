package farmguardian

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian/tools"
)

var (
	setupPackVerbIntent = regexp.MustCompile(`(?i)\b(add|create|set\s*up|start)\b`)
	setupPackGrowIntent = regexp.MustCompile(`(?i)\b(plant|philodendron|pothos|monstera|ficus|herb|basil|tomato|seedling)\b|fertigation\s+program|feeding\s+program|light\s+feed`)
	setupPlantNamePattern = regexp.MustCompile(`(?i)(?:add|create|set\s*up|start)\s+(?:my\s+|a\s+)?([a-z][a-z0-9\s'-]{1,32}?)\s+(?:to|in)\b`)
)

// matchSetupPackIntent builds a frozen apply_grow_setup_pack proposal from chat +
// snapshot (Phase 32 WS4). Returns false when zone/plant cannot be resolved safely.
func matchSetupPackIntent(
	ctx context.Context,
	q db.Querier,
	farmID int64,
	question string,
	snap Snapshot,
) (map[string]any, string, bool) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, "", false
	}
	if !setupPackVerbIntent.MatchString(question) || !setupPackGrowIntent.MatchString(question) {
		return nil, "", false
	}
	if _, ok := matchAlertToolIntent(question); ok {
		return nil, "", false
	}
	if createTaskIntent.MatchString(question) && strings.Contains(strings.ToLower(question), "task") {
		return nil, "", false
	}

	plantName, ok := extractPlantDisplayName(question)
	if !ok {
		return nil, "", false
	}
	zoneName, ok := resolveZoneNameForSetupPack(question, snap)
	if !ok {
		return nil, "", false
	}
	if plantAlreadyOnFarm(plantName, snap) {
		return nil, "", false
	}
	if zoneHasActiveCycle(zoneName, snap) {
		return nil, "", false
	}
	if q == nil || farmID <= 0 {
		return nil, "", false
	}
	zoneID, err := resolveZoneIDByName(ctx, q, farmID, zoneName)
	if err != nil || zoneID <= 0 {
		return nil, "", false
	}

	args := buildSetupPackArgs(inferSetupProfile(zoneName, question), zoneID, zoneName, plantName)
	return args, tools.GrowSetupPackSummary(args), true
}

func extractPlantDisplayName(question string) (string, bool) {
	if m := setupPlantNamePattern.FindStringSubmatch(question); len(m) > 1 {
		name := strings.TrimSpace(m[1])
		if name == "" || isReservedSetupWord(name) {
			return "", false
		}
		return titleWords(name), true
	}
	lower := strings.ToLower(question)
	for _, known := range []string{"philodendron", "pothos", "monstera", "ficus", "basil", "tomato"} {
		if strings.Contains(lower, known) {
			return titleWords(known), true
		}
	}
	if strings.Contains(lower, "plant") {
		if m := regexp.MustCompile(`(?i)plant\s+(?:called|named)\s+([a-z][a-z0-9\s'-]{1,32})`).FindStringSubmatch(question); len(m) > 1 {
			name := strings.TrimSpace(m[1])
			if name != "" && !isReservedSetupWord(name) {
				return titleWords(name), true
			}
		}
	}
	return "", false
}

func isReservedSetupWord(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "plant", "a", "my", "new", "the", "setup", "grow", "full", "light", "task":
		return true
	default:
		return false
	}
}

func titleWords(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	parts := strings.Fields(s)
	for i, p := range parts {
		runes := []rune(strings.ToLower(p))
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}

func resolveZoneNameForSetupPack(question string, snap Snapshot) (string, bool) {
	lowerQ := strings.ToLower(question)
	var best string
	bestLen := 0
	match := func(name string) {
		n := strings.ToLower(strings.TrimSpace(name))
		if n == "" {
			return
		}
		if strings.Contains(lowerQ, n) && len(n) > bestLen {
			best = strings.TrimSpace(name)
			bestLen = len(n)
		}
	}
	for _, name := range snap.ZoneNames {
		match(name)
	}
	for _, c := range snap.ActiveCycles {
		match(c.ZoneName)
	}
	for _, zp := range snap.ProgramsByZone {
		match(zp.ZoneName)
	}
	return best, best != ""
}

func resolveZoneIDByName(ctx context.Context, q db.Querier, farmID int64, name string) (int64, error) {
	zones, err := q.ListZonesByFarm(ctx, farmID)
	if err != nil {
		return 0, err
	}
	for _, z := range zones {
		if strings.EqualFold(strings.TrimSpace(z.Name), strings.TrimSpace(name)) {
			return z.ID, nil
		}
	}
	return 0, fmt.Errorf("zone %q not found", name)
}

func plantAlreadyOnFarm(displayName string, snap Snapshot) bool {
	target := strings.ToLower(strings.TrimSpace(displayName))
	for _, n := range snap.PlantNames {
		line := strings.ToLower(n)
		if strings.HasPrefix(line, target) || strings.Contains(line, target) {
			return true
		}
	}
	return false
}

func zoneHasActiveCycle(zoneName string, snap Snapshot) bool {
	for _, c := range snap.ActiveCycles {
		if strings.EqualFold(c.ZoneName, zoneName) {
			return true
		}
	}
	return false
}

func inferSetupProfile(zoneName, question string) string {
	lower := strings.ToLower(zoneName + " " + question)
	for _, kw := range []string{
		"veg room", "flower room", "greenhouse", "commercial", "canopy",
		"outdoor garden", "outdoor", "propagation",
	} {
		if strings.Contains(lower, kw) {
			return "commercial_zone"
		}
	}
	return "house_plant"
}

func buildSetupPackArgs(profile string, zoneID int64, zoneName, plantName string) map[string]any {
	today := time.Now().UTC().Format("2006-01-02")
	cycleName := plantName + " — " + zoneName
	programName := plantName + " light feed"
	volume := 0.5
	ecLow := 0.8
	phLo := 5.8
	phHi := 6.5
	stage := "vegetative"

	if profile == "commercial_zone" {
		programName = plantName + " feed program"
		volume = 95.0
		ecLow = 1.2
		phHi = 6.8
		stage = "late_veg"
	}

	return map[string]any{
		"profile":   profile,
		"zone_id":   zoneID,
		"zone_name": zoneName,
		"plant": map[string]any{
			"display_name": plantName,
		},
		"cycle": map[string]any{
			"name":          cycleName,
			"current_stage": stage,
			"started_at":    today,
		},
		"program": map[string]any{
			"name":                programName,
			"total_volume_liters": volume,
			"ec_trigger_low":      ecLow,
			"ph_trigger_low":      phLo,
			"ph_trigger_high":     phHi,
			"is_active":           true,
		},
		"optional_task": map[string]any{
			"title": "Monitor new " + plantName + " — first two weeks",
		},
	}
}
