-- ============================================================
-- Queries: gr33ncore.cost_transactions
-- ============================================================

-- name: CreateCostTransaction :one
INSERT INTO gr33ncore.cost_transactions (
    farm_id, transaction_date, category, subcategory, amount, currency,
    description, is_income, created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListCostTransactionsByFarm :many
SELECT * FROM gr33ncore.cost_transactions
WHERE farm_id = $1
ORDER BY transaction_date DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: ListCostTransactionsByFarmExport :many
SELECT id, farm_id, transaction_date, category, subcategory, amount, currency,
 description, is_income
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
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCostTransaction :exec
DELETE FROM gr33ncore.cost_transactions WHERE id = $1;
