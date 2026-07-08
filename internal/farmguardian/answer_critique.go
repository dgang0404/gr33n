// Phase 146 WS1 — optional lightweight self-critique judge (GPU profile).

package farmguardian

import (
	"context"
	"os"
	"strings"

	"gr33n-api/internal/rag/llm"
)

const critiqueSystemPrompt = `You are a strict QA gate for farm Guardian answers. Reply with exactly one line starting with YES or NO, a colon, then one sentence explaining why. Example: YES: The answer cites EC targets and stays on the question.`

// AnswerCritique is the outcome of an optional post-finalize critique pass.
type AnswerCritique struct {
	Enabled bool
	Skipped bool
	Pass    bool
	Reason  string
}

// AnswerCritiqueEnabled returns true when GUARDIAN_ANSWER_CRITIQUE=1.
func AnswerCritiqueEnabled() bool {
	return strings.TrimSpace(os.Getenv("GUARDIAN_ANSWER_CRITIQUE")) == "1"
}

// CritiqueAnswer runs one short LLM yes/no gate when enabled.
func CritiqueAnswer(ctx context.Context, client llm.ChatCompleter, question, answer string) AnswerCritique {
	if !AnswerCritiqueEnabled() {
		return AnswerCritique{Skipped: true}
	}
	if client == nil {
		return AnswerCritique{Enabled: true, Pass: true, Reason: "critique skipped: no chat client"}
	}
	q := strings.TrimSpace(question)
	a := strings.TrimSpace(answer)
	if q == "" || a == "" {
		return AnswerCritique{Enabled: true, Pass: true, Reason: "critique skipped: empty turn"}
	}
	if len(a) > 2400 {
		a = a[:2400] + "…"
	}
	user := "Question:\n" + q + "\n\nAnswer:\n" + a + "\n\nDoes the answer address the question using only cited farm or documentation facts?"
	raw, err := client.ChatCompletion(ctx, critiqueSystemPrompt, user)
	if err != nil {
		return AnswerCritique{Enabled: true, Pass: true, Reason: "critique error (non-blocking): " + err.Error()}
	}
	pass, reason := parseCritiqueLine(raw)
	return AnswerCritique{Enabled: true, Pass: pass, Reason: reason}
}

func parseCritiqueLine(raw string) (bool, string) {
	line := strings.TrimSpace(raw)
	if line == "" {
		return true, "empty critique response"
	}
	if i := strings.Index(line, "\n"); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	upper := strings.ToUpper(line)
	switch {
	case strings.HasPrefix(upper, "YES"):
		reason := strings.TrimSpace(strings.TrimPrefix(line, line[:3]))
		reason = strings.TrimPrefix(reason, ":")
		reason = strings.TrimSpace(reason)
		if reason == "" {
			reason = "critique yes"
		}
		return true, reason
	case strings.HasPrefix(upper, "NO"):
		reason := strings.TrimSpace(strings.TrimPrefix(line, line[:2]))
		reason = strings.TrimPrefix(reason, ":")
		reason = strings.TrimSpace(reason)
		if reason == "" {
			reason = "critique no"
		}
		return false, reason
	default:
		return true, "unparsed critique: " + truncateCritique(line, 120)
	}
}

func truncateCritique(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// CritiqueFromTurnDebug maps dev turn debug critique fields for QA archives.
func CritiqueFromTurnDebug(d *TurnDebug) AnswerCritique {
	if d == nil || !d.CritiqueEnabled {
		return AnswerCritique{}
	}
	out := AnswerCritique{Enabled: true, Reason: d.CritiqueReason}
	if d.CritiquePass != nil {
		out.Pass = *d.CritiquePass
	} else {
		out.Pass = true
	}
	return out
}
