package automation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	db "gr33n-api/internal/db"
)

type Status struct {
	Running        bool      `json:"running"`
	SimulationMode bool      `json:"simulation_mode"`
	LastTickAt     time.Time `json:"last_tick_at"`
	LastError      string    `json:"last_error,omitempty"`
}

// PushNotifier is the minimum interface the rule evaluator needs from
// pushnotify.Dispatcher. Declaring it here (instead of importing
// pushnotify directly) avoids a potential import cycle and lets tests
// stub it. The real implementation is internal/pushnotify.Dispatcher.
type PushNotifier interface {
	DispatchFarmAlert(ctx context.Context, alert db.Gr33ncoreAlertsNotification)
}

type Worker struct {
	q          *db.Queries
	simulation bool
	cooldown   time.Duration
	maxRetries int
	notifier   PushNotifier

	mu         sync.RWMutex
	running    bool
	lastTickAt time.Time
	lastError  string
	// Phase 20.7: track the last UTC date we rolled up electricity for
	// so the scheduled tick only fires once per day (the per-actuator
	// idempotency table handles retries separately).
	lastElecRollupDate time.Time
}

func NewWorker(pool *pgxpool.Pool, simulation bool, opts ...WorkerOption) *Worker {
	w := &Worker{
		q:          db.New(pool),
		simulation: simulation,
		cooldown:   2 * time.Minute,
		maxRetries: 2,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

type WorkerOption func(*Worker)

func WithCooldown(d time.Duration) WorkerOption {
	return func(w *Worker) { w.cooldown = d }
}

func WithMaxRetries(n int) WorkerOption {
	return func(w *Worker) { w.maxRetries = n }
}

// WithPushNotifier injects a push dispatcher for rule-driven
// send_notification actions. If nil, the worker inserts the
// alerts_notifications row but skips the push fan-out (the alert still
// shows up on the Alerts page).
func WithPushNotifier(n PushNotifier) WorkerOption {
	return func(w *Worker) { w.notifier = n }
}

func (w *Worker) Start(ctx context.Context) {
	w.mu.Lock()
	w.running = true
	w.mu.Unlock()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	w.runTick(ctx)
	for {
		select {
		case <-ctx.Done():
			w.mu.Lock()
			w.running = false
			w.mu.Unlock()
			return
		case <-ticker.C:
			w.runTick(ctx)
		}
	}
}

func (w *Worker) GetStatus() Status {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return Status{
		Running:        w.running,
		SimulationMode: w.simulation,
		LastTickAt:     w.lastTickAt,
		LastError:      w.lastError,
	}
}

func (w *Worker) setLastTick(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastTickAt = time.Now().UTC()
	if err != nil {
		w.lastError = err.Error()
	} else {
		w.lastError = ""
	}
}

func (w *Worker) runTick(ctx context.Context) {
	now := time.Now().UTC().Truncate(time.Minute)
	schedules, err := w.q.ListActiveSchedules(ctx)
	if err != nil {
		w.setLastTick(err)
		log.Printf("automation tick failed: %v", err)
		return
	}

	for _, s := range schedules {
		should, evalErr := shouldTriggerNow(s.CronExpression, s.LastTriggeredTime, now)
		if evalErr != nil {
			if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
				FarmID:     s.FarmID,
				ScheduleID: &s.ID,
				RuleID:     nil,
				Status:     "failed",
				Message:    ptr(fmt.Sprintf("cron parse error for %s: %v", s.Name, evalErr)),
				Details:    []byte(`{"phase":"cron_eval"}`),
				ExecutedAt: now,
			}); err != nil {
				log.Printf("failed to record automation run: %v", err)
			}
			continue
		}
		if !should {
			continue
		}

		if w.cooldown > 0 {
			if skipped := w.checkCooldown(ctx, s, now); skipped {
				continue
			}
		}

		idemKey := idempotencyKey(s.ID, now)
		if w.checkIdempotency(ctx, s, idemKey, now) {
			continue
		}

		// Interlock-lite: the latest reading for every sensor in
		// preconditions must satisfy its predicate before we touch actuators.
		if skipped := w.checkPreconditions(ctx, s, now); skipped {
			continue
		}

		w.executeSchedule(ctx, s, now, idemKey)
	}
	w.runRuleTick(ctx, now)

	// Phase 22 WS1 — program-tick. Evaluates every active fertigation
	// program whose bound schedule's cron fires at `now`. Deliberately
	// runs after the schedule/rule passes so actuator commands from
	// schedules land first (programs usually piggyback on the same
	// schedule and we want consistent ordering in the event feed).
	w.runProgramTick(ctx, now)

	// Phase 20.7 WS5 — cheap per-batch-per-day deduped alert sweep.
	// Safe to run every tick; the dedupe inside maybeFireLowStock
	// keeps the alert table from filling up.
	w.TickLowStockAlerts(ctx)

	// Phase 20.7 WS4 — once-per-day electricity rollup. We target
	// the *previous* UTC calendar day so all its events have landed.
	// The first tick after 01:00 UTC wins; subsequent ticks see
	// lastElecRollupDate == today and skip. The per-(actuator,date)
	// idempotency table is the second line of defence.
	yesterday := now.AddDate(0, 0, -1)
	yesterdayDate := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
	if now.Hour() >= 1 && !w.lastElecRollupDate.Equal(yesterdayDate) {
		w.TickElectricityRollup(ctx, yesterdayDate)
		w.lastElecRollupDate = yesterdayDate
	}

	w.setLastTick(nil)
}

