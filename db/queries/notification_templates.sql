-- ============================================================
-- Queries: gr33ncore.notification_templates (Phase 115 WS2)
-- ============================================================

-- name: ListNotificationTemplatesByFarm :many
SELECT * FROM gr33ncore.notification_templates
WHERE farm_id = $1 OR farm_id IS NULL
ORDER BY farm_id NULLS FIRST, template_key ASC;

-- name: CreateNotificationTemplate :one
INSERT INTO gr33ncore.notification_templates (
  farm_id, template_key, description, subject_template, body_template_text,
  body_template_html, default_delivery_channels, default_priority, is_system_template
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, FALSE)
RETURNING *;

-- name: UpdateNotificationTemplate :one
UPDATE gr33ncore.notification_templates
SET template_key = COALESCE(NULLIF($2::text, ''), template_key),
    description = COALESCE($3, description),
    subject_template = COALESCE($4, subject_template),
    body_template_text = COALESCE($5, body_template_text),
    body_template_html = COALESCE($6, body_template_html),
    default_delivery_channels = COALESCE($7, default_delivery_channels),
    default_priority = COALESCE($8, default_priority),
    updated_at = NOW()
WHERE id = $1 AND farm_id = $9 AND is_system_template = FALSE
RETURNING *;
