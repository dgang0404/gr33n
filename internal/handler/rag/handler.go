package rag

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/rag/embed"
	"gr33n-api/internal/rag/llm"
	"gr33n-api/internal/rag/synthesis"
)

const (
	maxRAGQueryRunes       = 8000
	maxRAGResults          = 50
	defaultRAGLimit        = 10
	defaultAnswerContext   = 8
	maxAnswerContextChunks = 15
)

type Handler struct {
	q            *db.Queries
	embedder     embed.Embedder
	llm          *llm.Client
	synthLimiter *minuteLimiter
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	h := &Handler{
		q:            db.New(pool),
		synthLimiter: newSynthLimiterFromEnv(),
	}
	if emb, err := embed.NewOpenAICompatibleFromEnv(); err == nil {
		h.embedder = emb
	}
	if c, err := llm.NewChatClientFromEnv(); err == nil {
		h.llm = c
	}
	return h
}

type searchArgs struct {
	Query        string
	Module       *string
	CreatedSince pgtype.Timestamptz
	CreatedUntil pgtype.Timestamptz
	Limit        int32
}

func (h *Handler) retrieveFilteredChunks(ctx context.Context, farmID int64, args searchArgs) ([]db.SearchRagNearestNeighborsFilteredRow, error) {
	vecs, err := h.embedder.Embed(ctx, []string{args.Query})
	if err != nil {
		return nil, err
	}
	if len(vecs) != 1 || len(vecs[0]) == 0 {
		return nil, errors.New("invalid embedding response")
	}
	qv := pgvector.NewVector(vecs[0])
	return h.q.SearchRagNearestNeighborsFiltered(ctx, db.SearchRagNearestNeighborsFilteredParams{
		QueryEmbedding: qv,
		FarmID:         farmID,
		Module:         args.Module,
		CreatedSince:   args.CreatedSince,
		CreatedUntil:   args.CreatedUntil,
		MatchLimit:     args.Limit,
	})
}

// Search handles GET and POST /farms/{id}/rag/search — JWT + farm membership; vector similarity with optional filters.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	if h.embedder == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "RAG search is not configured (set EMBEDDING_API_KEY)")
		return
	}

	var args searchArgs
	switch r.Method {
	case http.MethodGet:
		args, err = parseSearchGET(r)
	case http.MethodPost:
		args, err = parseSearchPOST(r)
	default:
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateSearchArgs(&args); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	rows, err := h.retrieveFilteredChunks(ctx, farmID, args)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "embedding request failed")
		return
	}

	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		item := map[string]any{
			"id":           row.ID,
			"farm_id":      row.FarmID,
			"source_type":  row.SourceType,
			"source_id":    row.SourceID,
			"chunk_index":  row.ChunkIndex,
			"content_text": row.ContentText,
			"model_id":     row.ModelID,
			"created_at":   row.CreatedAt.UTC().Format(time.RFC3339Nano),
			"updated_at":   row.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		if d, ok := distanceToFloat64(row.Distance); ok && !math.IsNaN(d) {
			item["distance"] = d
		}
		if len(row.Metadata) > 0 {
			var meta any
			if err := json.Unmarshal(row.Metadata, &meta); err == nil {
				item["metadata"] = meta
			} else {
				item["metadata"] = json.RawMessage(row.Metadata)
			}
		} else {
			item["metadata"] = map[string]any{}
		}
		out = append(out, item)
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"results":  out,
		"model_id": h.embedder.ModelID(),
	})
}

func validateSearchArgs(args *searchArgs) error {
	if strings.TrimSpace(args.Query) == "" {
		return errors.New("query is required")
	}
	if utf8.RuneCountInString(args.Query) > maxRAGQueryRunes {
		return errors.New("query too long")
	}
	if args.Limit <= 0 {
		args.Limit = defaultRAGLimit
	}
	if args.Limit > maxRAGResults {
		args.Limit = maxRAGResults
	}
	return nil
}

