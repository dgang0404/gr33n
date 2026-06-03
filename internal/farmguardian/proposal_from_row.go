package farmguardian

import (
	"encoding/json"

	db "gr33n-api/internal/db"
)

// ActionProposalFromRow maps a persisted proposal to the chat/inbox API shape,
// including Phase 34 revision lineage, operator-supplied facts, and the
// plain-language impact summary.
func ActionProposalFromRow(row db.Gr33ncoreGuardianActionProposal) ActionProposal {
	var args map[string]any
	if len(row.Args) > 0 {
		_ = json.Unmarshal(row.Args, &args)
	}
	if args == nil {
		args = map[string]any{}
	}

	var meta proposalMeta
	if len(row.Meta) > 0 {
		_ = json.Unmarshal(row.Meta, &meta)
	}

	ap := ActionProposal{
		ProposalID:       row.ProposalID.String(),
		Tool:             row.ToolID,
		Args:             args,
		Summary:          row.Summary,
		RiskTier:         row.RiskTier,
		ExpiresAt:        row.ExpiresAt,
		Revision:         int(row.Revision),
		Status:           row.Status,
		OperatorProvided: meta.OperatorProvided,
	}
	if row.SupersedesProposalID.Valid {
		ap.SupersedesProposalID = uuidString(row.SupersedesProposalID.Bytes)
	}
	ap.ImpactSummary = ImpactSummary(row.ToolID, args, meta.OperatorProvided)
	return ap
}

func uuidString(b [16]byte) string {
	const hexDigits = "0123456789abcdef"
	var buf [36]byte
	j := 0
	for i := 0; i < 16; i++ {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			buf[j] = '-'
			j++
		}
		buf[j] = hexDigits[b[i]>>4]
		buf[j+1] = hexDigits[b[i]&0x0f]
		j += 2
	}
	return string(buf[:])
}
