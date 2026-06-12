package farmguardian

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/croplibrary"
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

	cropKey, ok := resolveCropKeyForSetupPack(question)
	if !ok {
		return nil, "", false
	}
	zoneName, ok := resolveZoneNameForSetupPack(question, snap)
	if !ok {
		return nil, "", false
	}
	if exists, err := plantCropKeyOnFarm(ctx, q, farmID, cropKey); err != nil || exists {
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

	args := buildSetupPackArgs(inferSetupProfile(zoneName, question), zoneID, zoneName, cropKey)
	return args, tools.GrowSetupPackSummary(args), true
}

func resolveCropKeyForSetupPack(question string) (string, bool) {
	reg, err := defaultCropRegistry()
	if err != nil || reg == nil {
		return "", false
	}
	for _, m := range reg.FindMentions(question) {
		if m.Kind == croplibrary.MentionCrop {
			return m.Key, true
		}
	}
	if name, ok := extractPlantDisplayName(question); ok {
		term := strings.ToLower(strings.TrimSpace(name))
		term = strings.ReplaceAll(term, " ", "_")
		if m, ok := reg.ResolveTerm(term); ok && m.Kind == croplibrary.MentionCrop {
			return m.Key, true
		}
	}
	return "", false
}

func plantCropKeyOnFarm(ctx context.Context, q db.Querier, farmID int64, cropKey string) (bool, error) {
	cropKey = strings.TrimSpace(cropKey)
	if cropKey == "" || q == nil || farmID <= 0 {
		return false, nil
	}
	_, err := q.GetPlantByFarmCropKey(ctx, db.GetPlantByFarmCropKeyParams{
		FarmID:  farmID,
		CropKey: &cropKey,
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return false, err
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

func buildSetupPackArgs(profile string, zoneID int64, zoneName, cropKey string) map[string]any {
	label := titleWords(strings.ReplaceAll(cropKey, "_", " "))
	reg, _ := defaultCropRegistry()
	if reg != nil {
		if m, ok := reg.ResolveTerm(cropKey); ok && strings.TrimSpace(m.DisplayName) != "" {
			label = m.DisplayName
		}
	}
	today := time.Now().UTC().Format("2006-01-02")
	cycleName := label + " — " + zoneName
	programName := label + " light feed"
	volume := 0.5
	ecLow := 0.8
	phLo := 5.8
	phHi := 6.5
	stage := "early_veg"

	if profile == "commercial_zone" {
		programName = label + " feed program"
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
			"crop_key": cropKey,
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
			"title": "Monitor new " + label + " — first two weeks",
		},
	}
}

// InsertEmptyZoneSetupProposal creates a proactive apply_grow_setup_pack change request (Phase 73 WS2).
func InsertEmptyZoneSetupProposal(
	ctx context.Context,
	q db.Querier,
	userID uuid.UUID,
	farmID, zoneID int64,
	cropKey string,
) (db.Gr33ncoreGuardianActionProposal, error) {
	if q == nil || farmID <= 0 || zoneID <= 0 || userID == uuid.Nil {
		return db.Gr33ncoreGuardianActionProposal{}, errors.New("invalid empty-zone setup input")
	}
	zone, err := q.GetZoneByID(ctx, zoneID)
	if err != nil {
		return db.Gr33ncoreGuardianActionProposal{}, err
	}
	if zone.FarmID != farmID {
		return db.Gr33ncoreGuardianActionProposal{}, errors.New("zone not on farm")
	}
	if _, err := q.GetActiveCropCycleForZone(ctx, zoneID); err == nil {
		return db.Gr33ncoreGuardianActionProposal{}, errors.New("zone already has an active grow")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return db.Gr33ncoreGuardianActionProposal{}, err
	}
	cropKey = strings.TrimSpace(cropKey)
	profile := inferSetupProfile(zone.Name, "")
	if cropKey == "" {
		if profile == "commercial_zone" {
			cropKey = "tomato"
		} else {
			cropKey = "pothos"
		}
	}
	exists, err := plantCropKeyOnFarm(ctx, q, farmID, cropKey)
	if err != nil {
		return db.Gr33ncoreGuardianActionProposal{}, err
	}
	if exists {
		return db.Gr33ncoreGuardianActionProposal{}, errors.New("crop already on farm")
	}
	args := buildSetupPackArgs(profile, zoneID, zone.Name, cropKey)
	return insertProposal(ctx, q, insertProposalInput{
		userID:  userID,
		farmID:  farmID,
		toolID:  "apply_grow_setup_pack",
		args:    args,
		summary: tools.GrowSetupPackSummary(args),
	})
}
