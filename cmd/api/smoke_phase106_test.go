// Phase 106 — deficiency & pest symptom catalog smoke.
package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"gr33n-api/internal/farmguardian"
	db "gr33n-api/internal/db"
)

func TestPhase106_LookupCropSymptomsTomatoYellow(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	q := db.New(testPool)

	var count int64
	if err := testPool.QueryRow(ctx, `SELECT COUNT(*) FROM gr33ncrops.agronomy_symptom_entries WHERE published = TRUE`).Scan(&count); err != nil || count < 5 {
		t.Fatalf("symptom catalog not seeded (count=%d err=%v)", count, err)
	}

	block, err := farmguardian.LookupCropSymptoms(ctx, q, 1, "Yellow leaves on my tomato — interveinal pattern", nil)
	if err != nil {
		t.Fatalf("LookupCropSymptoms: %v", err)
	}
	lower := strings.ToLower(block)
	for _, want := range []string{
		"lookup_crop_symptoms",
		"tomato",
		"hypothes",
		"interveinal",
		"lookup_crop_targets",
		"mS/cm",
	} {
		if !strings.Contains(lower, strings.ToLower(want)) {
			t.Fatalf("block missing %q:\n%s", want, block)
		}
	}
}
