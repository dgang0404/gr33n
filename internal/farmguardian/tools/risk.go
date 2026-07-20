package tools

// Risk tier constants (Phase 30 WS2).
const (
	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"
)

// RiskTierForTool returns the proposal risk tier for a tool + frozen args.
func RiskTierForTool(toolID string, args map[string]any) string {
	switch toolID {
	case "mark_alert_read", "ack_alert":
		return RiskLow
	case "apply_bootstrap_template", "enqueue_actuator_command", "apply_grow_setup_pack":
		return RiskHigh
	case "patch_rule":
		if isActive, ok := args["is_active"].(bool); ok && !isActive {
			return RiskHigh
		}
		return RiskMedium
	case "create_task", "create_task_from_alert", "update_cycle_stage",
		"patch_schedule", "patch_fertigation_program",
		"create_plant", "create_crop_cycle", "create_fertigation_program", "create_lighting_program",
		"draft_input_definition", "draft_application_recipe", "draft_input_batch":
		return RiskMedium
	default:
		return RiskMedium
	}
}
