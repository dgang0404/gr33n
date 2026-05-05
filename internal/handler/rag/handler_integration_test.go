package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/embed"
	"gr33n-api/internal/rag/llm"
)

type fakeEmbedder struct {
	vec []float32
}

func (f *fakeEmbedder) ModelID() string { return "fake-embed-model" }

func (f *fakeEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = f.vec
	}
	return out, nil
}

type fakeLLM struct {
	reply string
}

func (f *fakeLLM) ModelLabel() string { return "fake-llm" }

func (f *fakeLLM) ChatCompletion(ctx context.Context, system, user string) (string, error) {
	if f.reply != "" {
		return f.reply, nil
	}
	return "Synthetic answer citing [1].", nil
}

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	u := os.Getenv("DATABASE_URL")
	if u == "" {
		t.Skip("DATABASE_URL not set — skipping RAG integration test")
	}
	pool, err := pgxpool.New(context.Background(), u)
	if err != nil {
		t.Fatalf("pgx pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func testVec() []float32 {
	v := make([]float32, embed.DefaultExpectedDims)
	v[0] = 1
	v[1] = 0.02
	return v
}

func TestIntegration_SearchReturnsUpsertedChunk(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	vec := testVec()
	q := db.New(pool)
	const srcType = "integration_ws5"
	const srcID int64 = 919191001
	_, err := q.UpsertRagEmbeddingChunk(ctx, db.UpsertRagEmbeddingChunkParams{
		FarmID:      1,
		SourceType:  srcType,
		SourceID:    srcID,
		ChunkIndex:  0,
		ContentText: "WS5 integration chunk content",
		Embedding:   pgvector.NewVector(vec),
		ModelID:     "fake-embed-model",
		Metadata:    []byte(`{"module":"tasks"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM gr33ncore.rag_embedding_chunks WHERE farm_id = $1 AND source_type = $2 AND source_id = $3`,
			1, srcType, srcID)
	})

	h := NewHandlerForTest(pool, &fakeEmbedder{vec: vec}, nil, allowAllSynth{})
	req := httptest.NewRequest(http.MethodGet, "/farms/1/rag/search?q=hello", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
	rr := httptest.NewRecorder()
	h.Search(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	results, _ := body["results"].([]any)
	if len(results) < 1 {
		t.Fatalf("expected >=1 result, got %#v", body)
	}
}

func TestIntegration_SearchOtherFarmDoesNotSeeChunks(t *testing.T) {
	pool := testPool(t)
	vec := testVec()
	h := NewHandlerForTest(pool, &fakeEmbedder{vec: vec}, nil, allowAllSynth{})
	req := httptest.NewRequest(http.MethodGet, "/farms/999/rag/search?q=hello", nil)
	req.SetPathValue("id", "999")
	req = req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
	rr := httptest.NewRecorder()
	h.Search(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	results, _ := body["results"].([]any)
	if len(results) != 0 {
		t.Fatalf("expected no chunks for other farm_id, got %d", len(results))
	}
}

func TestIntegration_AnswerWithFakeLLM(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	vec := testVec()
	q := db.New(pool)
	const srcType = "integration_ws5_answer"
	const srcID int64 = 919191002
	_, err := q.UpsertRagEmbeddingChunk(ctx, db.UpsertRagEmbeddingChunkParams{
		FarmID:      1,
		SourceType:  srcType,
		SourceID:    srcID,
		ChunkIndex:  0,
		ContentText: "Tomatoes in zone A need topping.",
		Embedding:   pgvector.NewVector(vec),
		ModelID:     "fake-embed-model",
		Metadata:    []byte(`{}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM gr33ncore.rag_embedding_chunks WHERE farm_id = $1 AND source_type = $2 AND source_id = $3`,
			1, srcType, srcID)
	})

	llm := &fakeLLM{reply: "Summary [1]."}
	h := NewHandlerForTest(pool, &fakeEmbedder{vec: vec}, llm, allowAllSynth{})
	body := map[string]any{"query": "What about tomatoes?"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/farms/1/rag/answer", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	req = req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
	rr := httptest.NewRecorder()
	h.Answer(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["answer"] != "Summary [1]." {
		t.Fatalf("answer mismatch: %#v", out["answer"])
	}
}

func TestIntegration_AnswerRateLimited(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	vec := testVec()
	q := db.New(pool)
	const srcType = "integration_ws5_rl"
	const srcID int64 = 919191003
	_, err := q.UpsertRagEmbeddingChunk(ctx, db.UpsertRagEmbeddingChunkParams{
		FarmID:      1,
		SourceType:  srcType,
		SourceID:    srcID,
		ChunkIndex:  0,
		ContentText: "rate limit body",
		Embedding:   pgvector.NewVector(vec),
		ModelID:     "fake",
		Metadata:    []byte(`{}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM gr33ncore.rag_embedding_chunks WHERE farm_id = $1 AND source_type = $2 AND source_id = $3`,
			1, srcType, srcID)
	})

	gate := NewTestSynthGlobalLimiter(2)
	llm := &fakeLLM{}
	h := NewHandlerForTest(pool, &fakeEmbedder{vec: vec}, llm, gate)

	doAnswer := func() int {
		body := map[string]any{"query": "q"}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/farms/1/rag/answer", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", "1")
		req = req.WithContext(authctx.WithFarmAuthzSkip(context.Background(), true))
		rr := httptest.NewRecorder()
		h.Answer(rr, req)
		return rr.Code
	}
	if c := doAnswer(); c != http.StatusOK {
		t.Fatalf("first status %d", c)
	}
	if c := doAnswer(); c != http.StatusOK {
		t.Fatalf("second status %d", c)
	}
	if c := doAnswer(); c != http.StatusTooManyRequests {
		t.Fatalf("third status want 429 got %d", c)
	}
}

var _ llm.ChatCompleter = (*fakeLLM)(nil)
