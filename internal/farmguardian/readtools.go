// Phase 31 WS6 — read-only Guardian tools that enrich the grounded system
// prompt before the LLM call. These never create proposals or require Confirm.

package farmguardian

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

const (
	// ReadToolsMaxAlerts is the cap for list_unread_alerts (snapshot keeps 3).
	ReadToolsMaxAlerts = 20

	// ReadToolsMaxPlants is the cap for list_plants (snapshot keeps SnapshotMaxPlantNames).
	ReadToolsMaxPlants = 20

	// ReadToolsMaxSensorReadings caps per-zone sensor lines in summarize_zone.
	ReadToolsMaxSensorReadings = 12

	// ReadToolsMaxZonePrograms caps programs listed in summarize_zone_fertigation.
	ReadToolsMaxZonePrograms = 8
)

var (
	zoneIDPattern = regexp.MustCompile(`(?i)\bzone\s*#?\s*(\d+)\b`)
	listAlertsIntent = regexp.MustCompile(`(?i)\b(list|show|what are|tell me about|any|how many)\b.*\b(unread\s+)?alerts?\b|\b(unread\s+)?alerts?\b.*\b(list|show|details?)\b`)
	listPlantsIntent = regexp.MustCompile(`(?i)\b(list|show|what are|tell me about|any|how many|do i have)\b.*\bplants?\b|\bplants?\b.*\b(list|show|catalog|inventory)\b`)
)

// ReadToolIDs returns registered read-only tool ids for platform context.
func ReadToolIDs() []string {
	return []string{"list_unread_alerts", "summarize_zone", "list_plants", "summarize_zone_fertigation"}
}

// EnrichPromptBlock runs matching read-only tools and returns extra system
// prompt text. Best-effort: query failures are logged; empty string means no
// enrichment for this turn.
func EnrichPromptBlock(ctx context.Context, q db.Querier, farmID int64, question string, snap Snapshot) string {
	if q == nil || farmID <= 0 {
		return ""
	}
	var blocks []string

	if matchListUnreadAlertsIntent(question) {
		if block, err := renderListUnreadAlerts(ctx, q, farmID); err != nil {
			slog.Warn("farm guardian read tool failed", "tool", "list_unread_alerts", "farm_id", farmID, "err", err)
		} else if block != "" {
			blocks = append(blocks, block)
		}
	}

	if matchListPlantsIntent(question) {
		if block, err := renderListPlants(ctx, q, farmID); err != nil {
			slog.Warn("farm guardian read tool failed", "tool", "list_plants", "farm_id", farmID, "err", err)
		} else if block != "" {
			blocks = append(blocks, block)
		}
	}

	if shouldRunSummarizeZoneReadIntent(question) {
		if zone, ok := resolveZoneForSummary(ctx, q, farmID, question, snap); ok {
			if block, err := renderSummarizeZone(ctx, q, farmID, zone); err != nil {
				slog.Warn("farm guardian read tool failed", "tool", "summarize_zone", "farm_id", farmID, "zone_id", zone.ID, "err", err)
			} else if block != "" {
				blocks = append(blocks, block)
			}
		}
	}

	if shouldRunSummarizeZoneFertigationReadIntent(question) {
		if zone, ok := resolveZoneForSummary(ctx, q, farmID, question, snap); ok {
			if block, err := renderSummarizeZoneFertigation(ctx, q, farmID, zone); err != nil {
				slog.Warn("farm guardian read tool failed", "tool", "summarize_zone_fertigation", "farm_id", farmID, "zone_id", zone.ID, "err", err)
			} else if block != "" {
				blocks = append(blocks, block)
			}
		}
	}

	if len(blocks) == 0 {
		return ""
	}
	return "Live read-tool results (background — do not cite as [n]):\n" + strings.Join(blocks, "\n\n")
}

func matchListUnreadAlertsIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if _, ok := matchAlertToolIntent(q); ok {
		return false
	}
	if snapUnreadCountIntent(q) {
		return false
	}
	return listAlertsIntent.MatchString(q)
}

