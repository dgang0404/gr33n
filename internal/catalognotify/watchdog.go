// Package catalognotify detects platform crop catalog version bumps and notifies farm admins.
package catalognotify

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/notifyprefs"
	"gr33n-api/internal/pushnotify"
)

const SourceType = "catalog_version_bump"

// SyncResult summarizes a catalog version sync pass.
type SyncResult struct {
	CurrentVersion int32
	PreviousVersion int32
	FarmsNotified  int
	AlertsCreated  int
}

// SyncOnStartup compares live catalog_version to platform state and notifies farm admins.
func SyncOnStartup(ctx context.Context, pool *pgxpool.Pool, push *pushnotify.Dispatcher) (SyncResult, error) {
	return Sync(ctx, db.New(pool), push)
}

// Sync is the testable entry point.
func Sync(ctx context.Context, q *db.Queries, push *pushnotify.Dispatcher) (SyncResult, error) {
	if q == nil {
		return SyncResult{}, errors.New("catalognotify: nil querier")
	}
	maxRow, err := q.GetMaxCropCatalogVersion(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("max catalog version: %w", err)
	}
	current := maxRow

	prevState, err := q.GetPlatformCatalogState(ctx)
	prev := int32(1)
	if err == nil {
		prev = prevState.CatalogVersion
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return SyncResult{}, fmt.Errorf("platform catalog state: %w", err)
	}

	out := SyncResult{CurrentVersion: current, PreviousVersion: prev}
	if current <= prev {
		return out, nil
	}

	farmIDs, err := q.ListAllFarmIDs(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("list farms: %w", err)
	}

	now := time.Now()
	srcType := SourceType
	severity := db.Gr33ncoreNotificationPriorityEnumMedium
	srcID := int64(current)

	for _, farmID := range farmIDs {
		seenRow, serr := q.GetFarmCatalogVersionSeen(ctx, farmID)
		seen := int32(0)
		if serr == nil {
			seen = seenRow.CatalogVersionSeen
		} else if !errors.Is(serr, pgx.ErrNoRows) {
			log.Printf("catalognotify: farm %d seen lookup: %v", farmID, serr)
			continue
		}
		if seen >= current {
			continue
		}

		if _, derr := q.GetCatalogVersionBumpAlertForFarm(ctx, db.GetCatalogVersionBumpAlertForFarmParams{
			FarmID:                    farmID,
			TriggeringEventSourceID:   &srcID,
		}); derr == nil {
			_, _ = q.UpsertFarmCatalogVersionSeen(ctx, db.UpsertFarmCatalogVersionSeenParams{
				FarmID:             farmID,
				CatalogVersionSeen: current,
				NotifiedAt:         pgtype.Timestamptz{Time: now, Valid: true},
			})
			continue
		} else if !errors.Is(derr, pgx.ErrNoRows) {
			log.Printf("catalognotify: farm %d debounce lookup: %v", farmID, derr)
			continue
		}

		adminIDs, err := q.ListFarmCatalogNotifyAdminUserIDs(ctx, farmID)
		if err != nil {
			log.Printf("catalognotify: farm %d admins: %v", farmID, err)
			continue
		}

		fromVer := prev
		if seen > 0 && seen < current {
			fromVer = seen
		}
		subject := fmt.Sprintf("gr33n knowledge base updated (v%d → v%d)", fromVer, current)
		message := fmt.Sprintf(
			"New crops may be available in the Plants picker. Review Settings → Crops & targets or re-run guardian-bootstrap-farm after migrate.",
		)

		createdAny := false
		for _, uid := range adminIDs {
			prof, perr := q.GetProfileByUserID(ctx, uid)
			if perr != nil {
				continue
			}
			if !notifyprefs.FromPreferencesJSON(prof.Preferences).CatalogUpdates {
				continue
			}
			alert, cerr := q.CreateAlert(ctx, db.CreateAlertParams{
				FarmID:                    farmID,
				RecipientUserID:           pgtype.UUID{Bytes: uid, Valid: true},
				TriggeringEventSourceType: &srcType,
				TriggeringEventSourceID:   &srcID,
				Severity:                  &severity,
				SubjectRendered:           &subject,
				MessageTextRendered:       &message,
			})
			if cerr != nil {
				log.Printf("catalognotify: create alert farm=%d user=%s: %v", farmID, uid, cerr)
				continue
			}
			createdAny = true
			out.AlertsCreated++
			if push != nil {
				push.DispatchCatalogUpdate(ctx, alert)
			}
		}

		if createdAny {
			out.FarmsNotified++
			_, _ = q.UpsertFarmCatalogVersionSeen(ctx, db.UpsertFarmCatalogVersionSeenParams{
				FarmID:             farmID,
				CatalogVersionSeen: current,
				NotifiedAt:         pgtype.Timestamptz{Time: now, Valid: true},
			})
		}
	}

	if _, err := q.UpsertPlatformCatalogState(ctx, current); err != nil {
		return out, fmt.Errorf("upsert platform catalog state: %w", err)
	}
	if out.AlertsCreated > 0 {
		log.Printf("catalognotify: catalog v%d → v%d — %d farms, %d alerts",
			prev, current, out.FarmsNotified, out.AlertsCreated)
	}
	return out, nil
}
