package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

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

type Worker struct {
	q          *db.Queries
	simulation bool

	mu         sync.RWMutex
	running    bool
	lastTickAt time.Time
	lastError  string
}

func NewWorker(pool *pgxpool.Pool, simulation bool) *Worker {
	return &Worker{
		q:          db.New(pool),
		simulation: simulation,
	}
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
			_, _ = w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
				FarmID:     s.FarmID,
				ScheduleID: &s.ID,
				RuleID:     nil,
				Status:     "failed",
				Message:    ptr(fmt.Sprintf("cron parse error for %s: %v", s.Name, evalErr)),
				Details:    []byte(`{"phase":"cron_eval"}`),
				ExecutedAt: now,
			})
			continue
		}
		if !should {
			continue
		}
		w.executeSchedule(ctx, s, now)
	}
	w.setLastTick(nil)
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

func (w *Worker) executeSchedule(ctx context.Context, s db.Gr33ncoreSchedule, now time.Time) {
	actions, err := w.q.ListExecutableActionsBySchedule(ctx, &s.ID)
	if err != nil {
		_, _ = w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     s.FarmID,
			ScheduleID: &s.ID,
			Status:     "failed",
			Message:    ptr(fmt.Sprintf("failed to list actions: %v", err)),
			Details:    []byte(`{"phase":"list_actions"}`),
			ExecutedAt: now,
		})
		return
	}

	if len(actions) == 0 {
		_, _ = w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
			FarmID:     s.FarmID,
			ScheduleID: &s.ID,
			Status:     "skipped",
			Message:    ptr("schedule has no executable actions"),
			Details:    []byte(`{"phase":"execute","actions":0}`),
			ExecutedAt: now,
		})
		_, _ = w.q.MarkScheduleTriggered(ctx, db.MarkScheduleTriggeredParams{
			ID: s.ID,
			LastTriggeredTime: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
		})
		return
	}

	successCount := 0
	errorMessages := []string{}
	for _, action := range actions {
		if err := w.executeAction(ctx, s, action, now); err != nil {
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
		"errors":          errorMessages,
	})

	msg := fmt.Sprintf("executed %d/%d actions", successCount, len(actions))
	if len(errorMessages) > 0 {
		msg = msg + ": " + strings.Join(errorMessages, " | ")
	}

	_, _ = w.q.CreateAutomationRun(ctx, db.CreateAutomationRunParams{
		FarmID:     s.FarmID,
		ScheduleID: &s.ID,
		Status:     status,
		Message:    ptr(msg),
		Details:    details,
		ExecutedAt: now,
	})

	_, _ = w.q.MarkScheduleTriggered(ctx, db.MarkScheduleTriggeredParams{
		ID: s.ID,
		LastTriggeredTime: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	})
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
			_, _ = w.q.UpdateActuatorState(ctx, db.UpdateActuatorStateParams{
				ID:                  *action.TargetActuatorID,
				CurrentStateNumeric: numeric,
				CurrentStateText:    &stateText,
			})
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
			EventTime:            now,
			ActuatorID:           *action.TargetActuatorID,
			CommandSent:          ptr(command),
			ParametersSent:       params,
			TriggeredByUserID:    pgtype.UUID{},
			TriggeredByScheduleID: &schedule.ID,
			TriggeredByRuleID:     nil,
			Source:               source,
			ExecutionStatus: db.NullGr33ncoreActuatorExecutionStatusEnum{
				Gr33ncoreActuatorExecutionStatusEnum: status,
				Valid:                                true,
			},
			MetaData:             []byte(`{}`),
		})
		return err

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
