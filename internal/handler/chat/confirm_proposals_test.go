package chat

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/authctx"
	"gr33n-api/internal/farmguardian"
)

func TestAttachProposals_SkipsWithoutUserOrFarm(t *testing.T) {
	h := NewHandlerWithDeps(ai.Config{}, nil, nil, nil)
	resp := &postResponse{}
	h.attachProposals(context.Background(), 1, false, uuid.Nil, uuid.Nil, "Set feed to 0.3 L", "{}", farmguardian.Snapshot{}, resp)
	if len(resp.Proposals) != 0 {
		t.Fatalf("expected no proposals without user, got %+v", resp.Proposals)
	}
	h.attachProposals(context.Background(), 0, true, uuid.New(), uuid.Nil, "Set feed to 0.3 L", "{}", farmguardian.Snapshot{}, resp)
	if len(resp.Proposals) != 0 {
		t.Fatalf("expected no proposals without farm, got %+v", resp.Proposals)
	}
}

func TestAttachProposals_SkipsLLMWhenFlagOff(t *testing.T) {
	t.Setenv("GUARDIAN_LLM_PROPOSALS", "false")
	h := NewHandlerWithDeps(ai.Config{}, nil, nil, nil)
	resp := &postResponse{}
	assistant := "```json\n{\"tool\":\"patch_fertigation_program\",\"args\":{\"program_id\":1},\"summary\":\"x\"}\n```"
	ctx := authctx.WithFarmAuthzSkip(context.Background(), true)
	h.attachProposals(ctx, 1, true, uuid.New(), uuid.Nil, "Update feed volume to 0.3 L", assistant, farmguardian.Snapshot{}, resp)
	if len(resp.Proposals) != 0 {
		t.Fatalf("expected no proposals without DB + disabled flag, got %+v", resp.Proposals)
	}
}
