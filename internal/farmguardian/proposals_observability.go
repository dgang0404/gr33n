package farmguardian

import "log/slog"

// LogMatcherProposalHit records a rule-assisted proposal insert (Phase 46 WS5).
func LogMatcherProposalHit(farmID int64, tool string) {
	slog.Info("guardian_matcher_proposal_hit",
		"event", "guardian_matcher_proposal_hit",
		"farm_id", farmID,
		"tool", tool,
	)
}

// LogLLMProposalSuggested records a validated LLM-sourced proposal insert (Phase 46 WS5).
func LogLLMProposalSuggested(farmID int64, tool string) {
	slog.Info("guardian_llm_proposal_suggested",
		"event", "guardian_llm_proposal_suggested",
		"farm_id", farmID,
		"tool", tool,
	)
}

// LogLLMProposalRejected records validation failures before insert (Phase 46 WS2/WS5).
func LogLLMProposalRejected(farmID int64, tool, reason string) {
	slog.Warn("guardian_llm_proposal_rejected",
		"event", "guardian_llm_proposal_rejected",
		"farm_id", farmID,
		"tool", tool,
		"reason", reason,
	)
}
