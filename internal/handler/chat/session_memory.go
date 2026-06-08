// Phase 63 — Guardian session memory HTTP handlers.

package chat

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

type sessionSummaryResponse struct {
	SessionID   string   `json:"session_id"`
	SummaryText string   `json:"summary_text"`
	Topics      []string `json:"topics"`
	CreatedAt   string   `json:"created_at,omitempty"`
}

// CloseSession handles POST /v1/chat/sessions/{session_id}/close — summarizes
// the session for cross-session memory when the operator starts a new chat.
func (h *Handler) CloseSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	var body struct {
		FarmID *int64 `json:"farm_id"`
	}
	if r.Body != nil {
		raw, _ := io.ReadAll(io.LimitReader(r.Body, 4<<10))
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &body)
		}
	}

	row, err := h.summarizeSession(r.Context(), userID, sessionID, body.FarmID)
	if err != nil {
		if err == errSessionNotFound {
			httputil.WriteError(w, http.StatusNotFound, "session not found")
			return
		}
		if err == errSessionEmpty {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		slog.Warn("session close summarize failed", "session_id", sessionID, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "summary failed")
		return
	}
	if row == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, sessionSummaryResponse{
		SessionID:   row.SessionID.String(),
		SummaryText: row.SummaryText,
		Topics:      row.Topics,
		CreatedAt:   row.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// RecentMemory handles GET /v1/farms/{id}/guardian-memory/recent — newest
// summary whose topics overlap the optional route query param.
func (h *Handler) RecentMemory(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	route := strings.TrimSpace(r.URL.Query().Get("route"))
	topics := farmguardian.TopicsForRoute(route)
	if len(topics) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	row, err := h.q.FindMatchingSessionSummary(r.Context(), db.FindMatchingSessionSummaryParams{
		FarmID: farmID,
		UserID: userID,
		Topics: topics,
	})
	if err != nil {
		if isNoRowsErr(err) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"session_id":   row.SessionID.String(),
		"summary_text": row.SummaryText,
		"topics":       row.Topics,
		"prompt":       farmguardian.RecentTopicPrompt(row),
		"created_at":   row.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// ExportMemory handles GET /v1/farms/{id}/guardian-memory/export — plain text.
func (h *Handler) ExportMemory(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	rows, err := h.q.ListSessionSummariesByFarmUser(r.Context(), db.ListSessionSummariesByFarmUserParams{
		FarmID:     farmID,
		UserID:     userID,
		MatchLimit: 200,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "export failed")
		return
	}
	farmName := ""
	if farm, ferr := h.q.GetFarmByID(r.Context(), farmID); ferr == nil {
		farmName = strings.TrimSpace(farm.Name)
	}
	text := farmguardian.FormatSessionSummariesExport(rows, farmName)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="guardian-memory-export.txt"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(text))
}

// ClearMemory handles DELETE /v1/farms/{id}/guardian-memory — removes all
// summaries for this operator on the farm (sessions/turns untouched).
func (h *Handler) ClearMemory(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if _, err := h.q.DeleteAllSessionSummariesForFarmUser(r.Context(), db.DeleteAllSessionSummariesForFarmUserParams{
		FarmID: farmID,
		UserID: userID,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "clear failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var (
	errSessionNotFound = errSentinel("session not found")
	errSessionEmpty    = errSentinel("session empty")
)

type errSentinel string

func (e errSentinel) Error() string { return string(e) }

// summarizeSession builds and stores a session summary when turns exist.
func (h *Handler) summarizeSession(ctx context.Context, userID, sessionID uuid.UUID, farmID *int64) (*db.Gr33ncoreSessionSummary, error) {
	turns, err := h.q.ListConversationTurnsBySession(ctx, db.ListConversationTurnsBySessionParams{
		SessionID: sessionID,
		UserID:    userID,
	})
	if err != nil {
		return nil, err
	}
	if len(turns) == 0 {
		return nil, errSessionEmpty
	}

	resolvedFarm := int64(0)
	if farmID != nil && *farmID > 0 {
		resolvedFarm = *farmID
	}
	for i := len(turns) - 1; i >= 0 && resolvedFarm <= 0; i-- {
		if turns[i].FarmID != nil && *turns[i].FarmID > 0 {
			resolvedFarm = *turns[i].FarmID
		}
	}
	if resolvedFarm <= 0 {
		return nil, errSessionEmpty
	}

	summaryText, topics, serr := farmguardian.BuildSessionSummaryText(turns, h.llm)
	if serr != nil {
		return nil, serr
	}
	if strings.TrimSpace(summaryText) == "" {
		return nil, errSessionEmpty
	}

	row, uerr := h.q.UpsertSessionSummary(ctx, db.UpsertSessionSummaryParams{
		SessionID:   sessionID,
		FarmID:      resolvedFarm,
		UserID:      userID,
		SummaryText: summaryText,
		Topics:      topics,
	})
	if uerr != nil {
		return nil, uerr
	}
	out := row
	return &out, nil
}

func (h *Handler) injectPriorSessionMemory(ctx context.Context, system *string, farmID int64, userID uuid.UUID, question string, ref *farmguardian.ContextRef) {
	if h.q == nil || farmID <= 0 || system == nil {
		return
	}
	topics := farmguardian.MatchingTopicsForTurn(question, ref)
	if len(topics) == 0 {
		return
	}
	row, err := h.q.FindMatchingSessionSummary(ctx, db.FindMatchingSessionSummaryParams{
		FarmID: farmID,
		UserID: userID,
		Topics: topics,
	})
	if err != nil {
		return
	}
	if block := farmguardian.PriorSessionContextBlock(row, time.Now().UTC()); block != "" {
		*system += block + "\n\n"
	}
}
