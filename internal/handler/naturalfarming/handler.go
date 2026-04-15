package naturalfarming

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	q *db.Queries
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// GET /farms/{id}/naturalfarming/inputs
func (h *Handler) ListInputs(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	rows, err := h.q.ListInputDefinitionsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list input definitions")
		return
	}
	if rows == nil {
		rows = []db.Gr33nnaturalfarmingInputDefinition{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// GET /farms/{id}/naturalfarming/batches
func (h *Handler) ListBatches(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	rows, err := h.q.ListInputBatchesByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list input batches")
		return
	}
	if rows == nil {
		rows = []db.Gr33nnaturalfarmingInputBatch{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}
