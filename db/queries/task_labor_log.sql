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

-- name: DeleteTaskLaborLog :exec
DELETE FROM gr33ncore.task_labor_log WHERE id = $1;

-- name: RecalcTaskTimeSpentMinutes :exec
-- Running SUM over surviving rows. Called by handler after INSERT/DELETE.
UPDATE gr33ncore.tasks
SET time_spent_minutes = COALESCE((
    SELECT SUM(minutes)::INTEGER FROM gr33ncore.task_labor_log WHERE task_id = $1
), 0)
WHERE id = $1;
