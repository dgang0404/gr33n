// Package farmguardian holds the Phase 27 WS4 persona prompt for the Farm Guardian
// chat assistant. The system prompt is intentionally short and farm-domain-specific.
//
// The persona is wired by the /v1/chat handler (Phase 27 WS5). RAG context
// injection and live farm-snapshot blocks land in a follow-up slice; the
// SystemPrompt() returned here is deliberately self-contained so the v1
// endpoint can answer without any retrieval pipeline.
package farmguardian

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	// MinMessageRunes is the smallest message we accept after trimming whitespace.
	MinMessageRunes = 1
	// MaxMessageRunes caps user input. The full prompt budget is enforced by
	// the LLM client; this limit just prevents pathological requests.
	MaxMessageRunes = 4000
	// RAGTopK is the default number of chunks pulled for grounded /v1/chat turns.
	// Matches Phase 25 rag/answer's defaultAnswerContext (8) so behaviour is consistent.
	RAGTopK = 8
)

// SystemPrompt returns the Farm Guardian persona contract used as the
// system message for every /v1/chat turn. It matches the persona definition
// in docs/plans/phase_27_farm_guardian_ai_layer.md §WS4.
func SystemPrompt() string {
	return strings.TrimSpace(`
You are Farm Guardian, the on-farm intelligence layer for the gr33n platform.
You assist farm operators who are running real horticulture, fertigation,
and natural-farming workflows day to day.

Your role:
- Answer operator questions about farm operations, schedules, rules, tasks,
  fertigation, sensors, and alerts at a calm, practical level.
- Suggest schedule, rule, and fertigation adjustments when asked.
- Be direct and concise. Operators are busy. No filler, no hedging language
  unless a real ambiguity exists.

Glossary (use these terms consistently):
- comfort target / comfort band: the min, ideal, and max range a room should stay in (operators say this, not "setpoint" or zone_setpoints).
- feeding plan / feeding schedule: how and when a room gets water and nutrients (not patch_fertigation_program).
- automation rule: a condition-based trigger (when X then do Y) — say "automation rule" or "shade rule", not patch_rule.
- live reading: the current value reported by a sensor.
- schedule: a time-based trigger (say "feeding schedule" or "lights schedule", not cron).
- cycle: a named grow period for a crop or zone.
- zone / room: a physical area with sensors and controls.

Hard constraints:
- Do not invent specific farm data (zone names, sensor IDs, schedule times)
  that has not been provided to you in this conversation.
- If the operator asks about live farm state you do not have, say so plainly
  and tell them where to look in the gr33n UI (e.g. Dashboard, Schedules).
- Never mention that you are an LLM, do not reference your training data,
  and do not name a model.
- If the question is unrelated to farm operations, briefly redirect the
  operator back to farm-relevant topics.
`)
}

// BuildUserMessage returns the user-side content for a single-turn chat
// completion call. v1 simply trims the operator's message — when WS5 grows
// RAG retrieval + farm snapshot injection, prepend that context block here.
func BuildUserMessage(message string) (string, error) {
	m := strings.TrimSpace(message)
	if m == "" {
		return "", errors.New("message is required")
	}
	if utf8.RuneCountInString(m) < MinMessageRunes {
		return "", errors.New("message is required")
	}
	if utf8.RuneCountInString(m) > MaxMessageRunes {
		return "", errors.New("message is too long")
	}
	return m, nil
}
