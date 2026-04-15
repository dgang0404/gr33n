package actuator

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
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

// GET /farms/{id}/actuators
func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	rows, err := h.q.ListActuatorsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list actuators")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreActuator{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// PATCH /actuators/{id}/state
func (h *Handler) UpdateState(w http.ResponseWriter, r *http.Request) {
	actuatorID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid actuator id")
		return
	}

	var body struct {
		StateText    string   `json:"state_text"`
		StateNumeric *float64 `json:"state_numeric"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var stateNumeric pgtype.Numeric
	if body.StateNumeric != nil {
		if err := stateNumeric.Scan(*body.StateNumeric); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid state_numeric")
			return
		}
	}
	var stateText *string
	if body.StateText != "" {
		stateText = &body.StateText
	}

	row, err := h.q.UpdateActuatorState(r.Context(), db.UpdateActuatorStateParams{
		ID:                  actuatorID,
		CurrentStateNumeric: stateNumeric,
		CurrentStateText:    stateText,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update actuator state")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}
