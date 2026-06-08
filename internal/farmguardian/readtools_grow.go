// Phase 62 WS1 — grow advisor read tool (VPD, DLI, stage targets, comfort bands).

package farmguardian

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

// GrowAdvisorPersonaRule is injected into the platform context block.
const GrowAdvisorPersonaRule = `Grow advisor (Phase 62): Use grow_advisor and lookup_crop_targets for VPD, DLI, stage targets, and transition readiness. Say "flip" not "transition to 12/12"; "light hours" not "photoperiod"; "harvest window" not "day of senescence". VPD is fine without definition unless the operator asks. Post-harvest: one concrete recommendation grounded in prior run data when available.`

var (
	growAdvisorIntent = regexp.MustCompile(`(?i)\b(vpd|dli|ppfd|flip|harvest|light hours?|photoperiod|grow advisor|summarize (this )?grow|ready to harvest|days? to flip|optimize light|comfort (band|target)|on target|stage advice|transition)\b|\bis my vpd\b|\bhow many days\b.*\bflip\b`)
	tempSensorTypes   = []string{"temperature", "air_temperature", "temp"}
	humidityTypes     = []string{"humidity", "relative_humidity", "rh"}
	ppfdSensorTypes   = []string{"ppfd", "par", "light"}
)

// CalcVPDKpa returns vapour pressure deficit in kPa from air temp (°C) and RH (%).
// Matches pi_client/gr33n_client.py compute_vpd_kpa (Tetens over water).
func CalcVPDKpa(tempC, rhPct float64) float64 {
	svp := 0.6108 * math.Exp((17.27*tempC)/(tempC+237.3))
	vpd := svp * (1.0 - rhPct/100.0)
	return math.Round(vpd*1000) / 1000
}

func shouldRunGrowAdvisorReadIntent(question string, ref *ContextRef) bool {
	q := strings.TrimSpace(question)
	if q == "" && ref == nil {
		return false
	}
	if ref != nil && ref.CropCycleID > 0 {
		if q == "" || growAdvisorIntent.MatchString(q) {
			return true
		}
	}
	if ref != nil && strings.EqualFold(ref.Type, "zone") && ref.ID > 0 {
		if q == "" || growAdvisorIntent.MatchString(q) {
			return true
		}
	}
	return growAdvisorIntent.MatchString(q)
}

type growAdvisorScope struct {
	zone      db.Gr33ncoreZone
	cycle     db.Gr33nfertigationCropCycle
	haveCycle bool
	profileID int64
	plantName string
	stage     db.Gr33nfertigationGrowthStageEnum
}

func resolveGrowAdvisorScope(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) (growAdvisorScope, error) {
	var out growAdvisorScope

	if ref != nil {
		if strings.EqualFold(ref.Type, "zone") && ref.ID > 0 {
			z, err := q.GetZoneByID(ctx, ref.ID)
			if err == nil && z.FarmID == farmID {
				out.zone = z
			}
		}
		if ref.CropCycleID > 0 {
			c, err := q.GetCropCycleByID(ctx, ref.CropCycleID)
			if err == nil && c.FarmID == farmID {
				out.cycle = c
				out.haveCycle = true
				if out.zone.ID == 0 {
					if z, zerr := q.GetZoneByID(ctx, c.ZoneID); zerr == nil {
						out.zone = z
					}
				}
			}
		}
	}

	if !out.haveCycle && out.zone.ID > 0 {
		if c, ok := activeCycleForZoneID(ctx, q, farmID, out.zone.ID); ok {
			out.cycle = c
			out.haveCycle = true
		}
	}
	if !out.haveCycle {
		if zone, ok := resolveZoneForSummary(ctx, q, farmID, question, Snapshot{}); ok {
			out.zone = zone
			if c, err := q.GetActiveCropCycleForZone(ctx, zone.ID); err == nil && c.FarmID == farmID {
				out.cycle = c
				out.haveCycle = true
			}
		}
	}
	if !out.haveCycle {
		return out, nil
	}

	if out.cycle.CurrentStage != nil {
		out.stage = *out.cycle.CurrentStage
	}
	if out.cycle.PlantID != nil && *out.cycle.PlantID > 0 {
		plant, perr := q.GetPlant(ctx, *out.cycle.PlantID)
		if perr == nil {
			out.plantName = plant.DisplayName
			if plant.CropProfileID != nil {
				out.profileID = *plant.CropProfileID
			}
		}
	}
	return out, nil
}

