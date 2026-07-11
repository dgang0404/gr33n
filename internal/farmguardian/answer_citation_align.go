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

// AnswerUncitedTailTrim records tail removal when grounded answers drift off citations.
type AnswerUncitedTailTrim struct {
	Trimmed      bool `json:"uncited_tail_trimmed,omitempty"`
	CharsRemoved int  `json:"uncited_tail_chars_removed,omitempty"`
}

// EcphCropDriftNote flags off-topic fruiting crops appended to leafy-greens EC/pH answers.
func EcphCropDriftNote(prompt, answer string) string {
	p := strings.ToLower(strings.TrimSpace(prompt))
	if !leafyGreensECPHPrompt(p) {
		return ""
	}
	_, tail := SplitAnswerOpeningTail(answer)
	check := strings.ToLower(answer)
	if tail != "" {
		check = strings.ToLower(tail)
	}
	for _, crop := range []string{
		"blueberry", "strawberry", "cannabis", "tomato", "cucumber", "pepper", "melon",
	} {
		if strings.Contains(check, crop) && !strings.Contains(p, crop) {
			return "topic_drift: off-topic crop in leafy greens EC/pH answer"
		}
	}
	return ""
}

func leafyGreensECPHPrompt(p string) bool {
	for _, m := range []string{"leafy green", "lettuce", "kale", "spinach"} {
		if strings.Contains(p, m) {
			return true
		}
	}
	return strings.Contains(p, "ec") && strings.Contains(p, "ph")
}

func shouldTrimUncitedTail(prompt, answer, corpus string) bool {
	if uncitedTailNote(answer, corpus) != "" {
		return true
	}
	q := strings.ToLower(strings.TrimSpace(prompt))
	if agronomyQuestion(q) && EcphCropDriftNote(prompt, answer) != "" {
		return true
	}
	return false
}

// TrimUncitedTail removes paragraphs that drift off cited excerpts or introduce
// off-topic crops into leafy-greens EC/pH answers (Phase 161).
func TrimUncitedTail(answer, prompt string, cites []CitationSummary) (string, AnswerUncitedTailTrim) {
	orig := answer
	if strings.TrimSpace(orig) == "" || len(cites) == 0 {
		return orig, AnswerUncitedTailTrim{}
	}
	corpus := citedCorpusLower(cites)
	if !shouldTrimUncitedTail(prompt, orig, corpus) {
		return orig, AnswerUncitedTailTrim{}
	}
	trimmed := trimAnswerBeforeDrift(orig, prompt, corpus)
	if trimmed == orig || strings.TrimSpace(trimmed) == "" {
		return orig, AnswerUncitedTailTrim{}
	}
	return trimmed, AnswerUncitedTailTrim{
		Trimmed:      true,
		CharsRemoved: len(orig) - len(trimmed),
	}
}

func trimAnswerBeforeDrift(answer, prompt, corpus string) string {
	opening, tail := SplitAnswerOpeningTail(answer)
	if tail != "" && opening != "" {
		return opening
	}
	paras := strings.Split(answer, "\n\n")
	for len(paras) > 1 {
		candidate := strings.Join(paras[:len(paras)-1], "\n\n")
		if !shouldTrimUncitedTail(prompt, candidate, corpus) {
			return strings.TrimSpace(candidate)
		}
		paras = paras[:len(paras)-1]
	}
	return answer
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