// Tick runs a single evaluation pass. Exported so integration tests can
// exercise the scheduler deterministically without the 30s ticker.
func (w *Worker) Tick(ctx context.Context) {
	w.runTick(ctx)
}

func shouldTriggerNow(expr string, lastTriggered pgtype.Timestamptz, now time.Time) (bool, error) {
	if lastTriggered.Valid && lastTriggered.Time.UTC().Truncate(time.Minute).Equal(now) {
		return false, nil
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(expr)
	if err != nil {
		return false, err
	}
	previousMinute := now.Add(-1 * time.Minute)
	next := sched.Next(previousMinute)
	return next.Equal(now), nil
}

// checkCooldown prevents re-execution if the last successful run is within the cooldown window.
// Returns true (skip) on DB errors to avoid running unchecked.
func (w *Worker) checkCooldown(ctx context.Context, s db.Gr33ncoreSchedule, now time.Time) bool {
	lastRun, err := w.q.GetLastSuccessfulRunBySchedule(ctx, &s.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false
		}
		log.Printf("cooldown check failed for schedule %d: %v", s.ID, err)
		return true
	}
	elapsed := now.Sub(lastRun.ExecutedAt)
	if elapsed < w.cooldown {
		log.Printf("schedule %d (%s) skipped: cooldown %v remaining", s.ID, s.Name, w.cooldown-elapsed)
		if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     s.FarmID,
			ScheduleID: &s.ID,
			Status:     "skipped",
			Message:    ptr(fmt.Sprintf("cooldown: %v since last success, requires %v", elapsed.Truncate(time.Second), w.cooldown)),
			Details:    []byte(`{"phase":"cooldown"}`),
			ExecutedAt: now,
		}); err != nil {
			log.Printf("failed to record automation run: %v", err)
		}
		return true
	}
	return false
}

