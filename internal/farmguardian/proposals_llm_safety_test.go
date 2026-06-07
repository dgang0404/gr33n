package farmguardian

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestTryBuildLLM_MatcherMatchedSkipsInsert(t *testing.T) {
	assistant := llmProposalJSON("patch_fertigation_program", map[string]any{
		"program_id":          1,
		"total_volume_liters": 0.3,
	}, "Set volume")
	props, err := TryBuildLLMProposalsFromAssistant(
		context.Background(), nil, uuid.New(), 1, uuid.Nil,
		"Update feed volume to 0.3 L",
		assistant,
		LLMProposalPolicy{Enabled: true},
		true, true,
		false, true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(props) != 0 {
		t.Fatalf("matcher hit should skip LLM insert, got %+v", props)
	}
}

func TestTryBuildLLM_NoOperateSkipsInsert(t *testing.T) {
	assistant := llmProposalJSON("create_task", map[string]any{"title": "Check tank"}, "Create task")
	props, err := TryBuildLLMProposalsFromAssistant(
		context.Background(), nil, uuid.New(), 1, uuid.Nil,
		"Create a task to check the tank",
		assistant,
		LLMProposalPolicy{Enabled: true},
		false, false,
		false, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(props) != 0 {
		t.Fatalf("viewer should not get LLM proposal, got %+v", props)
	}
}

func TestTryBuildLLM_NotOnAllowlist(t *testing.T) {
	assistant := llmProposalJSON("enqueue_actuator_command", map[string]any{
		"device_id": 1, "actuator_id": 1, "command": "on",
	}, "Turn on pump")
	props, err := TryBuildLLMProposalsFromAssistant(
		context.Background(), nil, uuid.New(), 1, uuid.Nil,
		"Turn on the pump",
		assistant,
		LLMProposalPolicy{Enabled: true},
		true, true,
		false, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(props) != 0 {
		t.Fatalf("unknown/blocked tool should not insert, got %+v", props)
	}
}

func TestTryBuildLLM_BootstrapTemplateRejected(t *testing.T) {
	assistant := llmProposalJSON("apply_bootstrap_template", map[string]any{"template": "greenhouse"}, "Bootstrap")
	props, err := TryBuildLLMProposalsFromAssistant(
		context.Background(), nil, uuid.New(), 1, uuid.Nil,
		"Apply bootstrap template",
		assistant,
		LLMProposalPolicy{Enabled: true},
		true, true,
		false, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(props) != 0 {
		t.Fatalf("bootstrap should be rejected, got %+v", props)
	}
}

func TestTryBuildLLM_NoProposalJSON(t *testing.T) {
	props, err := TryBuildLLMProposalsFromAssistant(
		context.Background(), nil, uuid.New(), 1, uuid.Nil,
		"Patch fertigation settings",
		"Here is some prose with no tool block.",
		LLMProposalPolicy{Enabled: true},
		true, true,
		false, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(props) != 0 {
		t.Fatalf("expected no proposal without JSON, got %+v", props)
	}
}

func TestTryBuildLLM_DisabledPolicySkips(t *testing.T) {
	assistant := llmProposalJSON("create_task", map[string]any{"title": "x"}, "Task")
	props, err := TryBuildLLMProposalsFromAssistant(
		context.Background(), nil, uuid.New(), 1, uuid.Nil,
		"Create a task",
		assistant,
		LLMProposalPolicy{Enabled: false},
		true, true,
		false, false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(props) != 0 {
		t.Fatalf("disabled policy should skip, got %+v", props)
	}
}

func llmProposalJSON(tool string, args map[string]any, summary string) string {
	raw, _ := json.Marshal(map[string]any{
		"tool":       tool,
		"args":       args,
		"summary":    summary,
		"confidence": "high",
	})
	return "```json\n" + string(raw) + "\n```"
}
