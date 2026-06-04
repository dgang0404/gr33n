package automation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/robfig/cron/v3"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/fertigation/mixplan"
	acthandler "gr33n-api/internal/handler/actuator"
	"gr33n-api/internal/platform/commontypes"
)

// runProgramTick is the Phase 22 WS1 counterpart to runTick (schedules)
// and runRuleTick (rules). It evaluates every active fertigation program
// whose bound schedule's cron expression fires at `now` and dispatches
// its actions through dispatchProgramAction.
//
// Programs without a bound schedule are skipped on cron ticks — use
// POST /farms/{id}/fertigation/programs/{id}/run-now for ad-hoc runs. Cron
// evaluation reuses shouldTriggerNow via a tiny shouldTriggerProgramNow
// wrapper so schedule dedup and program dedup share the same mental
// model.
//
// Idempotency: we stamp program.last_triggered_time on success and also
// write an automation_runs row with a deterministic idempotency key in
// `details.idempotency_key`. checkProgramIdempotency short-circuits on
// the second visit within the same minute so an operator manually
// invoking Tick() twice in a second doesn't fire every program twice.
func (w *Worker) runProgramTick(ctx context.Context, now time.Time) {
	programs, err := w.q.ListActivePrograms(ctx)
	if err != nil {
		slog.Warn("automation worker tick failed", "phase", "list_programs", "err", err)
		return
	}
	for _, p := range programs {
		w.evaluateProgram(ctx, p, now)
	}
}

// TickPrograms runs a single program-evaluation pass. Exported so
// integration tests can drive the program scheduler deterministically
// without the 30s ticker.
func (w *Worker) TickPrograms(ctx context.Context) {
	w.runProgramTick(ctx, time.Now().UTC().Truncate(time.Minute))
}

func (w *Worker) evaluateProgram(ctx context.Context, p db.Gr33nfertigationProgram, now time.Time) {
	if p.ScheduleID == nil {
		return
	}
	schedule, err := w.q.GetScheduleByID(ctx, *p.ScheduleID)
	if err != nil {
		// Program points at a schedule that's been deleted/detached.
		// Record a skip so the runs page surfaces the misconfiguration.
		w.recordProgramRun(ctx, p, "skipped",
			fmt.Sprintf("program %d references missing schedule %d: %v", p.ID, *p.ScheduleID, err),
			map[string]any{"phase": "schedule_lookup"}, now)
		return
	}
	if !schedule.IsActive {
		return
	}

	should, evalErr := shouldTriggerNow(schedule.CronExpression, schedule.Timezone, schedule.LastTriggeredTime, now)
	if evalErr != nil {
		w.recordProgramRun(ctx, p, "failed",
			fmt.Sprintf("cron parse error for program %q (schedule %s): %v", p.Name, schedule.Name, evalErr),
			map[string]any{"phase": "cron_eval"}, now)
		return
	}
	// The "already triggered this minute" guard uses the program's own
	// last_triggered_time, not the schedule's — a single schedule can
	// own multiple programs, and we want each program to decide for
	// itself.
	if p.LastTriggeredTime.Valid && p.LastTriggeredTime.Time.UTC().Truncate(time.Minute).Equal(now) {
		return
	}
	if !should {
		// Also honour the schedule's parser: if shouldTriggerNow says
		// the cron doesn't match this minute, we're done.
		_ = w.maybeStampSchedulePrevTrigger(schedule) // noop shim for future use
		return
	}

	idemKey := programIdempotencyKey(p.ID, now)
	if w.checkProgramIdempotency(ctx, p, idemKey, now) {
		return
	}

	dctx := programDispatchCtx{scheduleID: &schedule.ID, manualRun: false}
	w.executeProgram(ctx, p, dctx, now, idemKey)
}