// checkPreconditions fetches the latest reading for every sensor listed in
// schedule.Preconditions and verifies its predicate holds. On any failure
// it records an automation_runs row (status='skipped',
// message='precondition_failed') and returns true so the caller aborts
// execution. Missing readings are treated as a failure — the operator asked
// for an interlock and we can't assert safety without data.
//
// The underlying evaluator (EvaluatePredicates) is shared with the Phase 20
// rule evaluator in rules.go; schedule preconditions always use LogicAll.
func (w *Worker) checkPreconditions(ctx context.Context, s db.Gr33ncoreSchedule, now time.Time) bool {
	if len(s.Preconditions) == 0 {
		return false
	}
	var preds []Predicate
	if err := json.Unmarshal(s.Preconditions, &preds); err != nil {
		log.Printf("schedule %d: invalid preconditions JSON, skipping execution: %v", s.ID, err)
		details, _ := json.Marshal(map[string]any{
			"phase": "preconditions",
			"error": "invalid_preconditions_json",
		})
		if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     s.FarmID,
			ScheduleID: &s.ID,
			Status:     "skipped",
			Message:    ptr("precondition_failed"),
			Details:    details,
			ExecutedAt: now,
		}); err != nil {
			log.Printf("failed to record automation run: %v", err)
		}
		return true
	}
	if len(preds) == 0 {
		return false
	}

	passed, failed := EvaluatePredicates(ctx, w.q, LogicAll, preds)
	if passed {
		return false
	}

	details, _ := json.Marshal(map[string]any{
		"phase":  "preconditions",
		"failed": failed,
	})
	log.Printf("schedule %d (%s) skipped: %d precondition(s) failed", s.ID, s.Name, len(failed))
	if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
		FarmID:     s.FarmID,
		ScheduleID: &s.ID,
		Status:     "skipped",
		Message:    ptr("precondition_failed"),
		Details:    details,
		ExecutedAt: now,
	}); err != nil {
		log.Printf("failed to record automation run: %v", err)
	}
	return true
}

func idempotencyKey(scheduleID int64, now time.Time) string {
	raw := fmt.Sprintf("%d:%s", scheduleID, now.Format("2006-01-02T15:04"))
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h[:8])
}

// checkIdempotency prevents duplicate execution for the same schedule+minute.
func (w *Worker) checkIdempotency(ctx context.Context, s db.Gr33ncoreSchedule, key string, now time.Time) bool {
	detailsJSON, _ := json.Marshal(map[string]string{"idempotency_key": key})
	_, err := w.q.GetAutomationRunByDetails(ctx, db.GetAutomationRunByDetailsParams{
		ScheduleID: &s.ID,
		Column2:    detailsJSON,
	})
	if err == nil {
		log.Printf("schedule %d (%s) skipped: idempotent run already exists (key=%s)", s.ID, s.Name, key)
		return true
	}
	return false
}

func (w *Worker) executeSchedule(ctx context.Context, s db.Gr33ncoreSchedule, now time.Time, idemKey string) {
	actions, err := w.q.ListExecutableActionsBySchedule(ctx, &s.ID)
	if err != nil {
		if _, runErr := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     s.FarmID,
			ScheduleID: &s.ID,
			Status:     "failed",
			Message:    ptr(fmt.Sprintf("failed to list actions: %v", err)),
			Details:    []byte(`{"phase":"list_actions"}`),
			ExecutedAt: now,
		}); runErr != nil {
			log.Printf("failed to record automation run: %v", runErr)
		}
		return
	}

	if len(actions) == 0 {
		if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     s.FarmID,
			ScheduleID: &s.ID,
			Status:     "skipped",
			Message:    ptr("schedule has no executable actions"),
			Details:    []byte(`{"phase":"execute","actions":0}`),
			ExecutedAt: now,
		}); err != nil {
			log.Printf("failed to record automation run: %v", err)
		}
		if _, err := w.q.MarkScheduleTriggered(ctx, db.MarkScheduleTriggeredParams{
			ID: s.ID,
			LastTriggeredTime: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
		}); err != nil {
			log.Printf("failed to mark schedule triggered: %v", err)
		}
		return
	}

	successCount := 0
	errorMessages := []string{}
	for _, action := range actions {
		if err := w.executeActionWithRetry(ctx, s, action, now); err != nil {
			errorMessages = append(errorMessages, err.Error())
		} else {
			successCount++
		}
	}

	status := "success"
	if successCount == 0 && len(errorMessages) > 0 {
		status = "failed"
	} else if len(errorMessages) > 0 {
		status = "partial_success"
	}

	details, _ := json.Marshal(map[string]any{
		"actions_total":   len(actions),
		"actions_success": successCount,
		"actions_failed":  len(errorMessages),
		"simulation_mode": w.simulation,
		"idempotency_key": idemKey,
		"errors":          errorMessages,
	})

	msg := fmt.Sprintf("executed %d/%d actions", successCount, len(actions))
	if len(errorMessages) > 0 {
		msg = msg + ": " + strings.Join(errorMessages, " | ")
	}

	if _, err := w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
		FarmID:     s.FarmID,
		ScheduleID: &s.ID,
		Status:     status,
		Message:    ptr(msg),
		Details:    details,
		ExecutedAt: now,
	}); err != nil {
		log.Printf("failed to record automation run: %v", err)
	}

	if _, err := w.q.MarkScheduleTriggered(ctx, db.MarkScheduleTriggeredParams{
		ID: s.ID,
		LastTriggeredTime: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	}); err != nil {
		log.Printf("failed to mark schedule triggered: %v", err)
	}
}

