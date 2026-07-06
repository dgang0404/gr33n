// Phase 136 — plant_context_bundle read tool (Phase 82 WS7).

package farmguardian

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	db "gr33n-api/internal/db"
)

const (
	// PlantContextBundleMaxRunes caps fused grow block size (~800 tokens).
	PlantContextBundleMaxRunes = 3200

	// PlantContextBundleRule is injected into the platform context block.
	PlantContextBundleRule = `Plant context bundle (Phase 136): When plant_context_bundle appears, answer from that fused block first — one active grow per turn. Lead with live vs target deltas (EC/VPD/temp/RH) from structured rows; field-guide RAG is supplemental only (Phase 97).`
)

var plantContextBundleIntent = regexp.MustCompile(`(?i)\b(grow|stage|harvest|flip|how is my|how's my|how is this|how's this|this grow|my grow|my plant|this plant|this room|veg grow|flower run|on target|comfort)\b|\b(ec|vpd|dli|photoperiod|light hours?)\b`)

// ShouldRunPlantContextBundleIntent reports whether to fuse grow read tools this turn.
func ShouldRunPlantContextBundleIntent(question string, ref *ContextRef) bool {
	if ref != nil && ref.CropCycleID > 0 {
		return true
	}
	q := strings.TrimSpace(question)
	if q == "" && ref == nil {
		return false
	}
	if ref != nil && strings.EqualFold(ref.Type, "zone") && ref.ID > 0 {
		if q == "" || plantContextBundleIntent.MatchString(q) || questionMentionsCrop(q) {
			return true
		}
	}
	if plantContextBundleIntent.MatchString(q) || questionMentionsCrop(q) {
		if strings.Contains(strings.ToLower(q), "how") ||
			strings.Contains(strings.ToLower(q), "stage") ||
			strings.Contains(strings.ToLower(q), "grow") ||
			lookupCropTargetsIntent.MatchString(q) ||
			growAdvisorIntent.MatchString(q) {
			return true
		}
	}
	if shouldRunLookupCropSymptomsIntent(q, ref) {
		return true
	}
	return false
}

type plantBundleSection struct {
	priority int
	text     string
}

func renderPlantContextBundle(ctx context.Context, q db.Querier, farmID int64, question string, snap Snapshot, ref *ContextRef) (string, error) {
	scope, err := resolveGrowAdvisorScope(ctx, q, farmID, question, ref)
	if err != nil {
		return "", err
	}
	if !scope.haveCycle {
		return "plant_context_bundle: no active crop cycle in scope — ask which zone or start a grow.", nil
	}

	var sections []plantBundleSection
	sections = append(sections, plantBundleSection{priority: 0, text: renderPlantBundleHeader(ctx, q, scope)})

	if block, err := renderLookupCropTargets(ctx, q, farmID, question, bundleContextRef(ref, scope)); err == nil && strings.TrimSpace(block) != "" {
		sections = append(sections, plantBundleSection{priority: 1, text: block})
	}

	if block, err := renderGrowAdvisor(ctx, q, farmID, question, bundleContextRef(ref, scope)); err == nil && strings.TrimSpace(block) != "" {
		sections = append(sections, plantBundleSection{priority: 2, text: block})
	}

	if scope.zone.ID > 0 && !zoneContextRefCovers(ref, scope.zone) {
		if block, err := renderSummarizeZone(ctx, q, farmID, scope.zone); err == nil && strings.TrimSpace(block) != "" {
			sections = append(sections, plantBundleSection{priority: 3, text: block})
		}
	}

	if scope.zone.ID > 0 {
		if block, err := renderSummarizeZoneFertigation(ctx, q, farmID, scope.zone); err == nil && strings.TrimSpace(block) != "" {
			sections = append(sections, plantBundleSection{priority: 4, text: block})
		}
		if block, err := renderSummarizeZoneLightingRead(ctx, q, farmID, scope.zone.ID); err == nil && strings.TrimSpace(block) != "" {
			sections = append(sections, plantBundleSection{priority: 5, text: block})
		}
	}

	if shouldRunLookupCropSymptomsIntent(question, ref) || bundleSymptomFooterIntent(question) {
		if block, err := renderLookupCropSymptoms(ctx, q, farmID, question, bundleContextRef(ref, scope)); err == nil && strings.TrimSpace(block) != "" {
			sections = append(sections, plantBundleSection{priority: 6, text: block})
		}
	}

	out := trimPlantContextBundle(sections, PlantContextBundleMaxRunes)
	if out == "" {
		return "plant_context_bundle: cycle in scope but no detail rows loaded.", nil
	}
	return "plant_context_bundle:\n" + out, nil
}

func bundleContextRef(ref *ContextRef, scope growAdvisorScope) *ContextRef {
	if ref != nil {
		cp := *ref
		if cp.CropCycleID == 0 && scope.haveCycle {
			cp.CropCycleID = scope.cycle.ID
		}
		if cp.ID == 0 && scope.zone.ID > 0 {
			cp.Type = "zone"
			cp.ID = scope.zone.ID
			if cp.Name == "" {
				cp.Name = scope.zone.Name
			}
		}
		return &cp
	}
	if !scope.haveCycle {
		return nil
	}
	out := &ContextRef{Type: "zone", CropCycleID: scope.cycle.ID}
	if scope.zone.ID > 0 {
		out.ID = scope.zone.ID
		out.Name = scope.zone.Name
	}
	return out
}

