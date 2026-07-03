package farmguardian

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestProposalRepairSystemAddendum_includesError(t *testing.T) {
	msg := ProposalRepairSystemAddendum("unexpected end of JSON input")
	if msg == "" || !strings.Contains(msg, "Parse error") || !strings.Contains(msg, `"tool"`) {
		t.Fatalf("repair addendum missing expected content: %q", msg)
	}
}

func TestTryBuildLLMProposalsWithRepair_recoversMalformedJSON(t *testing.T) {
	question := "Set feed volume to 0.3 L for Veg Tent"
	bad := "Sure, I'll update that for you."
	good := "```json\n{\"tool\":\"patch_fertigation_program\",\"args\":{\"program_id\":1},\"summary\":\"Lower feed\",\"confidence\":\"medium\"}\n```"

	calls := 0
	_, outcome, err := TryBuildLLMProposalsWithRepair(
		nil, nil, uuid.Nil, 1, uuid.Nil, question, bad,
		LLMProposalPolicy{Enabled: true},
		true, true, false, false,
		func(repairSystem string) (string, error) {
			calls++
			if repairSystem == "" {
				t.Fatal("expected repair system message")
			}
			return good, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.Attempted || !outcome.Recovered {
		t.Fatalf("expected repair recovery, got %+v", outcome)
	}
	if calls != 1 {
		t.Fatalf("expected one repair call, got %d", calls)
	}
}

func TestParseLLMProposalFromAssistantDetailed_malformed(t *testing.T) {
	_, ok, reason := ParseLLMProposalFromAssistantDetailed("no json here")
	if ok || reason == "" {
		t.Fatalf("expected parse failure, ok=%v reason=%q", ok, reason)
	}
}
