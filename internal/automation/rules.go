package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

// RuleConditions is the canonical shape stored in
// `automation_rules.conditions_jsonb`. The scalar `condition_logic`
// column on the rule row is kept in sync with `Logic` so simple list
// queries can filter without JSON parsing; the full object here is what
// the evaluator reads.
type RuleConditions struct {
	Logic      string      `json:"logic"`
	Predicates []Predicate `json:"predicates"`
}

// runRuleTick evaluates every active rule once. Mirrors runTick for
// schedules: one DB fetch up front, per-rule work isolated so a bad
// row can't poison siblings. Every rule gets its `last_evaluated_time`
// stamped regardless of outcome (skip/fire/fail) so operators can see
// "yes, the worker looked at this rule".
func (w *Worker) runRuleTick(ctx context.Context, now time.Time) {
	rules, err := w.q.ListActiveAutomationRules(ctx)
	if err != nil {
		log.Printf("automation rule tick failed: %v", err)
		w.setLastTick(err)
		return
	}
	for _, r := range rules {
		w.evaluateRule(ctx, r, now)
	}
}

// TickRules runs a single rule-evaluation pass. Exported so integration
// tests can drive the evaluator deterministically without the 30s
// ticker — mirrors the Phase 19 WS4 treatment of Tick().
func (w *Worker) TickRules(ctx context.Context) {
	w.runRuleTick(ctx, time.Now().UTC())
}

// evaluateRule performs the full lifecycle for a single rule:
// cooldown check → parse conditions → evaluate predicates → dispatch
// actions → mark evaluated/triggered. Each terminal path records an
// `automation_runs` row so the runs page tells the whole story.
func (w *Worker) evaluateRule(ctx context.Context, rule db.Gr33ncoreAutomationRule, now time.Time) {
	// Always stamp last_evaluated_time at the end so operators can see
	// the worker visited the rule, even if nothing else happened.
	defer func() {
		if _, err := w.q.MarkAutomationRuleEvaluated(ctx, db.MarkAutomationRuleEvaluatedParams{
			ID:                rule.ID,
			LastEvaluatedTime: pgtype.Timestamptz{Time: now, Valid: true},
		}); err != nil {
			log.Printf("failed to mark rule %d evaluated: %v", rule.ID, err)
		}
	}()

	if w.rulePastCooldown(ctx, rule, now) {
		return
	}

	conds, ok := w.parseRuleConditions(ctx, rule, now)
	if !ok {
		return
	}

	passed, failed := EvaluatePredicates(ctx, w.q, conds.Logic, conds.Predicates)
	if !passed {
		details, _ := json.Marshal(map[string]any{
			"phase":          "conditions",
			"logic":          conds.Logic,
			"conditions_met": false,
			"failed":         failed,
		})
		if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     rule.FarmID,
			ScheduleID: nil,
			RuleID:     &rule.ID,
			Status:     "skipped",
			Message:    ptr("conditions_not_met"),
			Details:    details,
			ExecutedAt: now,
		}); err != nil {
			log.Printf("failed to record rule run: %v", err)
		}
		return
	}

	w.executeRule(ctx, rule, conds, now)
}

// rulePastCooldown returns true (AND records a skipped run) when the
// rule fired recently and is inside its cooldown window. Mirrors
// checkCooldown for schedules, but keyed off the rule's own
// `last_triggered_time` column rather than a runs lookup — we already
// have it on the row.
func (w *Worker) rulePastCooldown(ctx context.Context, rule db.Gr33ncoreAutomationRule, now time.Time) bool {
	if rule.CooldownPeriodSeconds == nil || *rule.CooldownPeriodSeconds <= 0 {
		return false
	}
	if !rule.LastTriggeredTime.Valid {
		return false
	}
	cooldown := time.Duration(*rule.CooldownPeriodSeconds) * time.Second
	elapsed := now.Sub(rule.LastTriggeredTime.Time)
	if elapsed >= cooldown {
		return false
	}
	remaining := cooldown - elapsed
	details, _ := json.Marshal(map[string]any{
		"phase":             "cooldown",
		"cooldown_seconds":  *rule.CooldownPeriodSeconds,
		"elapsed_seconds":   int64(elapsed.Seconds()),
		"remaining_seconds": int64(remaining.Seconds()),
	})
	log.Printf("rule %d (%s) skipped: cooldown %v remaining", rule.ID, rule.Name, remaining.Truncate(time.Second))
	if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
		FarmID:     rule.FarmID,
		ScheduleID: nil,
		RuleID:     &rule.ID,
		Status:     "skipped",
		Message:    ptr("cooldown"),
		Details:    details,
		ExecutedAt: now,
	}); err != nil {
		log.Printf("failed to record rule run: %v", err)
	}
	return true
}

