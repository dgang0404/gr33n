// Phase 132 WS4 — morning walkthrough findings → frozen proposals.

package farmguardian

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
)

const walkthroughMaxProposals = 2

// BuildWalkthroughProposals emits rule-based ack_alert proposals from walk_farm warn
// findings during morning_walkthrough mode (no LLM proposals flag required).
func BuildWalkthroughProposals(
	ctx context.Context,
	q db.Querier,
	userID uuid.UUID,
	farmID int64,
	sessionID uuid.UUID,
	ref *ContextRef,
) ([]ActionProposal, error) {
	if q == nil || farmID <= 0 || userID == uuid.Nil || sessionID == uuid.Nil {
		return nil, nil
	}
	if ref == nil || !strings.EqualFold(strings.TrimSpace(ref.GuardianMode), "morning_walkthrough") {
		return nil, nil
	}

	findings, err := collectWalkFarmFindings(ctx, q, farmID)
	if err != nil {
		return nil, err
	}
	warns := filterWalkFindings(findings, "warn")
	if len(warns) == 0 {
		return nil, nil
	}

	var out []ActionProposal
	for _, f := range warns {
		if f.Category != "alerts" || f.AlertID <= 0 {
			continue
		}
		summary := "Acknowledge: " + strings.TrimSpace(f.PlainText)
		if summary == "Acknowledge:" {
			summary = "Acknowledge alert #" + strconv.FormatInt(f.AlertID, 10)
		}
		row, err := insertProposal(ctx, q, insertProposalInput{
			userID:    userID,
			farmID:    farmID,
			sessionID: sessionID,
			toolID:    "ack_alert",
			args:      map[string]any{"alert_id": f.AlertID},
			summary:   summary,
			revision:  1,
		})
		if err != nil {
			return nil, err
		}
		LogMatcherProposalHit(farmID, "ack_alert")
		out = append(out, ActionProposalFromRow(row))
		if len(out) >= walkthroughMaxProposals {
			break
		}
	}
	return out, nil
}

// collectWalkFarmFindings gathers walk_farm categories without rendering prompt text.
func collectWalkFarmFindings(ctx context.Context, q db.Querier, farmID int64) ([]walkFinding, error) {
	zones, _ := q.ListZonesByFarm(ctx, farmID)
	zoneName := map[int64]string{}
	for _, z := range zones {
		zoneName[z.ID] = strings.TrimSpace(z.Name)
	}

	var findings []walkFinding

	alertFindings, err := walkFarmAlertFindings(ctx, q, farmID)
	if err != nil {
		return nil, err
	}
	findings = append(findings, alertFindings...)

	_, feedWarn, err := walkFarmFeedFindings(ctx, q, farmID, zoneName)
	if err != nil {
		return nil, err
	}
	findings = append(findings, feedWarn...)

	deviceFindings, err := walkFarmDeviceFindings(ctx, q, farmID)
	if err != nil {
		return nil, err
	}
	findings = append(findings, deviceFindings...)

	comfortFindings, err := walkFarmComfortFindings(ctx, q, farmID, zoneName)
	if err != nil {
		return nil, err
	}
	findings = append(findings, comfortFindings...)

	stockFindings, err := walkFarmLowStockFindings(ctx, q, farmID)
	if err != nil {
		return nil, err
	}
	findings = append(findings, stockFindings...)

	return findings, nil
}
