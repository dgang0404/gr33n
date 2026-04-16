-- ============================================================
-- Cost transaction idempotency
-- ============================================================

-- name: GetCostTransactionIDByIdempotencyKey :one
SELECT cost_transaction_id
FROM gr33ncore.cost_transaction_idempotency
WHERE farm_id = $1 AND idempotency_key = $2;

-- name: InsertCostTransactionIdempotency :exec
INSERT INTO gr33ncore.cost_transaction_idempotency (farm_id, idempotency_key, cost_transaction_id)
VALUES ($1, $2, $3);
