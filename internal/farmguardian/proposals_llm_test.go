package farmguardian

import (
	"context"
	"os"
	"testing"
)

func TestLoadLLMProposalPolicyFromEnv(t *testing.T) {
	t.Setenv("GUARDIAN_LLM_PROPOSALS", "true")
	if !LoadLLMProposalPolicyFromEnv().Enabled {
		t.Fatal("expected enabled")
	}
	t.Setenv("GUARDIAN_LLM_PROPOSALS", "0")
	if LoadLLMProposalPolicyFromEnv().Enabled {
		t.Fatal("expected disabled")
	}
}

func TestShouldAttemptLLMProposal(t *testing.T) {
	policy := LLMProposalPolicy{Enabled: true}
	base := LLMProposalAttemptInput{
		Question:    "Set feed volume to 0.3 L for Flower Room",
		HasOperate:  true,
		InProcedure: false,
		Policy:      policy,
	}
	if !ShouldAttemptLLMProposal(base) {
		t.Fatal("expected attempt on write intent")
	}
	base.MatcherMatched = true
	if ShouldAttemptLLMProposal(base) {
		t.Fatal("matcher hit should skip LLM path")
	}
	base.MatcherMatched = false
	base.Policy.Enabled = false
	if ShouldAttemptLLMProposal(base) {
		t.Fatal("disabled policy should skip")
	}
	base.Policy.Enabled = true
	base.HasOperate = false
	if ShouldAttemptLLMProposal(base) {
		t.Fatal("viewer should skip")
	}
	base.HasOperate = true
	base.Question = "Why is humidity high?"
	if ShouldAttemptLLMProposal(base) {
		t.Fatal("read-only Q&A should skip")
	}
	base.Question = "start procedure wire-pi-relay-light"
	base.InProcedure = true
	if ShouldAttemptLLMProposal(base) {
		t.Fatal("procedure turn should skip")
	}
}

func TestHasWriteIntent(t *testing.T) {
	if !HasWriteIntent("Update the feeding plan volume to 0.3 L") {
		t.Fatal("expected write")
	}
	if HasWriteIntent("What is the feeding plan?") {
		t.Fatal("expected read-only")
	}
}

func TestIsLLMToolAllowed(t *testing.T) {
	if !IsLLMToolAllowed("patch_fertigation_program") {
		t.Fatal("patch_fertigation_program should be allowed")
	}
	if IsLLMToolAllowed("apply_grow_setup_pack") {
		t.Fatal("setup pack should be blocked")
	}
	if IsLLMToolAllowed("enqueue_actuator_command") {
		t.Fatal("actuator enqueue should be blocked")
	}
}

func TestParseLLMProposalFromAssistant(t *testing.T) {
	text := `Here is the change:
` + "```json\n" + `{
  "tool": "patch_fertigation_program",
  "args": {"program_id": 12, "total_volume_liters": 0.3},
  "summary": "Set program Flower Feed volume to 0.3 L",
  "confidence": "high"
}` + "\n```"
	draft, ok := ParseLLMProposalFromAssistant(text)
	if !ok {
		t.Fatal("expected parse")
	}
	if draft.Tool != "patch_fertigation_program" {
		t.Fatalf("tool %q", draft.Tool)
	}
	if draft.Args["program_id"] != float64(12) {
		t.Fatalf("args %#v", draft.Args)
	}
}

func TestValidateLLMProposalDraft_LowConfidenceHighTier(t *testing.T) {
	draft := LLMProposalDraft{
		Tool:       "patch_rule",
		Args:       map[string]any{"rule_id": 1, "is_active": false},
		Summary:    "Pause shade rule",
		Confidence: "low",
	}
	if reason := ValidateLLMProposalDraft(context.Background(), nil, 0, draft, true); reason != "low confidence on high-tier tool" {
		t.Fatalf("got %q", reason)
	}
}

func TestValidateLLMProposalDraft_RequiresAdmin(t *testing.T) {
	_ = os.Getenv
	draft := LLMProposalDraft{
		Tool:    "create_task",
		Args:    map[string]any{"title": "Check tank"},
		Summary: "Create task",
	}
	if reason := ValidateLLMProposalDraft(context.Background(), nil, 0, draft, false); reason != "" {
		t.Fatalf("create_task should pass for operator: %s", reason)
	}
}
