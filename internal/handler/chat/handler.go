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
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/farmguardian/procedures"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/filestorage"
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
	cfg        ai.Config
	q          *db.Queries
	pool       *pgxpool.Pool
	llm        llm.ChatCompleter
	baseLLM    *llm.Client
	visionLLM  llm.ChatCompleter // optional multimodal client (Phase 30 WS6)
	modelCache *farmguardian.ModelCache
	fileStore  filestorage.Store
	embedder   embed.Embedder
	costGuard  farmguardian.CostGuardConfig // zero value = disabled
}

// NewHandler wires the configured chat + embedding clients when AI is enabled.
func NewHandler(pool *pgxpool.Pool, cfg ai.Config, fileStore filestorage.Store, modelCache *farmguardian.ModelCache) *Handler {
	h := &Handler{cfg: cfg, q: db.New(pool), pool: pool, fileStore: fileStore, modelCache: modelCache}
	if cfg.Enabled {
		if c, err := llm.NewChatClientFromEnv(); err == nil {
			h.llm = c
			h.baseLLM = c
		}
		if c, err := llm.NewVisionChatClientFromEnv(); err == nil {
			h.visionLLM = c
		}
		if e, err := embed.NewOpenAICompatibleFromEnv(); err == nil {
			h.embedder = e
		}
		h.costGuard = farmguardian.LoadCostGuardConfigFromEnv()
	}
	if h.modelCache == nil {
		h.modelCache = farmguardian.NewModelCache()
	}
	return h
}

// NewHandlerWithDeps is the test seam — inject any chat client or embedder
// (real or mock) without depending on env vars or a real pool.
func NewHandlerWithDeps(cfg ai.Config, q *db.Queries, client llm.ChatCompleter, embedder embed.Embedder) *Handler {
	h := &Handler{cfg: cfg, q: q, llm: client, embedder: embedder, modelCache: farmguardian.NewModelCache()}
	if c, ok := client.(*llm.Client); ok {
		h.baseLLM = c
	}
	return h
}

// WithModelCache overrides the Ollama model cache (tests).
func (h *Handler) WithModelCache(cache *farmguardian.ModelCache) *Handler {
	if cache != nil {
		h.modelCache = cache
	}
	return h
}

// WithVisionLLM sets the multimodal client for tests (Phase 30 WS6).
func (h *Handler) WithVisionLLM(client llm.ChatCompleter) *Handler {
	h.visionLLM = client
	return h
}

// WithFileStore sets file storage for vision attachment resolution in tests.
func (h *Handler) WithFileStore(store filestorage.Store) *Handler {
	h.fileStore = store
	return h
}

// WithCostGuard overrides the cost-guard config on h and returns h so the
// call can chain. Phase 27 WS5 follow-up test seam.
func (h *Handler) WithCostGuard(cfg farmguardian.CostGuardConfig) *Handler {
	h.costGuard = cfg
	return h
}

type postBody struct {
	Message string `json:"message"`
	FarmID  *int64 `json:"farm_id"`
	// ContextRef is the UI "Ask Guardian" anchor (Phase 29 WS6) — alert,
	// crop cycle, or zone the operator opened the drawer from.
	ContextRef *farmguardian.ContextRef `json:"context_ref,omitempty"`
	// NavHistory is the ordered list of recent routes the operator visited
	// before the current page (most recent first, max 3). Used to give the
	// Guardian breadcrumb context so starters don't need "I'm on page X".
	NavHistory []farmguardian.ContextRef `json:"nav_history,omitempty"`
	// SessionID is an opaque identifier the client can use to correlate turns.
	// When empty the handler generates a fresh UUID and returns it. When
	// supplied, prior turns in that session (owned by the same user) are
	// replayed into the prompt up to MaxHistoryTurns.
	SessionID string `json:"session_id,omitempty"`
	// Stream switches the response to Server-Sent Events when the LLM client
	// supports streaming (Phase 27 WS5 v3).
	Stream bool `json:"stream"`
	// AttachmentIDs lists zone reference photo file_attachments to include in a
	// multimodal user turn (Phase 30 WS6). Requires farm_id and vision LLM config.
	AttachmentIDs []int64 `json:"attachment_ids,omitempty"`
	// SetupMode opts into the setup-mode persona (Phase 44 WS4). Also activates
	// when the farm snapshot has zero zones or POST ?setup=1.
	SetupMode bool `json:"setup_mode,omitempty"`
	// Model overrides the chat model for this turn only (Phase 111).
	Model string `json:"model,omitempty"`
}

