package fileattachutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/filestorage"
)

// DeleteAttachmentIfUnreferenced removes a cost receipt attachment row and blob
// only when no cost transaction still points at it.
func DeleteAttachmentIfUnreferenced(ctx context.Context, pool *pgxpool.Pool, store filestorage.Store, attachmentID int64) error {
	var storagePath string
	err := pool.QueryRow(ctx, `
DELETE FROM gr33ncore.file_attachments fa
WHERE fa.id = $1
  AND fa.related_table_name = 'cost_transactions'
  AND NOT EXISTS (
    SELECT 1
    FROM gr33ncore.cost_transactions ct
    WHERE ct.receipt_file_id = fa.id
  )
RETURNING fa.storage_path
`, attachmentID).Scan(&storagePath)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("delete file attachment row: %w", err)
	}
	if err := store.Delete(ctx, storagePath); err != nil {
		return fmt.Errorf("delete file attachment blob: %w", err)
	}
	return nil
}
