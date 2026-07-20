// Phase 143 WS1 — strip instruction-template echoes from assistant answers.

package farmguardian

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// AnswerLeakTrim records instruction-leak detection applied before turn persist.
type AnswerLeakTrim struct {
	Trimmed      bool   `json:"leak_trimmed,omitempty"`
	CharsRemoved int    `json:"leak_chars_removed,omitempty"`
	Marker       string `json:"leak_marker,omitempty"`
}

var substituteQuestionPrefixRE = regexp.MustCompile(`(?i)^question:\s*`)

// AnswerIsSubstituteQuestionLeak reports when the model output a different exam-style Question instead of answering.
func AnswerIsSubstituteQuestionLeak(answer, question string) bool {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return false
	}
	rest := answer
	if substituteQuestionPrefixRE.MatchString(answer) {
		rest = strings.TrimSpace(substituteQuestionPrefixRE.ReplaceAllString(answer, ""))
	} else if !strings.HasPrefix(strings.ToLower(answer), "question") {
		return false
	}
	if len(rest) < 20 {
		return false
	}
	return normalizeLeakText(rest) != normalizeLeakText(question)
}

// TrimSubstituteQuestionLeak clears answers that are entirely a different Question prompt.
func TrimSubstituteQuestionLeak(answer, question string) (string, AnswerLeakTrim) {
	if !AnswerIsSubstituteQuestionLeak(answer, question) {
		return answer, AnswerLeakTrim{}
	}
	return "", AnswerLeakTrim{
		Trimmed:      true,
		CharsRemoved: len(strings.TrimSpace(answer)),
		Marker:       "substitute_question",
	}
}

// TrimInstructionLeak removes trailing prompt-template leaks (e.g. "## Your task", echoed Question:).
func TrimInstructionLeak(answer, question string) (string, AnswerLeakTrim) {
	orig := answer
	answer = strings.TrimRight(answer, " \t\r\n")
	if answer == "" {
		return orig, AnswerLeakTrim{}
	}

	cut := leakCutIndex(answer, question)
	if cut < 0 || cut >= len(answer) {
		return orig, AnswerLeakTrim{}
	}

	trimmed := strings.TrimRight(answer[:cut], " \t\r\n")
	if trimmed == "" {
		return orig, AnswerLeakTrim{}
	}

	return trimmed, AnswerLeakTrim{
		Trimmed:      true,
		CharsRemoved: len(answer) - len(trimmed),
		Marker:       leakMarkerAt(answer, cut),
	}
}

// leakTopMarkers are unambiguous prompt-template headings — real farm answers
// never contain these verbatim, so any match is cut regardless of context.
var leakTopMarkers = []string{"\n## your task", "## your task", "\n## instruction", "## instruction"}

// leakEssayTells mark a *different* few-shot template (a generic document-QA
// or essay-writing benchmark) leaking into the answer — e.g. the model
// completed "Question" with an unrelated "## Instruction> / Write an
// extensive essay... / Document: ..." block instead of answering the farm
// question (Phase 188, live turn: a "which zone should this task go in?"
// clarification came back discussing "The Great Gatsby" and a fabricated
// Faulkner novel). None of these phrases has any legitimate reason to appear
// in a gr33n farm answer.
var leakEssayTells = []string{
	"## instruction",
	"\ndocument:\n",
	"write an extensive essay",
	"write an essay",
}

func leakCutIndex(answer, question string) int {
	lower := strings.ToLower(answer)
	best := -1

	for _, marker := range leakTopMarkers {
		if idx := strings.Index(lower, marker); idx >= 0 {
			if best < 0 || idx < best {
				best = idx
				if marker[0] == '\n' {
					best++ // keep content before the blank line
				}
			}
		}
	}

	if idx := bareQuestionHeadingCutIndex(lower); idx >= 0 {
		if best < 0 || idx < best {
			best = idx
		}
	}

	if q := normalizeLeakText(question); q != "" {
		if idx := trailingQuestionEchoIndex(answer, q); idx >= 0 {
			if best < 0 || idx < best {
				best = idx
			}
		}
	}

	return best
}

// bareQuestionHeadingCutIndex finds a standalone "Question" line (no colon,
// no requirement that it echo the real question) that precedes one of
// leakEssayTells later in the answer. A bare "Question" heading alone is too
// common a word to cut on by itself, so this only fires once it's confirmed
// to be followed by an unrelated instruction-template tell.
func bareQuestionHeadingCutIndex(lower string) int {
	idx := strings.Index(lower, "\nquestion\n")
	if idx < 0 {
		return -1
	}
	tail := lower[idx:]
	for _, tell := range leakEssayTells {
		if strings.Contains(tail, tell) {
			return idx + 1 // keep content before the blank line
		}
	}
	return -1
}