type postResponse struct {
	Answer           string                      `json:"answer"`
	LLMModel         string                      `json:"llm_model"`
	Grounded         bool                        `json:"grounded"`
	Citations        []synthesis.Citation        `json:"citations,omitempty"`
	ContextCount     int                         `json:"context_count"`
	EmbeddingID      string                      `json:"embedding_model_id,omitempty"`
	SessionID        string                      `json:"session_id,omitempty"`
	TurnIndex        int32                       `json:"turn_index"`
	PromptTokens     int                         `json:"prompt_tokens"`
	CompletionTokens int                         `json:"completion_tokens"`
	Proposals        []farmguardian.ActionProposal `json:"proposals,omitempty"`
	Procedure        *procedures.TurnPayload       `json:"procedure,omitempty"`
	FieldDegraded    bool                          `json:"field_degraded,omitempty"`
	VisionUsed       bool                          `json:"vision_used,omitempty"`
	AttachmentIDs    []int64                       `json:"attachment_ids,omitempty"`
	ModelUsed        string                        `json:"model_used,omitempty"`
	ModelFallback    bool                          `json:"model_fallback,omitempty"`
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

	var farmID int64
	if pb.FarmID != nil {
		farmID = *pb.FarmID
	}
	grounded := farmID > 0
	if farmID > 0 && !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	if pb.FarmID != nil && farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm_id")
		return
	}

	modelPreview := h.previewModelOutcome(r.Context(), pb.Model, farmID, grounded)
	effectiveWindow := h.effectiveContextWindowForModel(modelPreview.ModelName)
	advertisedWindow := h.advertisedContextWindowForModel(modelPreview.ModelName)
	promptBudget, trimLog := farmguardian.ComputePromptBudget(effectiveWindow, MaxHistoryTurns)
	h.logPromptBudgetTrims(modelPreview.ModelName, effectiveWindow, advertisedWindow, trimLog)

	// Resolve grounded vs plain path. Plain path needs only an LLM. When farm_id
	// is set we always attach the live snapshot; RAG chunk retrieval is optional
	// and only runs when an embedder is configured (EMBEDDING_API_KEY).
	var (
		chunks   []db.SearchRagNearestNeighborsFilteredRow
		system   = farmguardian.ChatSystemPrompt(h.cfg, h.llm != nil)
		user     = question
		liveSnap farmguardian.Snapshot
	)

	if grounded {
		// Live farm-state snapshot (WS4 follow-up). Best-effort: a snapshot
		// failure is logged but never blocks the chat turn.
		snapshotBlock := ""
		if h.q != nil {
			snap, serr := farmguardian.BuildSnapshot(r.Context(), h.q, farmID)
			if serr != nil {
				slog.Warn("farm guardian snapshot failed", "farm_id", farmID, "err", serr)
			} else {
				liveSnap = snap
				liveSnap.ApplyBudgetLimits(promptBudget.Snapshot)
			}
			snapshotBlock = liveSnap.PromptBlock()
		}

		system = farmguardian.ChatSystemPrompt(h.cfg, h.llm != nil) + "\n\n"
		if snapshotBlock != "" {
			system += snapshotBlock + "\n\n"
		}
		var readBlock string
		if h.q != nil {
			readBlock = farmguardian.EnrichPromptBlock(r.Context(), h.q, farmID, question, liveSnap, pb.ContextRef)
			if readBlock != "" {
				system += readBlock + "\n\n"
			}
		}
		if pb.ContextRef != nil {
			if focus := farmguardian.ContextRefPromptBlock(r.Context(), h.q, farmID, *pb.ContextRef, pb.NavHistory); focus != "" {
				system += focus + "\n\n"
			}
		}
		if uid, uok := authctx.UserID(r.Context()); uok {
			h.injectPriorSessionMemory(r.Context(), &system, farmID, uid, question, pb.ContextRef)
		}
		setupExplicit := pb.SetupMode || strings.TrimSpace(r.URL.Query().Get("setup")) == "1"
		if farmguardian.SetupModeActive(liveSnap, setupExplicit) {
			if setupBlock := farmguardian.SetupModePromptBlock(liveSnap); setupBlock != "" {
				system += setupBlock + "\n\n"
			}
		}

		if h.embedder != nil {
			var rerr error
			chunks, rerr = h.retrieveChunks(r.Context(), farmID, question, promptBudget.RAGTopK)
			if rerr != nil {
				slog.Warn("farm guardian retrieval failed", "farm_id", farmID, "err", rerr)
				if !farmguardian.IsLocalInferenceURL(strings.TrimSpace(os.Getenv("LLM_BASE_URL"))) {
					httputil.WriteError(w, http.StatusBadGateway, "retrieval failed")
					return
				}
				// Phase 37 WS1 — offline field mode: snapshot + procedures still work without RAG.
			} else if len(chunks) > 0 {
				if farmguardian.ReadBlockHasCropTargets(readBlock) {
					chunks = synthesis.StripNutrientNumbersFromChunks(chunks)
					system += synthesis.StructuredTruthRAGBlock() + "\n\n"
				}
				system += synthesis.GuardianRAGInstructions(chunks)
				user = synthesis.BuildUserMessage(question, chunks)
			} else {
				system += synthesis.ZeroChunkGuardBlock()
			}
		}
		if farmguardian.IsLocalInferenceURL(strings.TrimSpace(os.Getenv("LLM_BASE_URL"))) ||
			synthesis.HasFieldGuideChunks(chunks) {
			system += "\n\n" + farmguardian.FieldAssistantPromptBlock()
		}
	}

	// Resolve / validate session_id and load any prior history (DB only when
	// we have a real authenticated user_id — dev-bypass requests skip persistence).
	userID, hasUser := authctx.UserID(r.Context())
	sessionID, sessionErr := parseOrNewSession(pb.SessionID)
	if sessionErr != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid session_id")
		return
	}

	// Cost-guard check happens after we resolve user + farm but before we
	// touch the LLM. Returning early avoids spending tokens just to reject
	// the turn. Fails open when the DB lookup itself errors so a transient
	// Postgres hiccup doesn't take chat offline. Phase 27 WS5 follow-up.
	if !h.checkCostBudget(r.Context(), w, userID, hasUser, farmID) {
		return
	}

	if h.tryProcedureOrSafetyTurn(w, r, question, pb, sessionID, userID, hasUser, farmID, pb.Stream) {
		return
	}
	if h.tryFieldDegradeTurn(w, r, question, pb, sessionID, userID, hasUser, farmID, pb.Stream) {
		return
	}

	if h.llm == nil {
		msg := "Farm Guardian chat is not configured (set LLM_BASE_URL and LLM_MODEL)"
		if procedures.IsFieldRelatedQuestion(question) {
			msg = "LLM not configured; for field install use: start procedure wire-pi-relay-light (or list procedures)"
		}
		httputil.WriteError(w, http.StatusServiceUnavailable, msg)
		return
	}
	if h.fieldDegradeEligible() && !h.llmReachable(r.Context()) {
		httputil.WriteError(w, http.StatusServiceUnavailable,
			"local LLM is unreachable; for physical install use: start procedure <id> or GET /v1/field-guides/procedures/{id}/print")
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
			history = replayHistory(rows, promptBudget.MaxHistoryTurns)
		}
	}

	attachmentIDs := attachmentIDsFromRequest(pb.AttachmentIDs)
	if len(attachmentIDs) > 0 && farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "attachment_ids require farm_id")
		return
	}
	if len(attachmentIDs) > 0 && h.visionLLM == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "vision chat is not configured (set LLM_VISION_MODEL and LLM_BASE_URL or LLM_VISION_BASE_URL)")
		return
	}

	var visionImages []llm.ImageAttachment
	if len(attachmentIDs) > 0 {
		var verr error
		visionImages, verr = h.resolveVisionAttachments(r.Context(), farmID, attachmentIDs)
		if verr != nil {
			httputil.WriteError(w, http.StatusBadRequest, verr.Error())
			return
		}
		system += "\n\n" + farmguardian.VisionContextBlock()
		if pb.ContextRef != nil {
			if cropBlock := farmguardian.FieldPhotoCropGroundingBlock(r.Context(), h.q, farmID, pb.ContextRef); cropBlock != "" {
				system += "\n\n" + cropBlock
			}
		}
	}

	chatClient := h.llm
	modelOutcome := farmguardian.ResolveOutcome{}
	if len(visionImages) > 0 {
		chatClient = h.visionLLM
		if chatClient != nil {
			modelOutcome = farmguardian.ResolveOutcome{ModelName: chatClient.ModelLabel()}
		}
	} else {
		var resolved llm.ChatCompleter
		resolved, modelOutcome = h.resolveChatClient(r.Context(), pb.Model, farmID, grounded, false)
		if modelOutcome.RejectReason != "" {
			httputil.WriteError(w, http.StatusBadRequest, modelOutcome.RejectReason)
			return
		}
		if resolved != nil {
			chatClient = resolved
		}
	}
	if chatClient == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "Farm Guardian chat is not configured (set LLM_BASE_URL and LLM_MODEL)")
		return
	}

	maybeUnloadEmbedBeforeChat(r.Context(), chatClient, grounded)
	chatClient = applyChatClientForTurn(chatClient, grounded)

	currentUser := llm.UserMessageWithImages(user, visionImages)
	messages := buildMessages(system, history, currentUser)

	turnMeta := chatTurnMeta{
		farmID:          farmID,
		grounded:        grounded,
		stream:          pb.Stream,
		model:           chatClient.ModelLabel(),
		question:        question,
		historyTurns:    len(history) / 2,
		contextChunks:   len(chunks),
		effectiveWindow: effectiveWindow,
		sessionID:       sessionID.String(),
	}
	turnStarted := time.Now()
	h.logChatTurnStarted(r.Context(), turnMeta)

	if pb.Stream {
		// Pick the most capable streaming surface. Phase 27 WS5 follow-up:
		// UsageAwareStreamingChatCompleter (preferred) returns the OpenAI-style
		// usage block from the terminal SSE chunk so we can persist tokens for
		// streaming turns too. Older clients that only implement the legacy
		// interface still work — the adapter returns Usage{} and the streaming
		// turn lands with zero tokens (matches pre-follow-up behaviour).
		var stream streamFn
		if usageAware, ok := chatClient.(llm.UsageAwareStreamingChatCompleter); ok {
			stream = usageAware.ChatCompletionStreamMessagesWithUsage
		} else if legacy, ok := chatClient.(llm.MessagesStreamingChatCompleter); ok {
			stream = func(ctx context.Context, messages []llm.Message, onDelta func(string)) (llm.Usage, error) {
				err := legacy.ChatCompletionStreamMessages(ctx, messages, onDelta)
				return llm.Usage{}, err
			}
		} else {
			httputil.WriteError(w, http.StatusNotImplemented, "configured LLM client does not support streaming")
			return
		}
		h.streamResponse(w, r, stream, chatClient, messages, farmID, grounded, chunks, sessionID, userID, hasUser, question, liveSnap, len(visionImages) > 0, attachmentIDs, modelOutcome, turnMeta, turnStarted)
		return
	}

	var (
		answer string
		usage  llm.Usage
	)
	switch client := chatClient.(type) {
	case llm.UsageAwareChatCompleter:
		answer, usage, err = client.ChatCompletionMessagesWithUsage(r.Context(), messages)
	case llm.MessagesChatCompleter:
		answer, err = client.ChatCompletionMessages(r.Context(), messages)
	default:
		// Legacy fallback — shouldn't happen for *llm.Client but keeps mocks honest.
		answer, err = chatClient.ChatCompletion(r.Context(), system, user)
	}
	if err != nil {
		if errors.Is(err, r.Context().Err()) {
			return
		}
		h.logChatTurnFailed(r.Context(), turnMeta, turnStarted, err)
		ResetLLMReachabilityCache()
		if h.tryFieldDegradeTurn(w, r, question, pb, sessionID, userID, hasUser, farmID, pb.Stream) {
			return
		}
		payload := classifyLLMError(err)
		httputil.WriteJSON(w, http.StatusBadGateway, payload)
		return
	}
	answer, usage, repairOutcome := h.maybeRepairProposalAnswer(r.Context(), chatClient, messages, question, answer, usage)
	if repairOutcome.Attempted {
		slog.Info("guardian: proposal repair",
			"model", chatClient.ModelLabel(),
			"recovered", repairOutcome.Recovered,
			"parse_err", repairOutcome.ParseErr,
		)
	}
	if grounded {
		answer = synthesis.StripOrphanCitationRefs(answer, len(chunks))
	}

	resp := postResponse{
		Answer:           answer,
		LLMModel:         chatClient.ModelLabel(),
		Grounded:         grounded,
		SessionID:        sessionID.String(),
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		VisionUsed:       len(visionImages) > 0,
		AttachmentIDs:    attachmentIDs,
	}
	applyModelMeta(&resp, modelOutcome)
	if grounded {
		resp.Citations = synthesis.BuildCitations(answer, chunks)
		resp.ContextCount = len(chunks)
		if h.embedder != nil {
			resp.EmbeddingID = h.embedder.ModelID()
		}
	}

	if turnIdx, perr := h.persistTurn(r.Context(), sessionID, userID, hasUser, farmID, grounded, question, answer, resp.Citations, len(chunks), usage, chatClient.ModelLabel()); perr == nil {
		resp.TurnIndex = turnIdx
	}
	h.attachProposals(r.Context(), farmID, hasUser, userID, sessionID, question, answer, liveSnap, &resp)

	slog.Info("guardian: chat turn completed",
		"request_id", authctx.RequestID(r.Context()),
		"farm_id", farmID,
		"session_id", sessionID,
		"model", chatClient.ModelLabel(),
		"vision", len(visionImages) > 0,
		"grounded", grounded,
		"context_chunks", len(chunks),
		"citations", len(resp.Citations),
		"history_turns", len(history)/2,
		"prompt_tokens", usage.PromptTokens,
		"completion_tokens", usage.CompletionTokens,
		"elapsed_ms", time.Since(turnStarted).Milliseconds(),
		"stream", false,
	)
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// streamFn is the per-request streaming closure used by streamResponse — it
// wraps either UsageAwareStreamingChatCompleter or the legacy
// MessagesStreamingChatCompleter into a single signature.
type streamFn func(ctx context.Context, messages []llm.Message, onDelta func(string)) (llm.Usage, error)