// snapUnreadCountIntent catches "how many unread alerts" style questions that
// the farm snapshot already answers — skip the heavier list tool.
func snapUnreadCountIntent(question string) bool {
	lower := strings.ToLower(question)
	if !strings.Contains(lower, "alert") {
		return false
	}
	countWords := []string{"how many", "count", "number of"}
	for _, w := range countWords {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

func matchListPlantsIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if matchPlantWriteIntent(q) {
		return false
	}
	if snapPlantCountIntent(q) {
		return false
	}
	return listPlantsIntent.MatchString(q)
}

// snapPlantCountIntent catches "how many plants" when the snapshot already lists them.
func snapPlantCountIntent(question string) bool {
	lower := strings.ToLower(question)
	if !strings.Contains(lower, "plant") {
		return false
	}
	for _, w := range []string{"how many", "count", "number of"} {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

// matchPlantWriteIntent skips list_plants when the operator is asking to create a plant.
func matchPlantWriteIntent(question string) bool {
	lower := strings.ToLower(strings.TrimSpace(question))
	if !strings.Contains(lower, "plant") {
		return false
	}
	for _, verb := range []string{"add ", "create ", "set up ", "setup ", "start ", "new plant"} {
		if strings.Contains(lower, verb) {
			return true
		}
	}
	return false
}

func matchSummarizeZoneFertigationIntent(question string) bool {
	lower := strings.ToLower(strings.TrimSpace(question))
	if lower == "" {
		return false
	}
	for _, term := range []string{
		"fertigation", "feeding program", "feed program", "nutrient program",
		"watering program", "irrigation program", "fert program",
		"ec trigger", "ph trigger", "ec target", "ph target",
		"setpoint", "set point", "feeding schedule",
	} {
		if strings.Contains(lower, term) {
			return true
		}
	}
	if strings.Contains(lower, "program") &&
		(strings.Contains(lower, "zone") || strings.Contains(lower, "room") || strings.Contains(lower, "garden")) {
		return true
	}
	return false
}

// shouldRunSummarizeZoneFertigationReadIntent gates summarize_zone_fertigation enrichment.
func shouldRunSummarizeZoneFertigationReadIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	return matchSummarizeZoneFertigationIntent(q)
}

func matchSummarizeZoneIntent(question string) bool {
	lower := strings.ToLower(strings.TrimSpace(question))
	if lower == "" {
		return false
	}
	if strings.Contains(lower, "summarize") && strings.Contains(lower, "zone") {
		return true
	}
	for _, term := range []string{
		"humidity", "temperature", "temp ", "temp?", " co2", "ph ", "ec ", "vpd",
		"dew point", "reading", "readings", "sensor", "sensors",
	} {
		if strings.Contains(lower, term) {
			return true
		}
	}
	for _, phrase := range []string{
		"what's in ", "what is in ", "what's going on in ", "status of ",
		"zone status", "zone summary", "tell me about ",
	} {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// shouldRunSummarizeZoneReadIntent gates summarize_zone enrichment. Alert write
// proposals and alert-list questions must not also inject a zone sensor dump
// (Phase 33 WS1).
func shouldRunSummarizeZoneReadIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if _, ok := matchAlertToolIntent(q); ok {
		return false
	}
	if listAlertsIntent.MatchString(q) {
		return false
	}
	return matchSummarizeZoneIntent(q)
}

func resolveZoneForSummary(ctx context.Context, q db.Querier, farmID int64, question string, snap Snapshot) (db.Gr33ncoreZone, bool) {
	zones, err := q.ListZonesByFarm(ctx, farmID)
	if err != nil || len(zones) == 0 {
		return db.Gr33ncoreZone{}, false
	}

	if m := zoneIDPattern.FindStringSubmatch(question); len(m) > 1 {
		id, perr := strconv.ParseInt(m[1], 10, 64)
		if perr == nil {
			for _, z := range zones {
				if z.ID == id {
					return z, true
				}
			}
		}
	}

	lowerQ := strings.ToLower(question)
	var best *db.Gr33ncoreZone
	bestLen := 0
	matchZone := func(name string, z *db.Gr33ncoreZone) {
		n := strings.ToLower(strings.TrimSpace(name))
		if n == "" {
			return
		}
		if strings.Contains(lowerQ, n) && len(n) > bestLen {
			best = z
			bestLen = len(n)
		}
	}
	for i := range zones {
		matchZone(zones[i].Name, &zones[i])
	}
	for _, name := range snap.ZoneNames {
		for i := range zones {
			if zones[i].Name == name {
				matchZone(name, &zones[i])
			}
		}
	}
	if best != nil {
		return *best, true
	}
	if len(zones) == 1 {
		return zones[0], true
	}
	return db.Gr33ncoreZone{}, false
}

func renderListUnreadAlerts(ctx context.Context, q db.Querier, farmID int64) (string, error) {
	cnt, err := q.CountUnreadAlertsByFarm(ctx, farmID)
	if err != nil {
		return "", err
	}
	if cnt == 0 {
		return "list_unread_alerts: no unread alerts.", nil
	}

	limit := int32(ReadToolsMaxAlerts)
	alerts, err := q.ListRecentUnreadAlertsByFarm(ctx, farmID, limit)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("list_unread_alerts (%d unread):", cnt))
	for _, a := range alerts {
		d := toUnreadAlertDetail(a)
		b.WriteByte('\n')
		b.WriteString(fmt.Sprintf("- #%d", d.ID))
		if d.Severity != "" {
			b.WriteString(" [" + d.Severity + "]")
		}
		if d.Subject != "" {
			b.WriteString(" " + d.Subject)
		} else if d.Message != "" {
			b.WriteString(" " + d.Message)
		}
		meta := []string{humanizeAge(timeSince(d.TriggeredAt))}
		if d.SourceType != "" {
			src := d.SourceType
			if d.SourceID > 0 {
				src = fmt.Sprintf("%s #%d", src, d.SourceID)
			}
			meta = append(meta, src)
		}
		b.WriteString(" (" + strings.Join(meta, "; ") + ")")
	}
	if extra := cnt - int64(len(alerts)); extra > 0 {
		b.WriteString(fmt.Sprintf("\n(+ %d more unread alerts not listed)", extra))
	}
	return b.String(), nil
}

func renderListPlants(ctx context.Context, q db.Querier, farmID int64) (string, error) {
	plants, err := q.ListPlantsByFarm(ctx, farmID)
	if err != nil {
		return "", err
	}
	if len(plants) == 0 {
		return "list_plants: no plants on file for this farm.", nil
	}

	total := len(plants)
	listed := plants
	if len(listed) > ReadToolsMaxPlants {
		listed = listed[:ReadToolsMaxPlants]
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("list_plants (%d on file):", total))
	for _, p := range listed {
		b.WriteByte('\n')
		b.WriteString(fmt.Sprintf("- #%d %s", p.ID, strings.TrimSpace(p.DisplayName)))
		if p.VarietyOrCultivar != nil && strings.TrimSpace(*p.VarietyOrCultivar) != "" {
			b.WriteString(" (" + strings.TrimSpace(*p.VarietyOrCultivar) + ")")
		}
	}
	if extra := total - len(listed); extra > 0 {
		b.WriteString(fmt.Sprintf("\n(+ %d more plants not listed)", extra))
	}
	return b.String(), nil
}

func renderSummarizeZone(ctx context.Context, q db.Querier, farmID int64, zone db.Gr33ncoreZone) (string, error) {
	if zone.FarmID != farmID {
		return "", errors.New("zone farm mismatch")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("summarize_zone — %s (#%d)", zone.Name, zone.ID))
	if zone.ZoneType != nil && strings.TrimSpace(*zone.ZoneType) != "" {
		b.WriteString(" (" + strings.TrimSpace(*zone.ZoneType) + ")")
	}
	b.WriteByte('\n')

	cycles, err := q.ListCropCyclesByFarm(ctx, farmID)
	if err == nil {
		var active []string
		for _, c := range cycles {
			if !c.IsActive || c.ZoneID != zone.ID {
				continue
			}
			line := c.Name
			if c.StrainOrVariety != nil && strings.TrimSpace(*c.StrainOrVariety) != "" {
				line += " — " + strings.TrimSpace(*c.StrainOrVariety)
			}
			if c.CurrentStage.Valid {
				line += " (stage: " + string(c.CurrentStage.Gr33nfertigationGrowthStageEnum) + ")"
			}
			active = append(active, line)
		}
		if len(active) > 0 {
			b.WriteString("Active cycles: " + strings.Join(active, "; ") + "\n")
		}
	}

	zoneID := zone.ID
	sensors, err := q.ListSensorsByZone(ctx, &zoneID)
	if err != nil {
		return "", err
	}
	if len(sensors) == 0 {
		b.WriteString("Sensors: none configured in this zone.")
		return b.String(), nil
	}

	type readingLine struct {
		sortKey string
		text    string
	}
	lines := make([]readingLine, 0, len(sensors))
	for _, s := range sensors {
		reading, rerr := q.GetLatestReadingBySensor(ctx, s.ID)
		if rerr != nil {
			if errors.Is(rerr, pgx.ErrNoRows) {
				lines = append(lines, readingLine{
					sortKey: s.SensorType + " " + s.Name,
					text:    fmt.Sprintf("- %s (%s): no readings yet", sensorLabel(s), s.SensorType),
				})
				continue
			}
			return "", rerr
		}
		lines = append(lines, readingLine{
			sortKey: s.SensorType + " " + s.Name,
			text:    fmt.Sprintf("- %s (%s): %s (%s)", sensorLabel(s), s.SensorType, formatSensorReading(s, reading), humanizeAge(timeSince(reading.ReadingTime))),
		})
	}
	sort.Slice(lines, func(i, j int) bool { return lines[i].sortKey < lines[j].sortKey })

	b.WriteString("Latest sensor readings:")
	extra := 0
	if len(lines) > ReadToolsMaxSensorReadings {
		extra = len(lines) - ReadToolsMaxSensorReadings
		lines = lines[:ReadToolsMaxSensorReadings]
	}
	for _, ln := range lines {
		b.WriteByte('\n')
		b.WriteString(ln.text)
	}
	if extra > 0 {
		b.WriteString(fmt.Sprintf("\n(+ %d more sensors not listed)", extra))
	}
	return b.String(), nil
}

func renderSummarizeZoneFertigation(ctx context.Context, q db.Querier, farmID int64, zone db.Gr33ncoreZone) (string, error) {
	if zone.FarmID != farmID {
		return "", errors.New("zone farm mismatch")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("summarize_zone_fertigation — %s (#%d)", zone.Name, zone.ID))
	b.WriteByte('\n')

	programByID := make(map[int64]db.Gr33nfertigationProgram)
	if programs, err := q.ListProgramsByFarm(ctx, farmID); err == nil {
		for _, p := range programs {
			programByID[p.ID] = p
		}
		var zonePrograms []db.Gr33nfertigationProgram
		for _, p := range programs {
			if p.TargetZoneID == nil || *p.TargetZoneID != zone.ID {
				continue
			}
			zonePrograms = append(zonePrograms, p)
		}
		sort.Slice(zonePrograms, func(i, j int) bool {
			if zonePrograms[i].IsActive != zonePrograms[j].IsActive {
				return zonePrograms[i].IsActive
			}
			return zonePrograms[i].Name < zonePrograms[j].Name
		})
		if len(zonePrograms) == 0 {
			b.WriteString("Programs targeting zone: none\n")
		} else {
			b.WriteString("Programs targeting zone:")
			limit := ReadToolsMaxZonePrograms
			extra := 0
			if len(zonePrograms) > limit {
				extra = len(zonePrograms) - limit
				zonePrograms = zonePrograms[:limit]
			}
			for _, p := range zonePrograms {
				b.WriteByte('\n')
				status := "inactive"
				if p.IsActive {
					status = "active"
				}
				b.WriteString(fmt.Sprintf("- %s (#%d, %s)", strings.TrimSpace(p.Name), p.ID, status))
				if hints := programSetpointHints(p); hints != "" {
					b.WriteString("; " + hints)
				}
			}
			if extra > 0 {
				b.WriteString(fmt.Sprintf("\n(+ %d more programs not listed)", extra))
			}
			b.WriteByte('\n')
		}
	}

	if cycles, err := q.ListCropCyclesByFarm(ctx, farmID); err == nil {
		var active []string
		for _, c := range cycles {
			if !c.IsActive || c.ZoneID != zone.ID {
				continue
			}
			line := c.Name
			if c.StrainOrVariety != nil && strings.TrimSpace(*c.StrainOrVariety) != "" {
				line += " — " + strings.TrimSpace(*c.StrainOrVariety)
			}
			if c.CurrentStage.Valid {
				line += " (stage: " + string(c.CurrentStage.Gr33nfertigationGrowthStageEnum) + ")"
			}
			if c.PrimaryProgramID != nil {
				if p, ok := programByID[*c.PrimaryProgramID]; ok {
					line += "; primary program: " + strings.TrimSpace(p.Name)
				} else {
					line += fmt.Sprintf("; primary program: #%d", *c.PrimaryProgramID)
				}
			}
			active = append(active, line)
		}
		if len(active) > 0 {
			b.WriteString("Active cycles: " + strings.Join(active, "; ") + "\n")
		}
	}

	if targets, err := q.ListEcTargetsByFarm(ctx, farmID); err == nil {
		var zoneTargets []db.Gr33nfertigationEcTarget
		for _, t := range targets {
			if t.ZoneID != nil && *t.ZoneID == zone.ID {
				zoneTargets = append(zoneTargets, t)
			}
		}
		sort.Slice(zoneTargets, func(i, j int) bool {
			return string(zoneTargets[i].GrowthStage) < string(zoneTargets[j].GrowthStage)
		})
		if len(zoneTargets) > 0 {
			b.WriteString("EC/pH targets by stage:")
			for _, t := range zoneTargets {
				b.WriteByte('\n')
				b.WriteString(fmt.Sprintf("- %s: EC %s–%s mS/cm, pH %s–%s",
					string(t.GrowthStage),
					targetNumeric(t.EcMinMscm),
					targetNumeric(t.EcMaxMscm),
					targetPH(t.PhMin),
					targetPH(t.PhMax),
				))
			}
		}
	}

	return strings.TrimRight(b.String(), "\n"), nil
}

func programSetpointHints(p db.Gr33nfertigationProgram) string {
	var parts []string
	if p.TotalVolumeLiters.Valid {
		parts = append(parts, "volume "+formatLiters(numericToFloat64(p.TotalVolumeLiters)))
	}
	if p.EcTriggerLow.Valid {
		parts = append(parts, "EC trigger low "+formatEC(numericToFloat64(p.EcTriggerLow))+" mS/cm")
	}
	if p.PhTriggerLow.Valid || p.PhTriggerHigh.Valid {
		lo, hi := "—", "—"
		if p.PhTriggerLow.Valid {
			lo = formatPH(numericToFloat64(p.PhTriggerLow))
		}
		if p.PhTriggerHigh.Valid {
			hi = formatPH(numericToFloat64(p.PhTriggerHigh))
		}
		parts = append(parts, "pH "+lo+"–"+hi)
	}
	return strings.Join(parts, ", ")
}

func targetNumeric(n pgtype.Numeric) string {
	if !n.Valid {
		return "—"
	}
	return formatEC(numericToFloat64(n))
}

func targetPH(n pgtype.Numeric) string {
	if !n.Valid {
		return "—"
	}
	return formatPH(numericToFloat64(n))
}

func sensorLabel(s db.Gr33ncoreSensor) string {
	if strings.TrimSpace(s.Name) != "" {
		return strings.TrimSpace(s.Name)
	}
	return s.SensorType
}

func formatSensorReading(sensor db.Gr33ncoreSensor, reading db.Gr33ncoreSensorReading) string {
	if reading.ValueText != nil && strings.TrimSpace(*reading.ValueText) != "" {
		return strings.TrimSpace(*reading.ValueText)
	}
	val := numericToFloat64(reading.ValueRaw)
	if val == 0 && !reading.ValueRaw.Valid {
		return "—"
	}
	return fmt.Sprintf("%.1f%s", val, sensorUnitSuffix(sensor.SensorType))
}

func sensorUnitSuffix(sensorType string) string {
	switch strings.ToLower(strings.TrimSpace(sensorType)) {
	case "humidity", "relative_humidity":
		return "% RH"
	case "temperature", "air_temperature":
		return "°C"
	case "co2", "carbon_dioxide":
		return " ppm"
	case "ph":
		return " pH"
	case "ec", "electrical_conductivity":
		return " mS/cm"
	case "vpd":
		return " kPa"
	case "dew_point", "dewpoint":
		return "°C"
	case "light", "par", "ppfd":
		return " µmol/m²/s"
	default:
		return ""
	}
}
