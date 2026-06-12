// Phase 106 — lookup_crop_symptoms read tool (deficiency & pest symptom catalog).

package farmguardian

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
)

// SymptomGroundingRule is injected into grounded chat when symptom tools may run.
const SymptomGroundingRule = `Symptom catalog (Phase 106): NEVER diagnose from leaf appearance or photos alone. Use lookup_crop_symptoms for ranked hypotheses and measurable checks. Pair with lookup_crop_targets for EC/pH/VPD numbers — structured profile rows beat RAG narrative. Offer inspection steps; do not prescribe pesticides or medical advice.`

var lookupCropSymptomsIntent = regexp.MustCompile(`(?i)\b(yellow|yellowing|brown|browning|spot|spots|spotting|wilting|wilt|drooping|curl|curling|cupping|tip burn|burnt tips|necrotic|deficien|chlorosis|interveinal|purple stem|mold|mildew|powdery|pest|chewing|holes in leaves|what'?s wrong|what is wrong|looks sick|unhealthy leaves|dying leaves|blossom.?end rot|necrosis)\b`)

var symptomFoliageIntent = regexp.MustCompile(`(?i)\b(leaves|foliage|plant looks|my plant)\b`)

var symptomKeywordHints = []struct {
	re   *regexp.Regexp
	keys []string
}{
	{regexp.MustCompile(`(?i)interveinal|between veins`), []string{"interveinal_yellowing"}},
	{regexp.MustCompile(`(?i)tip burn|burnt tips|brown tips|necrotic tips`), []string{"tip_burn"}},
	{regexp.MustCompile(`(?i)lower leaves|bottom leaves|old leaves`), []string{"yellow_lower_leaves", "interveinal_yellowing"}},
	{regexp.MustCompile(`(?i)purple`), []string{"purple_stems"}},
	{regexp.MustCompile(`(?i)wilt|drooping|limp`), []string{"wilting"}},
	{regexp.MustCompile(`(?i)spot|lesion|blotch`), []string{"leaf_spotting"}},
	{regexp.MustCompile(`(?i)chew|holes|notch`), []string{"chewing_damage"}},
	{regexp.MustCompile(`(?i)powdery|white coat|white dust`), []string{"powdery_white_coating"}},
	{regexp.MustCompile(`(?i)curl|cupping|twist`), []string{"leaf_curl"}},
	{regexp.MustCompile(`(?i)blossom.?end`), []string{"blossom_end_rot"}},
	{regexp.MustCompile(`(?i)yellow`), []string{"interveinal_yellowing", "yellow_lower_leaves"}},
}

// LookupCropSymptoms runs the lookup_crop_symptoms read tool (exported for smokes).
func LookupCropSymptoms(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) (string, error) {
	return renderLookupCropSymptoms(ctx, q, farmID, question, ref)
}

func shouldRunLookupCropSymptomsIntent(question string, ref *ContextRef) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if lookupCropSymptomsIntent.MatchString(q) {
		return true
	}
	if symptomFoliageIntent.MatchString(q) {
		if reg, err := defaultCropRegistry(); err == nil && reg != nil && len(reg.FindMentions(q)) > 0 {
			return true
		}
	}
	return false
}

func renderLookupCropSymptoms(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) (string, error) {
	cropKey, category, cropLabel, err := resolveSymptomCropContext(ctx, q, farmID, question, ref)
	if err != nil {
		return "", err
	}

	var rows []db.Gr33ncropsAgronomySymptomEntry
	if cropKey != "" || category != "" {
		rows, err = q.ListAgronomySymptomsForCrop(ctx, db.ListAgronomySymptomsForCropParams{
			CropKey:  cropKey,
			Category: nullableCategory(category),
		})
	} else {
		rows, err = q.ListAgronomySymptomEntries(ctx)
	}
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "lookup_crop_symptoms: no symptom catalog rows — run phase 106 migration.", nil
	}

	ranked := rankSymptomEntries(rows, question)
	if len(ranked) == 0 {
		ranked = rows
	}
	if len(ranked) > 4 {
		ranked = ranked[:4]
	}

	var b strings.Builder
	b.WriteString("lookup_crop_symptoms")
	if cropLabel != "" {
		b.WriteString(" — " + cropLabel)
	} else if cropKey != "" {
		b.WriteString(" — crop_key=" + cropKey)
	}
	b.WriteString("\nHypotheses only — confirm with EC/pH/VPD checks below.")

	for i, row := range ranked {
		if i >= 3 {
			break
		}
		b.WriteString(fmt.Sprintf("\n\n### %s (%s)", row.DisplayName, row.SymptomKey))
		if row.SeverityHint != nil && strings.TrimSpace(*row.SeverityHint) != "" {
			b.WriteString(" · severity " + strings.TrimSpace(*row.SeverityHint))
		}
		body := strings.TrimSpace(row.BodyMd)
		if body != "" {
			b.WriteString("\n")
			b.WriteString(body)
		}
	}

	targetsBlock, err := renderLookupCropTargets(ctx, q, farmID, question, ref)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(targetsBlock) != "" {
		b.WriteString("\n\n---\n")
		b.WriteString(targetsBlock)
	}
	return b.String(), nil
}