// RunProgramNow executes a fertigation program immediately (product backlog B1).
// Idempotency matches cron ticks (same minute → second call is a no-op).
// Returns duplicate=true when an automation_runs row already exists for the key.
func (w *Worker) RunProgramNow(ctx context.Context, p db.Gr33nfertigationProgram) (status string, message string, duplicate bool, err error) {
	now := time.Now().UTC().Truncate(time.Minute)
	idemKey := programIdempotencyKey(p.ID, now)
	if w.checkProgramIdempotency(ctx, p, idemKey, now) {
		return "skipped", "program already ran this minute (idempotent)", true, nil
	}
	var dctx programDispatchCtx
	if p.ScheduleID != nil {
		if sched, sErr := w.q.GetScheduleByID(ctx, *p.ScheduleID); sErr == nil {
			dctx.scheduleID = &sched.ID
		}
	}
	dctx.manualRun = true
	w.executeProgram(ctx, p, dctx, now, idemKey)
	return "accepted", "program run started", false, nil
}

// maybeStampSchedulePrevTrigger is a placeholder so the compiler keeps
// the `schedule` param live even when we take the "cron says no" branch.
// If/when we want to back-stamp the schedule's last-evaluated marker for
// program-only schedules, this is the hook.
func (w *Worker) maybeStampSchedulePrevTrigger(_ db.Gr33ncoreSchedule) error { return nil }

func programIdempotencyKey(programID int64, now time.Time) string {
	raw := fmt.Sprintf("program:%d:%s", programID, now.Format("2006-01-02T15:04"))
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h[:8])
}

// checkProgramIdempotency mirrors checkIdempotency but keys on
// program_id so the schedule-bound idempotency row (if any) doesn't
// shadow the program's.
func (w *Worker) checkProgramIdempotency(ctx context.Context, p db.Gr33nfertigationProgram, key string, _ time.Time) bool {
	detailsJSON, _ := json.Marshal(map[string]string{"idempotency_key": key})
	_, err := w.q.GetAutomationRunByProgramAndDetails(ctx, db.GetAutomationRunByProgramAndDetailsParams{
		ProgramID: &p.ID,
		Column2:   detailsJSON,
	})
	if err == nil {
		log.Printf("program %d (%s) skipped: idempotent run already exists (key=%s)", p.ID, p.Name, key)
		return true
	}
	return false
}

