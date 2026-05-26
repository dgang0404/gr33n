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
