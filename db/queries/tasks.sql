-- ============================================================
-- Queries: gr33ncore.tasks
-- ============================================================

-- name: CreateTask :one
INSERT INTO gr33ncore.tasks (
    farm_id, zone_id, schedule_id, title, description, task_type, status, priority,
    assigned_to_user_id, due_date, estimated_duration_minutes,
    source_alert_id, source_rule_id, created_by_user_id, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
RETURNING *;

-- name: ListTasksBySourceAlertID :many
SELECT * FROM gr33ncore.tasks
WHERE source_alert_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListTasksBySourceRuleID :many
SELECT * FROM gr33ncore.tasks
WHERE source_rule_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetTaskByID :one
SELECT * FROM gr33ncore.tasks
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListTasksByFarm :many
SELECT * FROM gr33ncore.tasks
WHERE farm_id = $1 AND deleted_at IS NULL
ORDER BY due_date ASC NULLS LAST, priority DESC;

-- name: ListTasksByAssignee :many
SELECT * FROM gr33ncore.tasks
WHERE assigned_to_user_id = $1 AND farm_id = $2 AND deleted_at IS NULL
ORDER BY due_date ASC NULLS LAST;

-- name: UpdateTaskStatus :one
UPDATE gr33ncore.tasks
SET status = $2, updated_by_user_id = $3, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTask :one
UPDATE gr33ncore.tasks
SET title = $2, description = $3, zone_id = $4, schedule_id = $5,
    task_type = $6, priority = $7, due_date = $8,
    assigned_to_user_id = $9, estimated_duration_minutes = $10,
    updated_by_user_id = $11, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteTask :exec
UPDATE gr33ncore.tasks
SET deleted_at = NOW(), updated_at = NOW(), updated_by_user_id = $2
WHERE id = $1;

-- name: CountTasksByStatusForFarm :many
SELECT status, COUNT(*)::bigint AS cnt
FROM gr33ncore.tasks
WHERE farm_id = $1 AND deleted_at IS NULL
GROUP BY status
ORDER BY status ASC;
