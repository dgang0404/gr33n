// Package chat exposes Phase 27 Farm Guardian endpoints.
//
//   - WS5 v1: non-streaming single-turn completion behind AI_ENABLED.
//   - WS5 v2: optional farm_id → pgvector retrieval → grounded prompt with
//     [n] citations, same shape as POST /farms/{id}/rag/answer.
//   - WS5 v3: streaming SSE response (stream:true).
//   - WS5 v4: DB-backed conversation_turns + multi-turn history replay.
package chat

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/rag/embed"
	"gr33n-api/internal/rag/llm"
	"gr33n-api/internal/rag/synthesis"
)

// MaxHistoryTurns caps how many prior turns get replayed into the prompt. Older
// turns still live in the DB and are visible via GET /v1/chat/sessions/{id},
// they just don't get re-sent to the LLM (keeps the context window predictable).
const MaxHistoryTurns = 20

// MaxRecentSessions caps GET /v1/chat/sessions results.
const MaxRecentSessions = 50

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
	// When empty the handler generates a fresh UUID and returns it. When
	// supplied, prior turns in that session (owned by the same user) are
	// replayed into the prompt up to MaxHistoryTurns.
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
	TurnIndex    int32                `json:"turn_index"`
}

// PostV1 handles POST /v1/chat — JWT required by route wiring.
//
// History replay: when session_id is present and the calling user owns prior
// turns in that session, those (user, assistant) pairs are interleaved into
// the prompt before the current user message, capped at MaxHistoryTurns.
//
// Persistence: after a successful turn (both streaming and non-streaming) the
// (user_message, assistant_message, llm_model, grounded, context_count,
// citations) tuple is inserted into gr33ncore.conversation_turns. The
// turn_index is assigned by the SQL COALESCE(MAX+1, 0) subquery so concurrent
// inserts can't collide (UNIQUE (session_id, turn_index)).
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

	// Resolve / validate session_id and load any prior history (DB only when
	// we have a real authenticated user_id — dev-bypass requests skip persistence).
	userID, hasUser := authctx.UserID(r.Context())
	sessionID, sessionErr := parseOrNewSession(pb.SessionID)
	if sessionErr != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid session_id")
		return
	}

	var history []llm.Message
	if hasUser && pb.SessionID != "" && h.q != nil {
		rows, herr := h.q.ListConversationTurnsBySession(r.Context(), db.ListConversationTurnsBySessionParams{
			SessionID: sessionID,
			UserID:    userID,
		})
		if herr != nil {
			slog.Warn("conversation history load failed", "session_id", sessionID, "err", herr)
		} else {
			history = replayHistory(rows, MaxHistoryTurns)
		}
	}

	messages := buildMessages(system, history, user)

	if pb.Stream {
		streamer, ok := h.llm.(llm.MessagesStreamingChatCompleter)
		if !ok {
			httputil.WriteError(w, http.StatusNotImplemented, "configured LLM client does not support streaming")
			return
		}
		h.streamResponse(w, r, streamer, messages, farmID, grounded, chunks, sessionID, userID, hasUser, question)
		return
	}

	mc, ok := h.llm.(llm.MessagesChatCompleter)
	var answer string
	if ok {
		answer, err = mc.ChatCompletionMessages(r.Context(), messages)
	} else {
		// Legacy fallback — shouldn't happen for *llm.Client but keeps mocks honest.
		answer, err = h.llm.ChatCompletion(r.Context(), system, user)
	}
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
		SessionID: sessionID.String(),
	}
	if grounded {
		resp.Citations = synthesis.BuildCitations(answer, chunks)
		resp.ContextCount = len(chunks)
		if h.embedder != nil {
			resp.EmbeddingID = h.embedder.ModelID()
		}
	}

	if turnIdx, perr := h.persistTurn(r.Context(), sessionID, userID, hasUser, farmID, grounded, question, answer, resp.Citations, len(chunks)); perr == nil {
		resp.TurnIndex = turnIdx
	}

	slog.Info("farm guardian chat completed",
		"farm_id", farmID,
		"session_id", sessionID,
		"model", h.llm.ModelLabel(),
		"grounded", grounded,
		"context_chunks", len(chunks),
		"citations", len(resp.Citations),
		"history_turns", len(history)/2,
	)
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) streamResponse(
	w http.ResponseWriter,
	r *http.Request,
	streamer llm.MessagesStreamingChatCompleter,
	messages []llm.Message,
	farmID int64,
	grounded bool,
	chunks []db.SearchRagNearestNeighborsFilteredRow,
	sessionID uuid.UUID,
	userID uuid.UUID,
	hasUser bool,
	question string,
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

	streamErr := streamer.ChatCompletionStreamMessages(r.Context(), messages, onDelta)
	if streamErr != nil {
		if errors.Is(streamErr, r.Context().Err()) {
			return
		}
		slog.Warn("farm guardian stream failed", "farm_id", farmID, "err", streamErr)
		sendEvent("error", map[string]string{"error": "LLM request failed"})
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
		return
	}

	answer := collected.String()
	done := postResponse{
		Answer:    answer,
		LLMModel:  h.llm.ModelLabel(),
		Grounded:  grounded,
		SessionID: sessionID.String(),
	}
	if grounded {
		done.Citations = synthesis.BuildCitations(answer, chunks)
		done.ContextCount = len(chunks)
		if h.embedder != nil {
			done.EmbeddingID = h.embedder.ModelID()
		}
	}

	if turnIdx, perr := h.persistTurn(r.Context(), sessionID, userID, hasUser, farmID, grounded, question, answer, done.Citations, len(chunks)); perr == nil {
		done.TurnIndex = turnIdx
	}

	sendEvent("done", done)
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()

	slog.Info("farm guardian chat streamed",
		"farm_id", farmID,
		"session_id", sessionID,
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

// persistTurn inserts the just-completed (user, assistant) pair when we have
// authenticated state and a DB. Returns the assigned turn_index or -1 if the
// turn was not persisted (so the caller can omit it from the response).
func (h *Handler) persistTurn(
	ctx context.Context,
	sessionID uuid.UUID,
	userID uuid.UUID,
	hasUser bool,
	farmID int64,
	grounded bool,
	question, answer string,
	citations []synthesis.Citation,
	contextCount int,
) (int32, error) {
	if !hasUser || h.q == nil {
		return -1, nil
	}
	var farmPtr *int64
	if farmID > 0 {
		f := farmID
		farmPtr = &f
	}
	citationsJSON := []byte("[]")
	if len(citations) > 0 {
		if b, err := json.Marshal(citations); err == nil {
			citationsJSON = b
		}
	}
	row, err := h.q.InsertConversationTurn(ctx, db.InsertConversationTurnParams{
		SessionID:        sessionID,
		UserID:           userID,
		FarmID:           farmPtr,
		UserMessage:      question,
		AssistantMessage: answer,
		LlmModel:         h.llm.ModelLabel(),
		Grounded:         grounded,
		ContextCount:     int32(contextCount),
		Citations:        citationsJSON,
	})
	if err != nil {
		slog.Warn("conversation_turns insert failed", "session_id", sessionID, "err", err)
		return -1, err
	}
	return row.TurnIndex, nil
}

// parseOrNewSession validates the inbound session_id (must be a UUID) and
// generates a fresh one when the client omits it.
func parseOrNewSession(s string) (uuid.UUID, error) {
	if strings.TrimSpace(s) == "" {
		return uuid.New(), nil
	}
	return uuid.Parse(s)
}

// replayHistory builds a flat (user, assistant) message slice from stored turns,
// most-recent capped at maxTurns. Older turns are dropped from the head.
func replayHistory(rows []db.ListConversationTurnsBySessionRow, maxTurns int) []llm.Message {
	if len(rows) == 0 {
		return nil
	}
	if maxTurns > 0 && len(rows) > maxTurns {
		rows = rows[len(rows)-maxTurns:]
	}
	out := make([]llm.Message, 0, len(rows)*2)
	for _, row := range rows {
		out = append(out,
			llm.Message{Role: "user", Content: row.UserMessage},
			llm.Message{Role: "assistant", Content: row.AssistantMessage},
		)
	}
	return out
}

// buildMessages assembles the final OpenAI-style messages slice (system, then
// history pairs, then current user message). When history is empty this
// matches the v1 / v2 shape exactly so existing callers and tests are unchanged.
func buildMessages(system string, history []llm.Message, currentUser string) []llm.Message {
	out := make([]llm.Message, 0, 2+len(history))
	out = append(out, llm.Message{Role: "system", Content: system})
	out = append(out, history...)
	out = append(out, llm.Message{Role: "user", Content: currentUser})
	return out
}

// ──────────────────────────────────────────────────────────────────────────
// History endpoints (GET /v1/chat/sessions, GET /v1/chat/sessions/{session_id})
// ──────────────────────────────────────────────────────────────────────────

type sessionSummary struct {
	SessionID            string `json:"session_id"`
	TurnCount            int32  `json:"turn_count"`
	LastTurnAt           string `json:"last_turn_at"`
	AnyGrounded          bool   `json:"any_grounded"`
	FirstUserMessage     string `json:"first_user_message"`
	LastAssistantMessage string `json:"last_assistant_message"`
	LastFarmID           *int64 `json:"last_farm_id,omitempty"`
}

type sessionTurn struct {
	TurnIndex        int32                `json:"turn_index"`
	UserMessage      string               `json:"user_message"`
	AssistantMessage string               `json:"assistant_message"`
	LLMModel         string               `json:"llm_model"`
	Grounded         bool                 `json:"grounded"`
	ContextCount     int32                `json:"context_count"`
	Citations        []synthesis.Citation `json:"citations,omitempty"`
	FarmID           *int64               `json:"farm_id,omitempty"`
	CreatedAt        string               `json:"created_at"`
}

// ListSessions handles GET /v1/chat/sessions — returns the calling user's most
// recently active sessions (capped at MaxRecentSessions).
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	rows, err := h.q.ListRecentConversationSessions(r.Context(), db.ListRecentConversationSessionsParams{
		UserID:     userID,
		MatchLimit: MaxRecentSessions,
	})
	if err != nil {
		slog.Warn("list sessions failed", "user_id", userID, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load sessions")
		return
	}
	out := make([]sessionSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, sessionSummary{
			SessionID:            row.SessionID.String(),
			TurnCount:            row.TurnCount,
			LastTurnAt:           row.LastTurnAt.UTC().Format("2006-01-02T15:04:05Z"),
			AnyGrounded:          row.AnyGrounded,
			FirstUserMessage:     row.FirstUserMessage,
			LastAssistantMessage: row.LastAssistantMessage,
			LastFarmID:           row.LastFarmID,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"sessions": out})
}

// GetSession handles GET /v1/chat/sessions/{session_id} — ordered turn history
// for the session, scoped to the calling user.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	idStr := r.PathValue("session_id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid session_id")
		return
	}
	rows, err := h.q.ListConversationTurnsBySession(r.Context(), db.ListConversationTurnsBySessionParams{
		SessionID: sessionID,
		UserID:    userID,
	})
	if err != nil {
		slog.Warn("list session turns failed", "session_id", sessionID, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load session")
		return
	}
	out := make([]sessionTurn, 0, len(rows))
	for _, row := range rows {
		var cites []synthesis.Citation
		if len(row.Citations) > 0 {
			_ = json.Unmarshal(row.Citations, &cites)
		}
		out = append(out, sessionTurn{
			TurnIndex:        row.TurnIndex,
			UserMessage:      row.UserMessage,
			AssistantMessage: row.AssistantMessage,
			LLMModel:         row.LlmModel,
			Grounded:         row.Grounded,
			ContextCount:     row.ContextCount,
			Citations:        cites,
			FarmID:           row.FarmID,
			CreatedAt:        row.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"session_id": sessionID.String(),
		"turns":      out,
	})
}
