// Phase 64 WS3 — crop knowledge read tool (lookup_crop_targets).
// Phase 82 WS3 — multi-crop compare + YAML alias registry.

package farmguardian

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
)

// CropTargetsGroundingRule is injected into every grounded chat system prompt.
const CropTargetsGroundingRule = `Crop targets (Phase 64): NEVER state an EC, pH, VPD, DLI, or photoperiod target unless lookup_crop_targets (or an explicit crop profile stage row in read-tool results) provides it. EC is in mS/cm. If no crop profile is assigned to the plant/cycle, say so and offer to set one in Start grow or Plants — do not guess from general knowledge.`

const (
	maxMultiCropProfiles    = 6
	maxMultiUnsupportedCrops = 2
)

var lookupCropTargetsIntent = regexp.MustCompile(`(?i)\b(ec|ph|vpd|dli|photoperiod|target|targets|nutrient|feed strength|feed|water|watering|wet|moisture|irrigation|fertigation|how wet|light|lighting|hours|cycle|program|compare|vs|versus|difference between|each plant)\b|\bcrop\s+profile\b|\bwhat should (ec|ph|vpd)\b|\b(is my ec|is my ph)\b`)

var compareCropQuestion = regexp.MustCompile(`(?i)\b(compare|vs|versus|difference between|each plant)\b`)

var (
	cropRegOnce sync.Once
	cropReg     *croplibrary.Registry
	cropRegErr  error
)

func defaultCropRegistry() (*croplibrary.Registry, error) {
	cropRegOnce.Do(func() {
		cat, err := croplibrary.DefaultCatalog()
		if err != nil {
			cropRegErr = err
			return
		}
		cropReg = croplibrary.NewRegistry(cat)
	})
	return cropReg, cropRegErr
}

func shouldRunLookupCropTargetsReadIntent(question string, ref *ContextRef) bool {
	q := strings.TrimSpace(question)
	if q == "" && ref == nil {
		return false
	}
	if ref != nil && (ref.CropCycleID > 0 || (strings.EqualFold(ref.Type, "zone") && ref.ID > 0)) {
		if q == "" || lookupCropTargetsIntent.MatchString(q) || questionMentionsCrop(q) {
			return true
		}
	}
	if lookupCropTargetsIntent.MatchString(q) || questionMentionsCrop(q) {
		return true
	}
	return false
}

func questionMentionsCrop(question string) bool {
	reg, err := defaultCropRegistry()
	if err != nil || reg == nil {
		return false
	}
	return len(reg.FindMentions(question)) > 0
}

func renderLookupCropTargets(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) (string, error) {
	profileID, stage, plantName, err := resolveCropProfileContext(ctx, q, farmID, question, ref)
	if err != nil {
		return "", err
	}

	reg, regErr := defaultCropRegistry()
	var mentions []croplibrary.ResolvedMention
	if regErr == nil && reg != nil {
		mentions = reg.FindMentions(question)
	}
	cropMentions, unsupMentions := splitMentions(mentions)

	useMulti := len(cropMentions) >= 2 ||
		(len(unsupMentions) > 0 && len(cropMentions) > 0) ||
		(profileID <= 0 && len(cropMentions)+len(unsupMentions) > 0)

	if useMulti {
		return renderLookupCropTargetsMulti(ctx, q, farmID, question, cropMentions, unsupMentions)
	}

	if profileID <= 0 && len(cropMentions) == 1 {
		if id, ok := lookupBuiltinProfileID(ctx, q, farmID, cropMentions[0].Key); ok {
			profileID = id
		}
	}
	if profileID <= 0 && len(unsupMentions) == 1 && len(cropMentions) == 0 {
		return formatUnsupportedCropBlock(unsupMentions[0]), nil
	}
	if profileID <= 0 {
		return "lookup_crop_targets: no crop profile assigned to the active grow or plant. Offer to pick a profile in Start grow or Plants.", nil
	}

	return renderSingleCropTargets(ctx, q, profileID, stage, plantName)
}

