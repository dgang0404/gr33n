-- ============================================================
-- Queries: gr33ncore.cost_transactions
-- ============================================================

-- name: CreateCostTransaction :one
INSERT INTO gr33ncore.cost_transactions (
    farm_id, transaction_date, category, subcategory, amount, currency,
    description, is_income, created_by_user_id, receipt_file_id,
    document_type, document_reference, counterparty, crop_cycle_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: ListCostTransactionsByFarm :many
SELECT * FROM gr33ncore.cost_transactions
WHERE farm_id = $1
ORDER BY transaction_date DESC, id DESC
LIMIT $2 OFFSET $3;

-- Cursor batch for RAG ingest (stable id order).
-- name: ListCostTransactionsByFarmAfterID :many
SELECT * FROM gr33ncore.cost_transactions
WHERE farm_id = $1 AND id > $2
ORDER BY id ASC
LIMIT $3;

-- Incremental RAG ingest by updated_at (first page).
-- name: ListCostTransactionsByFarmUpdatedAfterFirst :many
SELECT * FROM gr33ncore.cost_transactions
WHERE farm_id = sqlc.arg('farm_id') AND updated_at > sqlc.arg('since')::timestamptz
ORDER BY updated_at ASC, id ASC
LIMIT sqlc.arg('limit');

-- Subsequent pages keyed by (updated_at, id).
-- name: ListCostTransactionsByFarmUpdatedAfterNext :many
SELECT * FROM gr33ncore.cost_transactions
WHERE farm_id = sqlc.arg('farm_id')
  AND (
    updated_at > sqlc.arg('cursor_updated_at')::timestamptz
    OR (updated_at = sqlc.arg('cursor_updated_at')::timestamptz AND id > sqlc.arg('cursor_id'))
  )
ORDER BY updated_at ASC, id ASC
LIMIT sqlc.arg('limit');

-- name: CountCostTransactionsByFarmUpdatedAfter :one
SELECT COUNT(*)::bigint FROM gr33ncore.cost_transactions
WHERE farm_id = sqlc.arg('farm_id') AND updated_at > sqlc.arg('since')::timestamptz;

-- name: CountCostTransactionsByFarm :one
SELECT COUNT(*)::bigint FROM gr33ncore.cost_transactions
WHERE farm_id = $1;

-- name: ListCostTransactionsByFarmExport :many
SELECT id, farm_id, transaction_date, category, subcategory, amount, currency,
 description, is_income, document_type, document_reference, counterparty
FROM gr33ncore.cost_transactions
WHERE farm_id = $1
ORDER BY transaction_date ASC, id ASC;

-- name: GetCostSummaryByFarm :one
SELECT
    COALESCE(SUM(CASE WHEN is_income THEN amount ELSE 0 END), 0)::numeric AS total_income,
    COALESCE(SUM(CASE WHEN NOT is_income THEN amount ELSE 0 END), 0)::numeric AS total_expenses,
    COALESCE(SUM(CASE WHEN is_income THEN amount ELSE -amount END), 0)::numeric AS net
FROM gr33ncore.cost_transactions
WHERE farm_id = $1;

-- name: GetCostCategoryTotalsByFarm :many
SELECT
    category,
    currency,
    COALESCE(SUM(CASE WHEN is_income THEN amount ELSE 0 END), 0)::numeric AS income,
    COALESCE(SUM(CASE WHEN NOT is_income THEN amount ELSE 0 END), 0)::numeric AS expense,
    COALESCE(SUM(CASE WHEN is_income THEN amount ELSE -amount END), 0)::numeric AS net,
    COUNT(*)::bigint AS tx_count
FROM gr33ncore.cost_transactions
WHERE farm_id = $1
GROUP BY category, currency
ORDER BY category ASC, currency ASC;

-- name: GetCostCategoryTotalsByFarmForYear :many
SELECT
    category,
    currency,
    COALESCE(SUM(CASE WHEN is_income THEN amount ELSE 0 END), 0)::numeric AS income,
    COALESCE(SUM(CASE WHEN NOT is_income THEN amount ELSE 0 END), 0)::numeric AS expense,
    COALESCE(SUM(CASE WHEN is_income THEN amount ELSE -amount END), 0)::numeric AS net,
    COUNT(*)::bigint AS tx_count
FROM gr33ncore.cost_transactions
WHERE farm_id = $1
  AND transaction_date >= $2::date
  AND transaction_date < $3::date
GROUP BY category, currency
ORDER BY category ASC, currency ASC;

-- name: GetCostTransactionByID :one
SELECT * FROM gr33ncore.cost_transactions WHERE id = $1;

-- name: UpdateCostTransaction :one
UPDATE gr33ncore.cost_transactions SET
    transaction_date = $2,
    category = $3,
    subcategory = $4,
    amount = $5,
    currency = $6,
    description = $7,
    is_income = $8,
    receipt_file_id = $9,
    document_type = $10,
    document_reference = $11,
    counterparty = $12,
    crop_cycle_id = $13,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCostTransaction :exec
DELETE FROM gr33ncore.cost_transactions WHERE id = $1;

-- Phase 20.7 WS2 — autologger extension. The manual operator-facing
-- path (CreateCostTransaction above) can't set related_* because
-- those fields exist solely to link an auto-generated row back at the
-- telemetry that produced it. Keeping a separate query (instead of
-- adding three optional args to the existing one) avoids churning
-- every handler call site that doesn't care.
-- name: CreateCostTransactionAutoLogged :one
INSERT INTO gr33ncore.cost_transactions (
    farm_id, transaction_date, category, subcategory, amount, currency,
    description, is_income, created_by_user_id, receipt_file_id,
    document_type, document_reference, counterparty, crop_cycle_id,
    related_module_schema, related_table_name, related_record_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
RETURNING *;

-- name: GetCostTransactionByIdempotencyKey :one
-- Used by the autologger to short-circuit if a deterministic key has
-- already produced a row. (farm_id, idempotency_key) is the PK so
-- this is a cheap point lookup.
SELECT ct.*
FROM gr33ncore.cost_transaction_idempotency idem
JOIN gr33ncore.cost_transactions ct ON ct.id = idem.cost_transaction_id
WHERE idem.farm_id = $1 AND idem.idempotency_key = $2;

-- name: CreateCostTransactionIdempotency :exec
INSERT INTO gr33ncore.cost_transaction_idempotency
    (farm_id, idempotency_key, cost_transaction_id)
VALUES ($1, $2, $3);

-- Phase 20.7 WS4 — electricity rollup needs the active $/kWh for a
-- (farm, transaction_date). Greatest effective_from <= transaction_date
-- wins; effective_to (if set) caps the row. Returns ErrNoRows when no
-- pricing has been configured — the worker treats that as a skip, not
-- a failure.
-- name: GetActiveFarmEnergyPrice :one
SELECT * FROM gr33ncore.farm_energy_prices
WHERE farm_id = $1
  AND effective_from <= $2
  AND (effective_to IS NULL OR effective_to > $2)
ORDER BY effective_from DESC
LIMIT 1;

-- Phase 20.7 WS4 — list all actuators with watts > 0. The rollup
-- iterates these per farm and sums their on-intervals. Soft-deleted
-- actuators are excluded so retired hardware doesn't keep billing.
-- name: ListBillableActuatorsByFarm :many
SELECT id, name, watts FROM gr33ncore.actuators
WHERE farm_id = $1
  AND deleted_at IS NULL
  AND watts IS NOT NULL
  AND watts > 0
ORDER BY id;

-- Phase 20.7 WS4 — pull the day's actuator_events for a given
-- (actuator, utc_date_window). The worker reconstructs on/off
-- intervals from `command_sent`; we return the rows ordered by
-- event_time so the Go side can do a single linear pass.
-- name: ListActuatorEventsForRollup :many
SELECT event_time, actuator_id, command_sent, resulting_state_numeric_actual
FROM gr33ncore.actuator_events
WHERE actuator_id = $1
  AND event_time >= $2
  AND event_time <  $3
ORDER BY event_time ASC;

-- Phase 20.7 WS4 — latest "what state was the actuator in right
-- before the window opened?". Needed so a light that was turned on
-- the day before and left on gets credited for its morning runtime.
-- NULL row → assume OFF at window start.
-- name: GetLastActuatorEventBefore :one
SELECT event_time, actuator_id, command_sent, resulting_state_numeric_actual
FROM gr33ncore.actuator_events
WHERE actuator_id = $1 AND event_time < $2
ORDER BY event_time DESC
LIMIT 1;

-- Phase 20.7 WS6 — the "Cost to date" card on Crop Cycle detail.
-- name: GetCostTotalsByCropCycle :many
SELECT
    category,
    currency,
    COALESCE(SUM(CASE WHEN is_income THEN amount ELSE 0 END), 0)::numeric AS income,
    COALESCE(SUM(CASE WHEN NOT is_income THEN amount ELSE 0 END), 0)::numeric AS expense,
    COALESCE(SUM(CASE WHEN is_income THEN amount ELSE -amount END), 0)::numeric AS net,
    COUNT(*)::bigint AS tx_count
FROM gr33ncore.cost_transactions
WHERE crop_cycle_id = $1
GROUP BY category, currency
ORDER BY category ASC, currency ASC;

-- name: ListCostTransactionsByCropCycle :many
SELECT * FROM gr33ncore.cost_transactions
WHERE crop_cycle_id = $1
ORDER BY transaction_date DESC, id DESC;
