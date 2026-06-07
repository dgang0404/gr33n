package farmguardian

import (
	"context"
	"testing"
)

func TestValidateLLMProposalSchema(t *testing.T) {
	cases := []struct {
		tool   string
		args   map[string]any
		reason string
	}{
		{
			tool: "patch_fertigation_program",
			args: map[string]any{"program_id": 1, "total_volume_liters": 0.3},
		},
		{
			tool:   "patch_fertigation_program",
			args:   map[string]any{"total_volume_liters": 0.3},
			reason: "program_id required",
		},
		{
			tool:   "patch_fertigation_program",
			args:   map[string]any{"program_id": 1},
			reason: "at least one patch field required",
		},
		{
			tool: "patch_schedule",
			args: map[string]any{"schedule_id": 2, "is_active": false},
		},
		{
			tool:   "patch_rule",
			args:   map[string]any{"rule_id": 3, "is_active": false},
		},
		{
			tool:   "patch_rule",
			args:   map[string]any{"rule_id": 3, "is_active": true},
			reason: "LLM patch_rule only allows is_active false v1",
		},
		{
			tool:   "patch_rule",
			args:   map[string]any{"rule_id": 3, "threshold": 70.0},
			reason: "LLM patch_rule may not set threshold v1",
		},
		{
			tool: "ack_alert",
			args: map[string]any{"alert_id": 4},
		},
		{
			tool: "create_task",
			args: map[string]any{"title": "Check tank"},
		},
		{
			tool: "create_task_from_alert",
			args: map[string]any{"alert_id": 5},
		},
		{
			tool: "update_cycle_stage",
			args: map[string]any{"cycle_id": 6, "current_stage": "early_veg"},
		},
		{
			tool:   "update_cycle_stage",
			args:   map[string]any{"cycle_id": 6, "current_stage": "narnia"},
			reason: "invalid current_stage",
		},
	}
	for _, c := range cases {
		got := validateLLMProposalSchema(c.tool, c.args)
		want := c.reason
		if want == "" && got != "" {
			t.Fatalf("%s %#v: got %q want pass", c.tool, c.args, got)
		}
		if want != "" && got != want {
			t.Fatalf("%s %#v: got %q want %q", c.tool, c.args, got, want)
		}
	}
}

func TestValidateLLMProposalDraft_SchemaIntegration(t *testing.T) {
	draft := LLMProposalDraft{
		Tool:    "patch_fertigation_program",
		Args:    map[string]any{"program_id": 3, "total_volume_liters": 0.3},
		Summary: "Set volume to 0.3 L",
	}
	if reason := ValidateLLMProposalDraft(context.Background(), nil, 0, draft, true); reason != "" {
		t.Fatalf("expected pass without bind: %s", reason)
	}
	draft.Args = map[string]any{"program_id": 3}
	if reason := ValidateLLMProposalDraft(context.Background(), nil, 0, draft, true); reason != "at least one patch field required" {
		t.Fatalf("got %q", reason)
	}
}

func TestValidateLLMProposalDraft_Allowlist(t *testing.T) {
	draft := LLMProposalDraft{
		Tool:    "apply_bootstrap_template",
		Args:    map[string]any{"template": "x"},
		Summary: "Bootstrap",
	}
	if reason := ValidateLLMProposalDraft(context.Background(), nil, 0, draft, true); reason != "tool not on LLM allowlist" {
		t.Fatalf("got %q", reason)
	}
}
