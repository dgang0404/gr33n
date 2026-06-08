// Package guardian — lightweight Farm Guardian HTTP endpoints (Phase 61 nudges).
package guardian

import (
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	q *db.Queries
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// Nudge — GET /farms/{id}/guardian-nudge
func (h *Handler) Nudge(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	nudge, err := farmguardian.ComputeGuardianNudge(r.Context(), h.q, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if nudge == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, nudge)
}