func renderPlantBundleHeader(ctx context.Context, q db.Querier, scope growAdvisorScope) string {
	cycle := scope.cycle
	zoneLabel := scope.zone.Name
	if zoneLabel == "" {
		zoneLabel = fmt.Sprintf("zone #%d", cycle.ZoneID)
	}
	days := durationDaysSinceStart(cycle.StartedAt, cycle.HarvestedAt)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Active grow — %s · cycle #%d %s", zoneLabel, cycle.ID, strings.TrimSpace(cycle.Name)))
	if scope.plantName != "" {
		b.WriteString(fmt.Sprintf(" · plant %s", scope.plantName))
	}
	if cycle.PlantID != nil && *cycle.PlantID > 0 {
		if plant, err := q.GetPlant(ctx, *cycle.PlantID); err == nil {
			if plant.CropKey != nil && strings.TrimSpace(*plant.CropKey) != "" {
				b.WriteString(fmt.Sprintf(" · crop_key %s", strings.TrimSpace(*plant.CropKey)))
			}
		}
	}
	if cycle.BatchLabel != nil && strings.TrimSpace(*cycle.BatchLabel) != "" {
		b.WriteString(fmt.Sprintf(" · batch %s", strings.TrimSpace(*cycle.BatchLabel)))
	}
	if scope.stage != "" {
		b.WriteString(fmt.Sprintf("\nStage: %s", scope.stage))
	}
	if days > 0 {
		b.WriteString(fmt.Sprintf(" · day %d in run", days))
	}
	b.WriteString("\nAnswer from bundle sections below — compare live readings to lookup_crop_targets / grow_advisor targets.")
	return b.String()
}

func renderSummarizeZoneLightingRead(ctx context.Context, q db.Querier, farmID, zoneID int64) (string, error) {
	programs, err := q.ListLightingProgramsByFarm(ctx, farmID)
	if err != nil {
		return "", err
	}
	var zonePrograms []db.Gr33ncoreLightingProgram
	for _, p := range programs {
		if p.ZoneID == zoneID {
			zonePrograms = append(zonePrograms, p)
		}
	}
	if len(zonePrograms) == 0 {
		return "summarize_zone_lighting: no lighting programs for this zone.", nil
	}
	sort.Slice(zonePrograms, func(i, j int) bool {
		if zonePrograms[i].IsActive != zonePrograms[j].IsActive {
			return zonePrograms[i].IsActive
		}
		return zonePrograms[i].Name < zonePrograms[j].Name
	})

	var b strings.Builder
	b.WriteString(fmt.Sprintf("summarize_zone_lighting — zone #%d:", zoneID))
	for _, p := range zonePrograms {
		status := "inactive"
		if p.IsActive {
			status = "active"
		}
		offAt := lightingOffAt(p.LightsOnAt, p.OnHours)
		b.WriteString(fmt.Sprintf("\n- %s (#%d, %s): %dh ON / %dh OFF — on %s off %s (%s)",
			strings.TrimSpace(p.Name), p.ID, status, p.OnHours, p.OffHours, p.LightsOnAt, offAt, p.Timezone))
	}
	return b.String(), nil
}

func lightingOffAt(lightsOnAt string, onHours int32) string {
	parts := strings.SplitN(lightsOnAt, ":", 2)
	if len(parts) != 2 {
		return "unknown"
	}
	var h, m int
	if _, err := fmt.Sscanf(parts[0], "%d", &h); err != nil {
		return "unknown"
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &m); err != nil {
		return "unknown"
	}
	total := int(onHours)*60 + h*60 + m
	h = (total / 60) % 24
	m = total % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func trimPlantContextBundle(sections []plantBundleSection, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = PlantContextBundleMaxRunes
	}
	sort.Slice(sections, func(i, j int) bool {
		return sections[i].priority < sections[j].priority
	})
	for len(sections) > 1 && bundleRunes(sections) > maxRunes {
		sections = sections[:len(sections)-1]
	}
	var parts []string
	for _, s := range sections {
		if t := strings.TrimSpace(s.text); t != "" {
			parts = append(parts, t)
		}
	}
	out := strings.Join(parts, "\n\n")
	for utf8.RuneCountInString(out) > maxRunes && len(out) > 200 {
		out = out[:len(out)*9/10]
	}
	return strings.TrimSpace(out)
}

func bundleRunes(sections []plantBundleSection) int {
	var parts []string
	for _, s := range sections {
		if t := strings.TrimSpace(s.text); t != "" {
			parts = append(parts, t)
		}
	}
	return utf8.RuneCountInString(strings.Join(parts, "\n\n"))
}

func bundleSymptomFooterIntent(question string) bool {
	return strings.Contains(strings.ToLower(question), "plant_context_bundle")
}

func bundleCoversReadTool(bundleRan bool, tool string) bool {
	if !bundleRan {
		return false
	}
	switch tool {
	case "lookup_crop_targets", "grow_advisor", "summarize_zone", "summarize_zone_fertigation", "lookup_crop_symptoms":
		return true
	default:
		return false
	}
}
