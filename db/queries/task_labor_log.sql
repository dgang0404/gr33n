-- ============================================================
-- Queries: gr33ncore.task_labor_log (Phase 20.95 WS1)
-- ============================================================

-- name: ListTaskLaborLogsByTask :many
SELECT * FROM gr33ncore.task_labor_log
WHERE task_id = $1
ORDER BY started_at ASC, id ASC;

-- name: GetTaskLaborLogByID :one
SELECT * FROM gr33ncore.task_labor_log
WHERE id = $1;

-- name: CreateTaskLaborLog :one
INSERT INTO gr33ncore.task_labor_log (
    farm_id, task_id, user_id, started_at, ended_at, minutes,
    hourly_rate_snapshot, currency, notes
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetOpenTaskLaborLogForUser :one
-- Phase 20.9 WS1 — timer stop path. At most one open (ended_at IS NULL)
-- row per (task, user) is expected; the handler is the only writer.
SELECT * FROM gr33ncore.task_labor_log
WHERE task_id = $1 AND user_id = $2 AND ended_at IS NULL
ORDER BY started_at DESC
LIMIT 1;

-- name: CloseTaskLaborLog :one
-- Phase 20.9 WS1 — stops a running timer. Rate is captured at close
-- time, not start, so a rate change mid-shift applies to the rest of
-- the shift, not retroactively.
UPDATE gr33ncore.task_labor_log
SET ended_at = $2,
    minutes = $3,
    hourly_rate_snapshot = sqlc.narg('hourly_rate_snapshot')::numeric,
    currency = sqlc.narg('currency')::char(3),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTaskLaborLog :exec
DELETE FROM gr33ncore.task_labor_log WHERE id = $1;

-- name: RecalcTaskTimeSpentMinutes :exec
-- Running SUM over surviving rows. Called by handler after INSERT/DELETE.
UPDATE gr33ncore.tasks
SET time_spent_minutes = COALESCE((
    SELECT SUM(minutes)::INTEGER FROM gr33ncore.task_labor_log WHERE task_id = $1
), 0)
WHERE id = $1;
