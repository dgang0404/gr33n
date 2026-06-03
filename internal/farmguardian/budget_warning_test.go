// Phase 28 WS5 — unit coverage for the chat-budget-warning hook.
// SQL-touching paths are validated in
// cmd/api/smoke_phase28_ws5_test.go; this file focuses on the decision
// logic (threshold + debounce + farm-required gate + best-effort
// error-swallowing) with a hand-rolled in-memory fake querier.

package farmguardian

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

type fakeWarningQuerier struct {
	totals         db.SumChatTokensSinceForUserRow
	totalsErr      error
	existing       bool // GetRecent returns id=1 when true
	existingErr    error
	created        []db.CreateAlertParams
	createErr      error
	nextCreatedID  int64
	sumCalls       int
	getRecentCalls int
	createAlertCnt int
}

func (f *fakeWarningQuerier) SumChatTokensSinceForUser(_ context.Context, _ db.SumChatTokensSinceForUserParams) (db.SumChatTokensSinceForUserRow, error) {
	f.sumCalls++
	return f.totals, f.totalsErr
}

func (f *fakeWarningQuerier) GetRecentChatBudgetWarningForUser(_ context.Context, _ db.GetRecentChatBudgetWarningForUserParams) (int64, error) {
	f.getRecentCalls++
	if f.existingErr != nil {
		return 0, f.existingErr
	}
	if f.existing {
		return 1, nil
	}
	return 0, pgx.ErrNoRows
}

func (f *fakeWarningQuerier) CreateAlert(_ context.Context, arg db.CreateAlertParams) (db.Gr33ncoreAlertsNotification, error) {
	f.createAlertCnt++
	if f.createErr != nil {
		return db.Gr33ncoreAlertsNotification{}, f.createErr
	}
	f.created = append(f.created, arg)
	f.nextCreatedID++
	return db.Gr33ncoreAlertsNotification{ID: f.nextCreatedID}, nil
}

func makeUser(t *testing.T) uuid.UUID {
	t.Helper()
	u, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("uuid: %v", err)
	}
	return u
}

func TestMaybeFireBudgetWarning_NoCapNoOp(t *testing.T) {
	f := &fakeWarningQuerier{}
	cfg := CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 0}
	res, err := MaybeFireBudgetWarning(context.Background(), f, cfg, makeUser(t), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Fired {
		t.Fatalf("must not fire when PerUserMaxTokens=0")
	}
	if f.sumCalls != 0 || f.createAlertCnt != 0 {
		t.Fatalf("must not touch DB when uncapped: sums=%d creates=%d", f.sumCalls, f.createAlertCnt)
	}
}

func TestMaybeFireBudgetWarning_NoFarmNoOp(t *testing.T) {
	f := &fakeWarningQuerier{}
	cfg := CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 1000}
	res, err := MaybeFireBudgetWarning(context.Background(), f, cfg, makeUser(t), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Fired {
		t.Fatalf("must not fire without a farm_id")
	}
	if f.sumCalls != 0 {
		t.Fatalf("must not query DB without farm: sums=%d", f.sumCalls)
	}
}

func TestMaybeFireBudgetWarning_BelowThresholdNoOp(t *testing.T) {
	f := &fakeWarningQuerier{
		totals: db.SumChatTokensSinceForUserRow{TotalTokens: 500},
	}
	cfg := CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 1000}
	res, err := MaybeFireBudgetWarning(context.Background(), f, cfg, makeUser(t), 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Fired {
		t.Fatalf("must not fire at 50%%")
	}
	if res.PctUsed != 0.5 || res.UsedTokens != 500 {
		t.Fatalf("unexpected: %+v", res)
	}
	if f.getRecentCalls != 0 || f.createAlertCnt != 0 {
		t.Fatalf("debounce + create should not be reached below threshold")
	}
}

