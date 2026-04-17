package automation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/commontypes"
)

// ── Response DTOs ────────────────────────────────────────────────────────────
//
// The sqlc-generated Gr33ncoreAutomationRule / Gr33ncoreExecutableAction
// structs store jsonb columns as []byte, which encoding/json would emit as a
// base64 string. The UI needs real JSON objects to round-trip conditions and
// action parameters into the rule-builder form, so we wrap rows into these
// small view types with json.RawMessage fields before writing the response.

type ruleView struct {
	ID                    int64                                   `json:"id"`
	FarmID                int64                                   `json:"farm_id"`
	Name                  string                                  `json:"name"`
	Description           *string                                 `json:"description"`
	IsActive              bool                                    `json:"is_active"`
	TriggerSource         commontypes.AutomationTriggerSourceEnum `json:"trigger_source"`
	TriggerConfiguration  json.RawMessage                         `json:"trigger_configuration"`
	ConditionLogic        *string                                 `json:"condition_logic"`
	ConditionsJsonb       json.RawMessage                         `json:"conditions_jsonb"`
	LastEvaluatedTime     pgtype.Timestamptz                      `json:"last_evaluated_time"`
	LastTriggeredTime     pgtype.Timestamptz                      `json:"last_triggered_time"`
	CooldownPeriodSeconds *int32                                  `json:"cooldown_period_seconds"`
	CreatedAt             time.Time                               `json:"created_at"`
	UpdatedAt             time.Time                               `json:"updated_at"`
}

type actionView struct {
	ID                           int64                                `json:"id"`
	RuleID                       *int64                               `json:"rule_id"`
	ScheduleID                   *int64                               `json:"schedule_id"`
	ProgramID                    *int64                               `json:"program_id"`
	ExecutionOrder               int32                                `json:"execution_order"`
	ActionType                   commontypes.ExecutableActionTypeEnum `json:"action_type"`
	TargetActuatorID             *int64                               `json:"target_actuator_id"`
	TargetAutomationRuleID       *int64                               `json:"target_automation_rule_id"`
	TargetNotificationTemplateID *int64                               `json:"target_notification_template_id"`
	ActionCommand                *string                              `json:"action_command"`
	ActionParameters             json.RawMessage                      `json:"action_parameters"`
	DelayBeforeExecutionSeconds  *int32                               `json:"delay_before_execution_seconds"`
}

// rawJSONOrNull returns the raw jsonb bytes as json.RawMessage, or the JSON
// literal null when the column is empty. Without this, empty []byte values
// would marshal as an empty string and break JSON.parse on the client.
func rawJSONOrNull(b []byte) json.RawMessage {
	if len(b) == 0 {
		return json.RawMessage("null")
	}
	return json.RawMessage(b)
}