func (w *Worker) executeProgram(
	ctx context.Context,
	p db.Gr33nfertigationProgram,
	dctx programDispatchCtx,
	now time.Time,
	idemKey string,
) {
	// Phase 39 WS5 / 39b: mix_batch only for fertigation programs (not irrigation_only).
	if !w.simulation && !p.IrrigationOnly {
		if mixCmdID, mixErr := w.dispatchProgramMix(ctx, p, now); mixErr != nil {
			// ErrProgramHasNoRecipe is expected for plain-irrigation programs — not a failure.
			if !errors.Is(mixErr, mixplan.ErrProgramHasNoRecipe) {
				log.Printf("program %d: mix_batch enqueue: %v", p.ID, mixErr)
			}
		} else {
			log.Printf("program %d: enqueued mix_batch command_id=%d", p.ID, mixCmdID)
		}
	}

	actions, source, err := ResolveProgramActions(ctx, w.q, p)
	if err != nil {
		w.recordProgramRun(ctx, p, "failed",
			fmt.Sprintf("failed to resolve actions: %v", err),
			map[string]any{"phase": "resolve_actions", "idempotency_key": idemKey}, now)
		return
	}

	// Option C WS2 — structured warning whenever we fall back to the
	// legacy metadata.steps array. Operators can grep their worker
	// logs (or scrape automation_runs.details.action_source) to find
	// programs still awaiting backfill.
	if source == ProgramActionsFromMetadataStepsFallback {
		w.noteMetadataStepsFallback()
		log.Printf("program %d (%s) using metadata.steps fallback — run the 20260515 backfill or POST /fertigation/programs/%d/actions rows", p.ID, p.Name, p.ID)
	}

	if len(actions) == 0 {
		w.recordProgramRun(ctx, p, "skipped", "program has no executable actions",
			map[string]any{"phase": "execute", "actions": 0, "idempotency_key": idemKey, "action_source": string(source)}, now)
		w.markProgramTriggered(ctx, p, now)
		return
	}

	type actionError struct {
		ActionID int64  `json:"action_id"`
		Message  string `json:"message"`
	}
	successCount := 0
	errs := make([]actionError, 0)
	for _, a := range actions {
		if err := w.dispatchProgramActionWithRetry(ctx, p, dctx, a, now); err != nil {
			// Synthetic (fallback) actions have ID=0; surface the
			// execution_order instead so the error row still points
			// operators at a specific step.
			id := a.ID
			if id == 0 {
				id = int64(a.ExecutionOrder)
			}
			errs = append(errs, actionError{ActionID: id, Message: err.Error()})
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

	details := map[string]any{
		"phase":           "actions",
		"actions_total":   len(actions),
		"actions_success": successCount,
		"actions_failed":  len(errs),
		"simulation_mode": w.simulation,
		"idempotency_key": idemKey,
		"action_source":   string(source),
		"errors":          errs,
		"manual_run":      dctx.manualRun,
	}
	if dctx.scheduleID != nil {
		details["schedule_id"] = *dctx.scheduleID
	}
	msg := fmt.Sprintf("executed %d/%d actions", successCount, len(actions))
	w.recordProgramRun(ctx, p, status, msg, details, now)
	w.markProgramTriggered(ctx, p, now)
}

func (w *Worker) markProgramTriggered(ctx context.Context, p db.Gr33nfertigationProgram, now time.Time) {
	if _, err := w.q.MarkProgramTriggered(ctx, db.MarkProgramTriggeredParams{
		ID:                p.ID,
		LastTriggeredTime: pgtype.Timestamptz{Time: now, Valid: true},
	}); err != nil {
		log.Printf("failed to mark program %d triggered: %v", p.ID, err)
	}
}

func (w *Worker) recordProgramRun(
	ctx context.Context,
	p db.Gr33nfertigationProgram,
	status string,
	message string,
	details map[string]any,
	now time.Time,
) {
	payload, _ := json.Marshal(details)
	programID := p.ID
	if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
		FarmID:     p.FarmID,
		ScheduleID: p.ScheduleID,
		RuleID:     nil,
		ProgramID:  &programID,
		Status:     status,
		Message:    ptr(message),
		Details:    payload,
		ExecutedAt: now,
	}); err != nil {
		log.Printf("failed to record program run: %v", err)
	}
}

// dispatchProgramActionWithRetry mirrors executeActionWithRetry but
// keeps program-bound retry bookkeeping self-contained. Same
// isTransient classifier as the schedule path.
func (w *Worker) dispatchProgramActionWithRetry(
	ctx context.Context,
	p db.Gr33nfertigationProgram,
	dctx programDispatchCtx,
	action db.Gr33ncoreExecutableAction,
	now time.Time,
) error {
	var lastErr error
	for attempt := range w.maxRetries + 1 {
		lastErr = w.dispatchProgramAction(ctx, p, dctx, action, now)
		if lastErr == nil {
			return nil
		}
		if !isTransient(lastErr) {
			return fmt.Errorf("[permanent] %w", lastErr)
		}
		if attempt < w.maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * 500 * time.Millisecond
			log.Printf("program %d action %d transient error (attempt %d/%d), retrying in %v: %v",
				p.ID, action.ID, attempt+1, w.maxRetries+1, backoff, lastErr)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			}
		}
	}
	return fmt.Errorf("[transient after %d retries] %w", w.maxRetries+1, lastErr)
}

// dispatchProgramAction is the per-action-type switchboard for
// program-driven actions. Supports the same trio that rules support
// (control_actuator, create_task, send_notification). Anything else is
// rejected with a clear error so the automation_runs row surfaces the
// misconfiguration instead of silently dropping the step.
func (w *Worker) dispatchProgramAction(
	ctx context.Context,
	p db.Gr33nfertigationProgram,
	dctx programDispatchCtx,
	action db.Gr33ncoreExecutableAction,
	now time.Time,
) error {
	switch string(action.ActionType) {
	case "control_actuator":
		return w.dispatchProgramActuator(ctx, p, dctx, action, now)
	case "create_task":
		return w.dispatchProgramCreateTask(ctx, p, action, now)
	case "send_notification":
		return w.dispatchProgramSendNotification(ctx, p, action, now)
	default:
		return fmt.Errorf("unsupported action_type=%s for program actions", action.ActionType)
	}
}

