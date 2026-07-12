package fileattachutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/filestorage"
)

// DeleteFarmLayoutBackgroundIfUnreferenced removes a farm layout background
// attachment when no farm meta_data still references it.
func DeleteFarmLayoutBackgroundIfUnreferenced(ctx context.Context, pool *pgxpool.Pool, store filestorage.Store, attachmentID int64) error {
	var storagePath string
	err := pool.QueryRow(ctx, `
DELETE FROM gr33ncore.file_attachments fa
WHERE fa.id = $1
  AND fa.related_table_name = 'farms'
  AND fa.file_type = 'farm_layout_background'
  AND NOT EXISTS (
    SELECT 1
    FROM gr33ncore.farms f
    WHERE f.deleted_at IS NULL
      AND f.meta_data IS NOT NULL
      AND (f.meta_data->>'layout_background_attachment_id')::bigint = $1
  )
RETURNING fa.storage_path
`, attachmentID).Scan(&storagePath)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("delete farm layout background attachment row: %w", err)
	}
	if err := store.Delete(ctx, storagePath); err != nil {
		return fmt.Errorf("delete farm layout background blob: %w", err)
	}
	return nil
}
