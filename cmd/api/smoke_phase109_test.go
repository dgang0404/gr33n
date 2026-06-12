// Phase 109 — catalog version bump notifications for farm admins.
package main

import (
	"context"
	"testing"

	"gr33n-api/internal/catalognotify"
	db "gr33n-api/internal/db"
)

func TestPhase109_CatalogVersionNotify(t *testing.T) {
	ctx := context.Background()
	q := db.New(testPool)

	maxVer, err := q.GetMaxCropCatalogVersion(ctx)
	if err != nil {
		t.Fatalf("max catalog version: %v", err)
	}
	if maxVer < 2 {
		t.Skip("catalog version too low for bump test")
	}

	_, err = testPool.Exec(ctx, `
DELETE FROM gr33ncore.alerts_notifications
WHERE farm_id = 1
  AND triggering_event_source_type = 'catalog_version_bump'
  AND triggering_event_source_id = $1`, maxVer)
	if err != nil {
		t.Fatalf("cleanup alerts: %v", err)
	}
	_, _ = testPool.Exec(ctx, `DELETE FROM gr33ncore.farm_catalog_version_seen WHERE farm_id = 1`)
	_, err = q.UpsertPlatformCatalogState(ctx, maxVer-1)
	if err != nil {
		t.Fatalf("seed platform state: %v", err)
	}

	// Opt-in (default) — expect alert for dev farm admin.
	resp := authPatch(t, smokeJWT(t), "/profile/notification-preferences", map[string]any{
		"catalog_updates": true,
	})
	expectStatus(t, resp, 200)

	result, err := catalognotify.Sync(ctx, q, nil)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if result.AlertsCreated < 1 {
		t.Fatalf("expected alerts created, got %+v", result)
	}

	resp = authGet(t, smokeJWT(t), "/farms/1/alerts?limit=50")
	expectStatus(t, resp, 200)
	found := false
	for _, raw := range decodeSlice(t, resp) {
		row, _ := raw.(map[string]any)
		if row["triggering_event_source_type"] == catalognotify.SourceType &&
			int64(row["triggering_event_source_id"].(float64)) == int64(maxVer) {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("catalog_version_bump alert not found for farm 1")
	}

	// Opt-out — no new alert on re-sync after resetting seen/state.
	resp = authPatch(t, smokeJWT(t), "/profile/notification-preferences", map[string]any{
		"catalog_updates": false,
	})
	expectStatus(t, resp, 200)

	_, _ = testPool.Exec(ctx, `DELETE FROM gr33ncore.farm_catalog_version_seen WHERE farm_id = 1`)
	_, _ = testPool.Exec(ctx, `
DELETE FROM gr33ncore.alerts_notifications
WHERE farm_id = 1 AND triggering_event_source_type = 'catalog_version_bump' AND triggering_event_source_id = $1`, maxVer)
	_, _ = q.UpsertPlatformCatalogState(ctx, maxVer-1)

	result, err = catalognotify.Sync(ctx, q, nil)
	if err != nil {
		t.Fatalf("sync opt-out: %v", err)
	}
	if result.AlertsCreated != 0 {
		t.Fatalf("opt-out should skip alerts, got %+v", result)
	}

	// Restore opt-in for other tests.
	_ = authPatch(t, smokeJWT(t), "/profile/notification-preferences", map[string]any{
		"catalog_updates": true,
	})
}

func TestPhase109_CatalogVersionNotifyDebounced(t *testing.T) {
	ctx := context.Background()
	q := db.New(testPool)

	maxVer, err := q.GetMaxCropCatalogVersion(ctx)
	if err != nil {
		t.Fatalf("max catalog version: %v", err)
	}

	_, _ = q.UpsertPlatformCatalogState(ctx, maxVer)
	_, _ = q.UpsertFarmCatalogVersionSeen(ctx, db.UpsertFarmCatalogVersionSeenParams{
		FarmID:             1,
		CatalogVersionSeen: maxVer,
	})

	result, err := catalognotify.Sync(ctx, q, nil)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if result.AlertsCreated != 0 || result.FarmsNotified != 0 {
		t.Fatalf("already at version should not notify: %+v", result)
	}
}
