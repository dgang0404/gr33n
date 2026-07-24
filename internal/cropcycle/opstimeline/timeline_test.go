package opstimeline

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

func TestParseTimeQuery(t *testing.T) {
	t.Parallel()
	if _, err := ParseTimeQuery("not-a-date"); err == nil {
		t.Fatal("expected error for bad date")
	}
	got, err := ParseTimeQuery("2025-06-15")
	if err != nil {
		t.Fatal(err)
	}
	if got.Format("2006-01-02") != "2025-06-15" {
		t.Fatalf("date parse = %s", got.Format(time.RFC3339))
	}
	rfc, err := ParseTimeQuery("2025-06-15T08:30:00Z")
	if err != nil || !rfc.Equal(time.Date(2025, 6, 15, 8, 30, 0, 0, time.UTC)) {
		t.Fatalf("RFC3339 parse = %v err=%v", rfc, err)
	}
}

func TestDefaultRangeUsesCycleDates(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	cycle := db.Gr33nfertigationCropCycle{
		StartedAt:   pgtype.Date{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		HarvestedAt: pgtype.Date{Time: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), Valid: true},
	}
	from, to := DefaultRange(cycle, now)
	if !from.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("from = %s", from)
	}
	if !to.Equal(now) {
		t.Fatalf("to = %s want %s (harvest before now keeps now as upper bound)", to, now)
	}
}
