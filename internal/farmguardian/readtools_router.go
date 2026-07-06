// Phase 132 WS1 — farm-counsel read-tool router.

package farmguardian

import (
	"log/slog"
	"strings"
)

// ToolPlan is the ordered set of read tools to run for one grounded turn.
type ToolPlan struct {
	ToolIDs []string
	Reason  map[string]string
}

// PlanReadTools selects core + mode-forced read tools before intent regex matching.
func PlanReadTools(question string, ref *ContextRef, snap Snapshot) ToolPlan {
	plan := ToolPlan{Reason: make(map[string]string)}
	add := func(id, why string) {
		for _, existing := range plan.ToolIDs {
			if existing == id {
				return
			}
		}
		plan.ToolIDs = append(plan.ToolIDs, id)
		plan.Reason[id] = why
	}

	if ref != nil && strings.EqualFold(strings.TrimSpace(ref.GuardianMode), "morning_walkthrough") {
		add("walk_farm", "guardian_mode=morning_walkthrough")
		add("summarize_device_health", "guardian_mode=morning_walkthrough")
	}

	if snap.UnreadAlerts > 0 {
		add("list_unread_alerts", "farm has unread alerts")
	}

	if shouldRunSummarizeDeviceHealthReadIntent(question, ref) {
		add("summarize_device_health", "device health intent")
	}

	slog.Info("farm guardian read tool plan",
		"event", "guardian_tool_plan",
		"tool_ids", plan.ToolIDs,
		"reasons", plan.Reason,
	)
	return plan
}

func planContains(plan ToolPlan, toolID string) bool {
	for _, id := range plan.ToolIDs {
		if id == toolID {
			return true
		}
	}
	return false
}
