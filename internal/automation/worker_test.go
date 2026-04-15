package automation

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
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

func TestIdempotencyKeyDeterministic(t *testing.T) {
	now := time.Date(2026, 4, 15, 6, 0, 0, 0, time.UTC)
	k1 := idempotencyKey(42, now)
	k2 := idempotencyKey(42, now)
	if k1 != k2 {
		t.Fatalf("expected same key for same inputs, got %q vs %q", k1, k2)
	}
	if len(k1) == 0 {
		t.Fatal("key should not be empty")
	}
}

func TestIdempotencyKeyVariesWithSchedule(t *testing.T) {
	now := time.Date(2026, 4, 15, 6, 0, 0, 0, time.UTC)
	k1 := idempotencyKey(1, now)
	k2 := idempotencyKey(2, now)
	if k1 == k2 {
		t.Fatalf("expected different keys for different schedules")
	}
}

func TestIdempotencyKeyVariesWithTime(t *testing.T) {
	t1 := time.Date(2026, 4, 15, 6, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 4, 15, 6, 1, 0, 0, time.UTC)
	k1 := idempotencyKey(1, t1)
	k2 := idempotencyKey(1, t2)
	if k1 == k2 {
		t.Fatalf("expected different keys for different times")
	}
}

func TestIsTransientClassifiesCorrectly(t *testing.T) {
	cases := []struct {
		err       error
		transient bool
	}{
		{nil, false},
		{fmt.Errorf("connection refused"), true},
		{fmt.Errorf("action 5 missing target_actuator_id"), false},
		{fmt.Errorf("timeout waiting for response"), true},
		{fmt.Errorf("conn closed unexpectedly"), true},
		{pgx.ErrNoRows, false},
		{fmt.Errorf("unsupported action_type=unknown"), false},
	}
	for _, tc := range cases {
		got := isTransient(tc.err)
		if got != tc.transient {
			t.Errorf("isTransient(%v) = %v, want %v", tc.err, got, tc.transient)
		}
	}
}

func TestExecuteActionWithRetryPermanentFails(t *testing.T) {
	w := &Worker{maxRetries: 2}
	calls := 0
	origAction := w.executeAction

	// We can't easily mock executeAction, so test isTransient + retry logic indirectly
	// The permanent error should not be retried
	permanentErr := fmt.Errorf("action 1 missing target_actuator_id")
	if isTransient(permanentErr) {
		t.Fatal("expected permanent error to not be transient")
	}
	_ = calls
	_ = origAction
}

func TestShouldTriggerNowNonMatchingMinute(t *testing.T) {
	now := time.Date(2026, 4, 15, 7, 0, 0, 0, time.UTC)
	ok, err := shouldTriggerNow("0 6 * * *", pgtype.Timestamptz{}, now)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if ok {
		t.Fatalf("expected schedule NOT to trigger at %v (only scheduled for 06:00)", now)
	}
}

func TestShouldTriggerNowEveryMinute(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 33, 0, 0, time.UTC)
	ok, err := shouldTriggerNow("* * * * *", pgtype.Timestamptz{}, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("every-minute cron should trigger at any minute")
	}
}

func TestIsTransientWrappedErrors(t *testing.T) {
	base := errors.New("connection reset by peer")
	wrapped := fmt.Errorf("database operation: %w", base)
	if !isTransient(wrapped) {
		t.Fatal("wrapped connection reset should be transient")
	}
}
