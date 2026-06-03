package chat

import (
	"net/http"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian/procedures"
)

const fieldDegradeModelLabel = "field-degrade"

func (h *Handler) tryFieldDegradeTurn(
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
	if h.llm != nil && h.llmReachable(r.Context()) {
		return false
	}
	if !ProceduresAvailable() {
		return false
	}
	meta := h.loadSessionMeta(r, sessionID, userID, hasUser)
	answer, newMeta, payload, ok := procedures.TryFieldDegrade(procedures.RepoRoot(), question, meta)
	if !ok {
		return false
	}
	if hasUser && h.q != nil {
		_ = h.q.UpdateConversationSessionMeta(r.Context(), db.UpdateConversationSessionMetaParams{
			ID: sessionID, UserID: userID, Meta: newMeta.Marshal(),
		})
	}
	h.writeStaticTurn(w, r, question, pb, sessionID, userID, hasUser, farmID, stream, answer, payload, fieldDegradeModelLabel, true)
	return true
}