func toRuleView(r db.Gr33ncoreAutomationRule) ruleView {
	return ruleView{
		ID:                    r.ID,
		FarmID:                r.FarmID,
		Name:                  r.Name,
		Description:           r.Description,
		IsActive:              r.IsActive,
		TriggerSource:         r.TriggerSource,
		TriggerConfiguration:  rawJSONOrNull(r.TriggerConfiguration),
		ConditionLogic:        r.ConditionLogic,
		ConditionsJsonb:       rawJSONOrNull(r.ConditionsJsonb),
		LastEvaluatedTime:     r.LastEvaluatedTime,
		LastTriggeredTime:     r.LastTriggeredTime,
		CooldownPeriodSeconds: r.CooldownPeriodSeconds,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}

func toRuleViews(rows []db.Gr33ncoreAutomationRule) []ruleView {
	out := make([]ruleView, 0, len(rows))
	for _, r := range rows {
		out = append(out, toRuleView(r))
	}
	return out
}

func toActionView(a db.Gr33ncoreExecutableAction) actionView {
	return actionView{
		ID:                           a.ID,
		RuleID:                       a.RuleID,
		ScheduleID:                   a.ScheduleID,
		ProgramID:                    a.ProgramID,
		ExecutionOrder:               a.ExecutionOrder,
		ActionType:                   a.ActionType,
		TargetActuatorID:             a.TargetActuatorID,
		TargetAutomationRuleID:       a.TargetAutomationRuleID,
		TargetNotificationTemplateID: a.TargetNotificationTemplateID,
		ActionCommand:                a.ActionCommand,
		ActionParameters:             rawJSONOrNull(a.ActionParameters),
		DelayBeforeExecutionSeconds:  a.DelayBeforeExecutionSeconds,
	}
}

func toActionViews(rows []db.Gr33ncoreExecutableAction) []actionView {
	out := make([]actionView, 0, len(rows))
	for _, a := range rows {
		out = append(out, toActionView(a))
	}
	return out
}

// validTriggerSources and validActionTypes mirror the gr33ncore enums.
// Keeping them here lets us reject invalid rule payloads at the write
// path with a helpful message instead of an opaque Postgres error.
var validTriggerSources = map[string]struct{}{
	"sensor_reading_threshold":  {},
	"specific_time_cron":        {},
	"actuator_state_changed":    {},
	"manual_api_trigger":        {},
	"task_status_updated":       {},
	"new_system_log_event":      {},
	"external_webhook_received": {},
}

// Phase 20 ships dispatchers for these three action types. The others
// remain valid in the DB enum but are explicitly rejected at write-time
// so operators can't create unrunnable rules.
var supportedActionTypes = map[string]struct{}{
	"control_actuator":  {},
	"create_task":       {},
	"send_notification": {},
}

var deferredActionTypes = map[string]struct{}{
	"trigger_another_automation_rule": {},
	"http_webhook_call":               {},
	"update_record_in_gr33n":          {},
	"log_custom_event":                {},
}

// rulePredicate is the canonical predicate shape shared with Phase 19
// schedule preconditions. A rule's conditions_jsonb stores
// { "logic": "ALL"|"ANY", "predicates": [<rulePredicate>,...] }.
//
// Phase 20.6 WS3 — Type discriminates two variants. Empty / "hard" is
// the legacy {sensor_id, op, value} predicate. "setpoint" is the new
// stage-scoped variant that resolves `gr33ncore.zone_setpoints` at
// eval time against the rule's zone + active crop cycle.
type rulePredicate struct {
	Type string `json:"type,omitempty"`

	SensorID int64   `json:"sensor_id,omitempty"`
	Op       string  `json:"op"`
	Value    float64 `json:"value,omitempty"`

	SensorType string `json:"sensor_type,omitempty"`
	Scope      string `json:"scope,omitempty"`
}

type ruleConditions struct {
	Logic      string          `json:"logic"`
	Predicates []rulePredicate `json:"predicates"`
}

// parseRuleConditions validates the client-supplied conditions object
// and returns a canonicalised JSON blob ready for conditions_jsonb.
// An empty/absent payload normalises to {"logic":"ALL","predicates":[]}.
func parseRuleConditions(ctx context.Context, q *db.Queries, farmID int64, logic string, rawPreds json.RawMessage) (string, []byte, error) {
	if logic == "" {
		logic = "ALL"
	}
	if logic != "ALL" && logic != "ANY" {
		return "", nil, fmt.Errorf("condition_logic must be 'ALL' or 'ANY'")
	}
	preds := []rulePredicate{}
	if len(rawPreds) > 0 && string(rawPreds) != "null" {
		if err := json.Unmarshal(rawPreds, &preds); err != nil {
			return "", nil, fmt.Errorf("conditions must be an array of {sensor_id, op, value}")
		}
	}
	for i, p := range preds {
		ptype := p.Type
		if ptype == "" {
			ptype = "hard"
		}
		switch ptype {
		case "hard":
			if p.SensorID <= 0 {
				return "", nil, fmt.Errorf("conditions[%d]: sensor_id must be > 0", i)
			}
			if _, ok := validPreconditionOps[p.Op]; !ok {
				return "", nil, fmt.Errorf("conditions[%d]: op must be one of lt|lte|eq|gte|gt|ne", i)
			}
			sensor, err := q.GetSensorByID(ctx, p.SensorID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return "", nil, fmt.Errorf("conditions[%d]: sensor %d not found", i, p.SensorID)
				}
				return "", nil, fmt.Errorf("conditions[%d]: %w", i, err)
			}
			if sensor.FarmID != farmID {
				return "", nil, fmt.Errorf("conditions[%d]: sensor %d does not belong to this farm", i, p.SensorID)
			}
		case "setpoint":
			// Phase 20.6 WS3 — setpoint predicates key off sensor_type
			// + scope and resolve a zone_setpoints row at eval time.
			// We validate shape here; zone/cycle resolution is a
			// per-tick concern so rules configured before the operator
			// has landed any setpoint rows still save cleanly.
			if strings.TrimSpace(p.SensorType) == "" {
				return "", nil, fmt.Errorf("conditions[%d]: sensor_type required for setpoint predicate", i)
			}
			switch p.Scope {
			case "", "current_stage", "zone_default":
				// OK; empty normalises to current_stage in the evaluator.
			default:
				return "", nil, fmt.Errorf("conditions[%d]: scope must be 'current_stage' or 'zone_default'", i)
			}
			switch p.Op {
			case "out_of_range", "below_ideal", "above_ideal", "inside_range":
				// OK
			default:
				return "", nil, fmt.Errorf("conditions[%d]: op must be one of out_of_range|below_ideal|above_ideal|inside_range for setpoint predicate", i)
			}
			if p.SensorID != 0 || p.Value != 0 {
				return "", nil, fmt.Errorf("conditions[%d]: sensor_id and value must be omitted when type='setpoint'", i)
			}
		default:
			return "", nil, fmt.Errorf("conditions[%d]: type must be 'hard' or 'setpoint'", i)
		}
	}
	canon, err := json.Marshal(ruleConditions{Logic: logic, Predicates: preds})
	if err != nil {
		return "", nil, err
	}
	return logic, canon, nil
}