// parseRuleConditions unmarshals conditions_jsonb. On failure it logs a
// failed run and returns (_, false) so the caller bails. Empty jsonb is
// treated as "no predicates" (always-fires rule), consistent with
// EvaluatePredicates's empty-slice contract.
func (w *Worker) parseRuleConditions(ctx context.Context, rule db.Gr33ncoreAutomationRule, now time.Time) (RuleConditions, bool) {
	var conds RuleConditions
	if len(rule.ConditionsJsonb) == 0 || string(rule.ConditionsJsonb) == "null" {
		if rule.ConditionLogic != nil {
			conds.Logic = *rule.ConditionLogic
		}
		return conds, true
	}
	if err := json.Unmarshal(rule.ConditionsJsonb, &conds); err != nil {
		details, _ := json.Marshal(map[string]any{
			"phase": "conditions",
			"error": "invalid_conditions_jsonb",
		})
		if _, runErr := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     rule.FarmID,
			ScheduleID: nil,
			RuleID:     &rule.ID,
			Status:     "failed",
			Message:    ptr(fmt.Sprintf("invalid conditions_jsonb: %v", err)),
			Details:    details,
			ExecutedAt: now,
		}); runErr != nil {
			log.Printf("failed to record rule run: %v", runErr)
		}
		return conds, false
	}
	if conds.Logic == "" && rule.ConditionLogic != nil {
		conds.Logic = *rule.ConditionLogic
	}
	return conds, true
}

// executeRule handles the "conditions met" path: list actions, dispatch
// each in execution_order, record a single aggregate run, and stamp
// `last_triggered_time`. Mirrors executeSchedule for schedules.
//
// WS3 fills in the per-type dispatch logic in dispatchRuleAction;
// WS2 deliberately ships the bookkeeping now so the cooldown +
// ALL/ANY smoke tests have something to observe.
func (w *Worker) executeRule(ctx context.Context, rule db.Gr33ncoreAutomationRule, conds RuleConditions, now time.Time) {
	actions, err := w.q.ListExecutableActionsByRule(ctx, &rule.ID)
	if err != nil {
		details, _ := json.Marshal(map[string]any{
			"phase": "list_actions",
			"logic": conds.Logic,
		})
		if _, runErr := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     rule.FarmID,
			ScheduleID: nil,
			RuleID:     &rule.ID,
			Status:     "failed",
			Message:    ptr(fmt.Sprintf("failed to list actions: %v", err)),
			Details:    details,
			ExecutedAt: now,
		}); runErr != nil {
			log.Printf("failed to record rule run: %v", runErr)
		}
		return
	}

	if len(actions) == 0 {
		details, _ := json.Marshal(map[string]any{
			"phase":          "actions",
			"logic":          conds.Logic,
			"conditions_met": true,
			"actions_total":  0,
		})
		if _, runErr := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     rule.FarmID,
			ScheduleID: nil,
			RuleID:     &rule.ID,
			Status:     "skipped",
			Message:    ptr("rule has no executable actions"),
			Details:    details,
			ExecutedAt: now,
		}); runErr != nil {
			log.Printf("failed to record rule run: %v", runErr)
		}
		// A rule with no actions still "fired" semantically — stamp
		// last_triggered_time so the cooldown guard applies next tick.
		w.markRuleTriggered(ctx, rule, now)
		return
	}

	type actionError struct {
		ActionID int64  `json:"action_id"`
		Message  string `json:"message"`
	}
	successCount := 0
	errs := make([]actionError, 0)
	for _, a := range actions {
		if err := w.dispatchRuleAction(ctx, rule, a, now); err != nil {
			errs = append(errs, actionError{ActionID: a.ID, Message: err.Error()})
			continue
		}
		successCount++
	}

	status := "success"
	switch {
	case successCount == 0 && len(errs) > 0:
		status = "failed"
	case len(errs) > 0:
		status = "partial_success"
	}

	details, _ := json.Marshal(map[string]any{
		"phase":           "actions",
		"logic":           conds.Logic,
		"conditions_met":  true,
		"actions_total":   len(actions),
		"actions_success": successCount,
		"errors":          errs,
	})
	msg := fmt.Sprintf("executed %d/%d actions", successCount, len(actions))
	if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
		FarmID:     rule.FarmID,
		ScheduleID: nil,
		RuleID:     &rule.ID,
		Status:     status,
		Message:    ptr(msg),
		Details:    details,
		ExecutedAt: now,
	}); err != nil {
		log.Printf("failed to record rule run: %v", err)
	}
	w.markRuleTriggered(ctx, rule, now)
}

