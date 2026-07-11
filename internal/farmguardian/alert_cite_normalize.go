// Phase 151 follow-up — normalize stray [n] markers on alert list items (run #9:
// item 2 cited [2] correctly then appended "platform docs [3]").

package farmguardian

import (
	"regexp"
	"strconv"
	"strings"

	db "gr33n-api/internal/db"
)

var multiSpaceRE = regexp.MustCompile(`  +`)

// NormalizeAlertListCitations strips extra [n] markers from numbered alert list
// items so item N keeps only [N]. Returns the answer and whether it changed.
func NormalizeAlertListCitations(answer string, chunks []db.SearchRagNearestNeighborsFilteredRow) (string, bool) {
	if !hasOnlyAlertChunks(chunks) || strings.TrimSpace(answer) == "" {
		return answer, false
	}

	lines := strings.Split(answer, "\n")
	changed := false
	itemNum := 0
	alertCount := len(chunks)
	for i, line := range lines {
		m := numberedListLineRE.FindStringSubmatch(line)
		if len(m) < 3 {
			continue
		}
		itemNum++
		if itemNum > alertCount {
			break
		}
		body, itemChanged := keepOnlyListItemRef(m[2], itemNum)
		if itemChanged {
			lines[i] = m[1] + body
			changed = true
		}
	}
	if !changed {
		return answer, false
	}
	return strings.Join(lines, "\n"), true
}

func hasOnlyAlertChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) bool {
	if len(chunks) < 2 {
		return false
	}
	for _, c := range chunks {
		if c.SourceType != SourceTypeAlertNotification {
			return false
		}
	}
	return true
}

func keepOnlyListItemRef(body string, wantRef int) (string, bool) {
	refs := claimBracketRE.FindAllStringSubmatch(body, -1)
	if len(refs) == 0 {
		return body, false
	}
	want := strconv.Itoa(wantRef)
	hasWant := false
	changed := false
	out := claimBracketRE.ReplaceAllStringFunc(body, func(match string) string {
		sub := claimBracketRE.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		if sub[1] == want {
			hasWant = true
			return match
		}
		changed = true
		return ""
	})
	if !changed {
		return body, false
	}
	out = strings.TrimSpace(multiSpaceRE.ReplaceAllString(out, " "))
	out = strings.TrimRight(out, " ,;:")
	if !hasWant {
		out = strings.TrimSpace(out) + " [" + want + "]"
	}
	return out, true
}

// MultipleCitationsPerListItemNote flags numbered list items that cite more
// than one distinct [n] (run #9 item 2: [2] and [3]).
func MultipleCitationsPerListItemNote(answer string) string {
	matches := numberedListItemRE.FindAllStringSubmatch(answer, -1)
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		seen := make(map[string]struct{})
		for _, sub := range claimBracketRE.FindAllStringSubmatch(m[2], -1) {
			if len(sub) < 2 {
				continue
			}
			seen[sub[1]] = struct{}{}
		}
		if len(seen) >= 2 {
			return "multiple_citations_per_list_item: item " + m[1]
		}
	}
	return ""
}
