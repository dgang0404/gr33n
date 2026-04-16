-- ============================================================
-- Queries: gr33ncore.farm_finance_account_mappings
-- ============================================================

-- name: ListFarmFinanceAccountMappings :many
SELECT *
FROM gr33ncore.farm_finance_account_mappings
WHERE farm_id = $1 AND is_active = TRUE
ORDER BY cost_category;

-- name: UpsertFarmFinanceAccountMapping :one
INSERT INTO gr33ncore.farm_finance_account_mappings (
    farm_id, cost_category, account_code, account_name, is_active
) VALUES ($1, $2, $3, $4, TRUE)
ON CONFLICT (farm_id, cost_category)
DO UPDATE SET
    account_code = EXCLUDED.account_code,
    account_name = EXCLUDED.account_name,
    is_active = TRUE,
    updated_at = NOW()
RETURNING *;

-- name: ResetFarmFinanceAccountMappingByCategory :execrows
DELETE FROM gr33ncore.farm_finance_account_mappings
WHERE farm_id = $1 AND cost_category = $2;

-- name: ResetFarmFinanceAccountMappingsAll :execrows
DELETE FROM gr33ncore.farm_finance_account_mappings
WHERE farm_id = $1;
