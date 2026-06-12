-- ============================================================
-- Queries: gr33nfertigation.crop_cycles
-- ============================================================

-- name: CreateCropCycle :one
INSERT INTO gr33nfertigation.crop_cycles (
    farm_id, zone_id, name, batch_label, current_stage,
    is_active, started_at, cycle_notes, primary_program_id, plant_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListCropCyclesByFarm :many
SELECT * FROM gr33nfertigation.crop_cycles
WHERE farm_id = $1
ORDER BY started_at DESC;

-- name: ListCropCyclesByFarmUpdatedAfter :many
SELECT * FROM gr33nfertigation.crop_cycles
WHERE farm_id = sqlc.arg('farm_id') AND updated_at > sqlc.arg('updated_after')::timestamptz
ORDER BY updated_at ASC, id ASC;

-- name: GetCropCycleByID :one
SELECT * FROM gr33nfertigation.crop_cycles WHERE id = $1;

-- name: UpdateCropCycle :one
UPDATE gr33nfertigation.crop_cycles SET
    name = $2,
    batch_label = $3,
    zone_id = $4,
    is_active = $5,
    cycle_notes = $6,
    harvested_at = $7,
    yield_grams = $8,
    yield_notes = $9,
    primary_program_id = $10,
    plant_id = $11,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCropCycleStage :one
UPDATE gr33nfertigation.crop_cycles SET
    current_stage = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SoftDeleteCropCycle :exec
UPDATE gr33nfertigation.crop_cycles SET is_active = FALSE, updated_at = NOW() WHERE id = $1;

-- name: GetActiveCropCycleForZone :one
-- Phase 20.6 WS3 — the rule engine calls this to translate a rule's zone
-- into the active cycle so the setpoint resolver has a `current_stage`
-- to match against. At most one crop cycle per zone is active at a time
-- (enforced by uq_active_crop_cycle).
SELECT * FROM gr33nfertigation.crop_cycles
WHERE zone_id = $1 AND is_active = TRUE
LIMIT 1;

-- name: GetFertigationAggregatesByCropCycle :one
-- Phase 28 WS1 — rolling fertigation stats for the cycle-summary endpoint.
-- COALESCE keeps the JSON shape stable (zeros instead of NULLs) when a
-- cycle has no events yet. EC after-feed is the canonical "what the plants
-- actually experienced" reading; pH average blends pre + post so the
-- number reflects the working solution, not just the freshly-mixed batch.
SELECT
    COUNT(*)::bigint                                                  AS event_count,
    COALESCE(SUM(volume_applied_liters), 0)::numeric                  AS total_liters,
    COALESCE(AVG(ec_after_mscm), 0)::numeric                          AS avg_ec_mscm,
    COALESCE(MIN(ec_after_mscm), 0)::numeric                          AS min_ec_mscm,
    COALESCE(MAX(ec_after_mscm), 0)::numeric                          AS max_ec_mscm,
    COALESCE(AVG((COALESCE(ph_before,0) + COALESCE(ph_after,0)) / NULLIF(
        ((CASE WHEN ph_before IS NULL THEN 0 ELSE 1 END) +
         (CASE WHEN ph_after  IS NULL THEN 0 ELSE 1 END)), 0
    )), 0)::numeric                                                   AS avg_ph
FROM gr33nfertigation.fertigation_events
WHERE crop_cycle_id = $1;
