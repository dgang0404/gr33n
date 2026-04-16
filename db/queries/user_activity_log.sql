-- ============================================================
-- Queries: gr33ncore.user_activity_log (compliance / audit trail)
-- ============================================================

-- name: InsertUserActivityLog :exec
INSERT INTO gr33ncore.user_activity_log (
    user_id,
    farm_id,
    action_type,
    target_module_schema,
    target_table_name,
    target_record_id,
    target_record_description,
    user_agent,
    status,
    failure_reason,
    details
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    COALESCE(sqlc.arg('details')::jsonb, '{}'::jsonb)
);

-- name: ListUserActivityLogByFarm :many
SELECT
    id,
    user_id,
    farm_id,
    activity_time,
    action_type,
    target_module_schema,
    target_table_name,
    target_record_id,
    target_record_description,
    ip_address,
    user_agent,
    session_id,
    status,
    failure_reason,
    details,
    created_at
FROM gr33ncore.user_activity_log
WHERE farm_id = $1
ORDER BY activity_time DESC, id DESC
LIMIT $2 OFFSET $3;
