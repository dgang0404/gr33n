// Phase 145 WS2 — citation corpus alignment for grounded answers.

package farmguardian

import (
	"strings"
	"unicode"
)

// CitationSummary is a compact citation row stored on QA archives (Phase 145).
type CitationSummary struct {
	Ref        int    `json:"ref"`
	SourceType string `json:"source_type,omitempty"`
	Excerpt    string `json:"excerpt,omitempty"`
}

var citationAlignStopwords = map[string]struct{}{
	"about": {}, "after": {}, "according": {}, "their": {}, "there": {},
	"these": {}, "those": {}, "which": {}, "would": {}, "could": {},
	"should": {}, "sources": {}, "source": {}, "field": {}, "guide": {},
	"documentation": {}, "operational": {},
}

var offTopicCitationMarkers = []string{
	"endocrine-disruptor", "endocrine disruptor", "endocrine_disruptor",
	"lake erie", "lake superior", "mississippi river", "typha latifolia",
	"wildlife", "hormonal systems", "aquatic ecosystems",
}

var agronomyQuestionMarkers = []string{
	"ec", "ph", "leafy green", "lettuce", "kale", "spinach", "crop",
	"fertigation", "nutrient", "hydro", "ms/cm",
}

// CitationAlignmentNote returns a failure reason when answer tail or cited corpus drifts off question topic.
func CitationAlignmentNote(question, answer string, cites []CitationSummary) string {
	if len(cites) == 0 || strings.TrimSpace(answer) == "" {
		return ""
	}
	q := strings.ToLower(strings.TrimSpace(question))
	corpus := citedCorpusLower(cites)
	if corpus == "" {
		return ""
	}
	if agronomyQuestion(q) {
		for _, marker := range offTopicCitationMarkers {
			if strings.Contains(corpus, marker) {
				_, tail := SplitAnswerOpeningTail(answer)
				check := strings.ToLower(answer)
				if tail != "" {
					check = strings.ToLower(tail)
				}
				if strings.Contains(check, marker) || strings.Contains(strings.ToLower(answer), marker) {
					return "citation corpus misaligned with question (off-topic excerpts cited)"
				}
			}
		}
		if !corpusSupportsQuestion(q, corpus) {
			return "cited excerpts do not support question topic"
		}
	}
	if note := uncitedTailNote(answer, corpus); note != "" {
		return note
	}
	return ""
}

func agronomyQuestion(q string) bool {
	for _, m := range agronomyQuestionMarkers {
		if strings.Contains(q, m) {
			return true
		}
	}
	return false
}

func citedCorpusLower(cites []CitationSummary) string {
	var b strings.Builder
	for _, c := range cites {
		if s := strings.TrimSpace(c.Excerpt); s != "" {
			b.WriteString(strings.ToLower(s))
			b.WriteByte(' ')
		}
	}
	return strings.TrimSpace(b.String())
}

func corpusSupportsQuestion(q, corpus string) bool {
	matched := 0
	need := 0
	for _, m := range agronomyQuestionMarkers {
		if !strings.Contains(q, m) {
			continue
		}
		need++
		if strings.Contains(corpus, m) {
			matched++
		}
	}
	if need == 0 {
		return true
	}
	return matched > 0
}

func uncitedTailNote(answer, corpus string) string {
	_, tail := SplitAnswerOpeningTail(answer)
	if tail == "" || len(tail) < 120 {
		return ""
	}
	terms := significantTailTerms(tail)
	if len(terms) == 0 {
		return ""
	}
	uncited := 0
	for _, term := range terms {
		if strings.Contains(corpus, term) {
			continue
		}
		uncited++
		if uncited >= 2 {
			return "answer tail mentions terms absent from cited excerpts"
		}
	}
	return ""
}

func significantTailTerms(tail string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, word := range tokenizeWords(tail) {
		if len(word) < 5 {
			continue
		}
		if _, skip := citationAlignStopwords[word]; skip {
			continue
		}
		if _, ok := seen[word]; ok {
			continue
		}
		seen[word] = struct{}{}
		out = append(out, word)
	}
	return out
}

func tokenizeWords(s string) []string {
	s = strings.ToLower(s)
	var (
		cur  strings.Builder
		out  []string
		flush = func() {
			if cur.Len() == 0 {
				return
			}
			out = append(out, cur.String())
			cur.Reset()
		}
	)
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}
