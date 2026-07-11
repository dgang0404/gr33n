// Phase 148 WS1 — citation-claim accuracy, duplicate enumeration, and unit-confusion
// detectors. These catch small-model failure modes that Phase 145's topical
// CitationAlignmentNote does not: a claim's [n] pointing at the wrong chunk,
// the same list item repeated under a different number, garbled digit/word
// merges, and pH values mislabeled with mS/cm (EC) units.

package farmguardian

import (
	"regexp"
	"strconv"
	"strings"
)

var garbledDigitWordRE = regexp.MustCompile(`\b\d(?:\.\d+)?([a-zA-Z]{5,})\b`)

var garbledAllowlist = map[string]struct{}{
	"gigabit": {}, "gigabyte": {}, "gigabytes": {}, "megabit": {}, "megabyte": {}, "megabytes": {},
}

// GarbledTokenNote flags digit-glued-to-word tokens like "0sourced" that
// indicate a broken generation (missing space / dropped token).
func GarbledTokenNote(answer string) string {
	matches := garbledDigitWordRE.FindAllStringSubmatch(answer, -1)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		word := strings.ToLower(m[1])
		if _, ok := garbledAllowlist[word]; ok {
			continue
		}
		return "garbled_token: " + m[0]
	}
	return ""
}

var numberedListItemRE = regexp.MustCompile(`(?m)^\s*(\d+)[.)]\s+(.+)$`)

// DuplicateListItemNote flags numbered list items that repeat the same
// subject under a different number (e.g. the same alert listed twice).
func DuplicateListItemNote(answer string) string {
	matches := numberedListItemRE.FindAllStringSubmatch(answer, -1)
	if len(matches) < 2 {
		return ""
	}
	type item struct {
		num   string
		words map[string]struct{}
	}
	items := make([]item, 0, len(matches))
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		items = append(items, item{num: m[1], words: significantWordSet(m[2])})
	}
	for i := 0; i < len(items); i++ {
		if len(items[i].words) == 0 {
			continue
		}
		for j := i + 1; j < len(items); j++ {
			if items[i].num == items[j].num || len(items[j].words) == 0 {
				continue
			}
			if jaccard(items[i].words, items[j].words) >= 0.4 {
				return "duplicate_list_item: items " + items[i].num + " and " + items[j].num
			}
		}
	}
	return ""
}