func (w *Worker) markRuleTriggered(ctx context.Context, rule db.Gr33ncoreAutomationRule, now time.Time) {
	if _, err := w.q.MarkAutomationRuleTriggered(ctx, db.MarkAutomationRuleTriggeredParams{
		ID:                rule.ID,
		LastTriggeredTime: pgtype.Timestamptz{Time: now, Valid: true},
	}); err != nil {
		log.Printf("failed to mark rule %d triggered: %v", rule.ID, err)
	}
}

// dispatchRuleAction is the WS3 per-action-type switchboard. Each
// supported type has a small helper that (a) performs the real side
// effect, and (b) returns an error describing any failure so
// executeRule can record it in the run's `details.errors[]`. Deferred
// action types (webhook, log event, etc.) are rejected here too — the
// CRUD validator already blocks them at write time, but the worker is
// the last line of defense for rows inserted by older binaries.
func (w *Worker) dispatchRuleAction(ctx context.Context, rule db.Gr33ncoreAutomationRule, action db.Gr33ncoreExecutableAction, now time.Time) error {
	switch string(action.ActionType) {
	case "control_actuator":
		return w.dispatchRuleActuator(ctx, rule, action, now)
	case "create_task":
		return w.dispatchRuleCreateTask(ctx, rule, action, now)
	case "send_notification":
		return w.dispatchRuleSendNotification(ctx, rule, action, now)
	default:
		return fmt.Errorf("unsupported action_type=%s (deferred to a future phase)", action.ActionType)
	}
}

// dispatchRuleActuator mirrors executeAction(control_actuator) for
// schedules but stamps `triggered_by_rule_id` (not schedule) and uses
// source=automation_rule_trigger so /farms/{id}/automation/runs and
// the actuator event feed attribute the command back to this rule.
// Respects `delay_before_execution_seconds` by offsetting `event_time`.
func (w *Worker) dispatchRuleActuator(ctx context.Context, rule db.Gr33ncoreAutomationRule, action db.Gr33ncoreExecutableAction, now time.Time) error {
	if action.TargetActuatorID == nil {
		return fmt.Errorf("action %d missing target_actuator_id", action.ID)
	}
	if action.ActionCommand == nil || strings.TrimSpace(*action.ActionCommand) == "" {
		return fmt.Errorf("action %d missing action_command", action.ID)
	}
	actuator, err := w.q.GetActuatorByID(ctx, *action.TargetActuatorID)
	if err != nil {
		return fmt.Errorf("action %d: actuator lookup: %w", action.ID, err)
	}
	if actuator.FarmID != rule.FarmID {
		return fmt.Errorf("action %d actuator belongs to farm %d, not rule farm %d", action.ID, actuator.FarmID, rule.FarmID)
	}

	command := strings.TrimSpace(*action.ActionCommand)
	stateText := command
	switch command {
	case "on":
		stateText = "online"
	case "off":
		stateText = "offline"
	}

	eventTime := now
	if action.DelayBeforeExecutionSeconds != nil && *action.DelayBeforeExecutionSeconds > 0 {
		eventTime = now.Add(time.Duration(*action.DelayBeforeExecutionSeconds) * time.Second)
	}

	if w.simulation {
		var numeric pgtype.Numeric
		_ = numeric.Scan(0)
		if _, err := w.q.UpdateActuatorState(ctx, db.UpdateActuatorStateParams{
			ID:                  *action.TargetActuatorID,
			CurrentStateNumeric: numeric,
			CurrentStateText:    &stateText,
		}); err != nil {
			log.Printf("rule %d action %d: update actuator state: %v", rule.ID, action.ID, err)
		}
	}

	params, _ := json.Marshal(map[string]any{
		"command":         command,
		"simulation_mode": w.simulation,
		"rule_id":         rule.ID,
		"rule_name":       rule.Name,
	})

	status := db.Gr33ncoreActuatorExecutionStatusEnumPendingConfirmationFromFeedback
	if w.simulation {
		status = db.Gr33ncoreActuatorExecutionStatusEnumExecutionCompletedSuccessOnDevice
	}

	if _, err := w.q.InsertActuatorEvent(ctx, db.InsertActuatorEventParams{
		EventTime:             eventTime,
		ActuatorID:            *action.TargetActuatorID,
		CommandSent:           ptr(command),
		ParametersSent:        params,
		TriggeredByUserID:     pgtype.UUID{},
		TriggeredByScheduleID: nil,
		TriggeredByRuleID:     &rule.ID,
		Source:                db.Gr33ncoreActuatorEventSourceEnumAutomationRuleTrigger,
		ExecutionStatus: db.NullGr33ncoreActuatorExecutionStatusEnum{
			Gr33ncoreActuatorExecutionStatusEnum: status,
			Valid:                                true,
		},
		MetaData: []byte(`{}`),
	}); err != nil {
		return fmt.Errorf("action %d: insert actuator event: %w", action.ID, err)
	}

	// In non-simulation mode, queue the command onto the owning device
	// so the Pi client can pick it up on its next poll. Best-effort —
	// the rule still counts as "fired" even if the device is offline.
	if !w.simulation && actuator.DeviceID != nil {
		pendingJSON, _ := json.Marshal(map[string]any{
			"command": command,
			"rule_id": rule.ID,
		})
		if err := w.q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{
			ID:      *actuator.DeviceID,
			Column2: pendingJSON,
		}); err != nil {
			log.Printf("rule %d action %d: set pending command: %v", rule.ID, action.ID, err)
		}
	}
	return nil
}

