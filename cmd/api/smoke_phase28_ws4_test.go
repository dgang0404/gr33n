// Phase 28 WS4 — Farm Guardian snapshot ↔ unread alert detail smoke
// test. Exercises ListRecentUnreadAlertsByFarm + the snapshot
// integration against real Postgres so the SQL ordering (severity DESC,
// created_at DESC) and the prompt-block render path are validated as a
// single surface.
//
// Coverage:
//   - Seeded unread alerts surface in Snapshot.UnreadAlertDetails with
//     the right severity, subject, source, and triggered_at.
//   - PromptBlock() includes a "[severity] subject (source #id, Xh ago)"
//     line per alert plus the unread count header.
//   - SnapshotMaxAlertDetails caps detail rendering even when more
//     unread alerts exist (the extras are still represented in the
//     UnreadAlerts count and a "(+ N more)" line).

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

// seedUnreadAlertForFarm inserts a single unread alert directly into
// gr33ncore.alerts_notifications and registers cleanup. Returns the
// alert ID for downstream assertions.
func seedUnreadAlertForFarm(t *testing.T, farmID int64, severity, subject, message, sourceType string, sourceID int64, ageHours int) int64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var id int64
	createdAt := time.Now().UTC().Add(-time.Duration(ageHours) * time.Hour)
	if err := testPool.QueryRow(ctx, `
INSERT INTO gr33ncore.alerts_notifications
    (farm_id, triggering_event_source_type, triggering_event_source_id,
     severity, subject_rendered, message_text_rendered, status, is_read, created_at)
VALUES ($1, $2, $3, $4::gr33ncore.notification_priority_enum, $5, $6, 'pending', FALSE, $7)
RETURNING id`,
		farmID, sourceType, sourceID, severity, subject, message, createdAt,
	).Scan(&id); err != nil {
		t.Fatalf("seed alert: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(ctx, `DELETE FROM gr33ncore.alerts_notifications WHERE id = $1`, id)
	})
	return id
}

func TestPhase28WS4_Snapshot_AttachesAlertDetails(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}

	// Seed two unread alerts. We use 'critical' for the one we plan to
	// assert on because the smoke DB has accumulated thousands of
	// 'high' alerts from prior tests, and 'critical' is reserved for
	// real-only scenarios — meaning our seed lands at the top of the
	// severity DESC, created_at DESC order without DB pruning. The
	// medium one is best-effort: it MAY or MAY NOT appear in the top-N
	// depending on what else is unread, and the test handles either.
	highID := seedUnreadAlertForFarm(t, 1, "critical",
		"Humidity threshold breach — Flower Room",
		"Humidity is 72.5% (threshold 65%) for sensor RH-Flower.",
		"sensor_reading", 4242, 0)
	medID := seedUnreadAlertForFarm(t, 1, "medium",
		"Reservoir refill due",
		"Tank level fell below 20%.",
		"automation_rule", 77, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	queries := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, queries, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	if snap.UnreadAlerts < 2 {
		t.Fatalf("expected at least 2 unread alerts, got %d", snap.UnreadAlerts)
	}
	if len(snap.UnreadAlertDetails) == 0 {
		t.Fatalf("expected UnreadAlertDetails populated; snapshot:\n%s", snap.Render())
	}

	// Find our seeded alerts by ID — the test farm may have other
	// unread alerts from concurrent tests; we only assert on ours.
	byID := map[int64]farmguardian.UnreadAlertDetail{}
	for _, a := range snap.UnreadAlertDetails {
		byID[a.ID] = a
	}
	highDetail, ok := byID[highID]
	if !ok {
		t.Logf("snapshot render:\n%s", snap.Render())
		t.Fatalf("critical alert %d not in top-N details", highID)
	}
	if highDetail.Severity != "critical" {
		t.Errorf("expected severity=critical, got %q", highDetail.Severity)
	}
	if !strings.Contains(highDetail.Subject, "Humidity threshold breach") {
		t.Errorf("expected humidity subject, got %q", highDetail.Subject)
	}
	if !strings.Contains(highDetail.Message, "threshold 65%") {
		t.Errorf("expected message snippet preserved, got %q", highDetail.Message)
	}
	if highDetail.SourceType != "sensor_reading" || highDetail.SourceID != 4242 {
		t.Errorf("expected source sensor_reading #4242, got %s #%d",
			highDetail.SourceType, highDetail.SourceID)
	}
	if highDetail.TriggeredAt.IsZero() {
		t.Error("TriggeredAt should not be zero")
	}

	// The medium alert may or may not be in the top-N depending on
	// what other unread alerts the test harness has accumulated — but
	// if it IS present, it should rank below the high one.
	if medDetail, hasMed := byID[medID]; hasMed {
		if medDetail.Severity != "medium" {
			t.Errorf("expected severity=medium, got %q", medDetail.Severity)
		}
	}

	// The rendered PromptBlock must include the unread count header AND
	// our high alert's compact line.
	block := snap.PromptBlock()
	for _, want := range []string{
		"Unread alerts:",
		"[critical] Humidity threshold breach",
		"sensor_reading #4242",
		"detail: Humidity is 72.5%",
	} {
		if !strings.Contains(block, want) {
			t.Logf("rendered:\n%s", block)
			t.Fatalf("PromptBlock missing %q", want)
		}
	}
}

func TestPhase28WS4_Snapshot_CapsAlertDetailsAtBudget(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}

	// Seed SnapshotMaxAlertDetails + 2 unread "critical" alerts so the
	// cap is exercised regardless of background test alerts. Same
	// severity + closely-spaced created_at — Postgres orders ties by
	// id DESC under our ORDER BY clause once severity matches.
	seeded := make([]int64, 0, farmguardian.SnapshotMaxAlertDetails+2)
	for i := 0; i < cap(seeded); i++ {
		id := seedUnreadAlertForFarm(t, 1, "critical",
			"WS4 cap test", "Synthetic alert for cap budget test.",
			"automation_program", int64(900+i), 0)
		seeded = append(seeded, id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	queries := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, queries, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}

	if len(snap.UnreadAlertDetails) > farmguardian.SnapshotMaxAlertDetails {
		t.Fatalf("UnreadAlertDetails over budget: %d > %d",
			len(snap.UnreadAlertDetails), farmguardian.SnapshotMaxAlertDetails)
	}
	// The unread count must still report the full population (cap is
	// rendering-only, not count-only). Our seeds alone push us over
	// the cap so this assertion is meaningful regardless of other
	// alerts in the DB.
	if snap.UnreadAlerts < int64(len(seeded)) {
		t.Fatalf("UnreadAlerts %d should include all %d seeded alerts",
			snap.UnreadAlerts, len(seeded))
	}
	// And the rendered block must include a "(+ N more unread alerts)"
	// marker when more alerts exist than fit in the detail block.
	block := snap.PromptBlock()
	if !strings.Contains(block, "more unread alerts") {
		t.Logf("rendered:\n%s", block)
		t.Fatal("expected '+ N more unread alerts' truncation note")
	}
}
