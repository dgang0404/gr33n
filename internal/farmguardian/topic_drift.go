// Phase 145 WS5 — generalized smoke / eval topic drift scoring.

package farmguardian

import (
	"fmt"
	"strings"
)

// SmokeTopicDriftInput bundles inputs for drift detection on archived smoke answers.
type SmokeTopicDriftInput struct {
	QuestionID string
	Category   string
	Prompt     string
	Answer     string
	Citations  []CitationSummary
	Relevance  AnswerRelevance
}

// SmokeTopicDriftNote returns a standardized eval failure reason when an answer drifts or fails hygiene.
func SmokeTopicDriftNote(in SmokeTopicDriftInput) string {
	if strings.TrimSpace(in.Answer) == "" {
		return ""
	}
	if note := smokeAnswerHygieneNote(in.Prompt, in.Answer); note != "" {
		return note
	}
	if note := AnswerAccuracyNote(in.Answer, in.Citations); note != "" {
		return note
	}
	if relevanceScored(in.Relevance) && in.Relevance.LowRelevance {
		return fmt.Sprintf("low_relevance (q↔a %.2f, open↔tail %.2f)",
			in.Relevance.QuestionAnswerCosine, in.Relevance.OpeningTailCosine)
	}
	if in.Category == "field_guide" && len(in.Citations) > 0 {
		if note := normalizeCitationDriftNote(CitationAlignmentNote(in.Prompt, in.Answer, in.Citations)); note != "" {
			return note
		}
	}
	if shouldCheckECPHKeywordDrift(in.QuestionID, in.Prompt) {
		if note := ecphKeywordDriftNote(in.Answer); note != "" {
			return note
		}
	}
	return ""
}

func smokeAnswerHygieneNote(prompt, answer string) string {
	if AnswerLooksLikePromptLeak(answer, prompt) {
		return "instruction template leak"
	}
	if AnswerContainsMetaCorrection(answer) {
		return "model self-correction / apology tail"
	}
	if AnswerContainsFakeCitationURL(answer) {
		return "hallucinated citation URLs"
	}
	if AnswerContainsSourceDump(answer) {
		return "raw source metadata dump"
	}
	if AnswerContainsDevAPIJargon(answer) {
		return "raw developer API jargon (HTTP verb + path) in farmer-facing answer"
	}
	return ""
}

func relevanceScored(rel AnswerRelevance) bool {
	return rel.LowRelevance || rel.QuestionAnswerCosine > 0
}

func normalizeCitationDriftNote(note string) string {
	if note == "" {
		return ""
	}
	lower := strings.ToLower(note)
	if strings.Contains(lower, "absent from cited") || strings.Contains(lower, "uncited") {
		return "uncited_tail"
	}
	if strings.Contains(lower, "misaligned") || strings.Contains(lower, "do not support") {
		return "citation_misaligned"
	}
	return note
}

func shouldCheckECPHKeywordDrift(questionID, prompt string) bool {
	if questionID == "smoke-ec-ph" {
		return true
	}
	return AgronomyQueryIntent(prompt)
}

func ecphKeywordDriftNote(answer string) string {
	a := strings.ToLower(answer)
	for _, term := range []string{
		"endocrine-disruptor", "endocrine disruptor", "endocrine_disruptor",
		"lake erie", "lake superior", "typha latifolia",
	} {
		if strings.Contains(a, term) {
			return "topic_drift: off-topic from leafy greens EC/pH"
		}
	}
	if strings.Contains(a, "endocrine") && !strings.Contains(a, "leafy green") {
		return "topic_drift: off-topic from leafy greens EC/pH"
	}
	return ""
}

// RelevanceFromTurnDebug maps dev turn debug relevance fields for QA archives.
func RelevanceFromTurnDebug(d *TurnDebug) AnswerRelevance {
	if d == nil {
		return AnswerRelevance{}
	}
	return AnswerRelevance{
		QuestionAnswerCosine: d.QuestionAnswerRelevance,
		OpeningTailCosine:    d.OpeningTailRelevance,
		LowRelevance:         d.LowRelevance,
		MinThreshold:         d.RelevanceMinThreshold,
	}
}