// parseTriggerConfiguration enforces the enum + shape for trigger_source
// and returns the canonical []byte to persist. Empty payload allowed for
// manual / webhook triggers; sensor_reading_threshold must name a sensor
// on the same farm so the worker can resolve it cheaply.
func parseTriggerConfiguration(ctx context.Context, q *db.Queries, farmID int64, triggerSource string, raw json.RawMessage) ([]byte, error) {
	if _, ok := validTriggerSources[triggerSource]; !ok {
		return nil, fmt.Errorf("trigger_source must be one of sensor_reading_threshold|specific_time_cron|actuator_state_changed|manual_api_trigger|task_status_updated|new_system_log_event|external_webhook_received")
	}
	if len(raw) == 0 || string(raw) == "null" {
		raw = json.RawMessage(`{}`)
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("trigger_configuration must be a JSON object")
	}
	if triggerSource == "sensor_reading_threshold" {
		sid, ok := cfg["sensor_id"]
		if !ok {
			return nil, fmt.Errorf("trigger_configuration.sensor_id is required when trigger_source = sensor_reading_threshold")
		}
		sidInt, err := coerceInt64(sid)
		if err != nil {
			return nil, fmt.Errorf("trigger_configuration.sensor_id must be an integer")
		}
		sensor, err := q.GetSensorByID(ctx, sidInt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("trigger_configuration.sensor_id: sensor %d not found", sidInt)
			}
			return nil, err
		}
		if sensor.FarmID != farmID {
			return nil, fmt.Errorf("trigger_configuration.sensor_id: sensor %d does not belong to this farm", sidInt)
		}
	}
	canon, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return canon, nil
}

func coerceInt64(v any) (int64, error) {
	switch t := v.(type) {
	case float64:
		return int64(t), nil
	case int:
		return int64(t), nil
	case int64:
		return t, nil
	case json.Number:
		return t.Int64()
	}
	return 0, fmt.Errorf("not an integer")
}

// ── Rules ───────────────────────────────────────────────────────────────────

// GET /farms/{id}/automation/rules
func (h *Handler) ListAutomationRulesByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListAutomationRulesByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list automation rules")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreAutomationRule{}
	}
	httputil.WriteJSON(w, http.StatusOK, toRuleViews(rows))
}

type automationRuleBody struct {
	Name                  string          `json:"name"`
	Description           *string         `json:"description"`
	IsActive              bool            `json:"is_active"`
	TriggerSource         string          `json:"trigger_source"`
	TriggerConfiguration  json.RawMessage `json:"trigger_configuration"`
	ConditionLogic        string          `json:"condition_logic"`
	Conditions            json.RawMessage `json:"conditions"`
	CooldownPeriodSeconds *int32          `json:"cooldown_period_seconds"`
}

