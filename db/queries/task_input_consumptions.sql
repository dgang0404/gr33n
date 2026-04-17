-- Phase 20.7 WS3 — task_input_consumptions CRUD. The handler wraps
-- each Create in the autologger so the paired batch-decrement +
-- cost_transactions row write happen atomically; Delete calls the
-- compensating-refund path so the ledger stays append-only.

-- name: ListTaskInputConsumptionsByTask :many
SELECT * FROM gr33ncore.task_input_consumptions
WHERE task_id = $1
ORDER BY recorded_at DESC, id DESC;

-- name: GetTaskInputConsumptionByID :one
SELECT * FROM gr33ncore.task_input_consumptions
WHERE id = $1;

-- name: CreateTaskInputConsumption :one
INSERT INTO gr33ncore.task_input_consumptions (
    farm_id, task_id, input_batch_id, quantity, unit_id, notes,
    recorded_by, cost_transaction_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING *;

-- name: UpdateTaskInputConsumptionCostTx :exec
-- The autologger may write the cost_transactions row after the
-- consumption row exists (so the idempotency key can reference the
-- consumption id). This backfills the link.
UPDATE gr33ncore.task_input_consumptions
SET cost_transaction_id = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteTaskInputConsumption :exec
DELETE FROM gr33ncore.task_input_consumptions WHERE id = $1;
