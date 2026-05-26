// Phase 29 WS7 — Guardian demo seed alerts smoke.
//
// Verifies master_seed.sql inserted the three unread demo alerts for farm 1
// (OHN inventory, Flower Room humidity, schedule reminder). Skips when the
// test pool is unavailable; does not mutate rows.
package main

import (
	"context"
	"testing"
	"time"
)

var phase29WS7DemoAlertSubjects = []string{
	"OHN batch below minimum — reorder or brew soon",
	"Humidity high — Flower Room",
	"Light schedule change in 48 hours — Flower Room",
}

func TestPhase29WS7_SeededGuardianDemoAlerts(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, subject := range phase29WS7DemoAlertSubjects {
		var n int
		err := testPool.QueryRow(ctx, `
SELECT COUNT(*)::int
FROM gr33ncore.alerts_notifications
WHERE farm_id = 1 AND subject_rendered = $1`, subject).Scan(&n)
		if err != nil {
			t.Fatalf("query alert %q: %v", subject, err)
		}
		if n < 1 {
			t.Fatalf("expected at least 1 seed alert with subject %q (got %d); run make seed or dev-stack-fresh", subject, n)
		}
	}
}
