package actuator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
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
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
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
	a0, err := h.q.GetActuatorByID(r.Context(), actuatorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "actuator not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load actuator")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, a0.FarmID) {
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

// GET /schedules/{id}/actuator-events?since=RFC3339&limit=N
func (h *Handler) ListEventsBySchedule(w http.ResponseWriter, r *http.Request) {
	scheduleID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid schedule id")
		return
	}

	since := time.Now().UTC().Add(-7 * 24 * time.Hour)
	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			since = t
		}
	}

	limit := int32(100)
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.ParseInt(l, 10, 32); err == nil && n > 0 && n <= 500 {
			limit = int32(n)
		}
	}

	sch, err := h.q.GetScheduleByID(r.Context(), scheduleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "schedule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load schedule")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, sch.FarmID) {
		return
	}

	rows, err := h.q.ListActuatorEventsBySchedule(r.Context(), db.ListActuatorEventsByScheduleParams{
		TriggeredByScheduleID: &scheduleID,
		EventTime:             since,
		Limit:                 limit,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list actuator events by schedule")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreActuatorEvent{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// POST /actuators/{id}/events — Pi reports an executed command
func (h *Handler) RecordEvent(w http.ResponseWriter, r *http.Request) {
	actuatorID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid actuator id")
		return
	}

	var body struct {
		CommandSent             string          `json:"command_sent"`
		Source                  string          `json:"source"`
		EventTime               string          `json:"event_time"`
		ExecutionStatus         string          `json:"execution_status"`
		TriggeredByScheduleID   *int64          `json:"triggered_by_schedule_id"`
		TriggeredByRuleID       *int64          `json:"triggered_by_rule_id"`
		ProgramID               *int64          `json:"program_id"`
		ParametersSent          json.RawMessage `json:"parameters_sent"`
		MetaData                json.RawMessage `json:"meta_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.CommandSent == "" {
		httputil.WriteError(w, http.StatusBadRequest, "command_sent is required")
		return
	}
	if body.TriggeredByRuleID != nil && body.ProgramID != nil {
		httputil.WriteError(w, http.StatusBadRequest, "cannot set both triggered_by_rule_id and program_id")
		return
	}

	ctx := r.Context()
	a0, err := h.q.GetActuatorByID(ctx, actuatorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "actuator not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load actuator")
		return
	}
	if !farmauthz.RequireFarmMemberOrPiEdge(w, r, h.q, a0.FarmID) {
		return
	}

	if msg, ok := validatePiActuatorEventProvenance(ctx, h.q, a0.FarmID, body.TriggeredByScheduleID, body.TriggeredByRuleID, body.ProgramID); !ok {
		httputil.WriteError(w, http.StatusBadRequest, msg)
		return
	}

	params := []byte(`{}`)
	if len(bytes.TrimSpace(body.ParametersSent)) > 0 {
		if !json.Valid(body.ParametersSent) {
			httputil.WriteError(w, http.StatusBadRequest, "parameters_sent must be valid JSON")
			return
		}
		params = body.ParametersSent
	}

	meta := map[string]any{"reported_by": "pi_client"}
	if len(bytes.TrimSpace(body.MetaData)) > 0 {
		var extra map[string]any
		if err := json.Unmarshal(body.MetaData, &extra); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "meta_data must be a JSON object")
			return
		}
		for k, v := range extra {
			meta[k] = v
		}
	}
	if body.ProgramID != nil {
		meta["program_id"] = *body.ProgramID
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to marshal meta_data")
		return
	}

	evtTime := time.Now().UTC()
	if body.EventTime != "" {
		if t, err := time.Parse(time.RFC3339, body.EventTime); err == nil {
			evtTime = t
		}
	}

	src := db.Gr33ncoreActuatorEventSourceEnum(body.Source)
	if src == "" {
		src = db.Gr33ncoreActuatorEventSourceEnumManualApiCall
	}
	status := db.Gr33ncoreActuatorExecutionStatusEnum(body.ExecutionStatus)
	if status == "" {
		status = db.Gr33ncoreActuatorExecutionStatusEnumCommandSentToDevice
	}

	row, err := h.q.InsertActuatorEvent(ctx, db.InsertActuatorEventParams{
		EventTime:             evtTime,
		ActuatorID:            actuatorID,
		CommandSent:           &body.CommandSent,
		ParametersSent:        params,
		TriggeredByUserID:     pgtype.UUID{},
		TriggeredByScheduleID: body.TriggeredByScheduleID,
		TriggeredByRuleID:     body.TriggeredByRuleID,
		Source:                src,
		ExecutionStatus: db.NullGr33ncoreActuatorExecutionStatusEnum{
			Gr33ncoreActuatorExecutionStatusEnum: status,
			Valid:                                true,
		},
		MetaData: metaBytes,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to record actuator event")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// validatePiActuatorEventProvenance ensures schedule / rule / program
// foreign keys referenced on a Pi-reported actuator event belong to the
// same farm as the actuator (defence-in-depth on top of DB FKs).
func validatePiActuatorEventProvenance(
	ctx context.Context,
	q *db.Queries,
	actuatorFarmID int64,
	scheduleID, ruleID, programID *int64,
) (string, bool) {
	if scheduleID != nil {
		sch, err := q.GetScheduleByID(ctx, *scheduleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "triggered_by_schedule_id not found", false
			}
			return fmt.Sprintf("schedule lookup: %v", err), false
		}
		if sch.FarmID != actuatorFarmID {
			return "triggered_by_schedule_id does not belong to actuator farm", false
		}
	}
	if ruleID != nil {
		rule, err := q.GetAutomationRuleByID(ctx, *ruleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "triggered_by_rule_id not found", false
			}
			return fmt.Sprintf("rule lookup: %v", err), false
		}
		if rule.FarmID != actuatorFarmID {
			return "triggered_by_rule_id does not belong to actuator farm", false
		}
	}
	if programID != nil {
		prog, err := q.GetFertigationProgramByID(ctx, *programID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "program_id not found", false
			}
			return fmt.Sprintf("program lookup: %v", err), false
		}
		if prog.FarmID != actuatorFarmID {
			return "program_id does not belong to actuator farm", false
		}
	}
	return "", true
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

	a0, err := h.q.GetActuatorByID(r.Context(), actuatorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "actuator not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load actuator")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, a0.FarmID) {
		return
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
