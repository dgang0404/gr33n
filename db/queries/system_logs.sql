-- ============================================================
-- Queries: gr33ncore.system_logs (Phase 115 WS3)
-- ============================================================

-- name: InsertSystemLog :exec
INSERT INTO gr33ncore.system_logs (
  farm_id, user_id, log_level, event_type, message, source_component, context_data
) VALUES ($1, $2, $3, $4, $5, $6, coalesce($7::jsonb, '{}'::jsonb));

-- name: ListSystemLogsByFarm :many
SELECT * FROM gr33ncore.system_logs
WHERE farm_id = $1
  AND ($2::text IS NULL OR $2 = '' OR log_level::text = $2)
ORDER BY log_time DESC
LIMIT $3;
