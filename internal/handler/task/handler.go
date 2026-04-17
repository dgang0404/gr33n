package task

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/commontypes"
)

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

// ListByFarm — GET /farms/{id}/tasks
func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	q := db.New(h.pool)
	if !farmauthz.RequireFarmMember(w, r, q, farmID) {
		return
	}
	rows, err := q.ListTasksByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreTask{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// Create — POST /farms/{id}/tasks
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	q := db.New(h.pool)
	if !farmauthz.RequireFarmOperate(w, r, q, farmID) {
		return
	}
	var body struct {
		Title            string     `json:"title"`
		Description      *string    `json:"description"`
		ZoneID           *int64     `json:"zone_id"`
		ScheduleID       *int64     `json:"schedule_id"`
		TaskType         *string    `json:"task_type"`
		Priority         *int32     `json:"priority"`
		DueDate          *string    `json:"due_date"`
		AssignedToUserID *uuid.UUID `json:"assigned_to_user_id"`
		SourceAlertID    *int64     `json:"source_alert_id"`
		SourceRuleID     *int64     `json:"source_rule_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		httputil.WriteError(w, http.StatusBadRequest, "title required")
		return
	}
	priority := int32(1)
	if body.Priority != nil {
		priority = *body.Priority
		if priority < 0 || priority > 3 {
			httputil.WriteError(w, http.StatusBadRequest, "priority must be 0–3")
			return
		}
	}
	var dueDate pgtype.Date
	if body.DueDate != nil && strings.TrimSpace(*body.DueDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*body.DueDate))
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid due_date (use YYYY-MM-DD)")
			return
		}
		dueDate = pgtype.Date{Time: t, Valid: true}
	}
	var assignID pgtype.UUID
	if body.AssignedToUserID != nil {
		assignID = pgtype.UUID{Bytes: *body.AssignedToUserID, Valid: true}
	}
	if body.ScheduleID != nil {
		sch, err := q.GetScheduleByID(r.Context(), *body.ScheduleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, "schedule not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if sch.FarmID != farmID {
			httputil.WriteError(w, http.StatusBadRequest, "schedule does not belong to this farm")
			return
		}
	}
	if body.SourceAlertID != nil {
		a, err := q.GetAlertNotificationByID(r.Context(), *body.SourceAlertID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, "source alert not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if a.FarmID != farmID {
			httputil.WriteError(w, http.StatusBadRequest, "source alert does not belong to this farm")
			return
		}
	}
	if body.SourceRuleID != nil {
		rule, err := q.GetAutomationRuleByID(r.Context(), *body.SourceRuleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, "source rule not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if rule.FarmID != farmID {
			httputil.WriteError(w, http.StatusBadRequest, "source rule does not belong to this farm")
			return
		}
	}
	var createdBy pgtype.UUID
	if uid, ok := authctx.UserID(r.Context()); ok {
		createdBy = pgtype.UUID{Bytes: uid, Valid: true}
	}
	task, err := q.CreateTask(r.Context(), db.CreateTaskParams{
		FarmID:                   farmID,
		ZoneID:                   body.ZoneID,
		ScheduleID:               body.ScheduleID,
		Title:                    title,
		Description:              body.Description,
		TaskType:                 body.TaskType,
		Status:                   commontypes.TaskStatusEnum("todo"),
		Priority:                 &priority,
		AssignedToUserID:         assignID,
		DueDate:                  dueDate,
		EstimatedDurationMinutes: nil,
		SourceAlertID:            body.SourceAlertID,
		SourceRuleID:             body.SourceRuleID,
		CreatedByUserID:          createdBy,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, task)
}

// UpdateStatus — PATCH /tasks/{id}/status
// Body: { "status": "in_progress" }
// Valid: todo | in_progress | on_hold | completed | cancelled | blocked_requires_input | pending_review
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	q := db.New(h.pool)
	t0, err := q.GetTaskByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, q, t0.FarmID) {
		return
	}
	task, err := q.UpdateTaskStatus(r.Context(), db.UpdateTaskStatusParams{
		ID:              id,
		Status:          commontypes.TaskStatusEnum(body.Status),
		UpdatedByUserID: pgtype.UUID{}, // zero value = NULL in DB
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, task)
}

// Update — PUT /tasks/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	q := db.New(h.pool)
	t0, err := q.GetTaskByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, q, t0.FarmID) {
		return
	}
	var body struct {
		Title                    string     `json:"title"`
		Description              *string    `json:"description"`
		ZoneID                   *int64     `json:"zone_id"`
		ScheduleID               *int64     `json:"schedule_id"`
		TaskType                 *string    `json:"task_type"`
		Priority                 *int32     `json:"priority"`
		DueDate                  *string    `json:"due_date"`
		AssignedToUserID         *uuid.UUID `json:"assigned_to_user_id"`
		EstimatedDurationMinutes *int32     `json:"estimated_duration_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		httputil.WriteError(w, http.StatusBadRequest, "title required")
		return
	}
	priority := int32(1)
	if body.Priority != nil {
		priority = *body.Priority
		if priority < 0 || priority > 3 {
			httputil.WriteError(w, http.StatusBadRequest, "priority must be 0\u20133")
			return
		}
	}
	var dueDate pgtype.Date
	if body.DueDate != nil && strings.TrimSpace(*body.DueDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*body.DueDate))
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid due_date (use YYYY-MM-DD)")
			return
		}
		dueDate = pgtype.Date{Time: t, Valid: true}
	}
	var assignID pgtype.UUID
	if body.AssignedToUserID != nil {
		assignID = pgtype.UUID{Bytes: *body.AssignedToUserID, Valid: true}
	}
	var updatedBy pgtype.UUID
	if uid, ok := authctx.UserID(r.Context()); ok {
		updatedBy = pgtype.UUID{Bytes: uid, Valid: true}
	}
	task, err := q.UpdateTask(r.Context(), db.UpdateTaskParams{
		ID:                       id,
		Title:                    title,
		Description:              body.Description,
		ZoneID:                   body.ZoneID,
		ScheduleID:               body.ScheduleID,
		TaskType:                 body.TaskType,
		Priority:                 &priority,
		DueDate:                  dueDate,
		AssignedToUserID:         assignID,
		EstimatedDurationMinutes: body.EstimatedDurationMinutes,
		UpdatedByUserID:          updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, task)
}

