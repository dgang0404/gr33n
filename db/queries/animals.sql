-- ============================================================
-- Phase 20.8 WS2 — animal_groups + animal_lifecycle_events queries
-- ============================================================

-- name: ListAnimalGroupsByFarm :many
SELECT * FROM gr33nanimals.animal_groups
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY active DESC, label ASC, id ASC;

-- name: GetAnimalGroupByID :one
SELECT * FROM gr33nanimals.animal_groups
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateAnimalGroup :one
INSERT INTO gr33nanimals.animal_groups (
    farm_id, label, species, count, primary_zone_id, meta
) VALUES (
    sqlc.arg(farm_id),
    sqlc.arg(label),
    sqlc.narg(species),
    sqlc.narg(count),
    sqlc.narg(primary_zone_id),
    COALESCE(sqlc.narg(meta)::jsonb, '{}'::jsonb)
)
RETURNING *;

-- name: UpdateAnimalGroup :one
UPDATE gr33nanimals.animal_groups SET
    label           = sqlc.arg(label),
    species         = sqlc.narg(species),
    count           = sqlc.narg(count),
    primary_zone_id = sqlc.narg(primary_zone_id),
    meta            = COALESCE(sqlc.narg(meta)::jsonb, meta),
    updated_at      = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: ArchiveAnimalGroup :one
UPDATE gr33nanimals.animal_groups SET
    active          = FALSE,
    archived_at     = NOW(),
    archived_reason = $2,
    updated_at      = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteAnimalGroup :exec
UPDATE gr33nanimals.animal_groups
SET deleted_at = NOW()
WHERE id = $1;

-- name: ListLifecycleEventsByGroup :many
SELECT * FROM gr33nanimals.animal_lifecycle_events
WHERE animal_group_id = $1
ORDER BY event_time DESC, id DESC;

-- name: GetLifecycleEventByID :one
SELECT * FROM gr33nanimals.animal_lifecycle_events
WHERE id = $1;

-- name: CreateLifecycleEvent :one
INSERT INTO gr33nanimals.animal_lifecycle_events (
    farm_id, animal_group_id, event_type, event_time,
    delta_count, notes, recorded_by, related_task_id, meta
) VALUES (
    sqlc.arg(farm_id),
    sqlc.arg(animal_group_id),
    sqlc.arg(event_type),
    COALESCE(sqlc.narg(event_time)::timestamptz, NOW()),
    sqlc.narg(delta_count),
    sqlc.narg(notes),
    sqlc.narg(recorded_by),
    sqlc.narg(related_task_id),
    COALESCE(sqlc.narg(meta)::jsonb, '{}'::jsonb)
) RETURNING *;

-- name: DeleteLifecycleEvent :exec
DELETE FROM gr33nanimals.animal_lifecycle_events
WHERE id = $1;

-- name: SumLifecycleDeltasByGroup :one
SELECT COALESCE(SUM(delta_count)::bigint, 0)::bigint AS delta_total
FROM gr33nanimals.animal_lifecycle_events
WHERE animal_group_id = $1;

-- name: GetLatestLifecycleEventByGroup :one
-- Phase 210 — feeds the `animal_event` automation predicate: "the most
-- recent lifecycle event for this flock is type X". farm_id is included so
-- the predicate evaluator can reject a group that belongs to another farm
-- without a second round trip.
SELECT e.* FROM gr33nanimals.animal_lifecycle_events e
JOIN gr33nanimals.animal_groups g ON g.id = e.animal_group_id
WHERE e.animal_group_id = sqlc.arg(animal_group_id)
  AND g.farm_id = sqlc.arg(farm_id)
  AND g.deleted_at IS NULL
ORDER BY e.event_time DESC, e.id DESC
LIMIT 1;
