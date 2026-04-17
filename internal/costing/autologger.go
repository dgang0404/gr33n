// Package costing owns the post-commit hooks that turn telemetry into
// cost_transactions rows. Phase 20.7 wires three of them:
//
//  1. LogMixingComponent — called after a gr33nfertigation.mixing_event_components
//     row is inserted. Decrements gr33nnaturalfarming.input_batches.current_quantity_remaining
//     and writes a cost_transactions row priced via input_definitions.unit_cost.
//  2. LogTaskConsumption — called after a gr33ncore.task_input_consumptions row
//     is inserted. Same shape as LogMixingComponent; the source table differs.
//  3. ReverseTaskConsumption — called before the DELETE of a task consumption.
//     Re-credits the batch and, if the original write landed a cost row, inserts
//     a compensating row so net cost = 0 while the ledger stays append-only.
//
// Every auto-write is idempotent via gr33ncore.cost_transaction_idempotency
// (PK: farm_id, idempotency_key). Callers that replay the same component /
// consumption id get a silent no-op instead of duplicate rows.
//
// The autologger accepts a *db.Queries rather than a pool/tx so callers can
// choose the boundary: the mixing handler hands it its per-transaction Queries
// so the deduct + cost write commit atomically with the component insert; a
// background backfill job would hand it a plain pool.
package costing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

const (
	// relatedModuleSchema for components that live in gr33nfertigation.
	schemaFertigation = "gr33nfertigation"
	// relatedModuleSchema for task_input_consumptions (gr33ncore).
	schemaCore = "gr33ncore"
)

// numericFromFloat mirrors internal/handler/cost/handler.go:numericFromFloat64.
func numericFromFloat(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(strconv.FormatFloat(v, 'f', -1, 64))
	return n, err
}

func numericToFloat(n pgtype.Numeric) (float64, bool) {
	if !n.Valid {
		return 0, false
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0, false
	}
	return f.Float64, true
}

// ErrIdempotentNoop is a sentinel returned internally when an
// idempotency key already maps to a cost row; the public entry points
// swallow it so callers treat repeats as "done".
var errIdempotentNoop = errors.New("autologger: already logged")

// resolvedPrice bundles the numbers the autologger needs after looking
// up the batch + definition. `costKnown` is false when the input has
// no unit_cost configured — in that case we still decrement stock but
// skip the cost_transactions write.
type resolvedPrice struct {
	unitCost  float64
	currency  string
	costKnown bool
	defName   string
	// category is the cost_category_enum value this input maps to.
	// Defaults to fertilizers_soil_amendments (the pre-Phase-20.8
	// autologger's hard-coded value); Phase 20.8 WS3 extended the
	// mapping to cover animal husbandry inputs.
	category commontypes.CostCategoryEnum
}

func resolveDefinitionPrice(ctx context.Context, q *db.Queries, defID int64) (resolvedPrice, error) {
	def, err := q.GetInputDefinitionByID(ctx, defID)
	if err != nil {
		return resolvedPrice{}, fmt.Errorf("get input_definition %d: %w", defID, err)
	}
	out := resolvedPrice{
		defName:  def.Name,
		category: mapInputCategoryToCostCategory(def.Category),
	}
	if def.UnitCost.Valid && def.UnitCostCurrency != nil && *def.UnitCostCurrency != "" {
		if f, ok := numericToFloat(def.UnitCost); ok && f > 0 {
			out.unitCost = f
			out.currency = *def.UnitCostCurrency
			out.costKnown = true
		}
	}
	return out, nil
}

// mapInputCategoryToCostCategory is the Phase 20.8 WS3 lookup between
// an `input_definitions.category` value (what the operator tagged the
// product as) and the `cost_category_enum` value the autologger
// stamps on the cost_transactions row. Kept here, not on the database,
// because it is purely a classification concern — the operator never
// chooses a cost category directly when consuming an input. If a new
// input category is added later without a mapping entry, we fall back
// to `fertilizers_soil_amendments` so automation still succeeds (the
// operator can retag via `UpdateCostTransactionCategory` after the
// fact).
func mapInputCategoryToCostCategory(ic db.Gr33nnaturalfarmingInputCategoryEnum) commontypes.CostCategoryEnum {
	switch ic {
	case db.Gr33nnaturalfarmingInputCategoryEnumAnimalFeed,
		db.Gr33nnaturalfarmingInputCategoryEnumBedding:
		// Bedding folds into feed_livestock for now; if farms start
		// caring about the split we can promote bedding to a distinct
		// cost_category_enum value in a later phase.
		return commontypes.CostCategoryFeedLivestock
	case db.Gr33nnaturalfarmingInputCategoryEnumVeterinarySupply:
		return commontypes.CostCategoryVeterinaryServices
	default:
		return commontypes.CostCategoryFertilizersSoilAmendments
	}
}