// ListLabor — GET /tasks/{id}/labor (Phase 20.95 WS1)
func (h *Handler) ListLabor(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	q := db.New(h.pool)
	t0, err := q.GetTaskByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, q, t0.FarmID) {
		return
	}
	rows, err := q.ListTaskLaborLogsByTask(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreTaskLaborLog{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// CreateLabor — POST /tasks/{id}/labor (Phase 20.95 WS1)
func (h *Handler) CreateLabor(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	q := db.New(h.pool)
	t0, err := q.GetTaskByID(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, q, t0.FarmID) {
		return
	}
	var body struct {
		StartedAt          string   `json:"started_at"`
		EndedAt            *string  `json:"ended_at"`
		Minutes            int32    `json:"minutes"`
		HourlyRateSnapshot *float64 `json:"hourly_rate_snapshot"`
		Currency           *string  `json:"currency"`
		Notes              *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Minutes < 0 {
		httputil.WriteError(w, http.StatusBadRequest, "minutes must be >= 0")
		return
	}
	startedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(body.StartedAt))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid started_at (RFC3339 required)")
		return
	}
	var endedAt pgtype.Timestamptz
	if body.EndedAt != nil && strings.TrimSpace(*body.EndedAt) != "" {
		et, err := time.Parse(time.RFC3339, strings.TrimSpace(*body.EndedAt))
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid ended_at (RFC3339 required)")
			return
		}
		endedAt = pgtype.Timestamptz{Time: et, Valid: true}
	}
	var hourly pgtype.Numeric
	if body.HourlyRateSnapshot != nil {
		if err := hourly.Scan(strconv.FormatFloat(*body.HourlyRateSnapshot, 'f', -1, 64)); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid hourly_rate_snapshot")
			return
		}
	}
	if body.Currency != nil {
		cur := strings.ToUpper(strings.TrimSpace(*body.Currency))
		if cur == "" {
			body.Currency = nil
		} else {
			if len(cur) != 3 {
				httputil.WriteError(w, http.StatusBadRequest, "currency must be ISO 4217 (3 uppercase letters)")
				return
			}
			body.Currency = &cur
		}
	}
	var userID pgtype.UUID
	if uid, ok := authctx.UserID(r.Context()); ok {
		userID = pgtype.UUID{Bytes: uid, Valid: true}
	}
	row, err := q.CreateTaskLaborLog(r.Context(), db.CreateTaskLaborLogParams{
		FarmID:             t0.FarmID,
		TaskID:             taskID,
		UserID:             userID,
		StartedAt:          startedAt,
		EndedAt:            endedAt,
		Minutes:            body.Minutes,
		HourlyRateSnapshot: hourly,
		Currency:           body.Currency,
		Notes:              body.Notes,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := q.RecalcTaskTimeSpentMinutes(r.Context(), taskID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// DeleteLabor — DELETE /labor/{id} (Phase 20.95 WS1)
func (h *Handler) DeleteLabor(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid labor log id")
		return
	}
	q := db.New(h.pool)
	row, err := q.GetTaskLaborLogByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "labor log not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, q, row.FarmID) {
		return
	}
	if err := q.DeleteTaskLaborLog(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := q.RecalcTaskTimeSpentMinutes(r.Context(), row.TaskID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete — DELETE /tasks/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	q := db.New(h.pool)
	t0, err := q.GetTaskByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, q, t0.FarmID) {
		return
	}
	var updatedBy pgtype.UUID
	if uid, ok := authctx.UserID(r.Context()); ok {
		updatedBy = pgtype.UUID{Bytes: uid, Valid: true}
	}
	if err := q.SoftDeleteTask(r.Context(), db.SoftDeleteTaskParams{
		ID:              id,
		UpdatedByUserID: updatedBy,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