func (h *Handler) streamResponse(
	w http.ResponseWriter,
	r *http.Request,
	stream streamFn,
	chatClient llm.ChatCompleter,
	messages []llm.Message,
	farmID int64,
	grounded bool,
	chunks []db.SearchRagNearestNeighborsFilteredRow,
	sessionID uuid.UUID,
	userID uuid.UUID,
	hasUser bool,
	question string,
	liveSnap farmguardian.Snapshot,
	visionUsed bool,
	attachmentIDs []int64,
	modelOutcome farmguardian.ResolveOutcome,
	turnMeta chatTurnMeta,
	turnStarted time.Time,
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

	if grounded {
		sendEvent("status", map[string]string{
			"phase":   "generating",
			"message": "Generating answer — running on CPU (no GPU). Grounded turns may take several minutes; wait before sending another message.",
		})
	} else {
		sendEvent("status", map[string]string{
			"phase":   "generating",
			"message": "Generating answer — phi3 on CPU can take several minutes for the first token.",
		})
	}

	var collected strings.Builder
	var firstTokenAt time.Time
	onDelta := func(delta string) {
		if firstTokenAt.IsZero() && delta != "" {
			firstTokenAt = time.Now()
			slog.Info("guardian: first token",
				"request_id", authctx.RequestID(r.Context()),
				"model", chatClient.ModelLabel(),
				"grounded", grounded,
				"ttft_ms", firstTokenAt.Sub(turnStarted).Milliseconds(),
			)
		}
		collected.WriteString(delta)
		sendEvent("delta", map[string]string{"text": delta})
	}

	usage, streamErr := stream(r.Context(), messages, onDelta)
	if streamErr != nil {
		if errors.Is(streamErr, r.Context().Err()) {
			return
		}
		h.logChatTurnFailed(r.Context(), turnMeta, turnStarted, streamErr)
		ResetLLMReachabilityCache()
		payload := classifyLLMError(streamErr)
		sendEvent("error", payload)
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
		return
	}

	answer := collected.String()
	if grounded {
		answer = synthesis.StripOrphanCitationRefs(answer, len(chunks))
	}
	done := postResponse{
		Answer:           answer,
		LLMModel:         chatClient.ModelLabel(),
		Grounded:         grounded,
		SessionID:        sessionID.String(),
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		VisionUsed:       visionUsed,
		AttachmentIDs:    attachmentIDs,
	}
	applyModelMeta(&done, modelOutcome)
	if grounded {
		done.Citations = synthesis.BuildCitations(answer, chunks)
		done.ContextCount = len(chunks)
		if h.embedder != nil {
			done.EmbeddingID = h.embedder.ModelID()
		}
	}

	// Phase 27 WS5 follow-up: backends that support stream_options.include_usage
	// (OpenAI + recent Ollama) return real token counts; older Ollama builds
	// leave usage zero and the row still lands so the UI sidebar count stays
	// honest about "this session had N streaming turns".
	if turnIdx, perr := h.persistTurn(r.Context(), sessionID, userID, hasUser, farmID, grounded, question, answer, done.Citations, len(chunks), usage, chatClient.ModelLabel()); perr == nil {
		done.TurnIndex = turnIdx
	}
	h.attachProposals(r.Context(), farmID, hasUser, userID, sessionID, question, answer, liveSnap, &done)

	sendEvent("done", done)
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()

	slog.Info("guardian: chat turn completed",
		"request_id", authctx.RequestID(r.Context()),
		"farm_id", farmID,
		"session_id", sessionID,
		"model", chatClient.ModelLabel(),
		"vision", visionUsed,
		"grounded", grounded,
		"context_chunks", len(chunks),
		"citations", len(done.Citations),
		"prompt_tokens", usage.PromptTokens,
		"completion_tokens", usage.CompletionTokens,
		"elapsed_ms", time.Since(turnStarted).Milliseconds(),
		"ttft_ms", ttftMs(firstTokenAt, turnStarted),
	)
}

