package chat

import (
	"context"
	"testing"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/rag/llm"
)

type repairMockClient struct {
	calls int
	reply string
}

func (m *repairMockClient) ModelLabel() string { return "mock" }

func (m *repairMockClient) ChatCompletion(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (m *repairMockClient) ChatCompletionMessagesWithUsage(_ context.Context, _ []llm.Message) (string, llm.Usage, error) {
	m.calls++
	return m.reply, llm.Usage{PromptTokens: 10, CompletionTokens: 5}, nil
}

func TestMaybeRepairProposalAnswer_recovers(t *testing.T) {
	t.Setenv("GUARDIAN_LLM_PROPOSALS", "true")
	h := &Handler{}
	client := &repairMockClient{reply: "```json\n{\"tool\":\"ack_alert\",\"args\":{\"alert_id\":1},\"summary\":\"Ack\",\"confidence\":\"high\"}\n```"}
	answer, usage, outcome := h.maybeRepairProposalAnswer(
		context.Background(),
		client,
		[]llm.Message{{Role: "user", Content: "Acknowledge the alert"}},
		"Acknowledge the highest severity unread alert.",
		"Sure, I will do that for you.",
		llm.Usage{},
	)
	if !outcome.Attempted || !outcome.Recovered {
		t.Fatalf("expected recovery, got %+v", outcome)
	}
	if client.calls != 1 {
		t.Fatalf("expected one repair call, got %d", client.calls)
	}
	if _, ok := farmguardian.ParseLLMProposalFromAssistant(answer); !ok {
		t.Fatalf("repaired answer should parse: %q", answer)
	}
	if usage.PromptTokens != 10 {
		t.Fatalf("usage merge failed: %+v", usage)
	}
}