func renderGrowAdvisor(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) (string, error) {
	scope, err := resolveGrowAdvisorScope(ctx, q, farmID, question, ref)
	if err != nil {
		return "", err
	}
	if !scope.haveCycle {
		return "grow_advisor: no active crop cycle in scope — ask which zone or start a grow.", nil
	}

	cycle := scope.cycle
	zoneLabel := scope.zone.Name
	if zoneLabel == "" {
		zoneLabel = fmt.Sprintf("zone #%d", cycle.ZoneID)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("grow_advisor — %s · cycle #%d %s", zoneLabel, cycle.ID, strings.TrimSpace(cycle.Name)))
	if scope.plantName != "" {
		b.WriteString(fmt.Sprintf(" · plant %s", scope.plantName))
	}
	days := durationDaysSinceStart(cycle.StartedAt, cycle.HarvestedAt)
	if scope.stage != "" {
		b.WriteString(fmt.Sprintf("\nStage: %s", scope.stage))
	}
	if days > 0 {
		b.WriteString(fmt.Sprintf(" · day %d", days))
	}

	if scope.profileID <= 0 {
		b.WriteString("\nCrop profile: none assigned — offer to pick one in Start grow or Plants; do not guess targets.")
	} else if scope.stage != "" {
		st, serr := q.GetCropProfileStage(ctx, db.GetCropProfileStageParams{
			CropProfileID: scope.profileID,
			Stage:         scope.stage,
		})
		if serr != nil {
			if !errors.Is(serr, pgx.ErrNoRows) {
				return "", serr
			}
			b.WriteString(fmt.Sprintf("\nCrop targets: no stage row for %s on assigned profile.", scope.stage))
		} else {
			b.WriteString(fmt.Sprintf("\nVPD target: %s kPa", formatRange(st.VpdMinKpa, st.VpdMaxKpa, "kPa")))
			if st.DliTarget.Valid {
				b.WriteString(fmt.Sprintf("\nDLI target: %s mol/m²/day", formatNumeric(st.DliTarget)))
			}
			if st.PhotoperiodHrs.Valid {
				b.WriteString(fmt.Sprintf("\nLight hours (profile): %s h", formatNumeric(st.PhotoperiodHrs)))
			}
			if st.Notes != nil && strings.TrimSpace(*st.Notes) != "" {
				b.WriteString("\nStage notes: " + strings.TrimSpace(*st.Notes))
			}
		}
	}

	if band := comfortBandLine(ctx, q, cycle, "temperature"); band != "" {
		b.WriteString("\nComfort band (temp): " + band)
	}
	if band := comfortBandLine(ctx, q, cycle, "humidity"); band != "" {
		b.WriteString("\nComfort band (humidity): " + band)
	}

	tempC, tempAge, haveTemp := latestZoneSensorValue(ctx, q, cycle.ZoneID, tempSensorTypes)
	rhPct, rhAge, haveRH := latestZoneSensorValue(ctx, q, cycle.ZoneID, humidityTypes)
	if haveTemp && haveRH {
		vpd := CalcVPDKpa(tempC, rhPct)
		b.WriteString(fmt.Sprintf("\nCurrent VPD: %.3f kPa (from %.1f°C %s, %.1f%% RH %s)", vpd, tempC, tempAge, rhPct, rhAge))
		if scope.profileID > 0 && scope.stage != "" {
			if st, serr := q.GetCropProfileStage(ctx, db.GetCropProfileStageParams{
				CropProfileID: scope.profileID,
				Stage:         scope.stage,
			}); serr == nil {
				b.WriteString(" — " + vpdVsTarget(vpd, st.VpdMinKpa, st.VpdMaxKpa))
			}
		}
	} else {
		b.WriteString("\nCurrent VPD: unavailable (need temperature and humidity readings in this zone)")
	}

	ppfd, _, havePPFD := latestZoneSensorValue(ctx, q, cycle.ZoneID, ppfdSensorTypes)
	photoperiodHrs := profilePhotoperiodHrs(ctx, q, scope.profileID, scope.stage)
	if havePPFD && photoperiodHrs > 0 {
		dli := estimateDLI(ppfd, photoperiodHrs)
		b.WriteString(fmt.Sprintf("\nDLI estimate: %.1f mol/m²/day (PPFD %.0f µmol/m²/s × %.1f light hours)", dli, ppfd, photoperiodHrs))
	} else if photoperiodHrs > 0 {
		b.WriteString(fmt.Sprintf("\nDLI estimate: unavailable without PPFD/PAR sensor (profile light hours: %.1f h)", photoperiodHrs))
	} else {
		b.WriteString("\nDLI estimate: unavailable (no PPFD sensor and no profile light hours)")
	}

	stageStr := string(scope.stage)
	if isVegStage(stageStr) {
		b.WriteString(fmt.Sprintf("\nFlip readiness: day %d in veg — check canopy size and node spacing against profile notes before flip.", days))
	}
	if isLateFlowerStage(stageStr) {
		b.WriteString("\nHarvest window: late flower/flush — check trichome maturity; taper feed per profile notes.")
	}
	if days >= 14 {
		if costLine := cycleCostSummaryLine(ctx, q, cycle.ID); costLine != "" {
			b.WriteString("\n" + costLine)
		}
	}

	return b.String(), nil
}

