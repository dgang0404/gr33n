-- ============================================================
-- Queries: gr33ncore.file_attachments
-- ============================================================

-- name: CreateFileAttachment :one
INSERT INTO gr33ncore.file_attachments (
    farm_id, related_module_schema, related_table_name, related_record_id,
    file_name, file_type, file_size_bytes, storage_path, mime_type, uploaded_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetFileAttachmentByID :one
SELECT * FROM gr33ncore.file_attachments WHERE id = $1;
