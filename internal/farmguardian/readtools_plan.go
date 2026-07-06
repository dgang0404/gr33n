// Phase 132 WS1 — execute planned read tools before intent regex matching.

package farmguardian

import (
	"context"

	db "gr33n-api/internal/db"
)

func runPlannedReadTools(ctx context.Context, q db.Querier, farmID int64, question string, snap Snapshot, ref *ContextRef, plan ToolPlan) (blocks []string, ran map[string]bool) {
	ran = make(map[string]bool, len(plan.ToolIDs))
	for _, toolID := range plan.ToolIDs {
		block, skip := renderPlannedReadTool(ctx, q, farmID, question, snap, ref, toolID)
		if skip {
			ran[toolID] = true
			continue
		}
		if block != "" {
			blocks = append(blocks, block)
			logReadToolUse(ctx, toolID, farmID)
		}
		ran[toolID] = true
	}
	return blocks, ran
}

func renderPlannedReadTool(ctx context.Context, q db.Querier, farmID int64, question string, snap Snapshot, ref *ContextRef, toolID string) (block string, skip bool) {
	switch toolID {
	case "list_unread_alerts":
		if snapUnreadCountIntent(question) {
			return "", true
		}
		b, err := renderListUnreadAlerts(ctx, q, farmID)
		if err != nil {
			return "", false
		}
		return b, false
	case "walk_farm":
		b, err := renderWalkFarm(ctx, q, farmID)
		if err != nil {
			return "", false
		}
		return b, false
	case "summarize_device_health":
		b, err := renderSummarizeDeviceHealth(ctx, q, farmID, question, ref)
		if err != nil {
			return "", false
		}
		return b, false
	default:
		return "", true
	}
}

func skipIfPlanned(ran map[string]bool, toolID string) bool {
	return ran != nil && ran[toolID]
}
