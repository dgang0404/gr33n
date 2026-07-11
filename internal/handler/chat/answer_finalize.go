package chat

import (
	"context"
	"log/slog"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/rag/llm"
	"gr33n-api/internal/rag/synthesis"

	db "gr33n-api/internal/db"
)

// citationSummariesFromCitations adapts the response-level citation shape
// (which carries chunk_id/source_id for the UI) to the compact accuracy-note
// input shape used by farmguardian.AnswerAccuracyNote.
func citationSummariesFromCitations(cites []synthesis.Citation) []farmguardian.CitationSummary {
	if len(cites) == 0 {
		return nil
	}
	out := make([]farmguardian.CitationSummary, 0, len(cites))
	for _, c := range cites {
		out = append(out, farmguardian.CitationSummary{
			Ref:        c.Ref,
			SourceType: c.SourceType,
			Excerpt:    c.Excerpt,
		})
	}
	return out
}

// attachCitationRoutes resolves a click-through UI route (zone, crop-cycle
// summary, ...) for each citation in place (Phase 152 WS2). Best-effort —
// citations whose source type has no route mapping, or whose row lookup
// fails, are left with an empty Route and render as plain text in the UI.
func attachCitationRoutes(ctx context.Context, q *db.Queries, farmID int64, cites []synthesis.Citation) {
	if q == nil || farmID <= 0 {
		return
	}
	for i := range cites {
		if route, ok := farmguardian.ResolveCitationRoute(ctx, q, farmID, cites[i].SourceType, cites[i].SourceID); ok {
			cites[i].Route = route
		}
	}
}

// applyAnswerAccuracyNote runs the live Phase 148/151/152 accuracy detectors
// (garbled truncation, citation-number mismatch, invented assumption math,
// uncited timeline claims, etc.) so bad answers are flagged in the UI and in
// logs the moment they happen, not only when someone re-runs guardian-eval.
// This never mutates the answer text — the detectors are heuristic and could
// false-positive, so we surface a warning rather than silently rewriting or
// blocking a farmer-facing answer.
func applyAnswerAccuracyNote(answer string, cites []synthesis.Citation) string {
	note := farmguardian.AnswerAccuracyNote(answer, citationSummariesFromCitations(cites))
	if note != "" {
		slog.Info("guardian: answer_accuracy_flagged", "note", note)
	}
	return note
}

func finalizeGroundedAnswer(answer string, chunks []db.SearchRagNearestNeighborsFilteredRow) string {
	answer = synthesis.StripOrphanCitationRefs(answer, len(chunks))
	if injected, ok := farmguardian.InjectAlertCitationRefs(answer, chunks); ok {
		slog.Info("guardian: alert_citation_refs_injected")
		answer = injected
	}
	if normalized, ok := farmguardian.NormalizeAlertListCitations(answer, chunks); ok {
		slog.Info("guardian: alert_list_citations_normalized")
		answer = normalized
	}
	return answer
}

type answerHygiene struct {
	leak       farmguardian.AnswerLeakTrim
	meta       farmguardian.AnswerMetaTrim
	cite       farmguardian.CitationURLSanitize
	sourceDump farmguardian.AnswerSourceDumpTrim
	devJargon  farmguardian.AnswerDevJargonRedaction
	length     farmguardian.AnswerLengthTrim
}

