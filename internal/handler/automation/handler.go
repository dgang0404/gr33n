package automation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	automationworker "gr33n-api/internal/automation"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// validPreconditionOps enumerates the comparison operators accepted by the
// schedule-precondition evaluator. Keeping the list here (mirrored in the
// worker) lets us reject bad rules at the write path rather than silently
// at tick time.
var validPreconditionOps = map[string]struct{}{
	"lt":  {},
	"lte": {},
	"eq":  {},
	"gte": {},
	"gt":  {},
	"ne":  {},
}

type schedulePrecondition struct {
	SensorID int64   `json:"sensor_id"`
	Op       string  `json:"op"`
	Value    float64 `json:"value"`
}

// parsePreconditions validates the raw JSON payload sent by the client and
// returns a canonicalised []byte suitable for the DB. An empty/absent list
// normalises to "[]".
func parsePreconditions(ctx context.Context, q *db.Queries, farmID int64, raw json.RawMessage) ([]byte, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []byte("[]"), nil
	}
	var items []schedulePrecondition
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("preconditions must be an array of {sensor_id, op, value}")
	}
	for i, p := range items {
		if p.SensorID <= 0 {
			return nil, fmt.Errorf("preconditions[%d]: sensor_id must be > 0", i)
		}
		if _, ok := validPreconditionOps[p.Op]; !ok {
			return nil, fmt.Errorf("preconditions[%d]: op must be one of lt|lte|eq|gte|gt|ne", i)
		}
		sensor, err := q.GetSensorByID(ctx, p.SensorID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("preconditions[%d]: sensor %d not found", i, p.SensorID)
			}
			return nil, fmt.Errorf("preconditions[%d]: %w", i, err)
		}
		if sensor.FarmID != farmID {
			return nil, fmt.Errorf("preconditions[%d]: sensor %d does not belong to this farm", i, p.SensorID)
		}
	}
	canonical, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	return canonical, nil
}

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

// POST /farms/{id}/schedules
func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body struct {
		Name           string          `json:"name"`
		Description    *string         `json:"description"`
		ScheduleType   string          `json:"schedule_type"`
		CronExpression string          `json:"cron_expression"`
		Timezone       string          `json:"timezone"`
		IsActive       bool            `json:"is_active"`
		MetaData       []byte          `json:"meta_data"`
		Preconditions  json.RawMessage `json:"preconditions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" || body.CronExpression == "" || body.ScheduleType == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name, schedule_type, and cron_expression are required")
		return
	}
	if body.Timezone == "" {
		body.Timezone = "UTC"
	}
	if body.MetaData == nil {
		body.MetaData = []byte("{}")
	}
	preconds, err := parsePreconditions(r.Context(), h.q, farmID, body.Preconditions)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	row, err := h.q.CreateSchedule(r.Context(), db.CreateScheduleParams{
		FarmID:         farmID,
		Name:           body.Name,
		Description:    body.Description,
		ScheduleType:   body.ScheduleType,
		CronExpression: body.CronExpression,
		Timezone:       body.Timezone,
		IsActive:       body.IsActive,
		MetaData:       body.MetaData,
		Preconditions:  preconds,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create schedule")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// PUT /schedules/{id}
func (h *Handler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid schedule id")
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
	var body struct {
		Name           string          `json:"name"`
		Description    *string         `json:"description"`
		ScheduleType   string          `json:"schedule_type"`
		CronExpression string          `json:"cron_expression"`
		Timezone       string          `json:"timezone"`
		IsActive       bool            `json:"is_active"`
		MetaData       []byte          `json:"meta_data"`
		Preconditions  json.RawMessage `json:"preconditions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" || body.CronExpression == "" || body.ScheduleType == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name, schedule_type, and cron_expression are required")
		return
	}
	if body.Timezone == "" {
		body.Timezone = "UTC"
	}
	if body.MetaData == nil {
		body.MetaData = []byte("{}")
	}
	// An absent preconditions field means "don't touch" — preserve existing
	// interlock rules so a partial PUT doesn't accidentally clear them.
	var preconds []byte
	if len(body.Preconditions) == 0 {
		preconds = sch.Preconditions
		if len(preconds) == 0 {
			preconds = []byte("[]")
		}
	} else {
		preconds, err = parsePreconditions(r.Context(), h.q, sch.FarmID, body.Preconditions)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	row, err := h.q.UpdateSchedule(r.Context(), db.UpdateScheduleParams{
		ID:             id,
		Name:           body.Name,
		Description:    body.Description,
		ScheduleType:   body.ScheduleType,
		CronExpression: body.CronExpression,
		Timezone:       body.Timezone,
		IsActive:       body.IsActive,
		MetaData:       body.MetaData,
		Preconditions:  preconds,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update schedule")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// DELETE /schedules/{id}
func (h *Handler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid schedule id")
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
	if err := h.q.DeleteSchedule(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete schedule")
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
