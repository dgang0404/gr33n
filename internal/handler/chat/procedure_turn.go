package chat

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/farmguardian/procedures"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/rag/llm"
)

const procedureModelLabel = "field-procedure"

func (h *Handler) tryProcedureOrSafetyTurn(
	w http.ResponseWriter,
	r *http.Request,
	question string,
	pb postBody,
	sessionID uuid.UUID,
	userID uuid.UUID,
	hasUser bool,
	farmID int64,
	stream bool,
) bool {
	if farmguardian.UnsafeInstructionRequest(question) {
		h.writeStaticTurn(w, r, question, pb, sessionID, userID, hasUser, farmID, stream, farmguardian.SafetyStopMessage, nil, procedureModelLabel, false)
		return true
	}

	if strings.Contains(strings.ToLower(question), "list procedure") {
		summary, err := procedures.ListSummary(procedures.RepoRoot())
		if err == nil {
			h.writeStaticTurn(w, r, question, pb, sessionID, userID, hasUser, farmID, stream, summary, nil, procedureModelLabel, false)
			return true
		}
	}

	meta := h.loadSessionMeta(r, sessionID, userID, hasUser)
	handled, answer, newMeta, payload := procedures.HandleTurn(procedures.RepoRoot(), question, meta)
	if !handled {
		return false
	}
	if hasUser && h.q != nil {
		_ = h.q.UpsertConversationSession(r.Context(), db.UpsertConversationSessionParams{
			ID: sessionID, UserID: userID,
		})
		_ = h.q.UpdateConversationSessionMeta(r.Context(), db.UpdateConversationSessionMetaParams{
			ID:     sessionID,
			UserID: userID,
			Meta:   newMeta.Marshal(),
		})
	}
	h.writeStaticTurn(w, r, question, pb, sessionID, userID, hasUser, farmID, stream, answer, payload, procedureModelLabel, false)
	return true
}

func (h *Handler) loadSessionMeta(r *http.Request, sessionID, userID uuid.UUID, hasUser bool) procedures.SessionMeta {
	if !hasUser || h.q == nil {
		return procedures.SessionMeta{}
	}
	row, err := h.q.GetConversationSessionMeta(r.Context(), db.GetConversationSessionMetaParams{
		ID:     sessionID,
		UserID: userID,
	})
	if err != nil {
		return procedures.SessionMeta{}
	}
	return procedures.ParseSessionMeta(row)
}

func (h *Handler) writeStaticTurn(
	w http.ResponseWriter,
	r *http.Request,
	question string,
	pb postBody,
	sessionID uuid.UUID,
	userID uuid.UUID,
	hasUser bool,
	farmID int64,
	stream bool,
	answer string,
	procedure *procedures.TurnPayload,
	modelLabel string,
	fieldDegraded bool,
) {
	resp := postResponse{
		Answer:        answer,
		LLMModel:      modelLabel,
		Grounded:      farmID > 0,
		SessionID:     sessionID.String(),
		Procedure:     procedure,
		FieldDegraded: fieldDegraded,
	}
	if turnIdx, err := h.persistTurn(r.Context(), sessionID, userID, hasUser, farmID, resp.Grounded, question, answer, nil, 0, llm.Usage{}, modelLabel, ""); err == nil {
		resp.TurnIndex = turnIdx
	}
	if stream {
		h.writeStaticStream(w, resp)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) writeStaticStream(w http.ResponseWriter, resp postResponse) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		httputil.WriteJSON(w, http.StatusOK, resp)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	send := func(eventType string, payload any) {
		b, _ := json.Marshal(payload)
		_, _ = w.Write([]byte("event: " + eventType + "\ndata: " + string(b) + "\n\n"))
		flusher.Flush()
	}
	send("delta", map[string]string{"text": resp.Answer})
	send("done", resp)
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()
}