// POST /farms/{id}/automation/rules
func (h *Handler) CreateAutomationRule(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body automationRuleBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" || body.TriggerSource == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name and trigger_source are required")
		return
	}
	trigCfg, err := parseTriggerConfiguration(r.Context(), h.q, farmID, body.TriggerSource, body.TriggerConfiguration)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	logic, conds, err := parseRuleConditions(r.Context(), h.q, farmID, body.ConditionLogic, body.Conditions)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	logicPtr := logic
	row, err := h.q.CreateAutomationRule(r.Context(), db.CreateAutomationRuleParams{
		FarmID:                farmID,
		Name:                  body.Name,
		Description:           body.Description,
		IsActive:              body.IsActive,
		TriggerSource:         commontypes.AutomationTriggerSourceEnum(body.TriggerSource),
		TriggerConfiguration:  trigCfg,
		ConditionLogic:        &logicPtr,
		ConditionsJsonb:       conds,
		CooldownPeriodSeconds: body.CooldownPeriodSeconds,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create automation rule: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, toRuleView(row))
}

// GET /automation/rules/{id}
func (h *Handler) GetAutomationRule(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	rule, err := h.q.GetAutomationRuleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "automation rule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load automation rule")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, rule.FarmID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, toRuleView(rule))
}

// PUT /automation/rules/{id}
func (h *Handler) UpdateAutomationRule(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	rule, err := h.q.GetAutomationRuleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "automation rule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load automation rule")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, rule.FarmID) {
		return
	}
	var body automationRuleBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" || body.TriggerSource == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name and trigger_source are required")
		return
	}
	trigCfg, err := parseTriggerConfiguration(r.Context(), h.q, rule.FarmID, body.TriggerSource, body.TriggerConfiguration)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	logic, conds, err := parseRuleConditions(r.Context(), h.q, rule.FarmID, body.ConditionLogic, body.Conditions)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	logicPtr := logic
	row, err := h.q.UpdateAutomationRule(r.Context(), db.UpdateAutomationRuleParams{
		ID:                    id,
		Name:                  body.Name,
		Description:           body.Description,
		IsActive:              body.IsActive,
		TriggerSource:         commontypes.AutomationTriggerSourceEnum(body.TriggerSource),
		TriggerConfiguration:  trigCfg,
		ConditionLogic:        &logicPtr,
		ConditionsJsonb:       conds,
		CooldownPeriodSeconds: body.CooldownPeriodSeconds,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update automation rule: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, toRuleView(row))
}

// PATCH /automation/rules/{id}/active
func (h *Handler) UpdateAutomationRuleActive(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	rule, err := h.q.GetAutomationRuleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "automation rule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load automation rule")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, rule.FarmID) {
		return
	}
	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	row, err := h.q.UpdateAutomationRuleActive(r.Context(), db.UpdateAutomationRuleActiveParams{
		ID:       id,
		IsActive: body.IsActive,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update automation rule")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, toRuleView(row))
}

// DELETE /automation/rules/{id}
// Hard delete; ON DELETE CASCADE on executable_actions.rule_id cleans
// up any attached actions automatically.
func (h *Handler) DeleteAutomationRule(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	rule, err := h.q.GetAutomationRuleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "automation rule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load automation rule")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, rule.FarmID) {
		return
	}
	if err := h.q.DeleteAutomationRule(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete automation rule")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Executable actions (rule-bound) ────────────────────────────────────────

type executableActionBody struct {
	ExecutionOrder               int32           `json:"execution_order"`
	ActionType                   string          `json:"action_type"`
	TargetActuatorID             *int64          `json:"target_actuator_id"`
	TargetNotificationTemplateID *int64          `json:"target_notification_template_id"`
	ActionCommand                *string         `json:"action_command"`
	ActionParameters             json.RawMessage `json:"action_parameters"`
	DelayBeforeExecutionSeconds  *int32          `json:"delay_before_execution_seconds"`
	// Phase 20.95 WS3 — if the client tries to attach this action to a second
	// source (schedule or fertigation program) in addition to the rule this
	// POST is scoped to, reject with 400 before the DB check constraint can.
	ScheduleID *int64 `json:"schedule_id"`
	ProgramID  *int64 `json:"program_id"`
}

// validateActionTypeForCreate enforces the Phase 20 supported-action
// whitelist and returns a 400-friendly error for deferred types.
func validateActionType(actionType string) error {
	if _, ok := deferredActionTypes[actionType]; ok {
		return fmt.Errorf("action_type %q is defined in the schema but not yet supported", actionType)
	}
	if _, ok := supportedActionTypes[actionType]; !ok {
		return fmt.Errorf("action_type must be one of control_actuator|create_task|send_notification")
	}
	return nil
}

// validateActionShape mirrors the DB's chk_executable_action_details so
// operators get a readable error instead of an opaque 500 on insert.
func validateActionShape(body *executableActionBody, farmID int64, q *db.Queries, ctx context.Context) ([]byte, error) {
	params := body.ActionParameters
	if len(params) == 0 || string(params) == "null" {
		params = nil
	}
	switch body.ActionType {
	case "control_actuator":
		if body.TargetActuatorID == nil || *body.TargetActuatorID <= 0 {
			return nil, fmt.Errorf("target_actuator_id is required for control_actuator")
		}
		if body.ActionCommand == nil || *body.ActionCommand == "" {
			return nil, fmt.Errorf("action_command is required for control_actuator")
		}
		act, err := q.GetActuatorByID(ctx, *body.TargetActuatorID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("target_actuator_id: actuator %d not found", *body.TargetActuatorID)
			}
			return nil, err
		}
		if act.FarmID != farmID {
			return nil, fmt.Errorf("target_actuator_id: actuator %d does not belong to this farm", *body.TargetActuatorID)
		}
	case "create_task":
		if params == nil {
			return nil, fmt.Errorf("action_parameters is required for create_task")
		}
		var probe map[string]any
		if err := json.Unmarshal(params, &probe); err != nil {
			return nil, fmt.Errorf("action_parameters must be a JSON object")
		}
		if len(probe) == 0 {
			return nil, fmt.Errorf("action_parameters must be a non-empty object for create_task")
		}
	case "send_notification":
		if body.TargetNotificationTemplateID == nil || *body.TargetNotificationTemplateID <= 0 {
			return nil, fmt.Errorf("target_notification_template_id is required for send_notification")
		}
	}
	return params, nil
}

