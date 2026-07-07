// Phase 143 WS1 — strip instruction-template echoes from assistant answers.

package farmguardian

import (
	"strings"
	"unicode"
)

// AnswerLeakTrim records instruction-leak detection applied before turn persist.
type AnswerLeakTrim struct {
	Trimmed      bool   `json:"leak_trimmed,omitempty"`
	CharsRemoved int    `json:"leak_chars_removed,omitempty"`
	Marker       string `json:"leak_marker,omitempty"`
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

func leakCutIndex(answer, question string) int {
	lower := strings.ToLower(answer)
	best := -1

	for _, marker := range []string{"\n## your task", "## your task"} {
		if idx := strings.Index(lower, marker); idx >= 0 {
			if best < 0 || idx < best {
				best = idx
				if marker[0] == '\n' {
					best++ // keep content before the blank line
				}
			}
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
