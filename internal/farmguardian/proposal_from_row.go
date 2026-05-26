package farmguardian

import (
	"encoding/json"

	db "gr33n-api/internal/db"
)

// ActionProposalFromRow maps a persisted proposal to the chat/inbox API shape.
func ActionProposalFromRow(row db.Gr33ncoreGuardianActionProposal) ActionProposal {
	var args map[string]any
	if len(row.Args) > 0 {
		_ = json.Unmarshal(row.Args, &args)
	}
	if args == nil {
		args = map[string]any{}
	}
	return ActionProposal{
		ProposalID: row.ProposalID.String(),
		Tool:       row.ToolID,
		Args:       args,
		Summary:    row.Summary,
		RiskTier:   row.RiskTier,
		ExpiresAt:  row.ExpiresAt,
	}
}
