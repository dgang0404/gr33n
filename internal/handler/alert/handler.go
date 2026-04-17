package alert

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	q := db.New(h.pool)
	if !farmauthz.RequireFarmMember(w, r, q, farmID) {
		return
	}
	rows, err := q.ListAlertsByFarm(r.Context(), db.ListAlertsByFarmParams{
		FarmID: farmID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreAlertsNotification{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) CountUnread(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	q := db.New(h.pool)
	if !farmauthz.RequireFarmMember(w, r, q, farmID) {
		return
	}
	count, err := q.CountUnreadAlertsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]int64{"unread_count": count})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid alert id")
		return
	}
	q := db.New(h.pool)
	a0, err := q.GetAlertNotificationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "alert not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, q, a0.FarmID) {
		return
	}
	alert, err := q.MarkAlertRead(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, alert)
}

func (h *Handler) MarkAcknowledged(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid alert id")
		return
	}
	q := db.New(h.pool)
	a0, err := q.GetAlertNotificationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "alert not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, q, a0.FarmID) {
		return
	}
	alert, err := q.MarkAlertAcknowledged(r.Context(), db.MarkAlertAcknowledgedParams{
		ID:                   id,
		AcknowledgedByUserID: pgtype.UUID{},
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, alert)
}

// CreateTaskFromAlert — POST /alerts/{id}/create-task
//
// Synthesises a task from the alert (title/description/priority/zone) so
// the operator can turn an alert into tracked work in one click. The body
// is optional; any provided field overrides the derived default.
func (h *Handler) CreateTaskFromAlert(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid alert id")
		return
	}
	q := db.New(h.pool)
	alertRow, err := q.GetAlertNotificationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "alert not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, q, alertRow.FarmID) {
		return
	}

	var body struct {
		Title            *string    `json:"title"`
		Description      *string    `json:"description"`
		ZoneID           *int64     `json:"zone_id"`
		TaskType         *string    `json:"task_type"`
		Priority         *int32     `json:"priority"`
		DueDate          *string    `json:"due_date"`
		AssignedToUserID *uuid.UUID `json:"assigned_to_user_id"`
	}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid body")
			return
		}
	}

	// --- Title (override → subject_rendered → "Alert #N") ---
	title := ""
	if body.Title != nil {
		title = strings.TrimSpace(*body.Title)
	}
	if title == "" && alertRow.SubjectRendered != nil {
		title = strings.TrimSpace(*alertRow.SubjectRendered)
	}
	if title == "" {
		title = fmt.Sprintf("Follow up on alert #%d", alertRow.ID)
	}

	// --- Description (override → message_text_rendered) ---
	var description *string
	if body.Description != nil {
		description = body.Description
	} else if alertRow.MessageTextRendered != nil && strings.TrimSpace(*alertRow.MessageTextRendered) != "" {
		msg := *alertRow.MessageTextRendered
		description = &msg
	}

	// --- Priority (override → derived from severity → 1) ---
	priority := int32(1)
	if body.Priority != nil {
		priority = *body.Priority
		if priority < 0 || priority > 3 {
			httputil.WriteError(w, http.StatusBadRequest, "priority must be 0\u20133")
			return
		}
	} else if alertRow.Severity.Valid {
		switch alertRow.Severity.Gr33ncoreNotificationPriorityEnum {
		case "critical":
			priority = 3
		case "high":
			priority = 2
		case "medium":
			priority = 1
		case "low":
			priority = 0
		}
	}

	// --- Zone (override → derive from triggering sensor if any) ---
	zoneID := body.ZoneID
	if zoneID == nil &&
		alertRow.TriggeringEventSourceType != nil &&
		*alertRow.TriggeringEventSourceType == "sensor_reading" &&
		alertRow.TriggeringEventSourceID != nil {
		if sensor, err := q.GetSensorByID(r.Context(), *alertRow.TriggeringEventSourceID); err == nil {
			if sensor.FarmID == alertRow.FarmID && sensor.ZoneID != nil {
				zid := *sensor.ZoneID
				zoneID = &zid
			}
		}
	}

	// --- Due date (optional override) ---
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

	taskType := body.TaskType
	if taskType == nil {
		tt := "alert_follow_up"
		taskType = &tt
	}

	var createdBy pgtype.UUID
	if uid, ok := authctx.UserID(r.Context()); ok {
		createdBy = pgtype.UUID{Bytes: uid, Valid: true}
	}

	alertID := alertRow.ID
	task, err := q.CreateTask(r.Context(), db.CreateTaskParams{
		FarmID:                   alertRow.FarmID,
		ZoneID:                   zoneID,
		ScheduleID:               nil,
		Title:                    title,
		Description:              description,
		TaskType:                 taskType,
		Status:                   commontypes.TaskStatusEnum("todo"),
		Priority:                 &priority,
		AssignedToUserID:         assignID,
		DueDate:                  dueDate,
		EstimatedDurationMinutes: nil,
		SourceAlertID:            &alertID,
		CreatedByUserID:          createdBy,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, task)
}