// dispatchRuleCreateTask materialises a Gr33ncoreTask with
// source_rule_id pointing at the owning rule, so the Tasks page's
// "created by rule #X" chip (added in WS1 wiring) can attribute it.
//
// `action_parameters` shape (all optional except title):
//   { "title": "...", "description": "...", "zone_id": 42,
//     "task_type": "inspection", "priority": 2,
//     "due_in_days": 1, "estimated_duration_minutes": 30 }
func (w *Worker) dispatchRuleCreateTask(ctx context.Context, rule db.Gr33ncoreAutomationRule, action db.Gr33ncoreExecutableAction, now time.Time) error {
	if len(action.ActionParameters) == 0 {
		return fmt.Errorf("action %d missing action_parameters", action.ID)
	}
	var params map[string]any
	if err := json.Unmarshal(action.ActionParameters, &params); err != nil {
		return fmt.Errorf("action %d invalid action_parameters json: %w", action.ID, err)
	}

	title := strings.TrimSpace(stringField(params, "title"))
	if title == "" {
		title = fmt.Sprintf("Follow up on rule \"%s\"", rule.Name)
	}
	var description *string
	if v := strings.TrimSpace(stringField(params, "description")); v != "" {
		description = &v
	}
	taskType := "automation_follow_up"
	if v := strings.TrimSpace(stringField(params, "task_type")); v != "" {
		taskType = v
	}

	var priority *int32
	if n, ok := intField(params, "priority"); ok {
		if n < 0 || n > 3 {
			return fmt.Errorf("action %d priority %d out of range 0-3", action.ID, n)
		}
		p := int32(n)
		priority = &p
	}

	var zoneID *int64
	if n, ok := intField(params, "zone_id"); ok {
		z := n
		zoneID = &z
	}

	var dueDate pgtype.Date
	if n, ok := intField(params, "due_in_days"); ok && n >= 0 {
		d := now.AddDate(0, 0, int(n))
		dueDate = pgtype.Date{Time: d, Valid: true}
	}

	var estimated *int32
	if n, ok := intField(params, "estimated_duration_minutes"); ok && n > 0 {
		v := int32(n)
		estimated = &v
	}

	ruleID := rule.ID
	if _, err := w.q.CreateTask(ctx, db.CreateTaskParams{
		FarmID:                   rule.FarmID,
		ZoneID:                   zoneID,
		ScheduleID:               nil,
		Title:                    title,
		Description:              description,
		TaskType:                 &taskType,
		Status:                   commontypes.TaskStatusEnum("todo"),
		Priority:                 priority,
		AssignedToUserID:         pgtype.UUID{},
		DueDate:                  dueDate,
		EstimatedDurationMinutes: estimated,
		SourceAlertID:            nil,
		SourceRuleID:             &ruleID,
		CreatedByUserID:          pgtype.UUID{},
	}); err != nil {
		return fmt.Errorf("action %d: create task: %w", action.ID, err)
	}
	return nil
}