// GET /automation/rules/{id}/actions
func (h *Handler) ListActionsByRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	rule, err := h.q.GetAutomationRuleByID(r.Context(), ruleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "automation rule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load automation rule")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, rule.FarmID) {
		return
	}
	rows, err := h.q.ListExecutableActionsByRule(r.Context(), &ruleID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list actions")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreExecutableAction{}
	}
	httputil.WriteJSON(w, http.StatusOK, toActionViews(rows))
}

// POST /automation/rules/{id}/actions
func (h *Handler) CreateActionForRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	rule, err := h.q.GetAutomationRuleByID(r.Context(), ruleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "automation rule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load automation rule")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, rule.FarmID) {
		return
	}
	var body executableActionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// Phase 20.95 WS3 — reject two-source writes before hitting the DB CHECK.
	if body.ScheduleID != nil || body.ProgramID != nil {
		httputil.WriteError(w, http.StatusBadRequest,
			"executable_actions may bind to exactly one source; this endpoint binds to the rule in the URL, so schedule_id and program_id must be omitted")
		return
	}
	if err := validateActionType(body.ActionType); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	params, err := validateActionShape(&body, rule.FarmID, h.q, r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	row, err := h.q.CreateExecutableActionForRule(r.Context(), db.CreateExecutableActionForRuleParams{
		RuleID:                       &ruleID,
		ExecutionOrder:               body.ExecutionOrder,
		ActionType:                   commontypes.ExecutableActionTypeEnum(body.ActionType),
		TargetActuatorID:             body.TargetActuatorID,
		TargetAutomationRuleID:       nil,
		TargetNotificationTemplateID: body.TargetNotificationTemplateID,
		ActionCommand:                body.ActionCommand,
		ActionParameters:             params,
		DelayBeforeExecutionSeconds:  body.DelayBeforeExecutionSeconds,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create action: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, toActionView(row))
}

// PUT /automation/actions/{id}
func (h *Handler) UpdateAction(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid action id")
		return
	}
	existing, err := h.q.GetExecutableActionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "action not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load action")
		return
	}
	farmID, ok := h.resolveActionFarmID(w, r, existing)
	if !ok {
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body executableActionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validateActionType(body.ActionType); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	params, err := validateActionShape(&body, farmID, h.q, r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	row, err := h.q.UpdateExecutableAction(r.Context(), db.UpdateExecutableActionParams{
		ID:                           id,
		ExecutionOrder:               body.ExecutionOrder,
		ActionType:                   commontypes.ExecutableActionTypeEnum(body.ActionType),
		TargetActuatorID:             body.TargetActuatorID,
		TargetAutomationRuleID:       nil,
		TargetNotificationTemplateID: body.TargetNotificationTemplateID,
		ActionCommand:                body.ActionCommand,
		ActionParameters:             params,
		DelayBeforeExecutionSeconds:  body.DelayBeforeExecutionSeconds,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update action: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, toActionView(row))
}

// DELETE /automation/actions/{id}
func (h *Handler) DeleteAction(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid action id")
		return
	}
	existing, err := h.q.GetExecutableActionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "action not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load action")
		return
	}
	farmID, ok := h.resolveActionFarmID(w, r, existing)
	if !ok {
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	if err := h.q.DeleteExecutableAction(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete action")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// resolveActionFarmID walks the action's parent (rule/schedule/program) and
// returns the owning farm_id. It writes a 4xx/5xx response and returns ok=false
// on any failure.
func (h *Handler) resolveActionFarmID(w http.ResponseWriter, r *http.Request, a db.Gr33ncoreExecutableAction) (int64, bool) {
	switch {
	case a.RuleID != nil:
		rule, err := h.q.GetAutomationRuleByID(r.Context(), *a.RuleID)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to resolve parent rule")
			return 0, false
		}
		return rule.FarmID, true
	case a.ProgramID != nil:
		prog, err := h.q.GetFertigationProgramByID(r.Context(), *a.ProgramID)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to resolve parent program")
			return 0, false
		}
		return prog.FarmID, true
	case a.ScheduleID != nil:
		httputil.WriteError(w, http.StatusBadRequest, "schedule-bound actions are managed via the schedules API")
		return 0, false
	default:
		httputil.WriteError(w, http.StatusInternalServerError, "orphaned executable_action: no parent bound")
		return 0, false
	}
}