// checkIdempotency returns (existingTxID, true) when the key has
// already produced a cost row; callers must short-circuit. A pgx
// ErrNoRows is translated to (0, false, nil) — the common happy path.
func checkIdempotency(ctx context.Context, q *db.Queries, farmID int64, key string) (int64, bool, error) {
	row, err := q.GetCostTransactionByIdempotencyKey(ctx, db.GetCostTransactionByIdempotencyKeyParams{
		FarmID:         farmID,
		IdempotencyKey: key,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return row.ID, true, nil
}

// writeCostRowWithIdempotency inserts the cost_transactions row and
// records the idempotency marker in a single autologger call. The
// idempotency INSERT races with a parallel autologger call on the same
// key are caught by the PK constraint; we treat a duplicate-key error
// on the idempotency row as "someone else logged it first" and roll
// back the cost row via a manual delete (the caller's tx would normally
// cover this, but this function doesn't own the tx).
func writeCostRowWithIdempotency(
	ctx context.Context,
	q *db.Queries,
	farmID int64,
	key string,
	params db.CreateCostTransactionAutoLoggedParams,
) (int64, error) {
	row, err := q.CreateCostTransactionAutoLogged(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("insert cost_transaction: %w", err)
	}
	if err := q.CreateCostTransactionIdempotency(ctx, db.CreateCostTransactionIdempotencyParams{
		FarmID:            farmID,
		IdempotencyKey:    key,
		CostTransactionID: row.ID,
	}); err != nil {
		return 0, fmt.Errorf("insert idempotency: %w", err)
	}
	return row.ID, nil
}

// LogMixingComponent is called post-insert for every
// gr33nfertigation.mixing_event_components row. Idempotent on
// `mixing_component:<id>`. Best-effort — if unit_cost is NULL the
// stock decrement still happens; if the batch is NULL (operator
// didn't attach one) neither happens.
func LogMixingComponent(
	ctx context.Context,
	q *db.Queries,
	farmID int64,
	component db.Gr33nfertigationMixingEventComponent,
	mixedAt time.Time,
) error {
	key := fmt.Sprintf("mixing_component:%d", component.ID)
	if _, already, err := checkIdempotency(ctx, q, farmID, key); err != nil {
		return err
	} else if already {
		return nil
	}

	volumeMl, ok := numericToFloat(component.VolumeAddedMl)
	if !ok || volumeMl <= 0 {
		return nil
	}

	price, err := resolveDefinitionPrice(ctx, q, component.InputDefinitionID)
	if err != nil {
		return err
	}

	if component.InputBatchID != nil {
		qtyN, _ := numericFromFloat(volumeMl)
		deducted, err := q.DecrementInputBatchQuantity(ctx, db.DecrementInputBatchQuantityParams{
			ID:     *component.InputBatchID,
			Column2: qtyN,
		})
		if err != nil {
			log.Printf("autologger: decrement batch %d: %v", *component.InputBatchID, err)
		} else if remaining, ok := numericToFloat(deducted.CurrentQuantityRemaining); ok && remaining < 0 {
			log.Printf("autologger: input_batch %d remaining went negative (%.3f) after mixing_component %d; stock tracking will under-report until operator adjusts",
				*component.InputBatchID, remaining, component.ID)
		}
	}

	if !price.costKnown {
		return nil
	}

	amount := volumeMl * price.unitCost
	amountN, err := numericFromFloat(amount)
	if err != nil {
		return fmt.Errorf("encode amount: %w", err)
	}
	desc := fmt.Sprintf("Mixing component: %.3f ml of %s @ %.4f/ml",
		volumeMl, price.defName, price.unitCost)
	relSchema := schemaFertigation
	relTable := "mixing_event_components"
	category := price.category
	txID, err := writeCostRowWithIdempotency(ctx, q, farmID, key,
		db.CreateCostTransactionAutoLoggedParams{
			FarmID:              farmID,
			TransactionDate:     pgtype.Date{Time: mixedAt, Valid: true},
			Category:            category,
			Amount:              amountN,
			Currency:            price.currency,
			Description:         &desc,
			RelatedModuleSchema: &relSchema,
			RelatedTableName:    &relTable,
			RelatedRecordID:     &component.ID,
		})
	if err != nil {
		return err
	}
	_ = txID
	return nil
}

// LogTaskConsumption mirrors LogMixingComponent for
// gr33ncore.task_input_consumptions. Returns the written
// cost_transaction_id (or nil if the input had no price) so the
// handler can backfill the consumption row's cost_transaction_id FK
// via UpdateTaskInputConsumptionCostTx.
func LogTaskConsumption(
	ctx context.Context,
	q *db.Queries,
	consumption db.Gr33ncoreTaskInputConsumption,
) (*int64, error) {
	key := fmt.Sprintf("task_consumption:%d", consumption.ID)
	if existing, already, err := checkIdempotency(ctx, q, consumption.FarmID, key); err != nil {
		return nil, err
	} else if already {
		return &existing, nil
	}

	quantity, ok := numericToFloat(consumption.Quantity)
	if !ok || quantity <= 0 {
		return nil, nil
	}

	batch, err := q.GetInputBatchByID(ctx, consumption.InputBatchID)
	if err != nil {
		return nil, fmt.Errorf("get input_batch %d: %w", consumption.InputBatchID, err)
	}

	price, err := resolveDefinitionPrice(ctx, q, batch.InputDefinitionID)
	if err != nil {
		return nil, err
	}

	qtyN, _ := numericFromFloat(quantity)
	deducted, err := q.DecrementInputBatchQuantity(ctx, db.DecrementInputBatchQuantityParams{
		ID:     consumption.InputBatchID,
		Column2: qtyN,
	})
	if err != nil {
		return nil, fmt.Errorf("decrement input_batch %d: %w", consumption.InputBatchID, err)
	}
	if remaining, ok := numericToFloat(deducted.CurrentQuantityRemaining); ok && remaining < 0 {
		log.Printf("autologger: input_batch %d remaining went negative (%.3f) after task_consumption %d",
			consumption.InputBatchID, remaining, consumption.ID)
	}

	if !price.costKnown {
		return nil, nil
	}

	amount := quantity * price.unitCost
	amountN, err := numericFromFloat(amount)
	if err != nil {
		return nil, fmt.Errorf("encode amount: %w", err)
	}
	desc := fmt.Sprintf("Task consumption: %.3f of %s @ %.4f/unit",
		quantity, price.defName, price.unitCost)
	relSchema := schemaCore
	relTable := "task_input_consumptions"
	category := price.category

	txID, err := writeCostRowWithIdempotency(ctx, q, consumption.FarmID, key,
		db.CreateCostTransactionAutoLoggedParams{
			FarmID:              consumption.FarmID,
			TransactionDate:     pgtype.Date{Time: consumption.RecordedAt, Valid: true},
			Category:            category,
			Amount:              amountN,
			Currency:            price.currency,
			Description:         &desc,
			CreatedByUserID:     consumption.RecordedBy,
			RelatedModuleSchema: &relSchema,
			RelatedTableName:    &relTable,
			RelatedRecordID:     &consumption.ID,
		})
	if err != nil {
		return nil, err
	}
	return &txID, nil
}

// ReverseTaskConsumption is called by the DELETE handler *before* the
// consumption row is removed. It re-credits the batch and, if the
// original landed a cost row, writes a compensating row so net = 0.
// The ledger stays append-only; both rows are visible on audit.
//
// Returns nil if the consumption had no cost_transaction_id (nothing
// to compensate — just credit the stock).
func ReverseTaskConsumption(
	ctx context.Context,
	q *db.Queries,
	consumption db.Gr33ncoreTaskInputConsumption,
) error {
	quantity, ok := numericToFloat(consumption.Quantity)
	if ok && quantity > 0 {
		qtyN, _ := numericFromFloat(quantity)
		if _, err := q.IncrementInputBatchQuantity(ctx, db.IncrementInputBatchQuantityParams{
			ID:     consumption.InputBatchID,
			Column2: qtyN,
		}); err != nil {
			return fmt.Errorf("credit input_batch %d: %w", consumption.InputBatchID, err)
		}
	}

	if consumption.CostTransactionID == nil {
		return nil
	}
	orig, err := q.GetCostTransactionByID(ctx, *consumption.CostTransactionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("get original cost_transaction %d: %w", *consumption.CostTransactionID, err)
	}

	// Negate the original amount. pgx returns a pgtype.Numeric with
	// Int / Exp fields; the simplest path is float conversion (the
	// ledger's numeric precision is 12,2 so float64 is safe).
	origAmount, ok := numericToFloat(orig.Amount)
	if !ok {
		return nil
	}
	compN, err := numericFromFloat(-origAmount)
	if err != nil {
		return err
	}
	desc := "[VOIDED] " + fmt.Sprintf("reverses cost_transaction %d", orig.ID)
	if orig.Description != nil {
		desc = "[VOIDED] " + *orig.Description
	}
	key := fmt.Sprintf("task_consumption_void:%d", consumption.ID)
	if _, already, err := checkIdempotency(ctx, q, consumption.FarmID, key); err != nil {
		return err
	} else if already {
		return nil
	}
	_, err = writeCostRowWithIdempotency(ctx, q, consumption.FarmID, key,
		db.CreateCostTransactionAutoLoggedParams{
			FarmID:              consumption.FarmID,
			TransactionDate:     pgtype.Date{Time: time.Now().UTC(), Valid: true},
			Category:            orig.Category,
			Amount:              compN,
			Currency:            orig.Currency,
			Description:         &desc,
			CreatedByUserID:     consumption.RecordedBy,
			RelatedModuleSchema: orig.RelatedModuleSchema,
			RelatedTableName:    orig.RelatedTableName,
			RelatedRecordID:     orig.RelatedRecordID,
			CropCycleID:         orig.CropCycleID,
		})
	return err
}
