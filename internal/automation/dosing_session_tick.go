package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	acthandler "gr33n-api/internal/handler/actuator"
	"gr33n-api/internal/httputil"
)

// Phase 216 WS2 — dosing session state machine worker.
//
// Every dosing_sessions row persists its own status + next_action_at, so a
// worker restart mid-cycle resumes from wherever it was written instead of
// losing the session — that persistence is the entire reason WS1 exists as
// a table instead of in-memory state.
//
// Cycle: idle (dose an increment) -> dosing (pulse landing) -> circulating
// (homogenize) -> settling (let it stabilize) -> sampling (collect N EC
// readings) -> evaluating (average + decide) -> back to idle, or done/aborted.
const (
	// dosingSampleIntervalSeconds spaces out sampling ticks. EC sensors on the
	// LED rig poll ~60s; this just re-checks for a fresh reading periodically
	// rather than assuming one lands on a fixed schedule.
	dosingSampleIntervalSeconds = 20
	// dosingEcImprovementEpsilon — deltas smaller than this count as "no
	// improvement" for the fouled-sensor guard (WS5), not a real reading change.
	dosingEcImprovementEpsilon = 0.02
)

// TickDosingSessions scans for due, non-terminal sessions and advances each
// one exactly one state. Called from runTick alongside the other Tick* passes.
func (w *Worker) TickDosingSessions(ctx context.Context) {
	sessions, err := w.q.ListDueDosingSessions(ctx)
	if err != nil {
		slog.Warn("dosing session tick: list due sessions failed", "err", err)
		return
	}
	for _, s := range sessions {
		if err := w.advanceDosingSession(ctx, s); err != nil {
			slog.Warn("dosing session tick: advance failed",
				"dosing_session_id", s.ID, "status", string(s.Status), "err", err)
		}
	}
}

func (w *Worker) advanceDosingSession(ctx context.Context, s db.Gr33nfertigationDosingSession) error {
	switch s.Status {
	case db.Gr33nfertigationDosingSessionStatusEnumIdle:
		return w.dosingFireIncrement(ctx, s)
	case db.Gr33nfertigationDosingSessionStatusEnumDosing:
		return w.dosingStartCirculation(ctx, s)
	case db.Gr33nfertigationDosingSessionStatusEnumCirculating:
		return w.dosingBeginSettling(ctx, s)
	case db.Gr33nfertigationDosingSessionStatusEnumSettling:
		return w.dosingBeginSampling(ctx, s)
	case db.Gr33nfertigationDosingSessionStatusEnumSampling:
		return w.dosingTakeSample(ctx, s)
	case db.Gr33nfertigationDosingSessionStatusEnumEvaluating:
		return w.dosingEvaluateCycle(ctx, s)
	default:
		// done/aborted are excluded by ListDueDosingSessions; nothing to do.
		return nil
	}
}

// dosingFireIncrement (idle -> dosing): safety cap first (WS5) — if the
// budget is already spent, close the session out instead of firing another
// increment. Otherwise queue the delivery pump for however long it takes to
// add dose_increment_ml at its calibrated ml/second rate.
func (w *Worker) dosingFireIncrement(ctx context.Context, s db.Gr33nfertigationDosingSession) error {
	maxTotal, _ := numericToFloat64(s.MaxTotalDoseMl)
	totalDosed, _ := numericToFloat64(s.TotalDosedMl)
	if totalDosed >= maxTotal {
		return w.dosingComplete(ctx, s, false, "max_total_dose_ml reached before target was met")
	}

	res, err := w.q.GetFertigationReservoirByID(ctx, s.ReservoirID)
	if err != nil {
		return fmt.Errorf("load reservoir %d: %w", s.ReservoirID, err)
	}
	if res.DeliveryActuatorID == nil {
		return w.dosingAbort(ctx, s, "reservoir has no delivery_actuator_id — cannot dose")
	}

	doseIncrement, _ := numericToFloat64(s.DoseIncrementMl)
	doseMl := doseIncrement
	if remaining := maxTotal - totalDosed; remaining < doseMl {
		doseMl = remaining
	}
	if doseMl <= 0 {
		return w.dosingComplete(ctx, s, false, "no dose budget remaining")
	}
	// WS5 Guardian gate: an increment larger than 2x the calibrated normal
	// dose isn't auto-fired. Nothing in WS1/WS2 can produce doseMl > doseIncrement
	// yet (no manual-override path exists), but the guard stays here so a
	// future operator override can't silently bypass it.
	if doseIncrement > 0 && doseMl > 2*doseIncrement {
		return w.dosingAbort(ctx, s, "requested dose exceeds 2x calibrated increment — needs Guardian confirm")
	}

	actuator, err := w.q.GetActuatorByID(ctx, *res.DeliveryActuatorID)
	if err != nil {
		return fmt.Errorf("load delivery actuator %d: %w", *res.DeliveryActuatorID, err)
	}
	mlPerSecond := 1.0
	var cfg acthandler.ActuatorConfig
	if len(actuator.Config) > 0 {
		if err := json.Unmarshal(actuator.Config, &cfg); err == nil && cfg.MlPerSecond != nil && *cfg.MlPerSecond > 0 {
			mlPerSecond = *cfg.MlPerSecond
		} else {
			slog.Warn("dosing session: delivery actuator has no calibrated ml_per_second, defaulting to 1.0",
				"dosing_session_id", s.ID, "actuator_id", actuator.ID)
		}
	}
	durationSeconds := int(doseMl/mlPerSecond + 0.999) // round up — under-dosing is safer than over
	if durationSeconds < 1 {
		durationSeconds = 1
	}

	if err := w.fireDosingPulse(ctx, actuator, durationSeconds, s.ID, "dosing_increment"); err != nil {
		return fmt.Errorf("fire dose pulse: %w", err)
	}

	doseNumeric, err := httputil.NumericFromFloat64(doseMl)
	if err != nil {
		return fmt.Errorf("convert dose_ml: %w", err)
	}
	nextAt := pgtype.Timestamptz{Time: time.Now().UTC().Add(time.Duration(durationSeconds) * time.Second), Valid: true}
	_, err = w.q.RecordDosingSessionDose(ctx, db.RecordDosingSessionDoseParams{
		ID:           s.ID,
		TotalDosedMl: doseNumeric,
		Status:       db.Gr33nfertigationDosingSessionStatusEnumDosing,
		NextActionAt: nextAt,
	})
	return err
}

