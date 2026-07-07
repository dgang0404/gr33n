// Phase 145 WS1 — embedding-based answer relevance / topic drift signals.

package farmguardian

import (
	"context"
	"math"
	"os"
	"strconv"
	"strings"
)

const defaultRelevanceMin = 0.35

// TextEmbedder produces vectors for relevance scoring (implemented by rag/embed.Client).
type TextEmbedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// AnswerRelevance captures question↔answer and opening↔tail cosine scores.
type AnswerRelevance struct {
	QuestionAnswerCosine float64 `json:"question_answer_cosine,omitempty"`
	OpeningTailCosine    float64 `json:"opening_tail_cosine,omitempty"`
	LowRelevance         bool    `json:"low_relevance,omitempty"`
	MinThreshold         float64 `json:"relevance_min_threshold,omitempty"`
}

// RelevanceMinFromEnv returns GUARDIAN_RELEVANCE_MIN (default 0.35).
func RelevanceMinFromEnv() float64 {
	s := strings.TrimSpace(os.Getenv("GUARDIAN_RELEVANCE_MIN"))
	if s == "" {
		return defaultRelevanceMin
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v < 0 || v > 1 {
		return defaultRelevanceMin
	}
	return v
}

// CosineSimilarity returns cosine similarity for equal-length vectors in [0,1] (0 if undefined).
func CosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		af := float64(a[i])
		bf := float64(b[i])
		dot += af * bf
		normA += af * af
		normB += bf * bf
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// SplitAnswerOpeningTail splits long answers for tail-drift detection.
func SplitAnswerOpeningTail(answer string) (opening, tail string) {
	answer = strings.TrimSpace(answer)
	if len(answer) < 500 {
		return answer, ""
	}
	// Prefer first paragraph break after ~400 chars.
	start := 400
	if start > len(answer) {
		start = len(answer) / 2
	}
	if idx := strings.Index(answer[start:], "\n\n"); idx >= 0 {
		cut := start + idx
		return strings.TrimSpace(answer[:cut]), strings.TrimSpace(answer[cut:])
	}
	// Fallback: split at 60% on last newline.
	cut := (len(answer) * 6) / 10
	if nl := strings.LastIndex(answer[:cut], "\n"); nl > 200 {
		cut = nl
	}
	return strings.TrimSpace(answer[:cut]), strings.TrimSpace(answer[cut:])
}

// ScoreAnswerRelevance computes relevance metrics from precomputed embeddings.
func ScoreAnswerRelevance(questionVec, answerVec, openingVec, tailVec []float32, min float64) AnswerRelevance {
	if min <= 0 {
		min = defaultRelevanceMin
	}
	out := AnswerRelevance{
		QuestionAnswerCosine: CosineSimilarity(questionVec, answerVec),
		MinThreshold:         min,
	}
	if len(tailVec) > 0 && len(openingVec) == len(tailVec) {
		out.OpeningTailCosine = CosineSimilarity(openingVec, tailVec)
	}
	out.LowRelevance = out.QuestionAnswerCosine < min ||
		(len(tailVec) > 0 && out.OpeningTailCosine < min)
	return out
}

// ScoreAnswerRelevanceFromText embeds question/answer segments and scores relevance.
func ScoreAnswerRelevanceFromText(ctx context.Context, embedder TextEmbedder, question, answer string) (AnswerRelevance, error) {
	question = strings.TrimSpace(question)
	answer = strings.TrimSpace(answer)
	if embedder == nil || question == "" || answer == "" {
		return AnswerRelevance{}, nil
	}
	opening, tail := SplitAnswerOpeningTail(answer)
	texts := []string{question, answer, opening}
	var tailIdx int
	if tail != "" && len(tail) >= 80 {
		texts = append(texts, tail)
		tailIdx = 3
	}
	vecs, err := embedder.Embed(ctx, texts)
	if err != nil {
		return AnswerRelevance{}, err
	}
	if len(vecs) != len(texts) {
		return AnswerRelevance{}, nil
	}
	var tailVec []float32
	if tailIdx > 0 && tailIdx < len(vecs) {
		tailVec = vecs[tailIdx]
	}
	return ScoreAnswerRelevance(vecs[0], vecs[1], vecs[2], tailVec, RelevanceMinFromEnv()), nil
}

// AnswerLowRelevance reports whether relevance scores fail the configured threshold.
func AnswerLowRelevance(rel AnswerRelevance) bool {
	return rel.LowRelevance
}