func ttftMs(firstTokenAt, started time.Time) int64 {
	if firstTokenAt.IsZero() {
		return 0
	}
	return firstTokenAt.Sub(started).Milliseconds()
}

// checkCostBudget runs the per-user / per-farm rolling-window cap. Returns
// true when the request is allowed to continue; false when a 429 has already
// been written and the caller should return. Fails open (returns true) when
// guards are disabled, there's no authenticated user, no DB, or the DB
// lookup errors — operator-facing errors take priority over budget
// enforcement. Phase 27 WS5 follow-up.
func (h *Handler) checkCostBudget(ctx context.Context, w http.ResponseWriter, userID uuid.UUID, hasUser bool, farmID int64) bool {
	if !h.costGuard.AnyEnabled() {
		return true
	}
	if !hasUser || h.q == nil {
		return true
	}
	decision, err := farmguardian.CheckBudget(ctx, h.q, h.costGuard, userID, farmID)
	if err != nil {
		slog.Warn("chat cost guard query failed", "user_id", userID, "farm_id", farmID, "err", err)
		return true
	}
	if decision.Allowed {
		return true
	}
	retrySec := int(decision.RetryAfter.Seconds())
	if retrySec < 1 {
		retrySec = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(retrySec))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	body, _ := json.Marshal(map[string]any{
		"error":               "chat token budget exceeded; try again later",
		"reason":              decision.Reason,
		"used_tokens":         decision.UsedTokens,
		"max_tokens":          decision.MaxTokens,
		"window_seconds":      decision.WindowSeconds,
		"retry_after_seconds": retrySec,
	})
	_, _ = w.Write(body)
	slog.Info("chat cost guard rejected request",
		"user_id", userID,
		"farm_id", farmID,
		"reason", decision.Reason,
		"used_tokens", decision.UsedTokens,
		"max_tokens", decision.MaxTokens,
	)
	return false
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
// turn was not persisted (so the caller can omit it from the response). Also
// upserts the matching conversation_sessions row so updated_at tracks the
// latest activity (used for sidebar ordering).
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
	usage llm.Usage,
	llmModel string,
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
		LlmModel:         llmModel,
		Grounded:         grounded,
		ContextCount:     int32(contextCount),
		Citations:        citationsJSON,
		PromptTokens:     int32(usage.PromptTokens),
		CompletionTokens: int32(usage.CompletionTokens),
	})
	if err != nil {
		slog.Warn("conversation_turns insert failed", "session_id", sessionID, "err", err)
		return -1, err
	}
	if uerr := h.q.UpsertConversationSession(ctx, db.UpsertConversationSessionParams{
		ID:     sessionID,
		UserID: userID,
	}); uerr != nil {
		slog.Warn("conversation_sessions upsert failed", "session_id", sessionID, "err", uerr)
	}
	// Phase 28 WS5 — fire the chat-budget-warning alert when the
	// just-persisted turn pushes the user across 80% of their cap.
	// Best-effort: the helper swallows DB errors so a transient hiccup
	// here never breaks the chat turn.
	if h.costGuard.PerUserMaxTokens > 0 && farmID > 0 {
		if res, werr := farmguardian.MaybeFireBudgetWarning(ctx, h.q, h.costGuard, userID, farmID); werr != nil {
			slog.Warn("chat budget warning failed", "user_id", userID, "farm_id", farmID, "err", werr)
		} else if res.Fired {
			slog.Info("chat budget warning fired",
				"user_id", userID, "farm_id", farmID,
				"pct_used", res.PctUsed,
				"used_tokens", res.UsedTokens,
				"max_tokens", res.MaxTokens,
				"alert_id", res.AlertID,
			)
		}
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
func buildMessages(system string, history []llm.Message, currentUser llm.Message) []llm.Message {
	out := make([]llm.Message, 0, 2+len(history))
	out = append(out, llm.Message{Role: "system", Content: system})
	out = append(out, history...)
	out = append(out, currentUser)
	return out
}

// ──────────────────────────────────────────────────────────────────────────
// History endpoints (GET /v1/chat/sessions, GET /v1/chat/sessions/{session_id})
// ──────────────────────────────────────────────────────────────────────────

type sessionSummary struct {
	SessionID             string   `json:"session_id"`
	Title                 *string  `json:"title,omitempty"`
	TurnCount             int32    `json:"turn_count"`
	LastTurnAt            string   `json:"last_turn_at"`
	AnyGrounded           bool     `json:"any_grounded"`
	FirstUserMessage      string   `json:"first_user_message"`
	LastAssistantMessage  string   `json:"last_assistant_message"`
	LastFarmID            *int64   `json:"last_farm_id,omitempty"`
	TotalPromptTokens     int32    `json:"total_prompt_tokens"`
	TotalCompletionTokens int32    `json:"total_completion_tokens"`
	Topics                []string `json:"topics,omitempty"` // Phase 63
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
	PromptTokens     int32                `json:"prompt_tokens"`
	CompletionTokens int32                `json:"completion_tokens"`
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
	topicRows, _ := h.q.ListSessionSummaryTopicsForUser(r.Context(), userID)
	topicBySession := map[uuid.UUID][]string{}
	for _, tr := range topicRows {
		topicBySession[tr.SessionID] = tr.Topics
	}
	out := make([]sessionSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, sessionSummary{
			SessionID:             row.SessionID.String(),
			Title:                 row.Title,
			TurnCount:             row.TurnCount,
			LastTurnAt:            row.LastTurnAt.UTC().Format("2006-01-02T15:04:05Z"),
			AnyGrounded:           row.AnyGrounded,
			FirstUserMessage:      row.FirstUserMessage,
			LastAssistantMessage:  row.LastAssistantMessage,
			LastFarmID:            row.LastFarmID,
			TotalPromptTokens:     row.TotalPromptTokens,
			TotalCompletionTokens: row.TotalCompletionTokens,
			Topics:                topicBySession[row.SessionID],
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
			PromptTokens:     row.PromptTokens,
			CompletionTokens: row.CompletionTokens,
			CreatedAt:        row.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"session_id": sessionID.String(),
		"turns":      out,
	})
}

// PatchSession handles PATCH /v1/chat/sessions/{session_id} — operator rename.
// Empty / whitespace-only title resets to NULL so the UI falls back to the
// first user message.
func (h *Handler) PatchSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	sessionID, err := uuid.Parse(r.PathValue("session_id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid session_id")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<10))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var patch struct {
		Title *string `json:"title"`
	}
	if len(body) > 0 {
		if jerr := json.Unmarshal(body, &patch); jerr != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
	}

	titlePtr := normaliseTitle(patch.Title)

	row, err := h.q.UpdateConversationSessionTitle(r.Context(), db.UpdateConversationSessionTitleParams{
		Title:  titlePtr,
		ID:     sessionID,
		UserID: userID,
	})
	if err != nil {
		// UPDATE … RETURNING with zero matched rows surfaces as pgx.ErrNoRows
		// (Scan never executes). Map it to 404 so the UI can react cleanly.
		if isNoRowsErr(err) {
			httputil.WriteError(w, http.StatusNotFound, "session not found")
			return
		}
		slog.Warn("session rename failed", "session_id", sessionID, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "rename failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"session_id": row.ID.String(),
		"title":      row.Title,
		"updated_at": row.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// DeleteSession handles DELETE /v1/chat/sessions/{session_id} — removes every
// turn for the (session_id, user_id) pair plus the metadata row. Idempotent:
// re-deleting the same session id returns 204 either way (no information leak
// about whether the row ever existed).
func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	sessionID, err := uuid.Parse(r.PathValue("session_id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid session_id")
		return
	}
	if derr := h.q.DeleteConversationTurnsBySession(r.Context(), db.DeleteConversationTurnsBySessionParams{
		SessionID: sessionID,
		UserID:    userID,
	}); derr != nil {
		slog.Warn("session delete turns failed", "session_id", sessionID, "err", derr)
		httputil.WriteError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	if serr := h.q.DeleteSessionSummary(r.Context(), db.DeleteSessionSummaryParams{
		SessionID: sessionID,
		UserID:    userID,
	}); serr != nil {
		slog.Warn("session delete summary failed", "session_id", sessionID, "err", serr)
	}
	if _, derr := h.q.DeleteConversationSession(r.Context(), db.DeleteConversationSessionParams{
		ID:     sessionID,
		UserID: userID,
	}); derr != nil {
		slog.Warn("session delete metadata failed", "session_id", sessionID, "err", derr)
		// Turns already gone — surface a softer message rather than 500.
	}
	w.WriteHeader(http.StatusNoContent)
}

// normaliseTitle trims whitespace and treats "" as "clear the title".
func normaliseTitle(in *string) *string {
	if in == nil {
		return nil
	}
	t := strings.TrimSpace(*in)
	if t == "" {
		return nil
	}
	if utf8.RuneCountInString(t) > 120 {
		// Trim to the first 120 runes (byte-safe).
		out := []rune(t)[:120]
		t = string(out) + "…"
	}
	return &t
}

// isNoRowsErr matches the pgx error returned by Scan when an UPDATE … RETURNING
// touched zero rows. Avoids depending on a specific pgx import path at this
// layer.
func isNoRowsErr(err error) bool {
	return err != nil && (err.Error() == "no rows in result set" || strings.Contains(err.Error(), "no rows in result set"))
}
