// Phase 151 WS5 — post-generation [n] injection when the model lists alerts in
// order but omits bracket citation markers (run #8 failure mode).

package farmguardian

import (
	"regexp"
	"strconv"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/synthesis"
)

var numberedListLineRE = regexp.MustCompile(`(?m)^(\s*\d+[.)]\s+)(.+)$`)

// InjectAlertCitationRefs appends [1]…[n] to numbered list lines when the answer
// has 2+ alert chunks, no existing [n] refs, and a numbered alert-style list.
// Returns the (possibly modified) answer and whether any injection occurred.
func InjectAlertCitationRefs(answer string, chunks []db.SearchRagNearestNeighborsFilteredRow) (string, bool) {
	if countAlertChunks(chunks) < 2 || strings.TrimSpace(answer) == "" {
		return answer, false
	}
	if len(synthesis.RefNumbersInAnswer(answer)) > 0 {
		return answer, false
	}
	if !looksLikeAlertSummaryAnswer(answer) {
		return answer, false
	}

	lines := strings.Split(answer, "\n")
	injected := false
	itemNum := 0
	alertCount := countAlertChunks(chunks)
	for i, line := range lines {
		m := numberedListLineRE.FindStringSubmatch(line)
		if len(m) < 3 {
			continue
		}
		itemNum++
		if itemNum > alertCount {
			break
		}
		body := strings.TrimSpace(m[2])
		if claimBracketRE.MatchString(body) {
			continue
		}
		lines[i] = m[1] + body + " [" + strconv.Itoa(itemNum) + "]"
		injected = true
	}
	if !injected {
		return answer, false
	}
	return strings.Join(lines, "\n"), true
}

func looksLikeAlertSummaryAnswer(answer string) bool {
	lower := strings.ToLower(answer)
	return strings.Contains(lower, "alert") ||
		strings.Contains(lower, "humidity") ||
		strings.Contains(lower, "ohn") ||
		strings.Contains(lower, "photoperiod") ||
		strings.Contains(lower, "unread")
}