// growAdvisorBriefLine is a compact hint for context_ref focus blocks (best-effort).
func growAdvisorBriefLine(ctx context.Context, q *db.Queries, farmID, zoneID, cycleID int64) string {
	cycle, err := q.GetCropCycleByID(ctx, cycleID)
	if err != nil || cycle.FarmID != farmID || !cycle.IsActive {
		return ""
	}
	scope := growAdvisorScope{
		zone:      db.Gr33ncoreZone{ID: zoneID},
		cycle:     cycle,
		haveCycle: true,
	}
	if cycle.CurrentStage != nil {
		scope.stage = *cycle.CurrentStage
	}
	if cycle.PlantID != nil && *cycle.PlantID > 0 {
		if plant, perr := q.GetPlant(ctx, *cycle.PlantID); perr == nil {
			scope.plantName = plant.DisplayName
			if plant.CropProfileID != nil {
				scope.profileID = *plant.CropProfileID
			}
		}
	}

	days := durationDaysSinceStart(cycle.StartedAt, cycle.HarvestedAt)
	var b strings.Builder
	if scope.plantName != "" {
		b.WriteString(fmt.Sprintf("Active grow: %s", scope.plantName))
	} else {
		b.WriteString(fmt.Sprintf("Active grow: %s", strings.TrimSpace(cycle.Name)))
	}
	if scope.stage != "" {
		b.WriteString(fmt.Sprintf(", day %d of %s", days, scope.stage))
	} else if days > 0 {
		b.WriteString(fmt.Sprintf(", day %d", days))
	}
	b.WriteString(".\nPrefer grow_advisor and lookup_crop_targets for VPD/DLI/stage advice.")

	tempC, _, haveTemp := latestZoneSensorValue(ctx, q, zoneID, tempSensorTypes)
	rhPct, _, haveRH := latestZoneSensorValue(ctx, q, zoneID, humidityTypes)
	if haveTemp && haveRH {
		vpd := CalcVPDKpa(tempC, rhPct)
		b.WriteString(fmt.Sprintf("\nCurrent VPD: %.3f kPa", vpd))
		if scope.profileID > 0 && scope.stage != "" {
			if st, serr := q.GetCropProfileStage(ctx, db.GetCropProfileStageParams{
				CropProfileID: scope.profileID,
				Stage:         scope.stage,
			}); serr == nil && (st.VpdMinKpa.Valid || st.VpdMaxKpa.Valid) {
				b.WriteString(fmt.Sprintf(" (target %s kPa for %s)", formatRange(st.VpdMinKpa, st.VpdMaxKpa, "kPa"), scope.stage))
			}
		}
		b.WriteByte('.')
	}
	return strings.TrimSpace(b.String())
}