func renderSingleCropTargets(ctx context.Context, q db.Querier, profileID int64, stage db.Gr33nfertigationGrowthStageEnum, plantName string) (string, error) {
	profile, err := q.GetCropProfile(ctx, profileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "lookup_crop_targets: crop profile not found.", nil
		}
		return "", err
	}
	if stage == "" {
		if len(profile.DisplayName) > 0 {
			stages, err := q.ListCropProfileStages(ctx, profileID)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			b.WriteString(fmt.Sprintf("lookup_crop_targets — %s (%s); stages available:", profile.DisplayName, profile.CropKey))
			for _, st := range stages {
				b.WriteString(fmt.Sprintf("\n- %s: EC %s mS/cm", st.Stage, formatEcRange(st.EcMin, st.EcTarget, st.EcMax)))
			}
			return b.String(), nil
		}
		return "lookup_crop_targets: specify a growth stage or zone context.", nil
	}
	st, err := q.GetCropProfileStage(ctx, db.GetCropProfileStageParams{
		CropProfileID: profileID,
		Stage:         stage,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Sprintf("lookup_crop_targets — %s has no targets for stage %s.", profile.DisplayName, stage), nil
		}
		return "", err
	}
	label := profile.DisplayName
	if plantName != "" {
		label = plantName + " (" + profile.DisplayName + ")"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("lookup_crop_targets — %s · stage %s", label, stage))
	b.WriteString(fmt.Sprintf("\nEC target: %s mS/cm", formatEcRange(st.EcMin, st.EcTarget, st.EcMax)))
	b.WriteString(fmt.Sprintf("\npH: %s", formatPhRange(st.PhMin, st.PhMax)))
	b.WriteString(fmt.Sprintf("\nVPD: %s kPa", formatRange(st.VpdMinKpa, st.VpdMaxKpa, "kPa")))
	b.WriteString(fmt.Sprintf("\nTemp: %s °C", formatRange(st.TempMinC, st.TempMaxC, "°C")))
	b.WriteString(fmt.Sprintf("\nRH: %s%%", formatRange(st.RhMinPct, st.RhMaxPct, "%")))
	if st.DliTarget.Valid {
		b.WriteString(fmt.Sprintf("\nDLI target: %s mol/m²/day", formatNumeric(st.DliTarget)))
	}
	if st.PhotoperiodHrs.Valid {
		b.WriteString(fmt.Sprintf("\nPhotoperiod: %s h", formatNumeric(st.PhotoperiodHrs)))
	}
	if st.Notes != nil && strings.TrimSpace(*st.Notes) != "" {
		b.WriteString("\nNotes: " + strings.TrimSpace(*st.Notes))
	}
	if profile.Source != nil && strings.TrimSpace(*profile.Source) != "" {
		b.WriteString("\nSource: " + strings.TrimSpace(*profile.Source))
	}
	return b.String(), nil
}

func renderLookupCropTargetsMulti(ctx context.Context, q db.Querier, farmID int64, question string, cropMentions, unsupMentions []croplibrary.ResolvedMention) (string, error) {
	var b strings.Builder
	b.WriteString("lookup_crop_targets — multi-crop")
	compare := compareCropQuestion.MatchString(question)
	stages := defaultStagesForMulti(compare)

	profileCount := 0
	for _, m := range cropMentions {
		if profileCount >= maxMultiCropProfiles {
			break
		}
		profileID, ok := lookupBuiltinProfileID(ctx, q, farmID, m.Key)
		if !ok {
			b.WriteString(fmt.Sprintf("\n---\n%s: no built-in profile — clone from nearest cousin in Plants or request a profile.", m.DisplayName))
			continue
		}
		profile, err := q.GetCropProfile(ctx, profileID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return "", err
		}
		b.WriteString(fmt.Sprintf("\n---\n%s (%s)", profile.DisplayName, profile.CropKey))
		for _, stageName := range stages {
			st, err := q.GetCropProfileStage(ctx, db.GetCropProfileStageParams{
				CropProfileID: profileID,
				Stage:         stageName,
			})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					continue
				}
				return "", err
			}
			b.WriteString(fmt.Sprintf("\n  %s: EC %s mS/cm", stageName, formatEcRange(st.EcMin, st.EcTarget, st.EcMax)))
			if st.PhotoperiodHrs.Valid {
				b.WriteString(fmt.Sprintf("; photoperiod %s h", formatNumeric(st.PhotoperiodHrs)))
			}
			if st.DliTarget.Valid {
				b.WriteString(fmt.Sprintf("; DLI %s mol/m²/day", formatNumeric(st.DliTarget)))
			}
		}
		profileCount++
	}

	unsupCount := 0
	for _, m := range unsupMentions {
		if unsupCount >= maxMultiUnsupportedCrops {
			break
		}
		b.WriteString("\n")
		b.WriteString(formatUnsupportedCropBlock(m))
		unsupCount++
	}

	if profileCount == 0 && unsupCount == 0 {
		return "lookup_crop_targets: no crop profile assigned to the active grow or plant. Offer to pick a profile in Start grow or Plants.", nil
	}
	return b.String(), nil
}