func TestMaybeFireBudgetWarning_AboveThresholdFires(t *testing.T) {
	f := &fakeWarningQuerier{
		totals: db.SumChatTokensSinceForUserRow{TotalTokens: 850},
	}
	cfg := CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 1000}
	u := makeUser(t)
	res, err := MaybeFireBudgetWarning(context.Background(), f, cfg, u, 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !res.Fired {
		t.Fatalf("expected fire at 85%%")
	}
	if res.AlertID != 1 || res.PctUsed != 0.85 {
		t.Fatalf("unexpected: %+v", res)
	}
	if f.createAlertCnt != 1 {
		t.Fatalf("expected 1 CreateAlert call, got %d", f.createAlertCnt)
	}
	got := f.created[0]
	if got.FarmID != 7 {
		t.Errorf("expected farm_id=7, got %d", got.FarmID)
	}
	if !got.RecipientUserID.Valid {
		t.Errorf("expected recipient_user_id valid")
	}
	if got.TriggeringEventSourceType == nil || *got.TriggeringEventSourceType != ChatBudgetWarningSourceType {
		t.Errorf("expected source_type=%q, got %v", ChatBudgetWarningSourceType, got.TriggeringEventSourceType)
	}
	if !got.Severity.Valid || got.Severity.Gr33ncoreNotificationPriorityEnum != db.Gr33ncoreNotificationPriorityEnumMedium {
		t.Errorf("expected severity=medium, got %+v", got.Severity)
	}
	if got.SubjectRendered == nil || *got.SubjectRendered != "Chat token budget at 85%" {
		t.Errorf("unexpected subject: %v", got.SubjectRendered)
	}

	// Verify the recipient bytes round-trip from uuid → pgtype.UUID.
	var rid pgtype.UUID = got.RecipientUserID
	if rid.Bytes != [16]byte(u) {
		t.Errorf("recipient uuid mismatch: got %x want %x", rid.Bytes, [16]byte(u))
	}
}

func TestMaybeFireBudgetWarning_DebounceHit(t *testing.T) {
	f := &fakeWarningQuerier{
		totals:   db.SumChatTokensSinceForUserRow{TotalTokens: 900},
		existing: true,
	}
	cfg := CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 1000}
	res, err := MaybeFireBudgetWarning(context.Background(), f, cfg, makeUser(t), 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Fired {
		t.Fatalf("must not fire when an existing in-window warning is present")
	}
	if f.createAlertCnt != 0 {
		t.Fatalf("must not create when debounced")
	}
}

func TestMaybeFireBudgetWarning_DebounceLookupErrorFailsClosed(t *testing.T) {
	// If the debounce query itself fails, we'd rather skip the warning
	// (fail closed) than risk spamming on transient errors.
	f := &fakeWarningQuerier{
		totals:      db.SumChatTokensSinceForUserRow{TotalTokens: 900},
		existingErr: errors.New("boom"),
	}
	cfg := CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 1000}
	res, err := MaybeFireBudgetWarning(context.Background(), f, cfg, makeUser(t), 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Fired {
		t.Fatalf("must not fire when debounce lookup errors")
	}
}

func TestMaybeFireBudgetWarning_SumErrorFailsOpen(t *testing.T) {
	// If the SUM query itself errors, return Fired=false with no
	// error — chat must keep flowing. The cost guard will catch the
	// user at the hard cap on the next turn.
	f := &fakeWarningQuerier{
		totalsErr: errors.New("boom"),
	}
	cfg := CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 1000}
	res, err := MaybeFireBudgetWarning(context.Background(), f, cfg, makeUser(t), 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Fired {
		t.Fatalf("must not claim Fired when SUM errored")
	}
}

func TestMaybeFireBudgetWarning_CreateAlertErrorFailsOpen(t *testing.T) {
	f := &fakeWarningQuerier{
		totals:    db.SumChatTokensSinceForUserRow{TotalTokens: 950},
		createErr: errors.New("alert insert boom"),
	}
	cfg := CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 1000}
	res, err := MaybeFireBudgetWarning(context.Background(), f, cfg, makeUser(t), 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Fired {
		t.Fatalf("must not claim Fired when CreateAlert errored")
	}
	if res.PctUsed != 0.95 {
		t.Fatalf("PctUsed still computed for telemetry, got %v", res.PctUsed)
	}
}

func TestMaybeFireBudgetWarning_NilQuerierIsProgrammerError(t *testing.T) {
	// Unlike the transient errors above (which are best-effort), a nil
	// queries handle is a programmer mistake — we want tests to fail
	// loud rather than silently no-op.
	cfg := CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 1000}
	_, err := MaybeFireBudgetWarning(context.Background(), nil, cfg, makeUser(t), 7)
	if err == nil {
		t.Fatal("nil querier should return an error")
	}
}
