-- ============================================================
-- Queries: gr33ncore.farm_energy_prices (Phase 20.95 WS2)
-- ============================================================

-- name: ListFarmEnergyPrices :many
SELECT * FROM gr33ncore.farm_energy_prices
WHERE farm_id = $1
ORDER BY effective_from DESC, id DESC;

-- name: GetFarmEnergyPriceByID :one
SELECT * FROM gr33ncore.farm_energy_prices
WHERE id = $1;

-- name: CreateFarmEnergyPrice :one
INSERT INTO gr33ncore.farm_energy_prices (
    farm_id, effective_from, effective_to, price_per_kwh, currency, notes
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateFarmEnergyPrice :one
UPDATE gr33ncore.farm_energy_prices
SET effective_from = $2,
    effective_to   = $3,
    price_per_kwh  = $4,
    currency       = $5,
    notes          = $6
WHERE id = $1
RETURNING *;

-- name: DeleteFarmEnergyPrice :exec
DELETE FROM gr33ncore.farm_energy_prices WHERE id = $1;
