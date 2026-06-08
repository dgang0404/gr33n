-- ============================================================
-- Queries: gr33nfertigation.crop_cycle_stage_events (Phase 56 WS2)
-- ============================================================

-- name: InsertCropCycleStageEvent :one
INSERT INTO gr33nfertigation.crop_cycle_stage_events (
    crop_cycle_id, growth_stage, entered_at
) VALUES ($1, $2, $3)
RETURNING *;

-- name: ListCropCycleStageEventsByCycle :many
SELECT * FROM gr33nfertigation.crop_cycle_stage_events
WHERE crop_cycle_id = $1
ORDER BY entered_at ASC, id ASC;