// dispatchProgramActuator fires a control_actuator step. Provenance is
// recorded via triggered_by_schedule_id (the program is schedule-bound)
// plus a program_id stashed in actuator_events.meta_data so the
// actuator event feed can attribute back to the specific program.
func (w *Worker) dispatchProgramActuator(
	ctx context.Context,
	p db.Gr33nfertigationProgram,
	dctx programDispatchCtx,
	action db.Gr33ncoreExecutableAction,
	now time.Time,
) error {
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
	if actuator.FarmID != p.FarmID {
		return fmt.Errorf("action %d actuator belongs to farm %d, not program farm %d", action.ID, actuator.FarmID, p.FarmID)
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
			log.Printf("program %d action %d: update actuator state: %v", p.ID, action.ID, err)
		}
	}

	params := map[string]any{
		"command":         command,
		"simulation_mode": w.simulation,
		"program_id":      p.ID,
		"program_name":    p.Name,
	}
	if dctx.scheduleID != nil {
		params["schedule_id"] = *dctx.scheduleID
	}
	paramsJSON, _ := json.Marshal(params)
	meta, _ := json.Marshal(map[string]any{
		"program_id":   p.ID,
		"program_name": p.Name,
	})

	eventSource := db.Gr33ncoreActuatorEventSourceEnumScheduleTrigger
	if dctx.manualRun {
		eventSource = db.Gr33ncoreActuatorEventSourceEnumManualApiCall
	}

	status := db.Gr33ncoreActuatorExecutionStatusEnumPendingConfirmationFromFeedback
	if w.simulation {
		status = db.Gr33ncoreActuatorExecutionStatusEnumExecutionCompletedSuccessOnDevice
	}

	var triggeredByScheduleID *int64
	if dctx.scheduleID != nil {
		triggeredByScheduleID = dctx.scheduleID
	}

	if _, err := w.q.InsertActuatorEvent(ctx, db.InsertActuatorEventParams{
		EventTime:             eventTime,
		ActuatorID:            *action.TargetActuatorID,
		CommandSent:           ptr(command),
		ParametersSent:        paramsJSON,
		TriggeredByUserID:     pgtype.UUID{},
		TriggeredByScheduleID: triggeredByScheduleID,
		TriggeredByRuleID:     nil,
		Source:                eventSource,
		ExecutionStatus: &status,
		MetaData:        meta,
	}); err != nil {
		return fmt.Errorf("action %d: insert actuator event: %w", action.ID, err)
	}

	if !w.simulation && actuator.DeviceID != nil {
		var dur *int
		if p.RunDurationSeconds != nil && *p.RunDurationSeconds > 0 && command == "on" {
			d := int(*p.RunDurationSeconds)
			dur = &d
		}
		progID := p.ID
		cmdSource := "schedule"
		if dctx.manualRun {
			cmdSource = "operator"
		}
		pendingIn := acthandler.PendingCommandInput{
			ActuatorID:      *action.TargetActuatorID,
			Command:         command,
			Source:          cmdSource,
			DurationSeconds: dur,
			ProgramID:       &progID,
		}
		if dctx.scheduleID != nil {
			schedID := *dctx.scheduleID
			pendingIn.ScheduleID = &schedID
		}
		pendingJSON, err := acthandler.BuildPendingCommandJSONFull(pendingIn)
		if err != nil {
			log.Printf("program %d action %d: build pending command: %v", p.ID, action.ID, err)
		} else {
			cmdType := "actuator"
			if dur != nil {
				cmdType = "pulse"
			}
			aID := *action.TargetActuatorID
			queueSource := "program"
			if dctx.manualRun {
				queueSource = "operator"
			}
			enqueue := db.EnqueueDeviceCommandParams{
				DeviceID:    *actuator.DeviceID,
				FarmID:      p.FarmID,
				CommandType: cmdType,
				Payload:     pendingJSON,
				Source:      queueSource,
				ActuatorID:  &aID,
				ProgramID:   &progID,
			}
			if dctx.scheduleID != nil {
				enqueue.ScheduleID = dctx.scheduleID
			}
			// Phase 39 WS1: enqueue to FIFO queue; mirror pending_command for backward compat.
			if _, qErr := w.q.EnqueueDeviceCommand(ctx, enqueue); qErr != nil {
				log.Printf("program %d action %d: enqueue device command: %v", p.ID, action.ID, qErr)
			}
			// Keep writing legacy slot so pre-39 Pi clients still receive the command.
			if err := w.q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{
				ID:      *actuator.DeviceID,
				Column2: pendingJSON,
			}); err != nil {
				log.Printf("program %d action %d: set pending command: %v", p.ID, action.ID, err)
			}
		}
	}
	return nil
}