// dispatchRuleSendNotification renders the action's notification
// template, writes an alerts_notifications row (so it's visible on the
// Alerts page with `triggering_event_source_type='automation_rule'`),
// and then fans the push out through the injected PushNotifier. If no
// notifier was provided (tests, or FCM creds missing in prod), the DB
// row is still written so operators can see the alert in-app.
//
// Template rendering is intentionally minimal for WS3: we substitute
// `{{ key }}` placeholders from `action_parameters.variables` (a
// JSON object). Anything fancier (conditionals, loops) is future work.
func (w *Worker) dispatchRuleSendNotification(ctx context.Context, rule db.Gr33ncoreAutomationRule, action db.Gr33ncoreExecutableAction, now time.Time) error {
	if action.TargetNotificationTemplateID == nil {
		return fmt.Errorf("action %d missing target_notification_template_id", action.ID)
	}
	tmpl, err := w.q.GetNotificationTemplateByID(ctx, *action.TargetNotificationTemplateID)
	if err != nil {
		return fmt.Errorf("action %d: load template: %w", action.ID, err)
	}
	// Templates may be global (farm_id IS NULL, is_system_template=true)
	// or per-farm. Reject cross-farm templates defensively.
	if tmpl.FarmID != nil && *tmpl.FarmID != rule.FarmID {
		return fmt.Errorf("action %d template belongs to farm %d, not rule farm %d", action.ID, *tmpl.FarmID, rule.FarmID)
	}

	vars := map[string]string{
		"rule_name":    rule.Name,
		"rule_id":      strconv.FormatInt(rule.ID, 10),
		"triggered_at": now.Format(time.RFC3339),
	}
	if len(action.ActionParameters) > 0 {
		var raw map[string]any
		if err := json.Unmarshal(action.ActionParameters, &raw); err == nil {
			if extra, ok := raw["variables"].(map[string]any); ok {
				for k, v := range extra {
					vars[k] = fmt.Sprint(v)
				}
			}
		}
	}

	subject := renderTemplate(tmpl.SubjectTemplate, vars, "Automation rule "+rule.Name)
	body := renderTemplate(tmpl.BodyTemplateText, vars, "")

	severity := db.NullGr33ncoreNotificationPriorityEnum{
		Gr33ncoreNotificationPriorityEnum: db.Gr33ncoreNotificationPriorityEnumMedium,
		Valid:                             true,
	}
	if tmpl.DefaultPriority.Valid {
		severity = tmpl.DefaultPriority
	}

	ruleID := rule.ID
	alert, err := w.q.CreateAlertForRule(ctx, db.CreateAlertForRuleParams{
		FarmID:                  rule.FarmID,
		NotificationTemplateID:  &tmpl.ID,
		TriggeringEventSourceID: &ruleID,
		Severity:                severity,
		SubjectRendered:         &subject,
		MessageTextRendered:     &body,
	})
	if err != nil {
		return fmt.Errorf("action %d: create alert: %w", action.ID, err)
	}

	if w.notifier != nil {
		w.notifier.DispatchFarmAlert(ctx, alert)
	}
	return nil
}

// renderTemplate does a tiny {{ key }} substitution pass. If the
// supplied template pointer is nil or empty and a fallback is provided,
// the fallback is returned verbatim.
func renderTemplate(tmpl *string, vars map[string]string, fallback string) string {
	if tmpl == nil || strings.TrimSpace(*tmpl) == "" {
		return fallback
	}
	out := *tmpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
		out = strings.ReplaceAll(out, "{{ "+k+" }}", v)
	}
	return out
}

// stringField looks up `key` on a decoded JSON object and returns the
// value as a string (including number→string coercion). Missing keys
// return "".
func stringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprint(val)
	}
}

// intField extracts an int64 from a decoded JSON object. JSON numbers
// decode to float64, so we accept float64 (truncated) and strings that
// parse cleanly. Returns (0, false) when the key is missing or not a
// number — callers use that to decide whether to apply a default.
func intField(m map[string]any, key string) (int64, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int:
		return int64(val), true
	case int64:
		return val, true
	case string:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}
