package ingest

import (
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

func TestTaskDocument(t *testing.T) {
	desc := "harvest basil"
	tk := db.Gr33ncoreTask{
		ID:          1,
		Title:       "Pick herbs",
		Description: &desc,
		Status:      commontypes.TaskStatusEnum("todo"),
	}
	out := TaskDocument(tk)
	if out == "" || len(out) < 10 {
		t.Fatalf("unexpected: %q", out)
	}
}

func TestCropCycleDocument(t *testing.T) {
	c := db.Gr33nfertigationCropCycle{
		ID:          9,
		FarmID:      1,
		ZoneID:      42,
		Name:        "Basil Block A",
		IsActive:    true,
		StartedAt:   pgtype.Date{Valid: true, Time: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)},
		CurrentStage: db.NullGr33nfertigationGrowthStageEnum{
			Valid:                           true,
			Gr33nfertigationGrowthStageEnum: db.Gr33nfertigationGrowthStageEnumLateVeg,
		},
	}
	out := CropCycleDocument(c)
	if out == "" || len(out) < 20 {
		t.Fatalf("unexpected: %q", out)
	}
	for _, sub := range []string{"crop_cycle:", "Basil Block A", "zone_id: 42", "late_veg", "active: yes"} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in: %q", sub, out)
		}
	}
}
