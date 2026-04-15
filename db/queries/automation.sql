-- ============================================================
-- Queries: schedules, executable actions, automation runs
-- ============================================================

-- name: ListSchedulesByFarm :many
SELECT * FROM gr33ncore.schedules
WHERE farm_id = $1
ORDER BY name ASC;

-- name: GetScheduleByID :one
SELECT * FROM gr33ncore.schedules
WHERE id = $1;

-- name: UpdateScheduleActive :one
UPDATE gr33ncore.schedules
SET is_active = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListActiveSchedules :many
SELECT * FROM gr33ncore.schedules
WHERE is_active = TRUE;

-- name: MarkScheduleTriggered :one
UPDATE gr33ncore.schedules
SET last_triggered_time = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListExecutableActionsBySchedule :many
SELECT * FROM gr33ncore.executable_actions
WHERE schedule_id = $1
ORDER BY execution_order ASC, id ASC;

-- name: CreateAutomationRun :one
INSERT INTO gr33ncore.automation_runs (
    farm_id, schedule_id, rule_id, status, message, details, executed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAutomationRunsByFarm :many
SELECT * FROM gr33ncore.automation_runs
WHERE farm_id = $1
ORDER BY executed_at DESC
LIMIT $2;

-- name: GetLastSuccessfulRunBySchedule :one
SELECT * FROM gr33ncore.automation_runs
WHERE schedule_id = $1 AND status = 'success'
ORDER BY executed_at DESC
LIMIT 1;

-- name: GetAutomationRunByDetails :one
SELECT * FROM gr33ncore.automation_runs
WHERE schedule_id = $1 AND details @> $2::jsonb
LIMIT 1;
