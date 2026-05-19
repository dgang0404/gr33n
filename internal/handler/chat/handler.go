// Package chat exposes Phase 27 Farm Guardian endpoints.
//
//   - WS5 v1 (shipped): non-streaming single-turn completion behind AI_ENABLED.
//   - WS5 v2 (shipped): optional farm_id → pgvector retrieval → grounded prompt
//     with [n] citations, same shape as POST /farms/{id}/rag/answer.
//
// Streaming (SSE) and DB-backed session history remain follow-ups in the plan.
package chat

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"gr33n-api/internal/ai"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/rag/embed"
	"gr33n-api/internal/rag/llm"
	"gr33n-api/internal/rag/synthesis"
)

// Handler exposes Phase 27 Farm Guardian routes.
type Handler struct {
	cfg      ai.Config
	q        *db.Queries
	llm      llm.ChatCompleter
	embedder embed.Embedder
}

// NewHandler wires the configured chat + embedding clients when AI is enabled.
// When AI is off, both stay nil and POST /v1/chat answers 503 — same contract
// as POST /farms/{id}/rag/answer in Lite mode.
func NewHandler(pool *pgxpool.Pool, cfg ai.Config) *Handler {
	h := &Handler{cfg: cfg, q: db.New(pool)}
	if cfg.Enabled {
		if c, err := llm.NewChatClientFromEnv(); err == nil {
			h.llm = c
		}
		if e, err := embed.NewOpenAICompatibleFromEnv(); err == nil {
			h.embedder = e
		}
	}
	return h
}

// NewHandlerWithDeps is the test seam — inject any chat client or embedder
// (real or mock) without depending on env vars or a real pool.
func NewHandlerWithDeps(cfg ai.Config, q *db.Queries, client llm.ChatCompleter, embedder embed.Embedder) *Handler {
	return &Handler{cfg: cfg, q: q, llm: client, embedder: embedder}
}

type postBody struct {
	Message string `json:"message"`
	FarmID  *int64 `json:"farm_id"`
	// SessionID is an opaque identifier the client can use to correlate turns.
	// v2 of /v1/chat does not persist turns yet — the value is echoed back so
	// clients can adopt a session model today and gain history when the
	// conversation_turns table lands in a follow-up slice (Phase 27 WS5 v4).
	SessionID string `json:"session_id,omitempty"`
	// Stream switches the response to Server-Sent Events when the LLM client
	// supports streaming (Phase 27 WS5 v3).
	Stream bool `json:"stream"`
}

type postResponse struct {
	Answer       string               `json:"answer"`
	LLMModel     string               `json:"llm_model"`
	Grounded     bool                 `json:"grounded"`
	Citations    []synthesis.Citation `json:"citations,omitempty"`
	ContextCount int                  `json:"context_count"`
	EmbeddingID  string               `json:"embedding_model_id,omitempty"`
	SessionID    string               `json:"session_id,omitempty"`
}