// executeActionWithRetry wraps executeAction with retries for transient errors.
func (w *Worker) executeActionWithRetry(ctx context.Context, schedule db.Gr33ncoreSchedule, action db.Gr33ncoreExecutableAction, now time.Time) error {
	var lastErr error
	for attempt := range w.maxRetries + 1 {
		lastErr = w.executeAction(ctx, schedule, action, now)
		if lastErr == nil {
			return nil
		}
		if !isTransient(lastErr) {
			return fmt.Errorf("[permanent] %w", lastErr)
		}
		if attempt < w.maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * 500 * time.Millisecond
			log.Printf("action %d transient error (attempt %d/%d), retrying in %v: %v",
				action.ID, attempt+1, w.maxRetries+1, backoff, lastErr)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			}
		}
	}
	return fmt.Errorf("[transient after %d retries] %w", w.maxRetries+1, lastErr)
}

// isTransient classifies errors as retryable (connection, timeout) vs permanent (bad config, missing target).
func isTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	msg := err.Error()
	transientPatterns := []string{
		"connection refused", "connection reset", "broken pipe",
		"timeout", "temporarily unavailable", "too many connections",
		"pgconn", "conn closed",
	}
	for _, p := range transientPatterns {
		if strings.Contains(strings.ToLower(msg), p) {
			return true
		}
	}
	// pgx ErrNoRows is not transient — it's a missing record
	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}
	return false
}

