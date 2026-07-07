package chat

import (
	"log/slog"

	"gr33n-api/internal/farmguardian"
)

type answerHygiene struct {
	leak farmguardian.AnswerLeakTrim
	meta farmguardian.AnswerMetaTrim
	cite farmguardian.CitationURLSanitize
}

func sanitizeAssistantAnswer(answer, question string) (string, answerHygiene) {
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
}