// dosingStartCirculation (dosing -> circulating): the dose pulse should have
// landed by now. Kick the circulation pump to homogenize before sampling.
// A reservoir without a circulation pump wired yet just skips straight to
// settling — circulation is a quality improvement, not a hard requirement.
func (w *Worker) dosingStartCirculation(ctx context.Context, s db.Gr33nfertigationDosingSession) error {
	if s.CirculationActuatorID != nil {
		actuator, err := w.q.GetActuatorByID(ctx, *s.CirculationActuatorID)
		if err != nil {
			return fmt.Errorf("load circulation actuator %d: %w", *s.CirculationActuatorID, err)
		}
		if err := w.fireDosingPulse(ctx, actuator, int(s.CirculateSeconds), s.ID, "circulate"); err != nil {
			return fmt.Errorf("fire circulation pulse: %w", err)
		}
	}
	nextAt := pgtype.Timestamptz{Time: time.Now().UTC().Add(time.Duration(s.CirculateSeconds) * time.Second), Valid: true}
	_, err := w.q.AdvanceDosingSessionState(ctx, db.AdvanceDosingSessionStateParams{
		ID: s.ID, Status: db.Gr33nfertigationDosingSessionStatusEnumCirculating, NextActionAt: nextAt,
	})
	return err
}

// dosingBeginSettling (circulating -> settling): circulation has finished;
// let the solution stabilize before trusting an EC reading.
func (w *Worker) dosingBeginSettling(ctx context.Context, s db.Gr33nfertigationDosingSession) error {
	nextAt := pgtype.Timestamptz{Time: time.Now().UTC().Add(time.Duration(s.SettleSeconds) * time.Second), Valid: true}
	_, err := w.q.AdvanceDosingSessionState(ctx, db.AdvanceDosingSessionStateParams{
		ID: s.ID, Status: db.Gr33nfertigationDosingSessionStatusEnumSettling, NextActionAt: nextAt,
	})
	return err
}

// dosingBeginSampling (settling -> sampling): no hardware action — the first
// actual sample is taken on the next sampling tick.
func (w *Worker) dosingBeginSampling(ctx context.Context, s db.Gr33nfertigationDosingSession) error {
	nextAt := pgtype.Timestamptz{Time: time.Now().UTC().Add(dosingSampleIntervalSeconds * time.Second), Valid: true}
	_, err := w.q.AdvanceDosingSessionState(ctx, db.AdvanceDosingSessionStateParams{
		ID: s.ID, Status: db.Gr33nfertigationDosingSessionStatusEnumSampling, NextActionAt: nextAt,
	})
	return err
}