func sanitizeAssistantAnswer(answer, question string, grounded bool, effectiveContextWindow int) (string, answerHygiene) {
	var h answerHygiene
	answer, h.leak = farmguardian.TrimInstructionLeak(answer, question)
	if h.leak.Trimmed {
		slog.Info("guardian: answer_leak_trimmed",
			"chars_removed", h.leak.CharsRemoved,
			"marker", h.leak.Marker,
		)
	}
	answer, h.meta = farmguardian.TrimMetaCorrection(answer)
	if h.meta.Trimmed {
		slog.Info("guardian: answer_meta_correction_trimmed",
			"chars_removed", h.meta.CharsRemoved,
			"marker", h.meta.Marker,
		)
	}
	answer, h.cite = farmguardian.SanitizeCitationURLs(answer)
	if h.cite.Sanitized {
		slog.Info("guardian: citation_url_sanitized",
			"links_rewritten", h.cite.LinksRewritten,
		)
	}
	answer, h.sourceDump = farmguardian.TrimSourceDump(answer)
	if h.sourceDump.Trimmed {
		slog.Info("guardian: answer_source_dump_trimmed",
			"chars_removed", h.sourceDump.CharsRemoved,
			"marker", h.sourceDump.Marker,
		)
	}
	answer, h.devJargon = farmguardian.RedactDevAPIJargon(answer)
	if h.devJargon.Redacted {
		slog.Info("guardian: answer_dev_jargon_redacted",
			"occurrences", h.devJargon.Occurrences,
			"chars_removed", h.devJargon.CharsRemoved,
		)
	}
	if grounded {
		answer, h.length = farmguardian.TrimGroundedAnswerLength(answer, effectiveContextWindow)
		if h.length.Trimmed {
			slog.Info("guardian: answer_length_trimmed",
				"chars_removed", h.length.CharsRemoved,
				"max_chars", h.length.MaxChars,
			)
		}
	}
	return answer, h
}

func applyAnswerHygieneDebug(dbg *farmguardian.TurnDebug, h answerHygiene) {
	if dbg == nil {
		return
	}
	if h.leak.Trimmed {
		dbg.LeakTrimmed = true
		dbg.LeakCharsRemoved = h.leak.CharsRemoved
	}
	if h.meta.Trimmed {
		dbg.MetaCorrectionTrimmed = true
		dbg.MetaCorrectionCharsRemoved = h.meta.CharsRemoved
	}
	if h.cite.Sanitized {
		dbg.CitationURLsSanitized = true
		dbg.CitationLinksRewritten = h.cite.LinksRewritten
	}
	if h.sourceDump.Trimmed {
		dbg.SourceDumpTrimmed = true
		dbg.SourceDumpCharsRemoved = h.sourceDump.CharsRemoved
	}
	if h.devJargon.Redacted {
		dbg.DevJargonRedacted = true
		dbg.DevJargonCharsRemoved = h.devJargon.CharsRemoved
	}
	if h.length.Trimmed {
		dbg.AnswerLengthTrimmed = true
		dbg.AnswerLengthCharsRemoved = h.length.CharsRemoved
		dbg.AnswerLengthMax = h.length.MaxChars
	}
}

func applyAnswerRelevanceDebug(ctx context.Context, dbg *farmguardian.TurnDebug, embedder farmguardian.TextEmbedder, question, answer string) {
	if dbg == nil || embedder == nil {
		return
	}
	rel, err := farmguardian.ScoreAnswerRelevanceFromText(ctx, embedder, question, answer)
	if err != nil {
		slog.Warn("guardian: answer_relevance_failed", "err", err)
		return
	}
	dbg.QuestionAnswerRelevance = rel.QuestionAnswerCosine
	dbg.OpeningTailRelevance = rel.OpeningTailCosine
	dbg.LowRelevance = rel.LowRelevance
	dbg.RelevanceMinThreshold = rel.MinThreshold
	if rel.LowRelevance {
		slog.Info("guardian: answer_low_relevance",
			"question_answer_cosine", rel.QuestionAnswerCosine,
			"opening_tail_cosine", rel.OpeningTailCosine,
			"min", rel.MinThreshold,
		)
	}
}

func applyAnswerCritiqueDebug(ctx context.Context, dbg *farmguardian.TurnDebug, client llm.ChatCompleter, question, answer string) {
	if dbg == nil || !farmguardian.AnswerCritiqueEnabled() {
		return
	}
	out := farmguardian.CritiqueAnswer(ctx, client, question, answer)
	dbg.CritiqueEnabled = out.Enabled
	if out.Skipped {
		return
	}
	pass := out.Pass
	dbg.CritiquePass = &pass
	dbg.CritiqueReason = out.Reason
	if !out.Pass {
		slog.Info("guardian: answer_critique_fail", "reason", out.Reason)
	}
}
