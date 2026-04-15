package automation

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestShouldTriggerNowMatchesCurrentMinute(t *testing.T) {
	now := time.Date(2026, 4, 15, 6, 0, 0, 0, time.UTC)
	ok, err := shouldTriggerNow("0 6 * * *", pgtype.Timestamptz{}, now)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !ok {
		t.Fatalf("expected schedule to trigger at %v", now)
	}
}

func TestShouldTriggerNowSkipsDuplicateMinute(t *testing.T) {
	now := time.Date(2026, 4, 15, 6, 0, 0, 0, time.UTC)
	last := pgtype.Timestamptz{Time: now, Valid: true}
	ok, err := shouldTriggerNow("0 6 * * *", last, now)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if ok {
		t.Fatalf("expected duplicate trigger in same minute to be skipped")
	}
}

func TestShouldTriggerNowInvalidCron(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Minute)
	_, err := shouldTriggerNow("invalid cron", pgtype.Timestamptz{}, now)
	if err == nil {
		t.Fatalf("expected parse error for invalid cron expression")
	}
}
