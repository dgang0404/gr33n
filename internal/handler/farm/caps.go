package farm

import (
	"net/http"
	"strconv"

	"gr33n-api/internal/authctx"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// MeCaps — GET /farms/{id}/me/caps
func (h *Handler) MeCaps(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	out, err := farmauthz.MeCapsForUser(r.Context(), h.q, uid, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to resolve farm caps")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}