func significantWordSet(s string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, w := range tokenizeWords(s) {
		if len(w) < 4 {
			continue
		}
		if _, skip := citationAlignStopwords[w]; skip {
			continue
		}
		out[w] = struct{}{}
	}
	return out
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	for w := range a {
		if _, ok := b[w]; ok {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

var claimBracketRE = regexp.MustCompile(`\[(\d+)\]`)

// CitationClaimMismatchNote flags [n] references whose nearby claim terms are
// absent from cite n's excerpt but present in a *different* cite's excerpt —
// a strong signal the model attached the wrong source number to a claim.
// Terms shared by every cited excerpt (e.g. a repeated zone name) are not
// discriminating and are excluded so shared context doesn't mask a mismatch.
func CitationClaimMismatchNote(answer string, cites []CitationSummary) string {
	if len(cites) < 2 {
		return ""
	}
	termsByRef := make(map[int]map[string]struct{}, len(cites))
	docFreq := make(map[string]int)
	for _, c := range cites {
		set := significantWordSet(c.Excerpt)
		termsByRef[c.Ref] = set
		for t := range set {
			docFreq[t]++
		}
	}

	locs := claimBracketRE.FindAllStringSubmatchIndex(answer, -1)
	for _, loc := range locs {
		refStr := answer[loc[2]:loc[3]]
		ref, err := strconv.Atoi(refStr)
		if err != nil {
			continue
		}
		ownTerms, ok := termsByRef[ref]
		if !ok {
			continue
		}
		claimStart := loc[0] - 70
		if claimStart < 0 {
			claimStart = 0
		}
		claim := answer[claimStart:loc[0]]
		discTerms := discriminatingTerms(significantWordSet(claim), docFreq, len(cites))
		if len(discTerms) == 0 {
			continue
		}
		if hasAnyTerm(ownTerms, discTerms) {
			continue
		}
		for otherRef, otherTerms := range termsByRef {
			if otherRef == ref {
				continue
			}
			if hasAnyTerm(otherTerms, discTerms) {
				return "citation_number_mismatch: claim near [" + refStr + "] matches [" + strconv.Itoa(otherRef) + "] instead"
			}
		}
	}
	return ""
}

// discriminatingTerms keeps claim terms that don't appear in every cited
// excerpt (i.e. terms with signal about which specific source is meant).
func discriminatingTerms(claimTerms map[string]struct{}, docFreq map[string]int, citeCount int) map[string]struct{} {
	out := make(map[string]struct{})
	for t := range claimTerms {
		if df := docFreq[t]; df > 0 && df < citeCount {
			out[t] = struct{}{}
		}
	}
	return out
}

func hasAnyTerm(set, terms map[string]struct{}) bool {
	for t := range terms {
		if _, ok := set[t]; ok {
			return true
		}
	}
	return false
}

var phMsCmRangeRE = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*[–-]\s*(\d+(?:\.\d+)?)\s*mS/cm`)

// ECPHUnitConfusionNote flags a numeric range labeled mS/cm (EC units) in the
// answer when that same range appears in a cited excerpt as a pH value —
// i.e. the model relabeled a pH target with EC units.
func ECPHUnitConfusionNote(answer string, cites []CitationSummary) string {
	matches := phMsCmRangeRE.FindAllStringSubmatch(answer, -1)
	if len(matches) == 0 {
		return ""
	}
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		rangeStr := m[1] + "–" + m[2]
		rangeStrHyphen := m[1] + "-" + m[2]
		for _, c := range cites {
			excerpt := c.Excerpt
			idx := strings.Index(excerpt, rangeStr)
			if idx < 0 {
				idx = strings.Index(excerpt, rangeStrHyphen)
			}
			if idx < 0 {
				continue
			}
			context := strings.ToLower(excerpt)
			windowStart := idx - 20
			if windowStart < 0 {
				windowStart = 0
			}
			windowEnd := idx + len(rangeStr) + 10
			if windowEnd > len(context) {
				windowEnd = len(context)
			}
			window := strings.ToLower(excerpt[windowStart:windowEnd])
			if strings.Contains(window, "ph") && !strings.Contains(window, "ec ") && !strings.Contains(window, "mS/cm") {
				return "ph_ec_unit_confusion: " + rangeStr + " labeled mS/cm but sourced as pH"
			}
		}
	}
	return ""
}

// MissingNumberedCitationsNote flags alert-style numbered lists that omit all
// [n] bracket markers (run #8: markdown links instead of [1]/[2]/[3]).
func MissingNumberedCitationsNote(answer string) string {
	if len(claimBracketRE.FindAllString(answer, -1)) > 0 {
		return ""
	}
	items := numberedListItemRE.FindAllStringSubmatch(answer, -1)
	if len(items) < 2 {
		return ""
	}
	if !looksLikeAlertSummaryAnswer(answer) {
		return ""
	}
	return "missing_numbered_citations"
}

// truncatedTailRE matches a word fragment glued to trailing digits (and an
// optional colon) at the very end of an answer — a generation was cut off
// mid-token, e.g. "...consistent and ade0:" (Phase 152 WS2, live UI run).
// A short allowlist covers legitimate unit-style words so "2CO2" style
// chemistry mentions never trip this at the boundary.
var truncatedTailRE = regexp.MustCompile(`(?i)\b([a-z]{2,}\d{1,2}):?\s*$`)

var truncatedTailAllowlist = map[string]struct{}{
	"co2": {}, "h2o": {}, "no2": {}, "so2": {}, "n2": {}, "o2": {}, "3d": {}, "24x7": {},
}

// TruncatedAnswerTailNote flags an answer whose final token looks like a
// broken mid-word cutoff rather than a completed sentence.
func TruncatedAnswerTailNote(answer string) string {
	trimmed := strings.TrimSpace(answer)
	if trimmed == "" {
		return ""
	}
	m := truncatedTailRE.FindStringSubmatch(trimmed)
	if m == nil {
		return ""
	}
	if _, ok := truncatedTailAllowlist[strings.ToLower(m[1])]; ok {
		return ""
	}
	return "truncated_answer_tail: " + m[0]
}

var weekOrDayClaimRE = regexp.MustCompile(`(?i)\b(week|day)\s+\d+\b`)

// UncitedTimelineClaimNote flags a "Week N" / "Day N" progress claim that has
// no [n] citation nearby. Crop-cycle chunks never carry a week/day-of-cycle
// field themselves (only started_at/stage), so any such claim can only be
// grounded by citing the specific task/note chunk it came from — otherwise
// it's likely borrowed from an unrelated (e.g. already-completed) chunk.
func UncitedTimelineClaimNote(answer string) string {
	loc := weekOrDayClaimRE.FindStringIndex(answer)
	if loc == nil {
		return ""
	}
	winStart := loc[0] - 60
	if winStart < 0 {
		winStart = 0
	}
	winEnd := loc[1] + 60
	if winEnd > len(answer) {
		winEnd = len(answer)
	}
	if claimBracketRE.MatchString(answer[winStart:winEnd]) {
		return ""
	}
	return "uncited_timeline_claim: " + strings.TrimSpace(answer[loc[0]:loc[1]])
}

var assumptionHedgeRE = regexp.MustCompile(`(?i)\b(assuming|if we assume|let'?s assume|for the sake of (?:estimate|argument)|hypothetically|for estimation purposes)\b`)

// InventedAssumptionMathNote flags a numeric claim justified by a hedge
// phrase like "assuming an average yield density" — the model is disclosing,
// in its own words, that a number was derived rather than sourced. The
// system prompt forbids inventing facts, so any assumption-qualified number
// is itself the violation.
func InventedAssumptionMathNote(answer string) string {
	loc := assumptionHedgeRE.FindStringIndex(answer)
	if loc == nil {
		return ""
	}
	winStart := loc[0] - 80
	if winStart < 0 {
		winStart = 0
	}
	winEnd := loc[1] + 80
	if winEnd > len(answer) {
		winEnd = len(answer)
	}
	window := answer[winStart:winEnd]
	if !regexp.MustCompile(`\d`).MatchString(window) {
		return ""
	}
	return "invented_assumption_math: " + strings.TrimSpace(answer[loc[0]:loc[1]])
}

// AnswerAccuracyNote runs all Phase 148/151/152 accuracy detectors and
// returns the first failure reason, or "" when the answer passes all checks.
func AnswerAccuracyNote(answer string, cites []CitationSummary) string {
	if note := MissingNumberedCitationsNote(answer); note != "" {
		return note
	}
	if note := MultipleCitationsPerListItemNote(answer); note != "" {
		return note
	}
	if note := GarbledTokenNote(answer); note != "" {
		return note
	}
	if note := TruncatedAnswerTailNote(answer); note != "" {
		return note
	}
	if note := DuplicateListItemNote(answer); note != "" {
		return note
	}
	if note := CitationClaimMismatchNote(answer, cites); note != "" {
		return note
	}
	if note := ECPHUnitConfusionNote(answer, cites); note != "" {
		return note
	}
	if note := InventedAssumptionMathNote(answer); note != "" {
		return note
	}
	if len(cites) > 0 {
		if note := UncitedTimelineClaimNote(answer); note != "" {
			return note
		}
	}
	return ""
}
