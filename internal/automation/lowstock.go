package automation

// Phase 20.7 WS5 — low-stock alert sweep. Every tick we list every
// input_batch on every farm whose current_quantity_remaining dropped
// below its (opt-in) low_stock_threshold, and fire one alert per
// batch per day. Severity is fixed to `medium`; the existing
// Phase 19 WS3 "create task from alert" flow converts it into a
// refill task with a single click.
//
// Idempotency: we consult GetLatestAlertCreatedAtForSource before
// inserting; if the most recent alert on (farm, "inventory_low_stock",
// batch_id) is still same-UTC-day, we skip.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

const lowStockSourceType = "inventory_low_stock"

// TickLowStockAlerts scans every farm for batches below their
// configured threshold and emits alerts via the same pipeline the
// rule engine uses. Exported so integration tests can exercise it
// deterministically (the scheduled tick is wired in Start()).
func (w *Worker) TickLowStockAlerts(ctx context.Context) {
	farms, err := w.q.ListAllFarms(ctx)
	if err != nil {
		log.Printf("low-stock tick: list farms: %v", err)
		return
	}
	today := time.Now().UTC()
	for _, f := range farms {
		if err := w.sweepLowStockForFarm(ctx, f, today); err != nil {
			log.Printf("low-stock tick: farm %d: %v", f.ID, err)
		}
	}
}

func (w *Worker) sweepLowStockForFarm(ctx context.Context, farm db.Gr33ncoreFarm, now time.Time) error {
	rows, err := w.q.ListLowStockBatchesByFarm(ctx, farm.ID)
	if err != nil {
		return fmt.Errorf("list low-stock batches: %w", err)
	}
	for _, row := range rows {
		if err := w.maybeFireLowStock(ctx, row, now); err != nil {
			log.Printf("low-stock tick: batch %d: %v", row.ID, err)
		}
	}
	return nil
}

func (w *Worker) maybeFireLowStock(ctx context.Context, b db.ListLowStockBatchesByFarmRow, now time.Time) error {
	// Per-batch-per-day dedupe.
	last, err := w.q.GetLatestAlertCreatedAtForSource(ctx, db.GetLatestAlertCreatedAtForSourceParams{
		FarmID:                    b.FarmID,
		TriggeringEventSourceType: ptrStringLocal(lowStockSourceType),
		TriggeringEventSourceID:   &b.ID,
	})
	if err == nil {
		if sameUTCDate(last, now) {
			return nil
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check recent alert: %w", err)
	}

	remaining, _ := numericToFloat64(b.CurrentQuantityRemaining)
	threshold, _ := numericToFloat64(b.LowStockThreshold)
	name := b.InputName
	subject := fmt.Sprintf("Inventory low: %s at %.2f (threshold %.2f)", name, remaining, threshold)
	body := fmt.Sprintf("Batch %d of %s has dropped below its low-stock threshold. Remaining: %.2f, threshold: %.2f.",
		b.ID, name, remaining, threshold)
	severity := db.NullGr33ncoreNotificationPriorityEnum{
		Gr33ncoreNotificationPriorityEnum: db.Gr33ncoreNotificationPriorityEnumMedium,
		Valid:                             true,
	}
	sourceType := lowStockSourceType
	sourceID := b.ID

	alert, err := w.q.CreateAlert(ctx, db.CreateAlertParams{
		FarmID:                    b.FarmID,
		RecipientUserID:           pgtype.UUID{},
		TriggeringEventSourceType: &sourceType,
		TriggeringEventSourceID:   &sourceID,
		Severity:                  severity,
		SubjectRendered:           &subject,
		MessageTextRendered:       &body,
	})
	if err != nil {
		return fmt.Errorf("insert alert: %w", err)
	}
	if w.notifier != nil {
		w.notifier.DispatchFarmAlert(ctx, alert)
	}
	return nil
}

func sameUTCDate(a, b time.Time) bool {
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}

func ptrStringLocal(s string) *string { return &s }
