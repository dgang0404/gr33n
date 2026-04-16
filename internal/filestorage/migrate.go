package filestorage

import (
	"context"
	"fmt"
)

type MigrationAttachment struct {
	ID          int64
	StoragePath string
	FileName    string
}

type MigrationSummary struct {
	Scanned int
	Copied  int
	Failed  int
}

// MigrateAttachments copies blobs from a source local store into the target store
// using the existing storage_path keys, so DB rows do not need to change.
func MigrateAttachments(ctx context.Context, source *Local, target Store, attachments []MigrationAttachment, dryRun bool) (MigrationSummary, error) {
	var summary MigrationSummary
	for _, att := range attachments {
		summary.Scanned++
		if att.StoragePath == "" {
			summary.Failed++
			return summary, fmt.Errorf("attachment %d has empty storage_path", att.ID)
		}
		if dryRun {
			summary.Copied++
			continue
		}
		rc, err := source.Open(ctx, att.StoragePath)
		if err != nil {
			summary.Failed++
			return summary, fmt.Errorf("open source blob for attachment %d: %w", att.ID, err)
		}
		_, err = target.Put(ctx, att.StoragePath, rc, maxMigrationObjectSize)
		_ = rc.Close()
		if err != nil {
			summary.Failed++
			return summary, fmt.Errorf("write target blob for attachment %d: %w", att.ID, err)
		}
		summary.Copied++
	}
	return summary, nil
}

const maxMigrationObjectSize = 128 << 20 // 128 MiB safety cap for backfill jobs.
