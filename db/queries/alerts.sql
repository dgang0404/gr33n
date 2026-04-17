-- ============================================================
-- Queries: gr33ncore.alerts_notifications
-- ============================================================

-- name: CreateAlert :one
INSERT INTO gr33ncore.alerts_notifications (
  farm_id, recipient_user_id, triggering_event_source_type,
  triggering_event_source_id, severity, subject_rendered,
  message_text_rendered, status
) VALUES ($1,$2,$3,$4,$5,$6,$7,'pending') RETURNING *;

-- name: GetAlertNotificationByID :one
SELECT * FROM gr33ncore.alerts_notifications WHERE id = $1;

-- name: ListAlertsByFarm :many
SELECT * FROM gr33ncore.alerts_notifications
WHERE farm_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUnreadAlertsByFarm :one
SELECT COUNT(*) FROM gr33ncore.alerts_notifications
WHERE farm_id = $1 AND is_read = FALSE;

-- name: MarkAlertRead :one
UPDATE gr33ncore.alerts_notifications
SET is_read = TRUE, read_at = NOW(), status = 'read_by_user'
WHERE id = $1 RETURNING *;

-- name: MarkAlertAcknowledged :one
UPDATE gr33ncore.alerts_notifications
SET is_acknowledged = TRUE, acknowledged_at = NOW(),
    acknowledged_by_user_id = $2, status = 'acknowledged_by_user'
WHERE id = $1 RETURNING *;

-- name: ListAlertsByRecipient :many
SELECT * FROM gr33ncore.alerts_notifications
WHERE recipient_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentUnacknowledgedAlertForSource :one
SELECT id FROM gr33ncore.alerts_notifications
WHERE farm_id = $1
  AND triggering_event_source_type = $2
  AND triggering_event_source_id = $3
  AND is_acknowledged = FALSE
  AND created_at > NOW() - INTERVAL '30 minutes'
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLatestAlertCreatedAtForSource :one
-- Returns the created_at of the most recent alert for this (farm, source_type, source_id)
-- regardless of ack status. Used by the sensor threshold evaluator to enforce per-sensor
-- cooldown windows.
SELECT created_at FROM gr33ncore.alerts_notifications
WHERE farm_id = $1
  AND triggering_event_source_type = $2
  AND triggering_event_source_id = $3
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateAlertForRule :one
-- Inserts an alerts_notifications row for a Phase 20 rule-driven
-- send_notification action. Distinct from CreateAlert because it
-- carries notification_template_id (so the Alerts page can show which
-- template rendered this notification) and sets
-- triggering_event_source_type = 'automation_rule'.
INSERT INTO gr33ncore.alerts_notifications (
  farm_id, notification_template_id,
  triggering_event_source_type, triggering_event_source_id,
  severity, subject_rendered, message_text_rendered, status
) VALUES ($1, $2, 'automation_rule', $3, $4, $5, $6, 'pending')
RETURNING *;

-- name: GetNotificationTemplateByID :one
SELECT * FROM gr33ncore.notification_templates WHERE id = $1;
