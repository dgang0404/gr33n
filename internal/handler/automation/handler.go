package automation

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	automationworker "gr33n-api/internal/automation"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	q      *db.Queries
	worker *automationworker.Worker
}

func NewHandler(pool *pgxpool.Pool, worker *automationworker.Worker) *Handler {
	return &Handler{
		q:      db.New(pool),
		worker: worker,
	}
}

// GET /farms/{id}/schedules
func (h *Handler) ListSchedulesByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListSchedulesByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list schedules")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreSchedule{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// PATCH /schedules/{id}/active
func (h *Handler) UpdateScheduleActive(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid schedule id")
		return
	}
	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sch, err := h.q.GetScheduleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "schedule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load schedule")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, sch.FarmID) {
		return
	}
	row, err := h.q.UpdateScheduleActive(r.Context(), db.UpdateScheduleActiveParams{
		ID:       id,
		IsActive: body.IsActive,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update schedule")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// GET /farms/{id}/automation/runs
func (h *Handler) ListRunsByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListAutomationRunsByFarm(r.Context(), db.ListAutomationRunsByFarmParams{
		FarmID: farmID,
		Limit:  100,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list automation runs")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreAutomationRun{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// GET /automation/worker/health
func (h *Handler) WorkerHealth(w http.ResponseWriter, r *http.Request) {
	if h.worker == nil {
		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"running":         false,
			"simulation_mode": false,
			"status":          "disabled",
		})
		return
	}
	s := h.worker.GetStatus()
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"running":         s.Running,
		"simulation_mode": s.SimulationMode,
		"last_tick_at":    s.LastTickAt,
		"last_error":      s.LastError,
		"status":          "ok",
	})
}