// ── Executable actions (program-bound) ─────────────────────────────────────
//
// Phase 20.9 WS4 mirrors the rule-bound CRUD onto fertigation programs so the
// program editor can attach structured actions (control_actuator, create_task,
// send_notification) instead of storing them in programs.metadata.steps.

// GET /fertigation/programs/{id}/actions
func (h *Handler) ListActionsByProgram(w http.ResponseWriter, r *http.Request) {
	programID, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid program id")
		return
	}
	prog, err := h.q.GetFertigationProgramByID(r.Context(), programID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load program")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, prog.FarmID) {
		return
	}
	rows, err := h.q.ListExecutableActionsByProgram(r.Context(), &programID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list actions")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreExecutableAction{}
	}
	httputil.WriteJSON(w, http.StatusOK, toActionViews(rows))
}

// POST /fertigation/programs/{id}/actions
func (h *Handler) CreateActionForProgram(w http.ResponseWriter, r *http.Request) {
	programID, err := httputil.PathID(r.URL.Path, 3)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid program id")
		return
	}
	prog, err := h.q.GetFertigationProgramByID(r.Context(), programID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load program")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, prog.FarmID) {
		return
	}
	var body executableActionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// Reject two-source writes before hitting the DB CHECK.
	if body.ScheduleID != nil || body.ProgramID != nil {
		httputil.WriteError(w, http.StatusBadRequest,
			"executable_actions may bind to exactly one source; this endpoint binds to the program in the URL, so schedule_id and program_id must be omitted")
		return
	}
	if err := validateActionType(body.ActionType); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	params, err := validateActionShape(&body, prog.FarmID, h.q, r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	row, err := h.q.CreateExecutableActionForProgram(r.Context(), db.CreateExecutableActionForProgramParams{
		ProgramID:                    &programID,
		ExecutionOrder:               body.ExecutionOrder,
		ActionType:                   commontypes.ExecutableActionTypeEnum(body.ActionType),
		TargetActuatorID:             body.TargetActuatorID,
		TargetAutomationRuleID:       nil,
		TargetNotificationTemplateID: body.TargetNotificationTemplateID,
		ActionCommand:                body.ActionCommand,
		ActionParameters:             params,
		DelayBeforeExecutionSeconds:  body.DelayBeforeExecutionSeconds,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create action: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, toActionView(row))
}