func latestZoneSensorValue(ctx context.Context, q db.Querier, zoneID int64, types []string) (float64, string, bool) {
	zid := zoneID
	for _, sensorType := range types {
		reading, err := q.GetLatestReadingForZoneSensorType(ctx, db.GetLatestReadingForZoneSensorTypeParams{
			ZoneID:     &zid,
			SensorType: sensorType,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return 0, "", false
		}
		val := numericToFloat64(reading.ValueRaw)
		if !reading.ValueRaw.Valid && reading.ValueText != nil {
			if f, ok := parseFloatLoose(strings.TrimSpace(*reading.ValueText)); ok {
				val = f
			}
		}
		age := humanizeAge(timeSince(reading.ReadingTime))
		return val, age, true
	}
	return 0, "", false
}

func parseFloatLoose(s string) (float64, bool) {
	s = strings.TrimSuffix(s, "%")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err == nil
}

func comfortBandLine(ctx context.Context, q db.Querier, cycle db.Gr33nfertigationCropCycle, sensorType string) string {
	zid := cycle.ZoneID
	cid := cycle.ID
	var stage *string
	if cycle.CurrentStage != nil {
		s := string(*cycle.CurrentStage)
		stage = &s
	}
	row, err := q.GetActiveSetpointForScope(ctx, db.GetActiveSetpointForScopeParams{
		SensorType:  sensorType,
		CropCycleID: &cid,
		Stage:       stage,
		ZoneID:      &zid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ""
		}
		return ""
	}
	return formatRange(row.MinValue, row.MaxValue, sensorUnitSuffix(sensorType))
}

func profilePhotoperiodHrs(ctx context.Context, q db.Querier, profileID int64, stage db.Gr33nfertigationGrowthStageEnum) float64 {
	if profileID <= 0 || stage == "" {
		return 0
	}
	st, err := q.GetCropProfileStage(ctx, db.GetCropProfileStageParams{
		CropProfileID: profileID,
		Stage:         stage,
	})
	if err != nil || !st.PhotoperiodHrs.Valid {
		return 0
	}
	return numericToFloat64(st.PhotoperiodHrs)
}

// estimateDLI converts average PPFD (µmol/m²/s) and photoperiod hours to DLI (mol/m²/day).
func estimateDLI(ppfd float64, photoperiodHrs float64) float64 {
	if ppfd <= 0 || photoperiodHrs <= 0 {
		return 0
	}
	return math.Round(ppfd*photoperiodHrs*0.0036*10) / 10
}

func vpdVsTarget(vpd float64, min, max pgtype.Numeric) string {
	if min.Valid && vpd < numericToFloat64(min) {
		return "below stage target"
	}
	if max.Valid && vpd > numericToFloat64(max) {
		return "above stage target"
	}
	if min.Valid || max.Valid {
		return "within stage target"
	}
	return "no VPD target on profile"
}

func isVegStage(stage string) bool {
	switch stage {
	case "clone", "seedling", "early_veg", "late_veg":
		return true
	default:
		return false
	}
}

func isLateFlowerStage(stage string) bool {
	switch stage {
	case "late_flower", "flush":
		return true
	default:
		return false
	}
}
