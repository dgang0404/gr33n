package automation

// Phase 20.7 WS4 — nightly electricity rollup. For each farm with a
// configured $/kWh and each actuator with `watts > 0`, we reconstruct
// on-intervals from `gr33ncore.actuator_events` across the target day
// and write a single cost_transactions row per (actuator, date).
//
// Design notes:
//
//   - We reconstruct intervals from `command_sent` alone ("on" vs
//     "off"). Degenerate ON→ON or OFF→OFF pairs are logged and treated
//     as no-ops — they don't fail the rollup, which matters when the Pi
//     reconfirms state after a reboot.
//   - If the last event before the window began with "on", the window
//     opens in the ON state. Symmetrically, if the window ends in ON,
//     the runtime is capped at window end (the *next* day's rollup
//     picks up the rest).
//   - Idempotency key: "electricity:<actuator_id>:<YYYY-MM-DD>". A
//     pre-check against gr33ncore.cost_transaction_idempotency short-
//     circuits repeat invocations so the smoke tests can call Tick
//     twice safely.
//   - No farm_energy_prices row → silent skip for that farm. An
//     inline hint on the Costs page (WS6) tells operators to set a
//     rate to enable this loop.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

// TickElectricityRollup walks every farm and writes one
// cost_transactions row per (actuator, date) for actuators with
// watts > 0 and an active farm_energy_prices row. Idempotent.
//
// `date` is interpreted as a UTC calendar day. Multi-timezone farms
// are out-of-scope until a later phase (see 20.7 plan §Risks).
func (w *Worker) TickElectricityRollup(ctx context.Context, date time.Time) {
	farms, err := w.q.ListAllFarms(ctx)
	if err != nil {
		log.Printf("electricity rollup: list farms: %v", err)
		return
	}
	for _, f := range farms {
		if err := w.rollupFarmElectricity(ctx, f, date); err != nil {
			log.Printf("electricity rollup: farm %d: %v", f.ID, err)
		}
	}
}

// rollupFarmElectricity is factored out so the smoke test can target
// a single farm without scanning every row in the integration DB.
func (w *Worker) rollupFarmElectricity(ctx context.Context, farm db.Gr33ncoreFarm, date time.Time) error {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	price, err := w.q.GetActiveFarmEnergyPrice(ctx, db.GetActiveFarmEnergyPriceParams{
		FarmID: farm.ID,
		// Comparing against the last instant of the day covers the
		// tail of it while still respecting effective_to > arg semantics.
		EffectiveFrom: pgtype.Date{Time: dayStart, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("get farm energy price: %w", err)
	}
	priceFloat, ok := numericToFloat64(price.PricePerKwh)
	if !ok || priceFloat <= 0 {
		return nil
	}

	actuators, err := w.q.ListBillableActuatorsByFarm(ctx, farm.ID)
	if err != nil {
		return fmt.Errorf("list billable actuators: %w", err)
	}
	for _, a := range actuators {
		if err := w.rollupActuatorElectricity(ctx, farm.ID, a, dayStart, dayEnd, priceFloat, price.Currency); err != nil {
			log.Printf("electricity rollup: actuator %d: %v", a.ID, err)
		}
	}
	return nil
}

func (w *Worker) rollupActuatorElectricity(
	ctx context.Context,
	farmID int64,
	a db.ListBillableActuatorsByFarmRow,
	windowStart time.Time,
	windowEnd time.Time,
	pricePerKwh float64,
	currency string,
) error {
	key := fmt.Sprintf("electricity:%d:%s", a.ID, windowStart.Format("2006-01-02"))
	if existing, err := w.q.GetCostTransactionByIdempotencyKey(ctx, db.GetCostTransactionByIdempotencyKeyParams{
		FarmID:         farmID,
		IdempotencyKey: key,
	}); err == nil {
		_ = existing
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check idempotency: %w", err)
	}

	watts, ok := numericToFloat64(a.Watts)
	if !ok || watts <= 0 {
		return nil
	}

	// Initial state = last event before the window (if it was 'on',
	// the actuator enters the day running).
	initialOn := false
	var lastOnAt time.Time
	prev, err := w.q.GetLastActuatorEventBefore(ctx, db.GetLastActuatorEventBeforeParams{
		ActuatorID: a.ID,
		EventTime:  windowStart,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("get prior event: %w", err)
	} else if err == nil && isOnCommand(prev.CommandSent) {
		initialOn = true
		lastOnAt = windowStart
	}

	events, err := w.q.ListActuatorEventsForRollup(ctx, db.ListActuatorEventsForRollupParams{
		ActuatorID: a.ID,
		EventTime:  windowStart,
		EventTime_2: windowEnd,
	})
	if err != nil {
		return fmt.Errorf("list actuator events: %w", err)
	}

	isOn := initialOn
	var onDuration time.Duration
	for _, e := range events {
		cmdOn := isOnCommand(e.CommandSent)
		cmdOff := isOffCommand(e.CommandSent)
		switch {
		case cmdOn && !isOn:
			isOn = true
			lastOnAt = e.EventTime
		case cmdOff && isOn:
			onDuration += e.EventTime.Sub(lastOnAt)
			isOn = false
		case cmdOn && isOn, cmdOff && !isOn:
			// noise: degenerate transition, log once and skip
			log.Printf("electricity rollup: actuator %d degenerate %s at %s", a.ID, safeStr(e.CommandSent), e.EventTime.Format(time.RFC3339))
		}
	}
	if isOn {
		onDuration += windowEnd.Sub(lastOnAt)
	}

	if onDuration <= 0 {
		return nil
	}

	hours := onDuration.Hours()
	kwh := watts * hours / 1000.0
	cost := kwh * pricePerKwh

	amountN, err := numericFromFloat(cost)
	if err != nil {
		return fmt.Errorf("encode amount: %w", err)
	}
	desc := fmt.Sprintf("Electricity: %s ran %s × %.0fW @ %.4f/kWh (%.3f kWh)",
		a.Name, formatHoursMinutes(onDuration), watts, pricePerKwh, kwh)
	relSchema := "gr33ncore"
	relTable := "actuators"
	category := commontypes.CostCategoryUtilitiesElectricityGas

	tx, err := w.q.CreateCostTransactionAutoLogged(ctx, db.CreateCostTransactionAutoLoggedParams{
		FarmID:              farmID,
		TransactionDate:     pgtype.Date{Time: windowStart, Valid: true},
		Category:            category,
		Amount:              amountN,
		Currency:            currency,
		Description:         &desc,
		RelatedModuleSchema: &relSchema,
		RelatedTableName:    &relTable,
		RelatedRecordID:     &a.ID,
	})
	if err != nil {
		return fmt.Errorf("insert cost_transaction: %w", err)
	}
	if err := w.q.CreateCostTransactionIdempotency(ctx, db.CreateCostTransactionIdempotencyParams{
		FarmID:            farmID,
		IdempotencyKey:    key,
		CostTransactionID: tx.ID,
	}); err != nil {
		return fmt.Errorf("insert idempotency row: %w", err)
	}
	return nil
}

// --- tiny helpers -----------------------------------------------------------

func isOnCommand(cmd *string) bool {
	if cmd == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(*cmd)) {
	case "on", "online", "true", "1", "open":
		return true
	}
	return false
}

func isOffCommand(cmd *string) bool {
	if cmd == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(*cmd)) {
	case "off", "offline", "false", "0", "close", "closed":
		return true
	}
	return false
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatHoursMinutes(d time.Duration) string {
	total := int64(d.Minutes())
	h := total / 60
	m := total % 60
	return fmt.Sprintf("%dh%02dm", h, m)
}
