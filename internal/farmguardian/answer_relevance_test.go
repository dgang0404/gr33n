package farmguardian

import (
	"context"
	"math"
	"strings"
	"testing"
)

type stubEmbedder struct {
	fn func(texts []string) [][]float32
}

func (s stubEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	return s.fn(texts), nil
}

func unitVec(dim int, hot int) []float32 {
	v := make([]float32, dim)
	if hot >= 0 && hot < dim {
		v[hot] = 1
	}
	return v
}

func TestCosineSimilarity(t *testing.T) {
	t.Parallel()
	a := unitVec(4, 0)
	b := unitVec(4, 0)
	if got := CosineSimilarity(a, b); math.Abs(got-1) > 1e-6 {
		t.Fatalf("identical want 1 got %v", got)
	}
	orth := unitVec(4, 1)
	if got := CosineSimilarity(a, orth); math.Abs(got) > 1e-6 {
		t.Fatalf("orthogonal want 0 got %v", got)
	}
}

func TestSplitAnswerOpeningTail_short(t *testing.T) {
	t.Parallel()
	open, tail := SplitAnswerOpeningTail("short answer")
	if open != "short answer" || tail != "" {
		t.Fatalf("open=%q tail=%q", open, tail)
	}
}

func TestSplitAnswerOpeningTail_long(t *testing.T) {
	t.Parallel()
	body := strings.Repeat("word ", 120)
	answer := body + "\n\n" + strings.Repeat("tail ", 80)
	open, tail := SplitAnswerOpeningTail(answer)
	if tail == "" || open == "" {
		t.Fatal("expected split")
	}
	if !strings.Contains(tail, "tail") {
		t.Fatalf("tail=%q", tail)
	}
}

func TestScoreAnswerRelevanceFromText_lowOnTailDrift(t *testing.T) {
	t.Parallel()
	dim := 8
	emb := stubEmbedder{fn: func(texts []string) [][]float32 {
		out := make([][]float32, len(texts))
		for i := range texts {
			if i == len(texts)-1 && len(texts) > 3 {
				out[i] = unitVec(dim, 1)
			} else {
				out[i] = unitVec(dim, 0)
			}
		}
		return out
	}}
	question := "What EC and pH targets for leafy greens?"
	opening := strings.Repeat("Lettuce EC 1.0-1.3 and pH 5.5. ", 30)
	tail := strings.Repeat("Endocrine disruptors in Lake Erie wildlife. ", 20)
	answer := opening + "\n\n" + tail
	rel, err := ScoreAnswerRelevanceFromText(context.Background(), emb, question, answer)
	if err != nil {
		t.Fatal(err)
	}
	if rel.QuestionAnswerCosine < 0.9 {
		t.Fatalf("opening should align with question: %v", rel.QuestionAnswerCosine)
	}
	if rel.OpeningTailCosine > 0.1 {
		t.Fatalf("tail should drift from opening: %v", rel.OpeningTailCosine)
	}
	if !rel.LowRelevance {
		t.Fatalf("expected low relevance due to tail drift: %+v", rel)
	}
}

func TestRelevanceMinFromEnv_default(t *testing.T) {
	t.Setenv("GUARDIAN_RELEVANCE_MIN", "")
	if got := RelevanceMinFromEnv(); got != defaultRelevanceMin {
		t.Fatalf("got %v", got)
	}
}
