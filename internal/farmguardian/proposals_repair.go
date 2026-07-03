package farmguardian

import (
	"context"
	"strings"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
)

// ProposalRepairOutcome records a one-shot JSON repair attempt (Phase 122).
type ProposalRepairOutcome struct {
	Attempted bool
	Recovered bool
	ParseErr  string
}

// TryBuildLLMProposalsWithRepair parses the assistant reply and, on failure for a
// write-intent turn, may invoke repairFn once to obtain a corrected assistant reply.
func TryBuildLLMProposalsWithRepair(
	ctx context.Context,
	q db.Querier,
	userID uuid.UUID,
	farmID int64,
	sessionID uuid.UUID,
	question string,
	assistantText string,
	policy LLMProposalPolicy,
	hasOperate bool,
	hasAdmin bool,
	inProcedure bool,
	matcherMatched bool,
	repairFn func(repairSystem string) (string, error),
) ([]ActionProposal, ProposalRepairOutcome, error) {
	outcome := ProposalRepairOutcome{}
	if !ShouldAttemptLLMProposal(LLMProposalAttemptInput{
		Question:       question,
		MatcherMatched: matcherMatched,
		HasOperate:     hasOperate,
		InProcedure:    inProcedure,
		Policy:         policy,
	}) {
		return nil, outcome, nil
	}

	text := assistantText
	if _, ok, errMsg := ParseLLMProposalFromAssistantDetailed(text); !ok {
		outcome.ParseErr = errMsg
		if repairFn != nil {
			outcome.Attempted = true
			repaired, rerr := repairFn(ProposalRepairSystemAddendum(errMsg))
			if rerr == nil && strings.TrimSpace(repaired) != "" {
				text = repaired
				if _, ok2, _ := ParseLLMProposalFromAssistantDetailed(text); ok2 {
					outcome.Recovered = true
				}
			}
		}
	}

	props, err := TryBuildLLMProposalsFromAssistant(
		ctx, q, userID, farmID, sessionID, question, text,
		policy, hasOperate, hasAdmin, inProcedure, matcherMatched,
	)
	return props, outcome, err
}
