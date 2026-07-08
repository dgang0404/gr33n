package chat

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

const maxFeedbackReasonLen = 500

// PatchTurnFeedback handles PATCH /v1/chat/sessions/{session_id}/turns/{turn_index}/feedback.
func (h *Handler) PatchTurnFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.q == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "database not configured")
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
	turnIndex, err := strconv.ParseInt(r.PathValue("turn_index"), 10, 32)
	if err != nil || turnIndex < 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid turn_index")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<10))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var patch struct {
		Rating *string `json:"rating"`
		Reason *string `json:"reason"`
	}
	if len(body) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "empty body")
		return
	}
	if jerr := json.Unmarshal(body, &patch); jerr != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	rating := strings.TrimSpace(ptrStr(patch.Rating))
	if rating != "up" && rating != "down" {
		httputil.WriteError(w, http.StatusBadRequest, "rating must be up or down")
		return
	}
	reason := strings.TrimSpace(ptrStr(patch.Reason))
	if len(reason) > maxFeedbackReasonLen {
		httputil.WriteError(w, http.StatusBadRequest, "reason too long (max 500 chars)")
		return
	}

	turns, err := h.q.ListConversationTurnsBySession(r.Context(), db.ListConversationTurnsBySessionParams{
		SessionID: sessionID,
		UserID:    userID,
	})
	if err != nil {
		slog.Warn("feedback list turns failed", "session_id", sessionID, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load session")
		return
	}
	var farmID *int64
	found := false
	for _, t := range turns {
		if t.TurnIndex == int32(turnIndex) {
			found = true
			farmID = t.FarmID
			break
		}
	}
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "turn not found")
		return
	}
	if farmID != nil && *farmID > 0 {
		if !farmauthz.RequireFarmMember(w, r, h.q, *farmID) {
			return
		}
	}

	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}
	row, err := h.q.UpdateConversationTurnFeedback(r.Context(), db.UpdateConversationTurnFeedbackParams{
		FeedbackRating: &rating,
		FeedbackReason: reasonPtr,
		SessionID:      sessionID,
		UserID:         userID,
		TurnIndex:      int32(turnIndex),
	})
	if err != nil {
		if isNoRowsErr(err) {
			httputil.WriteError(w, http.StatusNotFound, "turn not found")
			return
		}
		slog.Warn("feedback update failed", "session_id", sessionID, "turn_index", turnIndex, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "feedback update failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"session_id":       row.SessionID.String(),
		"turn_index":       row.TurnIndex,
		"feedback_rating":  row.FeedbackRating,
		"feedback_reason":  row.FeedbackReason,
		"feedback_at":      formatPGTime(row.FeedbackAt),
	})
}

// ExportFeedback handles GET /v1/chat/feedback/export?farm_id=&since=7d.
func (h *Handler) ExportFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.q == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}
	farmID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("farm_id")), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "farm_id required")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	since, err := parseFeedbackSince(r.URL.Query().Get("since"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	rows, err := h.q.ListConversationFeedbackForFarm(r.Context(), db.ListConversationFeedbackForFarmParams{
		FarmID: &farmID,
		Since:  pgtype.Timestamptz{Time: since, Valid: true},
	})
	if err != nil {
		slog.Warn("feedback export failed", "farm_id", farmID, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "export failed")
		return
	}

	ratingFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("rating")))

	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "csv" {
		writeFeedbackCSV(w, rows)
		return
	}

	out := make([]feedbackExportRow, 0, len(rows))
	for _, row := range rows {
		exp := feedbackExportRowFromDB(row)
		if ratingFilter != "" && !strings.EqualFold(exp.Rating, ratingFilter) {
			continue
		}
		out = append(out, exp)
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"farm_id": farmID,
		"since":   since.UTC().Format(time.RFC3339),
		"rows":    out,
	})
}

type feedbackExportRow struct {
	SessionID      string  `json:"session_id"`
	TurnIndex      int32   `json:"turn_index"`
	Question       string  `json:"question"`
	AnswerExcerpt  string  `json:"answer_excerpt"`
	Rating         string  `json:"rating"`
	Reason         *string `json:"reason,omitempty"`
	Grounded       bool    `json:"grounded"`
	Model          string  `json:"model"`
	CreatedAt      string  `json:"created_at"`
	FeedbackAt     string  `json:"feedback_at,omitempty"`
}

func feedbackExportRowFromDB(row db.ListConversationFeedbackForFarmRow) feedbackExportRow {
	rating := ""
	if row.FeedbackRating != nil {
		rating = *row.FeedbackRating
	}
	return feedbackExportRow{
		SessionID:     row.SessionID.String(),
		TurnIndex:     row.TurnIndex,
		Question:      row.UserMessage,
		AnswerExcerpt: excerptText(row.AssistantMessage, 240),
		Rating:        rating,
		Reason:        row.FeedbackReason,
		Grounded:      row.Grounded,
		Model:         row.LlmModel,
		CreatedAt:     row.CreatedAt.UTC().Format(time.RFC3339),
		FeedbackAt:    formatPGTime(row.FeedbackAt),
	}
}

func writeFeedbackCSV(w http.ResponseWriter, rows []db.ListConversationFeedbackForFarmRow) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("session_id,turn_index,question,answer_excerpt,rating,reason,grounded,model,created_at,feedback_at\n"))
	for _, row := range rows {
		exp := feedbackExportRowFromDB(row)
		line := csvEscape(exp.SessionID) + "," +
			strconv.FormatInt(int64(exp.TurnIndex), 10) + "," +
			csvEscape(exp.Question) + "," +
			csvEscape(exp.AnswerExcerpt) + "," +
			csvEscape(exp.Rating) + "," +
			csvEscape(ptrStr(exp.Reason)) + "," +
			strconv.FormatBool(exp.Grounded) + "," +
			csvEscape(exp.Model) + "," +
			csvEscape(exp.CreatedAt) + "," +
			csvEscape(exp.FeedbackAt) + "\n"
		_, _ = w.Write([]byte(line))
	}
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func excerptText(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func parseFeedbackSince(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = "7d"
	}
	if strings.HasSuffix(raw, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(raw, "d"))
		if err != nil || days < 1 {
			return time.Time{}, errInvalidSince()
		}
		return time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour), nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, errInvalidSince()
	}
	return t.UTC(), nil
}

func errInvalidSince() error {
	return errSinceParse{msg: "since must be like 7d or RFC3339 timestamp"}
}

type errSinceParse struct{ msg string }

func (e errSinceParse) Error() string { return e.msg }

func formatPGTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
