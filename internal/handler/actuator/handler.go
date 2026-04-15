package actuator

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

// GET /actuators/{id}/events?since=RFC3339&limit=N
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	actuatorID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid actuator id")
		return
	}

	since := time.Now().UTC().Add(-24 * time.Hour)
	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			since = t
		}
	}

	limit := int32(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.ParseInt(l, 10, 32); err == nil && n > 0 && n <= 200 {
			limit = int32(n)
		}
	}

	rows, err := h.q.ListActuatorEventsByActuator(r.Context(), db.ListActuatorEventsByActuatorParams{
		ActuatorID: actuatorID,
		EventTime:  since,
		Limit:      limit,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list actuator events")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreActuatorEvent{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}
