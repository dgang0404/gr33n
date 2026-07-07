// Phase 143 WS2 — strip hallucinated citation URLs from assistant answers.

package farmguardian

import (
	"regexp"
	"strings"
)

var markdownLinkRE = regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`)

// CitationURLSanitize records fake-link rewrites applied before turn persist.
type CitationURLSanitize struct {
	Sanitized      bool `json:"citation_urls_sanitized,omitempty"`
	LinksRewritten int  `json:"citation_links_rewritten,omitempty"`
}

// SanitizeCitationURLs rewrites markdown links to gr33n.com, gr33n-docs, or bare # anchors as plain labels.
func SanitizeCitationURLs(answer string) (string, CitationURLSanitize) {
	if strings.TrimSpace(answer) == "" {
		return answer, CitationURLSanitize{}
	}
	rewritten := 0
	out := markdownLinkRE.ReplaceAllStringFunc(answer, func(match string) string {
		sub := markdownLinkRE.FindStringSubmatch(match)
		if len(sub) != 3 {
			return match
		}
		label := strings.TrimSpace(sub[1])
		url := strings.TrimSpace(sub[2])
		if !isHallucinatedCitationURL(url) {
			return match
		}
		rewritten++
		if label == "" {
			return ""
		}
		return label
	})
	if rewritten == 0 {
		return answer, CitationURLSanitize{}
	}
	return out, CitationURLSanitize{Sanitized: true, LinksRewritten: rewritten}
}

// AnswerContainsFakeCitationURL reports whether answer still has hallucinated markdown links.
func AnswerContainsFakeCitationURL(answer string) bool {
	for _, sub := range markdownLinkRE.FindAllStringSubmatch(answer, -1) {
		if len(sub) == 3 && isHallucinatedCitationURL(strings.TrimSpace(sub[2])) {
			return true
		}
	}
	return false
}

func isHallucinatedCitationURL(url string) bool {
	u := strings.ToLower(strings.TrimSpace(url))
	if u == "" || u == "#" {
		return true
	}
	if strings.HasPrefix(u, "#") {
		return true
	}
	return strings.Contains(u, "gr33n.com") || strings.Contains(u, "gr33n-docs")
}