// dosingTakeSample (sampling -> sampling|evaluating): reads the reservoir's
// EC sensor once, accumulates it, and either schedules another sample or
// moves on once samples_per_cycle readings are in.
func (w *Worker) dosingTakeSample(ctx context.Context, s db.Gr33nfertigationDosingSession) error {
	res, err := w.q.GetFertigationReservoirByID(ctx, s.ReservoirID)
	if err != nil {
		return fmt.Errorf("load reservoir %d: %w", s.ReservoirID, err)
	}
	if res.EcSensorID == nil {
		return w.dosingAbort(ctx, s, "reservoir has no ec_sensor_id — cannot sample")
	}
	reading, err := w.q.GetLatestReadingBySensor(ctx, *res.EcSensorID)
	if err != nil {
		return fmt.Errorf("read ec sensor %d: %w", *res.EcSensorID, err)
	}
	ecValue, ok := numericToFloat64(reading.ValueRaw)
	if !ok {
		return fmt.Errorf("ec sensor %d reading not numeric", *res.EcSensorID)
	}
	ecNumeric, err := httputil.NumericFromFloat64(ecValue)
	if err != nil {
		return fmt.Errorf("convert ec value: %w", err)
	}
	updated, err := w.q.RecordDosingSessionSample(ctx, db.RecordDosingSessionSampleParams{
		ID: s.ID, LastSampleEcMscm: ecNumeric,
	})
	if err != nil {
		return err
	}

	if updated.SamplesCollectedThisCycle < updated.SamplesPerCycle {
		nextAt := pgtype.Timestamptz{Time: time.Now().UTC().Add(dosingSampleIntervalSeconds * time.Second), Valid: true}
		_, err := w.q.AdvanceDosingSessionState(ctx, db.AdvanceDosingSessionStateParams{
			ID: s.ID, Status: db.Gr33nfertigationDosingSessionStatusEnumSampling, NextActionAt: nextAt,
		})
		return err
	}
	_, err = w.q.AdvanceDosingSessionState(ctx, db.AdvanceDosingSessionStateParams{
		ID: s.ID, Status: db.Gr33nfertigationDosingSessionStatusEnumEvaluating,
		NextActionAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
	return err
}

// dosingEvaluateCycle (evaluating -> idle|done|aborted): the decision point.
// Averages this cycle's samples, checks it against the mixing_event's
// ec_target, and either loops for another increment, finishes, or aborts
// per the WS5 safety rules (already-over-target, fouled/no-effect sensor).
func (w *Worker) dosingEvaluateCycle(ctx context.Context, s db.Gr33nfertigationDosingSession) error {
	sampleSum, _ := numericToFloat64(s.SamplesSumEcMscm)
	sampleCount := s.SamplesCollectedThisCycle
	if sampleCount <= 0 {
		return w.dosingAbort(ctx, s, "evaluating with zero samples collected")
	}
	avgEc := sampleSum / float64(sampleCount)
	prevAvg, hadPrev := numericToFloat64(s.LastSampleAvgEcMscm)

	noImprovement := s.NoImprovementCycles
	if hadPrev && avgEc-prevAvg < dosingEcImprovementEpsilon {
		noImprovement++
	} else {
		noImprovement = 0
	}

	avgNumeric, err := httputil.NumericFromFloat64(avgEc)
	if err != nil {
		return fmt.Errorf("convert avg ec: %w", err)
	}
	if _, err := w.q.EvaluateDosingSessionCycle(ctx, db.EvaluateDosingSessionCycleParams{
		ID: s.ID, LastSampleAvgEcMscm: avgNumeric, NoImprovementCycles: noImprovement,
	}); err != nil {
		return fmt.Errorf("persist evaluate cycle: %w", err)
	}

	mixEvent, err := w.q.GetMixingEventByID(ctx, s.MixingEventID)
	if err != nil {
		return fmt.Errorf("load mixing event %d: %w", s.MixingEventID, err)
	}
	if mixEvent.EcTargetID == nil {
		return w.dosingAbort(ctx, s, "mixing event has no ec_target_id — nothing to evaluate against")
	}
	target, err := w.q.GetEcTargetByID(ctx, *mixEvent.EcTargetID)
	if err != nil {
		return fmt.Errorf("load ec target %d: %w", *mixEvent.EcTargetID, err)
	}
	ecMin, _ := numericToFloat64(target.EcMinMscm)
	ecMax, _ := numericToFloat64(target.EcMaxMscm)

	switch {
	case avgEc > ecMax:
		return w.dosingAbort(ctx, s, fmt.Sprintf("ec %.3f already above target max %.3f — dosing cannot reduce EC", avgEc, ecMax))
	case avgEc >= ecMin:
		return w.dosingComplete(ctx, s, true, "")
	case noImprovement >= s.MaxNoImprovementCycles:
		return w.dosingAbort(ctx, s, fmt.Sprintf("ec did not improve for %d consecutive cycles — fouled sensor or dosing has no effect", noImprovement))
	default:
		totalDosed, _ := numericToFloat64(s.TotalDosedMl)
		maxTotal, _ := numericToFloat64(s.MaxTotalDoseMl)
		if totalDosed >= maxTotal {
			return w.dosingComplete(ctx, s, false, "max_total_dose_ml reached before target was met")
		}
		_, err := w.q.AdvanceDosingSessionState(ctx, db.AdvanceDosingSessionStateParams{
			ID: s.ID, Status: db.Gr33nfertigationDosingSessionStatusEnumIdle,
			NextActionAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		})
		return err
	}
}

// dosingComplete writes the finished reading back onto the parent
// mixing_event (the append-once record) and marks the session done.
func (w *Worker) dosingComplete(ctx context.Context, s db.Gr33nfertigationDosingSession, targetMet bool, cappedReason string) error {
	finalEc, _ := numericToFloat64(s.LastSampleAvgEcMscm)
	finalEcNumeric, _ := httputil.NumericFromFloat64(finalEc)
	if _, err := w.q.UpdateMixingEventFinal(ctx, db.UpdateMixingEventFinalParams{
		ID: s.MixingEventID, FinalEcMscm: finalEcNumeric, FinalPh: pgtype.Numeric{}, EcTargetMet: &targetMet,
	}); err != nil {
		slog.Warn("dosing session: update mixing_event final failed", "dosing_session_id", s.ID, "err", err)
	}
	if cappedReason != "" {
		slog.Info("dosing session capped without reaching target", "dosing_session_id", s.ID, "reason", cappedReason)
	}
	_, err := w.q.CompleteDosingSession(ctx, s.ID)
	return err
}

func (w *Worker) dosingAbort(ctx context.Context, s db.Gr33nfertigationDosingSession, reason string) error {
	slog.Warn("dosing session aborted", "dosing_session_id", s.ID, "reason", reason)
	_, err := w.q.AbortDosingSession(ctx, db.AbortDosingSessionParams{ID: s.ID, AbortedReason: &reason})
	return err
}

// fireDosingPulse queues a timed "on" pulse for actuatorID and logs the
// provenance on the actuator_events feed. Mirrors dispatchProgramActuator's
// simulation/non-simulation split: in full worker-simulation mode (no real
// Pi at all) the actuator state is flipped directly; otherwise the command
// goes through the normal device_commands queue for the real Pi to drain.
func (w *Worker) fireDosingPulse(ctx context.Context, actuator db.Gr33ncoreActuator, durationSeconds int, dosingSessionID int64, reason string) error {
	meta, _ := json.Marshal(map[string]any{
		"dosing_session_id": dosingSessionID,
		"reason":            reason,
		"duration_seconds":  durationSeconds,
	})
	status := db.Gr33ncoreActuatorExecutionStatusEnumPendingConfirmationFromFeedback
	if w.simulation {
		status = db.Gr33ncoreActuatorExecutionStatusEnumExecutionCompletedSuccessOnDevice
	}
	if _, err := w.q.InsertActuatorEvent(ctx, db.InsertActuatorEventParams{
		EventTime:       time.Now().UTC(),
		ActuatorID:      actuator.ID,
		CommandSent:     ptr("on"),
		ParametersSent:  meta,
		Source:          db.Gr33ncoreActuatorEventSourceEnumAutomationRuleTrigger,
		ExecutionStatus: &status,
		MetaData:        meta,
	}); err != nil {
		slog.Warn("dosing session: insert actuator event failed", "actuator_id", actuator.ID, "err", err)
	}

	if w.simulation {
		var zero pgtype.Numeric
		_ = zero.Scan(0)
		onState := "online"
		_, err := w.q.UpdateActuatorState(ctx, db.UpdateActuatorStateParams{
			ID: actuator.ID, CurrentStateNumeric: zero, CurrentStateText: &onState,
		})
		return err
	}
	if actuator.DeviceID == nil {
		return fmt.Errorf("actuator %d has no device_id", actuator.ID)
	}

	d := durationSeconds
	pendingJSON, err := acthandler.BuildPendingCommandJSONFull(acthandler.PendingCommandInput{
		ActuatorID: actuator.ID, Command: "on", Source: "automation", Reason: reason, DurationSeconds: &d,
	})
	if err != nil {
		return fmt.Errorf("build pending command: %w", err)
	}
	actuatorID := actuator.ID
	_, err = w.q.EnqueueDeviceCommand(ctx, db.EnqueueDeviceCommandParams{
		DeviceID: *actuator.DeviceID, FarmID: actuator.FarmID, CommandType: "pulse",
		Payload: pendingJSON, Source: "automation", ActuatorID: &actuatorID,
	})
	return err
}