func nullableCategory(category string) *string {
	category = strings.TrimSpace(category)
	if category == "" {
		return nil
	}
	return &category
}

func resolveSymptomCropContext(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) (cropKey, category, label string, err error) {
	profileID, _, plantName, activeMissingPlant, err := resolveCropProfileContext(ctx, q, farmID, question, ref)
	if err != nil {
		return "", "", "", err
	}
	if activeMissingPlant {
		return "", "", "", nil
	}
	if profileID > 0 {
		profile, err := q.GetCropProfile(ctx, profileID)
		if err == nil {
			cropKey = strings.TrimSpace(profile.CropKey)
			label = profile.DisplayName
			if plantName != "" {
				label = plantName + " (" + profile.DisplayName + ")"
			}
			if entry, err := q.GetCropCatalogEntry(ctx, cropKey); err == nil && entry.Category != nil {
				category = strings.TrimSpace(*entry.Category)
			}
			return cropKey, category, label, nil
		}
	}
	reg, rerr := defaultCropRegistry()
	if rerr == nil && reg != nil {
		for _, m := range reg.FindMentions(question) {
			if m.Kind != croplibrary.MentionCrop {
				continue
			}
			cropKey = m.Key
			label = m.DisplayName
			if entry, err := q.GetCropCatalogEntry(ctx, cropKey); err == nil && entry.Category != nil {
				category = strings.TrimSpace(*entry.Category)
			}
			return cropKey, category, label, nil
		}
	}
	return "", "", "", nil
}

func rankSymptomEntries(rows []db.Gr33ncropsAgronomySymptomEntry, question string) []db.Gr33ncropsAgronomySymptomEntry {
	q := strings.ToLower(question)
	scored := make([]struct {
		row   db.Gr33ncropsAgronomySymptomEntry
		score int
	}, 0, len(rows))

	prefer := map[string]int{}
	for _, hint := range symptomKeywordHints {
		if hint.re.MatchString(question) {
			for i, k := range hint.keys {
				prefer[k] += 10 - i
			}
		}
	}

	for _, row := range rows {
		score := prefer[row.SymptomKey]
		key := strings.ToLower(row.SymptomKey)
		name := strings.ToLower(row.DisplayName)
		if strings.Contains(q, strings.ReplaceAll(key, "_", " ")) {
			score += 8
		}
		for _, token := range strings.FieldsFunc(q, func(r rune) bool {
			return r <= ' ' || r == ',' || r == '.'
		}) {
			if len(token) < 4 {
				continue
			}
			if strings.Contains(name, token) || strings.Contains(strings.ToLower(row.BodyMd), token) {
				score += 2
			}
		}
		if score > 0 {
			scored = append(scored, struct {
				row   db.Gr33ncropsAgronomySymptomEntry
				score int
			}{row, score})
		}
	}
	if len(scored) == 0 {
		return nil
	}
	sortSymptomScores(scored)
	out := make([]db.Gr33ncropsAgronomySymptomEntry, len(scored))
	for i, s := range scored {
		out[i] = s.row
	}
	return out
}

func sortSymptomScores(scored []struct {
	row   db.Gr33ncropsAgronomySymptomEntry
	score int
}) {
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
}