func trailingQuestionEchoIndex(answer, normQuestion string) int {
	lower := strings.ToLower(answer)
	for _, prefix := range []string{"\nquestion:\n", "\nquestion: \n", "\nquestion:\r\n"} {
		idx := strings.LastIndex(lower, prefix)
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(answer[idx+len(prefix):])
		if normalizeLeakText(rest) == normQuestion {
			return idx
		}
	}
	return -1
}

func leakMarkerAt(answer string, cut int) string {
	if cut <= 0 || cut > len(answer) {
		return "## your task"
	}
	snip := strings.TrimSpace(answer[cut:])
	if len(snip) > 40 {
		snip = snip[:40]
	}
	return snip
}

func normalizeLeakText(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsSpace(r) {
			if b.Len() > 0 && b.String()[b.Len()-1] != ' ' {
				b.WriteRune(' ')
			}
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return strings.TrimSpace(b.String())
}

// AnswerLooksLikePromptLeak reports whether answer still contains a template leak marker.
func AnswerLooksLikePromptLeak(answer, question string) bool {
	_, meta := TrimInstructionLeak(answer, question)
	return meta.Trimmed
}

// AnswerMetaTrim records model self-correction / apology tails removed before persist.
type AnswerMetaTrim struct {
	Trimmed      bool   `json:"meta_correction_trimmed,omitempty"`
	CharsRemoved int    `json:"meta_correction_chars_removed,omitempty"`
	Marker       string `json:"meta_correction_marker,omitempty"`
}

var metaCorrectionMarkers = []string{
	"\ni apologize for misunderstanding",
	"\ni apologise for misunderstanding",
	"\ni apologize for the misunderstanding",
	"\nhere's an updated answer:",
	"\nhere is an updated answer:",
	"\nplease disregard",
	"\ndisregard any references",
}

// TrimMetaCorrection removes trailing model self-correction blocks (e.g. apology + "updated answer").
func TrimMetaCorrection(answer string) (string, AnswerMetaTrim) {
	orig := answer
	answer = strings.TrimRight(answer, " \t\r\n")
	if answer == "" {
		return orig, AnswerMetaTrim{}
	}
	lower := strings.ToLower(answer)
	best := -1
	marker := ""
	for _, m := range metaCorrectionMarkers {
		if idx := strings.Index(lower, m); idx >= 0 {
			if best < 0 || idx < best {
				best = idx
				marker = strings.TrimSpace(m)
			}
		}
	}
	if best < 0 {
		return orig, AnswerMetaTrim{}
	}
	trimmed := strings.TrimRight(answer[:best], " \t\r\n")
	if trimmed == "" {
		return orig, AnswerMetaTrim{}
	}
	return trimmed, AnswerMetaTrim{
		Trimmed:      true,
		CharsRemoved: len(answer) - len(trimmed),
		Marker:       marker,
	}
}

// AnswerContainsMetaCorrection reports whether answer still has a self-correction tail.
func AnswerContainsMetaCorrection(answer string) bool {
	_, meta := TrimMetaCorrection(answer)
	return meta.Trimmed
}

var (
	sourceDumpLineRE  = regexp.MustCompile(`(?i)\n\[\d+\]\s+type=(?:field_guide|platform_doc)\s+source_id=`)
	sourceDumpMarkers = []string{"\nsources:\n", "\nsources (cite", "\n\nsources:", "\n[type=field_guide", "\n[type=platform_doc"}
)

// AnswerSourceDumpTrim records raw RAG source metadata dumps removed before persist.
type AnswerSourceDumpTrim struct {
	Trimmed      bool   `json:"source_dump_trimmed,omitempty"`
	CharsRemoved int    `json:"source_dump_chars_removed,omitempty"`
	Marker       string `json:"source_dump_marker,omitempty"`
}

// TrimSourceDump removes echoed Sources blocks and raw chunk metadata tails from answers.
func TrimSourceDump(answer string) (string, AnswerSourceDumpTrim) {
	orig := answer
	answer = strings.TrimRight(answer, " \t\r\n")
	if answer == "" {
		return orig, AnswerSourceDumpTrim{}
	}
	cut := sourceDumpCutIndex(answer)
	if cut < 0 || cut >= len(answer) {
		return orig, AnswerSourceDumpTrim{}
	}
	trimmed := strings.TrimRight(answer[:cut], " \t\r\n")
	if trimmed == "" {
		return orig, AnswerSourceDumpTrim{}
	}
	return trimmed, AnswerSourceDumpTrim{
		Trimmed:      true,
		CharsRemoved: len(answer) - len(trimmed),
		Marker:       sourceDumpMarkerAt(answer, cut),
	}
}

func sourceDumpCutIndex(answer string) int {
	lower := strings.ToLower(answer)
	best := -1
	for _, marker := range sourceDumpMarkers {
		if idx := strings.Index(lower, marker); idx >= 0 {
			if best < 0 || idx < best {
				best = idx
			}
		}
	}
	if loc := sourceDumpLineRE.FindStringIndex(answer); loc != nil && loc[0] >= 200 {
		if best < 0 || loc[0] < best {
			best = loc[0]
		}
	}
	return best
}

func sourceDumpMarkerAt(answer string, cut int) string {
	if cut <= 0 || cut > len(answer) {
		return "sources"
	}
	snip := strings.TrimSpace(answer[cut:])
	if len(snip) > 48 {
		snip = snip[:48]
	}
	return snip
}

// AnswerContainsSourceDump reports whether answer still echoes raw source metadata blocks.
func AnswerContainsSourceDump(answer string) bool {
	_, meta := TrimSourceDump(answer)
	return meta.Trimmed
}

// Phase 150 WS1 — strip raw developer HTTP verb+path jargon (e.g. from
// docs/local-operator-bootstrap.md onboarding command blocks) that leaks
// verbatim into farmer-facing grounded answers, such as
// "`PATCH /alerts/{id}/acknowledge`".
var devAPIPathRE = regexp.MustCompile("`?\\b(GET|POST|PATCH|PUT|DELETE)\\s+/[^\\s`)]+`?")

// AnswerDevJargonRedaction records raw HTTP verb+path removals before persist.
type AnswerDevJargonRedaction struct {
	Redacted     bool `json:"dev_jargon_redacted,omitempty"`
	Occurrences  int  `json:"dev_jargon_occurrences,omitempty"`
	CharsRemoved int  `json:"dev_jargon_chars_removed,omitempty"`
}

// RedactDevAPIJargon removes literal "METHOD /path" developer jargon from an
// otherwise-farmer-facing answer, collapsing empty parens and doubled spaces
// left behind.
func RedactDevAPIJargon(answer string) (string, AnswerDevJargonRedaction) {
	matches := devAPIPathRE.FindAllString(answer, -1)
	if len(matches) == 0 {
		return answer, AnswerDevJargonRedaction{}
	}
	out := devAPIPathRE.ReplaceAllString(answer, "")
	charsRemoved := len(answer) - len(out)
	out = collapseDevJargonArtifacts(out)
	return out, AnswerDevJargonRedaction{Redacted: true, Occurrences: len(matches), CharsRemoved: charsRemoved}
}

var (
	emptyParensRE   = regexp.MustCompile(`\(\s*\)`)
	doubleSpaceRE   = regexp.MustCompile(`[ \t]{2,}`)
	danglingArrowRE = regexp.MustCompile(`:\s*→`)
)

func collapseDevJargonArtifacts(s string) string {
	s = emptyParensRE.ReplaceAllString(s, "")
	s = danglingArrowRE.ReplaceAllString(s, ":")
	s = doubleSpaceRE.ReplaceAllString(s, " ")
	return s
}

// AnswerContainsDevAPIJargon reports whether answer still has raw HTTP verb+path text.
func AnswerContainsDevAPIJargon(answer string) bool {
	return devAPIPathRE.MatchString(answer)
}

// AnswerLengthTrim records grounded answer length caps applied before persist.
type AnswerLengthTrim struct {
	Trimmed      bool `json:"answer_length_trimmed,omitempty"`
	CharsRemoved int  `json:"answer_length_chars_removed,omitempty"`
	MaxChars     int  `json:"answer_length_max,omitempty"`
}

// GroundedAnswerMaxChars returns the post-finalize length cap (0 = no cap).
func GroundedAnswerMaxChars(effectiveContextWindow int) int {
	if raw := strings.TrimSpace(os.Getenv("GUARDIAN_GROUNDED_ANSWER_MAX_CHARS")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	if effectiveContextWindow > 0 && effectiveContextWindow <= 4096 {
		return 2500
	}
	return 0
}

// TrimGroundedAnswerLength caps long grounded answers on small-context profiles.
func TrimGroundedAnswerLength(answer string, effectiveContextWindow int) (string, AnswerLengthTrim) {
	orig := answer
	max := GroundedAnswerMaxChars(effectiveContextWindow)
	if max <= 0 || len(answer) <= max {
		return orig, AnswerLengthTrim{}
	}
	trimmed := trimAnswerToMaxChars(answer, max)
	if trimmed == "" || trimmed == orig {
		return orig, AnswerLengthTrim{}
	}
	return trimmed, AnswerLengthTrim{
		Trimmed:      true,
		CharsRemoved: len(orig) - len(trimmed),
		MaxChars:     max,
	}
}

func trimAnswerToMaxChars(answer string, max int) string {
	if len(answer) <= max {
		return answer
	}
	cut := max
	if idx := strings.LastIndex(answer[:max], "\n\n"); idx > max/2 {
		cut = idx
	} else if idx := strings.LastIndex(answer[:max], "\n"); idx > max/2 {
		cut = idx
	}
	trimmed := strings.TrimRight(answer[:cut], " \t\r\n")
	if trimmed == "" {
		return answer[:max]
	}
	return trimmed + "…"
}
