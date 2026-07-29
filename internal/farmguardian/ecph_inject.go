// Phase 211.06 WS5 — inject explicit pH from cited chunks when answer has EC but omits "pH".

package farmguardian

import (
	"regexp"
	"strings"

	db "gr33n-api/internal/db"
)

var phSnippetRE = regexp.MustCompile(`(?i)\bpH\b[^.\n]{0,80}`)

// InjectPHFromChunks appends a short pH line from RAG chunk text when the answer
// already discusses EC (or cites nutrient docs) but never says "pH".
func InjectPHFromChunks(answer string, chunks []db.SearchRagNearestNeighborsFilteredRow) (string, bool) {
	a := strings.TrimSpace(answer)
	if a == "" || len(chunks) == 0 {
		return answer, false
	}
	lower := strings.ToLower(a)
	if strings.Contains(lower, "ph") {
		return answer, false
	}
	hasEC := strings.Contains(lower, "ec") || strings.Contains(lower, "ms/cm") ||
		strings.Contains(lower, "electrical conductivity")
	if !hasEC {
		return answer, false
	}
	snippet := firstPHSnippetFromChunks(chunks)
	if snippet == "" {
		return answer, false
	}
	return strings.TrimRight(a, " \t\r\n") + "\n\nFrom cited documentation: " + snippet, true
}

// CitationExcerptsContain reports whether any citation excerpt contains needle (case-insensitive).
func CitationExcerptsContain(cites []CitationSummary, needle string) bool {
	n := strings.ToLower(strings.TrimSpace(needle))
	if n == "" {
		return false
	}
	for _, c := range cites {
		if strings.Contains(strings.ToLower(c.Excerpt), n) {
			return true
		}
	}
	return false
}

func firstPHSnippetFromChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) string {
	for _, ch := range chunks {
		text := strings.TrimSpace(ch.ContentText)
		if text == "" {
			continue
		}
		if m := phSnippetRE.FindString(text); strings.TrimSpace(m) != "" {
			s := strings.TrimSpace(m)
			s = strings.TrimRight(s, " ,;:")
			if len(s) > 120 {
				s = s[:120] + "…"
			}
			return s
		}
	}
	return ""
}
