package chat

import (
	"log/slog"

	"gr33n-api/internal/farmguardian"
)

type answerHygiene struct {
	leak farmguardian.AnswerLeakTrim
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
	if h.cite.Sanitized {
		dbg.CitationURLsSanitized = true
		dbg.CitationLinksRewritten = h.cite.LinksRewritten
	}
}