// dispatchProgramCreateTask materialises a Gr33ncoreTask. Programs don't
// have a dedicated source_program_id column on tasks (yet), so we stash
// the program ID in the task description as a breadcrumb and tie the
// task back to the program's schedule via schedule_id. Operators who
// need strict program attribution can filter tasks by description
// prefix; a future schema pass can promote it to a dedicated column.
func (w *Worker) dispatchProgramCreateTask(
	ctx context.Context,
	p db.Gr33nfertigationProgram,
	action db.Gr33ncoreExecutableAction,
	now time.Time,
) error {
	if len(action.ActionParameters) == 0 {
		return fmt.Errorf("action %d missing action_parameters", action.ID)
	}
	var params map[string]any
	if err := json.Unmarshal(action.ActionParameters, &params); err != nil {
		return fmt.Errorf("action %d invalid action_parameters json: %w", action.ID, err)
	}

	title := strings.TrimSpace(stringField(params, "title"))
	if title == "" {
		title = fmt.Sprintf("Follow up on program %q", p.Name)
	}
	descStr := fmt.Sprintf("[program_id=%d] %s", p.ID, strings.TrimSpace(stringField(params, "description")))
	description := &descStr
	taskType := "automation_follow_up"
	if v := strings.TrimSpace(stringField(params, "task_type")); v != "" {
		taskType = v
	}

	var priority *int32
	if n, ok := intField(params, "priority"); ok {
		if n < 0 || n > 3 {
			return fmt.Errorf("action %d priority %d out of range 0-3", action.ID, n)
		}
		v := int32(n)
		priority = &v
	}

	var zoneID *int64
	if n, ok := intField(params, "zone_id"); ok {
		z := n
		zoneID = &z
	} else if p.TargetZoneID != nil {
		z := *p.TargetZoneID
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

	if _, err := w.q.CreateTask(ctx, db.CreateTaskParams{
		FarmID:                   p.FarmID,
		ZoneID:                   zoneID,
		ScheduleID:               p.ScheduleID,
		Title:                    title,
		Description:              description,
		TaskType:                 &taskType,
		Status:                   commontypes.TaskStatusEnum("todo"),
		Priority:                 priority,
		AssignedToUserID:         pgtype.UUID{},
		DueDate:                  dueDate,
		EstimatedDurationMinutes: estimated,
		SourceAlertID:            nil,
		SourceRuleID:             nil,
		CreatedByUserID:          pgtype.UUID{},
	}); err != nil {
		return fmt.Errorf("action %d: create task: %w", action.ID, err)
	}
	return nil
}

// dispatchProgramSendNotification mirrors dispatchRuleSendNotification
// but uses CreateAlertForProgram so the Alerts page can show
// `triggering_event_source_type='automation_program'`.
func (w *Worker) dispatchProgramSendNotification(
	ctx context.Context,
	p db.Gr33nfertigationProgram,
	action db.Gr33ncoreExecutableAction,
	now time.Time,
) error {
	if action.TargetNotificationTemplateID == nil {
		return fmt.Errorf("action %d missing target_notification_template_id", action.ID)
	}
	tmpl, err := w.q.GetNotificationTemplateByID(ctx, *action.TargetNotificationTemplateID)
	if err != nil {
		return fmt.Errorf("action %d: load template: %w", action.ID, err)
	}
	if tmpl.FarmID != nil && *tmpl.FarmID != p.FarmID {
		return fmt.Errorf("action %d template belongs to farm %d, not program farm %d", action.ID, *tmpl.FarmID, p.FarmID)
	}

	vars := map[string]string{
		"program_name": p.Name,
		"program_id":   strconv.FormatInt(p.ID, 10),
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

	subject := renderTemplate(tmpl.SubjectTemplate, vars, "Automation program "+p.Name)
	body := renderTemplate(tmpl.BodyTemplateText, vars, "")

	severityVal := db.Gr33ncoreNotificationPriorityEnumMedium
	if tmpl.DefaultPriority != nil {
		severityVal = *tmpl.DefaultPriority
	}

	progID := p.ID
	alert, err := w.q.CreateAlertForProgram(ctx, db.CreateAlertForProgramParams{
		FarmID:                  p.FarmID,
		NotificationTemplateID:  &tmpl.ID,
		TriggeringEventSourceID: &progID,
		Severity:                &severityVal,
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

// cron parser sanity check — compile-time guard that the import of
// robfig/cron stays needed even if the schedule path refactors away
// from shouldTriggerNow in the future.
var _ = cron.New

// sentinel used by tests that want to verify the fallback warning path
// fired for a specific program during the last tick.
var ErrProgramTickMetadataFallback = errors.New("program action resolved via metadata.steps fallback")

// dispatchProgramMix is the Phase 39 WS5 addition: before the normal
// control_actuator actions fire, calculate a MixPlan and enqueue a
// mix_batch command onto the reservoir's delivery device.
//
// Returns ErrProgramHasNoRecipe when the program is plain irrigation —
// callers should treat this as a non-error skip.
// Returns the command ID on success so the caller can log provenance.
func (w *Worker) dispatchProgramMix(ctx context.Context, p db.Gr33nfertigationProgram, _ time.Time) (int64, error) {
	// BuildFromProgramRow does all the guard checks (recipe, reservoir, base EC).
	in, err := mixplan.BuildFromProgramRow(ctx, w.q, p, mixplan.BuildOptions{})
	if err != nil {
		return 0, err
	}
	plan, err := mixplan.Calculate(in)
	if err != nil {
		return 0, fmt.Errorf("calculate mix plan: %w", err)
	}

	// Resolve device from the reservoir's delivery_actuator.
	res, err := w.q.GetFertigationReservoirByID(ctx, plan.ReservoirID)
	if err != nil {
		return 0, fmt.Errorf("load reservoir: %w", err)
	}
	if res.DeliveryActuatorID == nil {
		return 0, fmt.Errorf("reservoir %d has no delivery_actuator_id — cannot enqueue mix_batch", res.ID)
	}
	actuator, err := w.q.GetActuatorByID(ctx, *res.DeliveryActuatorID)
	if err != nil || actuator.DeviceID == nil {
		return 0, fmt.Errorf("delivery actuator not bound to a device")
	}

	payload, _ := json.Marshal(map[string]any{
		"command_type": "mix_batch",
		"program_id":   p.ID,
		"reservoir_id": plan.ReservoirID,
		"mix_plan":     plan,
	})

	cmd, err := w.q.EnqueueDeviceCommand(ctx, db.EnqueueDeviceCommandParams{
		DeviceID:    *actuator.DeviceID,
		FarmID:      p.FarmID,
		CommandType: "mix_batch",
		Payload:     payload,
		Source:      "program",
		ProgramID:   &p.ID,
	})
	if err != nil {
		return 0, fmt.Errorf("enqueue mix_batch: %w", err)
	}
	return cmd.ID, nil
}
