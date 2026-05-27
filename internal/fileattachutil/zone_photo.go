package fileattachutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/filestorage"
)

// DeleteZonePhotoIfUnreferenced removes a zone photo attachment when no zone
// meta_data still lists it in photo_attachment_ids.
func DeleteZonePhotoIfUnreferenced(ctx context.Context, pool *pgxpool.Pool, store filestorage.Store, attachmentID int64) error {
	var storagePath string
	err := pool.QueryRow(ctx, `
DELETE FROM gr33ncore.file_attachments fa
WHERE fa.id = $1
  AND fa.related_table_name = 'zones'
  AND fa.file_type = 'zone_photo'
  AND NOT EXISTS (
    SELECT 1
    FROM gr33ncore.zones z
    WHERE z.deleted_at IS NULL
      AND z.meta_data IS NOT NULL
      AND (z.meta_data->'photo_attachment_ids') @> to_jsonb($1::bigint)
  )
RETURNING fa.storage_path
`, attachmentID).Scan(&storagePath)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("delete zone photo attachment row: %w", err)
	}
	if err := store.Delete(ctx, storagePath); err != nil {
		return fmt.Errorf("delete zone photo blob: %w", err)
	}
	return nil
}
