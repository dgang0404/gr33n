package tools

import "testing"

func TestRiskTierForTool(t *testing.T) {
	tests := []struct {
		tool string
		args map[string]any
		want string
	}{
		{"mark_alert_read", nil, RiskLow},
		{"ack_alert", nil, RiskLow},
		{"create_task", map[string]any{"title": "x"}, RiskMedium},
		{"create_plant", map[string]any{"display_name": "x"}, RiskMedium},
		{"create_crop_cycle", map[string]any{"zone_id": 1}, RiskMedium},
		{"create_fertigation_program", map[string]any{"name": "x"}, RiskMedium},
		{"create_lighting_program", map[string]any{"preset_key": "veg_18_6", "zone_id": float64(1), "actuator_id": float64(1)}, RiskMedium},
		{"apply_grow_setup_pack", map[string]any{"zone_id": 1}, RiskHigh},
		{"apply_bootstrap_template", map[string]any{"template": "x"}, RiskHigh},
		{"enqueue_actuator_command", map[string]any{"command": "on"}, RiskHigh},
		{"patch_rule", map[string]any{"is_active": false}, RiskHigh},
		{"patch_rule", map[string]any{"is_active": true}, RiskMedium},
	}
	for _, tc := range tests {
		if got := RiskTierForTool(tc.tool, tc.args); got != tc.want {
			t.Fatalf("%s %+v => %q want %q", tc.tool, tc.args, got, tc.want)
		}
	}
}