// Answer handles POST /farms/{id}/rag/answer — retrieval + optional LLM synthesis with bracket citations [n].
func (h *Handler) Answer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	if h.embedder == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "RAG search is not configured (set EMBEDDING_API_KEY)")
		return
	}
	if h.llm == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "answer synthesis is not configured (set LLM_BASE_URL and LLM_MODEL)")
		return
	}
	if !h.synthLimiter.Allow() {
		httputil.WriteError(w, http.StatusTooManyRequests, "synthesis rate limit exceeded (see RAG_SYNTHESIS_MAX_PER_MINUTE)")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var pb answerPostBody
	if len(strings.TrimSpace(string(body))) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "request body required")
		return
	}
	if err := json.Unmarshal(body, &pb); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	args, err := pb.toSearchArgs()
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateSearchArgs(&args); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctxLimit := defaultAnswerContext
	if pb.MaxContextChunks != nil && *pb.MaxContextChunks > 0 {
		ctxLimit = int(*pb.MaxContextChunks)
	}
	if ctxLimit > maxAnswerContextChunks {
		ctxLimit = maxAnswerContextChunks
	}
	args.Limit = int32(ctxLimit)

	chunks, err := h.retrieveFilteredChunks(ctx, farmID, args)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "retrieval failed")
		return
	}
	if len(chunks) == 0 {
		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"answer":             "No indexed chunks matched your filters or query; ingest data with rag-ingest or broaden filters.",
			"citations":          []synthesis.Citation{},
			"embedding_model_id": h.embedder.ModelID(),
			"llm_model":          h.llm.Model,
			"context_count":      0,
		})
		return
	}

	userMsg := synthesis.BuildUserMessage(args.Query, chunks)
	answerText, err := h.llm.ChatCompletion(ctx, synthesis.SystemPrompt(), userMsg)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "LLM request failed")
		return
	}

	cites := synthesis.BuildCitations(answerText, chunks)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"answer":             answerText,
		"citations":          cites,
		"embedding_model_id": h.embedder.ModelID(),
		"llm_model":          h.llm.Model,
		"context_count":      len(chunks),
	})
}

type answerPostBody struct {
	Query            string  `json:"query"`
	Module           *string `json:"module"`
	CreatedSince     *string `json:"created_since"`
	CreatedUntil     *string `json:"created_until"`
	MaxContextChunks *int32  `json:"max_context_chunks"`
}

func (pb *answerPostBody) toSearchArgs() (searchArgs, error) {
	var a searchArgs
	a.Query = strings.TrimSpace(pb.Query)
	if pb.Module != nil {
		m := strings.TrimSpace(*pb.Module)
		if m != "" {
			a.Module = &m
		}
	}
	if pb.CreatedSince != nil && strings.TrimSpace(*pb.CreatedSince) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*pb.CreatedSince))
		if err != nil {
			return a, err
		}
		a.CreatedSince = pgtype.Timestamptz{Time: t, Valid: true}
	}
	if pb.CreatedUntil != nil && strings.TrimSpace(*pb.CreatedUntil) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*pb.CreatedUntil))
		if err != nil {
			return a, err
		}
		a.CreatedUntil = pgtype.Timestamptz{Time: t, Valid: true}
	}
	a.Limit = defaultAnswerContext
	return a, nil
}

func parseSearchGET(r *http.Request) (searchArgs, error) {
	q := r.URL.Query()
	var a searchArgs
	a.Query = strings.TrimSpace(q.Get("q"))
	if m := strings.TrimSpace(q.Get("module")); m != "" {
		a.Module = &m
	}
	if s := strings.TrimSpace(q.Get("since")); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return a, err
		}
		a.CreatedSince = pgtype.Timestamptz{Time: t, Valid: true}
	}
	if s := strings.TrimSpace(q.Get("until")); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return a, err
		}
		a.CreatedUntil = pgtype.Timestamptz{Time: t, Valid: true}
	}
	if lim := strings.TrimSpace(q.Get("limit")); lim != "" {
		n, err := strconv.ParseInt(lim, 10, 32)
		if err != nil || n < 1 {
			return a, errors.New("invalid limit")
		}
		a.Limit = int32(n)
	}
	return a, nil
}

type postBody struct {
	Query        string  `json:"query"`
	Module       *string `json:"module"`
	CreatedSince *string `json:"created_since"`
	CreatedUntil *string `json:"created_until"`
	Limit        *int32  `json:"limit"`
}

func parseSearchPOST(r *http.Request) (searchArgs, error) {
	var a searchArgs
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return a, err
	}
	var pb postBody
	if len(strings.TrimSpace(string(body))) == 0 {
		return a, errors.New("request body required")
	}
	if err := json.Unmarshal(body, &pb); err != nil {
		return a, err
	}
	a.Query = strings.TrimSpace(pb.Query)
	if pb.Module != nil {
		m := strings.TrimSpace(*pb.Module)
		if m != "" {
			a.Module = &m
		}
	}
	if pb.CreatedSince != nil && strings.TrimSpace(*pb.CreatedSince) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*pb.CreatedSince))
		if err != nil {
			return a, err
		}
		a.CreatedSince = pgtype.Timestamptz{Time: t, Valid: true}
	}
	if pb.CreatedUntil != nil && strings.TrimSpace(*pb.CreatedUntil) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*pb.CreatedUntil))
		if err != nil {
			return a, err
		}
		a.CreatedUntil = pgtype.Timestamptz{Time: t, Valid: true}
	}
	if pb.Limit != nil && *pb.Limit > 0 {
		a.Limit = *pb.Limit
	}
	return a, nil
}

func distanceToFloat64(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int64:
		return float64(x), true
	case int32:
		return float64(x), true
	case int:
		return float64(x), true
	case string:
		f, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}
