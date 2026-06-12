package synthesis

import (
	"regexp"
	"strings"

	db "gr33n-api/internal/db"
)

const structuredTruthRAGBlock = `Structured truth (Phase 97): lookup_crop_targets results in the system prompt above are authoritative for EC, pH, VPD, DLI, and photoperiod on this turn. Field-guide sources below may contain older narrative numbers — use read-tool mS/cm values only; cite field guides for qualitative context (deficiency signs, timing, mistakes) not conflicting EC/pH targets.`

var (
	nutrientMetricLineRE = regexp.MustCompile(`(?i)(ec|ph|vpd|dli|photoperiod|feed\s+strength|mS/cm|ms/cm|mol/m²|mol/m2|kpa)`)
	nutrientNumberRE     = regexp.MustCompile(`\d+\.?\d*`)
)

// StructuredTruthRAGBlock is appended when RAG chunks coexist with lookup_crop_targets output.
func StructuredTruthRAGBlock() string { return structuredTruthRAGBlock }

// StripNutrientNumbersFromChunks removes stale EC/pH/VPD/DLI numeric claims from field_guide chunks.
func StripNutrientNumbersFromChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) []db.SearchRagNearestNeighborsFilteredRow {
	if len(chunks) == 0 {
		return chunks
	}
	out := make([]db.SearchRagNearestNeighborsFilteredRow, len(chunks))
	copy(out, chunks)
	for i := range out {
		if !strings.EqualFold(strings.TrimSpace(out[i].SourceType), "field_guide") {
			continue
		}
		out[i].ContentText = stripNutrientNumbersFromText(out[i].ContentText)
	}
	return out
}

func stripNutrientNumbersFromText(text string) string {
	lines := strings.Split(text, "\n")
	changed := false
	for i, line := range lines {
		if !nutrientMetricLineRE.MatchString(line) || !nutrientNumberRE.MatchString(line) {
			continue
		}
		cleaned := nutrientNumberRE.ReplaceAllString(line, "—")
		if cleaned != line {
			lines[i] = cleaned + " (use lookup_crop_targets in system prompt for numeric targets)"
			changed = true
		}
	}
	if !changed {
		return text
	}
	return strings.Join(lines, "\n")
}
