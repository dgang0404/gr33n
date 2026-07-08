package farmguardian

import (
	"context"
	"strings"
	"testing"

	"gr33n-api/internal/rag/llm"
)

type mockCritiqueLLM struct {
	reply string
}

func (m mockCritiqueLLM) ChatCompletion(_ context.Context, _, _ string) (string, error) {
	return m.reply, nil
}

func (m mockCritiqueLLM) ModelLabel() string { return "mock-critique" }

func TestCritiqueAnswer_disabledByDefault(t *testing.T) {
	t.Setenv("GUARDIAN_ANSWER_CRITIQUE", "0")
	out := CritiqueAnswer(context.Background(), mockCritiqueLLM{reply: "NO: drift"}, "q", "a")
	if !out.Skipped || out.Enabled {
		t.Fatalf("got %+v", out)
	}
}

func TestCritiqueAnswer_parsesNo(t *testing.T) {
	t.Setenv("GUARDIAN_ANSWER_CRITIQUE", "1")
	out := CritiqueAnswer(context.Background(), mockCritiqueLLM{reply: "NO: Answer drifts to unrelated endocrine content."}, "EC?", "endocrine")
	if out.Pass || !strings.Contains(out.Reason, "endocrine") {
		t.Fatalf("got %+v", out)
	}
}

func TestCritiqueAnswer_parsesYes(t *testing.T) {
	t.Setenv("GUARDIAN_ANSWER_CRITIQUE", "1")
	out := CritiqueAnswer(context.Background(), mockCritiqueLLM{reply: "YES: cites lettuce EC targets."}, "EC?", "lettuce EC 1.0")
	if !out.Pass {
		t.Fatalf("got %+v", out)
	}
}

func TestCritiqueAnswer_run3ECPHWouldFail(t *testing.T) {
	t.Setenv("GUARDIAN_ANSWER_CRITIQUE", "1")
	answer := `Our operational documentation for leafy greens indicates lettuce EC 1.0–1.3 mS/cm and pH 5.5–6.0.
Sources on endocrine disruptors in Lake Erie wildlife show profound effects.`
	out := CritiqueAnswer(context.Background(), mockCritiqueLLM{reply: "NO: tail discusses endocrine disruptors unrelated to EC/pH."},
		"What does our operational documentation say about EC and pH targets for leafy greens here?", answer)
	if out.Pass {
		t.Fatalf("run #3 ec-ph tail should fail critique: %+v", out)
	}
}

var _ llm.ChatCompleter = mockCritiqueLLM{}