func formatUnsupportedCropBlock(m croplibrary.ResolvedMention) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("lookup_crop_targets — %s — not supported as a gr33n indoor fertigation crop.", m.DisplayName))
	if strings.TrimSpace(m.Reason) != "" {
		b.WriteString("\nReason: " + strings.TrimSpace(m.Reason))
	}
	if m.CousinOf != nil && strings.TrimSpace(*m.CousinOf) != "" {
		b.WriteString("\nNearest cousin crop key: " + strings.TrimSpace(*m.CousinOf))
	}
	b.WriteString("\nDo not state EC, pH, VPD, DLI, or photoperiod targets for this crop.")
	return b.String()
}

func splitMentions(mentions []croplibrary.ResolvedMention) (crops, unsupported []croplibrary.ResolvedMention) {
	for _, m := range mentions {
		switch m.Kind {
		case croplibrary.MentionCrop:
			crops = append(crops, m)
		case croplibrary.MentionUnsupported:
			unsupported = append(unsupported, m)
		}
	}
	return crops, unsupported
}

func defaultStagesForMulti(compare bool) []db.Gr33nfertigationGrowthStageEnum {
	if compare {
		return []db.Gr33nfertigationGrowthStageEnum{
			db.Gr33nfertigationGrowthStageEnumEarlyVeg,
			db.Gr33nfertigationGrowthStageEnumEarlyFlower,
		}
	}
	return []db.Gr33nfertigationGrowthStageEnum{
		db.Gr33nfertigationGrowthStageEnumEarlyVeg,
	}
}

func lookupBuiltinProfileID(ctx context.Context, q db.Querier, farmID int64, cropKey string) (int64, bool) {
	farmPtr := farmID
	p, err := q.GetCropProfileByKey(ctx, db.GetCropProfileByKeyParams{CropKey: cropKey, FarmID: &farmPtr})
	if err != nil {
		return 0, false
	}
	return p.ID, true
}

func resolveCropProfileContext(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) (profileID int64, stage db.Gr33nfertigationGrowthStageEnum, plantName string, err error) {
	var cycle db.Gr33nfertigationCropCycle
	var haveCycle bool

	if ref != nil {
		if ref.CropCycleID > 0 {
			cycle, err = q.GetCropCycleByID(ctx, ref.CropCycleID)
			if err == nil && cycle.FarmID == farmID {
				haveCycle = true
			}
		}
		if !haveCycle && strings.EqualFold(ref.Type, "zone") && ref.ID > 0 {
			cycle, err = q.GetActiveCropCycleForZone(ctx, ref.ID)
			if err == nil && cycle.FarmID == farmID {
				haveCycle = true
			}
		}
	}
	if !haveCycle {
		if zone, ok := resolveZoneForSummary(ctx, q, farmID, question, Snapshot{}); ok {
			cycle, err = q.GetActiveCropCycleForZone(ctx, zone.ID)
			if err == nil {
				haveCycle = true
			}
		}
	}
	if haveCycle {
		if cycle.CurrentStage != nil {
			stage = *cycle.CurrentStage
		}
		if cycle.PlantID != nil && *cycle.PlantID > 0 {
			plant, perr := q.GetPlant(ctx, *cycle.PlantID)
			if perr == nil {
				plantName = plant.DisplayName
				if plant.CropProfileID != nil {
					profileID = *plant.CropProfileID
				}
			}
		}
	}
	if profileID <= 0 {
		reg, rerr := defaultCropRegistry()
		if rerr == nil && reg != nil {
			for _, m := range reg.FindMentions(question) {
				if m.Kind != croplibrary.MentionCrop {
					continue
				}
				if id, ok := lookupBuiltinProfileID(ctx, q, farmID, m.Key); ok {
					profileID = id
					break
				}
			}
		}
	}
	return profileID, stage, plantName, nil
}

func formatEcRange(min, target, max pgtype.Numeric) string {
	if target.Valid {
		if min.Valid && max.Valid {
			return fmt.Sprintf("%s–%s (target %s)", formatNumeric(min), formatNumeric(max), formatNumeric(target))
		}
		return formatNumeric(target)
	}
	return formatRange(min, max, "mS/cm")
}

func formatPhRange(min, max pgtype.Numeric) string {
	return formatRange(min, max, "")
}

func formatRange(min, max pgtype.Numeric, unit string) string {
	if min.Valid && max.Valid {
		s := fmt.Sprintf("%s–%s", formatNumeric(min), formatNumeric(max))
		if unit != "" && unit != "mS/cm" {
			s += " " + unit
		}
		return s
	}
	if min.Valid {
		return "≥ " + formatNumeric(min)
	}
	if max.Valid {
		return "≤ " + formatNumeric(max)
	}
	return "—"
}

func formatNumeric(n pgtype.Numeric) string {
	if !n.Valid {
		return "—"
	}
	v := numericToFloat64(n)
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.2f", v)
}