// PostV1 handles POST /v1/chat — JWT required by route wiring. When body
// includes a `farm_id` the caller must be a member of that farm; the API
// retrieves the top farmguardian.RAGTopK chunks via pgvector and injects
// them into the user message exactly like /farms/{id}/rag/answer does.
func (h *Handler) PostV1(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.cfg.Enabled {
		httputil.WriteError(w, http.StatusServiceUnavailable, "AI features are disabled on this installation")
		return
	}
	if h.llm == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "Farm Guardian chat is not configured (set LLM_BASE_URL and LLM_MODEL)")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "request body required")
		return
	}
	var pb postBody
	if err := json.Unmarshal(body, &pb); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	question, err := farmguardian.BuildUserMessage(pb.Message)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Resolve grounded vs plain path. Plain path needs only an LLM; grounded
	// path also requires farm membership + embedder.
	var (
		grounded bool
		chunks   []db.SearchRagNearestNeighborsFilteredRow
		system   = farmguardian.SystemPrompt()
		user     = question
		farmID   int64
	)

	if pb.FarmID != nil {
		farmID = *pb.FarmID
		if farmID <= 0 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid farm_id")
			return
		}
		if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
			return
		}
		if h.embedder == nil {
			httputil.WriteError(w, http.StatusServiceUnavailable, "RAG retrieval is not configured (set EMBEDDING_API_KEY)")
			return
		}
		var rerr error
		chunks, rerr = h.retrieveChunks(r.Context(), farmID, question, farmguardian.RAGTopK)
		if rerr != nil {
			slog.Warn("farm guardian retrieval failed", "farm_id", farmID, "err", rerr)
			httputil.WriteError(w, http.StatusBadGateway, "retrieval failed")
			return
		}
		grounded = true
		system = farmguardian.SystemPrompt() + "\n\n" + synthesis.SystemPrompt()
		user = synthesis.BuildUserMessage(question, chunks)
	}

	if pb.Stream {
		streamer, ok := h.llm.(llm.StreamingChatCompleter)
		if !ok {
			httputil.WriteError(w, http.StatusNotImplemented, "configured LLM client does not support streaming")
			return
		}
		h.streamResponse(w, r, streamer, system, user, farmID, grounded, chunks, pb.SessionID)
		return
	}

	answer, err := h.llm.ChatCompletion(r.Context(), system, user)
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			return
		}
		slog.Warn("farm guardian chat failed", "farm_id", farmID, "grounded", grounded, "err", err)
		httputil.WriteError(w, http.StatusBadGateway, "LLM request failed")
		return
	}
	resp := postResponse{
		Answer:    answer,
		LLMModel:  h.llm.ModelLabel(),
		Grounded:  grounded,
		SessionID: pb.SessionID,
	}
	if grounded {
		resp.Citations = synthesis.BuildCitations(answer, chunks)
		resp.ContextCount = len(chunks)
		if h.embedder != nil {
			resp.EmbeddingID = h.embedder.ModelID()
		}
	}
	slog.Info("farm guardian chat completed",
		"farm_id", farmID,
		"model", h.llm.ModelLabel(),
		"grounded", grounded,
		"context_chunks", len(chunks),
		"citations", len(resp.Citations),
	)
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) streamResponse(
	w http.ResponseWriter,
	r *http.Request,
	streamer llm.StreamingChatCompleter,
	system, user string,
	farmID int64,
	grounded bool,
	chunks []db.SearchRagNearestNeighborsFilteredRow,
	sessionID string,
) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		httputil.WriteError(w, http.StatusInternalServerError, "server does not support streaming")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	sendEvent := func(eventType string, payload any) bool {
		b, _ := json.Marshal(payload)
		_, werr := w.Write([]byte("event: " + eventType + "\ndata: " + string(b) + "\n\n"))
		if werr != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	var collected strings.Builder
	onDelta := func(delta string) {
		collected.WriteString(delta)
		sendEvent("delta", map[string]string{"text": delta})
	}

	streamErr := streamer.ChatCompletionStream(r.Context(), system, user, onDelta)
	if streamErr != nil {
		if errors.Is(streamErr, r.Context().Err()) {
			// Client gone — nothing more to send.
			return
		}
		slog.Warn("farm guardian stream failed", "farm_id", farmID, "err", streamErr)
		sendEvent("error", map[string]string{"error": "LLM request failed"})
		// Terminate the stream cleanly even on error.
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
		return
	}

	answer := collected.String()
	done := postResponse{
		Answer:    answer,
		LLMModel:  h.llm.ModelLabel(),
		Grounded:  grounded,
		SessionID: sessionID,
	}
	if grounded {
		done.Citations = synthesis.BuildCitations(answer, chunks)
		done.ContextCount = len(chunks)
		if h.embedder != nil {
			done.EmbeddingID = h.embedder.ModelID()
		}
	}
	sendEvent("done", done)
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()

	slog.Info("farm guardian chat streamed",
		"farm_id", farmID,
		"model", h.llm.ModelLabel(),
		"grounded", grounded,
		"context_chunks", len(chunks),
		"citations", len(done.Citations),
	)
}

func (h *Handler) retrieveChunks(ctx context.Context, farmID int64, query string, topK int) ([]db.SearchRagNearestNeighborsFilteredRow, error) {
	vecs, err := h.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(vecs) != 1 || len(vecs[0]) == 0 {
		return nil, errors.New("invalid embedding response")
	}
	if topK <= 0 {
		topK = farmguardian.RAGTopK
	}
	return h.q.SearchRagNearestNeighborsFiltered(ctx, db.SearchRagNearestNeighborsFilteredParams{
		QueryEmbedding: pgvector.NewVector(vecs[0]),
		FarmID:         farmID,
		MatchLimit:     int32(topK),
	})
}
