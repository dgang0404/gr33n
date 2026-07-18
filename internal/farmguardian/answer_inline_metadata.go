// Phase 189 — strip raw RAG bookkeeping (source_id/chunk_id/type/doc_path)
// that leaks *inline*, mid-sentence, into farmer-facing answers, and clean up
// citation-placeholder artifacts where the model echoes the citation format
// instruction's own "[n]" token literally instead of substituting a real
// number.
//
// TrimSourceDump (Phase 143) already catches a *trailing block* of raw
// metadata (a "Sources:" dump or a "[6] type=field_guide source_id=…" line at
// the very end). Live turns show the same jargon woven into the middle of a
// sentence instead — "(field_guide source id=8, chunk id=66)",
// "field_guide source_id=17 chunk_id=18", "(field_guide: doc_path=field-
// guides/crop-microgreens-nutrition.md)" — which TrimSourceDump's
// end-of-answer markers never see.

package farmguardian

import "regexp"

var (
	inlineSourceIDRE = regexp.MustCompile(
		`(?i)\(?(?:type\s*=\s*)?(?:field_guide|symptom_guide|platform_doc)?[,:]?\s*source[_ ]id\s*=?\s*\d+(?:\s*,?\s*chunk[_ ]id\s*=?\s*\d+)?\)?`)
	inlineTypeFieldRE = regexp.MustCompile(`(?i)\(?type\s*=\s*(?:field_guide|symptom_guide|platform_doc)\)?`)
	inlineDocPathRE   = regexp.MustCompile(`(?i)\(?(?:(?:field_guide|symptom_guide|platform_doc)\s*:\s*)?doc_path\s*=\s*\S+\)?`)
)

// AnswerInlineMetadataRedaction records inline RAG-bookkeeping removals
// applied before persist.
type AnswerInlineMetadataRedaction struct {
	Redacted     bool `json:"inline_metadata_redacted,omitempty"`
	Occurrences  int  `json:"inline_metadata_occurrences,omitempty"`
	CharsRemoved int  `json:"inline_metadata_chars_removed,omitempty"`
}

// RedactInlineSourceMetadata strips mid-sentence "source_id=N", "chunk_id=N",
// "type=field_guide", and "doc_path=…" fragments that leak from the RAG
// citation payload into the visible answer, then collapses the empty
// parens / doubled spaces / dangling punctuation left behind.
func RedactInlineSourceMetadata(answer string) (string, AnswerInlineMetadataRedaction) {
	occurrences := 0
	out := answer
	for _, re := range []*regexp.Regexp{inlineSourceIDRE, inlineTypeFieldRE, inlineDocPathRE} {
		matches := re.FindAllString(out, -1)
		if len(matches) == 0 {
			continue
		}
		occurrences += len(matches)
		out = re.ReplaceAllString(out, "")
	}
	if occurrences == 0 {
		return answer, AnswerInlineMetadataRedaction{}
	}
	out = collapseDevJargonArtifacts(out)
	return out, AnswerInlineMetadataRedaction{
		Redacted:     true,
		Occurrences:  occurrences,
		CharsRemoved: len(answer) - len(out),
	}
}

// AnswerContainsInlineSourceMetadata reports whether answer still leaks raw
// RAG bookkeeping fields inline.
func AnswerContainsInlineSourceMetadata(answer string) bool {
	return inlineSourceIDRE.MatchString(answer) || inlineTypeFieldRE.MatchString(answer) || inlineDocPathRE.MatchString(answer)
}

var (
	// A literal "[n]" (the letter n, not a digit) means the model echoed the
	// citation-format instruction's own placeholder token instead of
	// substituting a real citation number — there's no information to
	// preserve, so these are removed outright.
	placeholderCiteLiteralNRE = regexp.MustCompile(`(?i)\(?\bsource(?:_id)?\s*[:=]?\s*\[n\]\)?`)
	// "source:[5]" / "source [1]" / "source[3]" use a real digit — these are
	// a valid citation just written in an inconsistent format. Normalize
	// down to the bare "[N]" form used everywhere else in the answer.
	placeholderCiteDigitRE = regexp.MustCompile(`(?i)\(?\bsource\s*[:=]?\s*(\[\d+\])\)?`)
)

// AnswerPlaceholderCitationRedaction records citation-placeholder cleanup
// applied before persist.
type AnswerPlaceholderCitationRedaction struct {
	Redacted     bool `json:"placeholder_citation_redacted,omitempty"`
	Occurrences  int  `json:"placeholder_citation_occurrences,omitempty"`
	CharsRemoved int  `json:"placeholder_citation_chars_removed,omitempty"`
}

// RedactPlaceholderCitationMarkers strips literal "source[n]"/"source_id=[n]"
// template-placeholder leaks and normalizes "source:[N]"/"source [N]" style
// citations down to the plain "[N]" form used elsewhere in the answer.
func RedactPlaceholderCitationMarkers(answer string) (string, AnswerPlaceholderCitationRedaction) {
	occurrences := len(placeholderCiteLiteralNRE.FindAllString(answer, -1))
	out := placeholderCiteLiteralNRE.ReplaceAllString(answer, "")

	digitMatches := placeholderCiteDigitRE.FindAllString(out, -1)
	occurrences += len(digitMatches)
	out = placeholderCiteDigitRE.ReplaceAllString(out, "$1")

	if occurrences == 0 {
		return answer, AnswerPlaceholderCitationRedaction{}
	}
	out = collapseDevJargonArtifacts(out)
	return out, AnswerPlaceholderCitationRedaction{
		Redacted:     true,
		Occurrences:  occurrences,
		CharsRemoved: len(answer) - len(out),
	}
}

// AnswerContainsPlaceholderCitation reports whether answer still has a
// literal "[n]" placeholder or a "source:[N]"-style inconsistent citation.
func AnswerContainsPlaceholderCitation(answer string) bool {
	return placeholderCiteLiteralNRE.MatchString(answer) || placeholderCiteDigitRE.MatchString(answer)
}