func (w *Worker) executeAction(ctx context.Context, schedule db.Gr33ncoreSchedule, action db.Gr33ncoreExecutableAction, now time.Time) error {
	switch string(action.ActionType) {
	case "control_actuator":
		if action.TargetActuatorID == nil {
			return fmt.Errorf("action %d missing target_actuator_id", action.ID)
		}
		command := "toggle"
		if action.ActionCommand != nil && *action.ActionCommand != "" {
			command = *action.ActionCommand
		}
		stateText := command
		if command == "on" {
			stateText = "online"
		} else if command == "off" {
			stateText = "offline"
		}
		if w.simulation {
			var numeric pgtype.Numeric
			_ = numeric.Scan(0)
			if _, err := w.q.UpdateActuatorState(ctx, db.UpdateActuatorStateParams{
				ID:                  *action.TargetActuatorID,
				CurrentStateNumeric: numeric,
				CurrentStateText:    &stateText,
			}); err != nil {
				log.Printf("failed to update actuator state: %v", err)
			}
		}
		params, _ := json.Marshal(map[string]any{
			"command":         command,
			"simulation_mode": w.simulation,
			"schedule_name":   schedule.Name,
		})
		status := db.Gr33ncoreActuatorExecutionStatusEnumPendingConfirmationFromFeedback
		if w.simulation {
			status = db.Gr33ncoreActuatorExecutionStatusEnumExecutionCompletedSuccessOnDevice
		}
		source := db.Gr33ncoreActuatorEventSourceEnumScheduleTrigger
		_, err := w.q.InsertActuatorEvent(ctx, db.InsertActuatorEventParams{
			EventTime:             now,
			ActuatorID:            *action.TargetActuatorID,
			CommandSent:           ptr(command),
			ParametersSent:        params,
			TriggeredByUserID:     pgtype.UUID{},
			TriggeredByScheduleID: &schedule.ID,
			TriggeredByRuleID:     nil,
			Source:                source,
			ExecutionStatus: db.NullGr33ncoreActuatorExecutionStatusEnum{
				Gr33ncoreActuatorExecutionStatusEnum: status,
				Valid:                                true,
			},
			MetaData: []byte(`{}`),
		})
		if err != nil {
			return err
		}

		if !w.simulation {
			actuator, lookupErr := w.q.GetActuatorByID(ctx, *action.TargetActuatorID)
			if lookupErr == nil && actuator.DeviceID != nil {
				pendingJSON, _ := json.Marshal(map[string]any{
					"command":     command,
					"schedule_id": schedule.ID,
				})
				_ = w.q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{
					ID:      *actuator.DeviceID,
					Column2: pendingJSON,
				})
			}
		}
		return nil

	case "update_record_in_gr33n":
		if len(action.ActionParameters) == 0 {
			return fmt.Errorf("action %d missing action_parameters", action.ID)
		}
		var payload map[string]any
		if err := json.Unmarshal(action.ActionParameters, &payload); err != nil {
			return fmt.Errorf("action %d has invalid action_parameters json", action.ID)
		}
		module, _ := payload["target_module_schema"].(string)
		table, _ := payload["target_table_name"].(string)
		if module != "gr33nfertigation" || table != "fertigation_events" {
			return fmt.Errorf("action %d unsupported target %s.%s", action.ID, module, table)
		}
		zoneID, err := toInt64(payload["zone_id"])
		if err != nil {
			return fmt.Errorf("action %d missing valid zone_id", action.ID)
		}
		volume := toFloat64(payload["volume_applied_liters"], 0)
		ecBefore := toFloat64(payload["ec_before_mscm"], 0)
		ecAfter := toFloat64(payload["ec_after_mscm"], 0)
		phBefore := toFloat64(payload["ph_before"], 6)
		phAfter := toFloat64(payload["ph_after"], 6)

		volN, _ := numericFromFloat(volume)
		ecBeforeN, _ := numericFromFloat(ecBefore)
		ecAfterN, _ := numericFromFloat(ecAfter)
		phBeforeN, _ := numericFromFloat(phBefore)
		phAfterN, _ := numericFromFloat(phAfter)

		trigger := db.NullGr33nfertigationProgramTriggerEnum{
			Gr33nfertigationProgramTriggerEnum: db.Gr33nfertigationProgramTriggerEnumScheduleCron,
			Valid:                              true,
		}
		_, err = w.q.CreateFertigationEvent(ctx, db.CreateFertigationEventParams{
			FarmID:              schedule.FarmID,
			ProgramID:           nil,
			ReservoirID:         nil,
			ZoneID:              zoneID,
			AppliedAt:           now,
			GrowthStage:         db.NullGr33nfertigationGrowthStageEnum{},
			VolumeAppliedLiters: volN,
			RunDurationSeconds:  nil,
			EcBeforeMscm:        ecBeforeN,
			EcAfterMscm:         ecAfterN,
			PhBefore:            phBeforeN,
			PhAfter:             phAfterN,
			TriggerSource:       trigger,
			Notes:               ptr("fertigation event created by automation worker"),
			Metadata:            []byte(`{"source":"automation_worker"}`),
		})
		return err

	default:
		return fmt.Errorf("action %d unsupported action_type=%s", action.ID, action.ActionType)
	}
}

func numericFromFloat(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(strconv.FormatFloat(v, 'f', -1, 64))
	return n, err
}

func toFloat64(v any, fallback float64) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return fallback
	}
}

func toInt64(v any) (int64, error) {
	switch val := v.(type) {
	case float64:
		return int64(val), nil
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	default:
		return 0, fmt.Errorf("not an integer")
	}
}

func ptr[T any](v T) *T { return &v }
