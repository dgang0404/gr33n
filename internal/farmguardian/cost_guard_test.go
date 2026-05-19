package farmguardian

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
)

func TestLoadCostGuardConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv("CHAT_COST_WINDOW_HOURS", "")
	t.Setenv("CHAT_COST_MAX_TOKENS_PER_USER", "")
	t.Setenv("CHAT_COST_MAX_TOKENS_PER_FARM", "")

	cfg := LoadCostGuardConfigFromEnv()
	if got, want := cfg.Window, time.Hour; got != want {
		t.Fatalf("window default: got %s want %s", got, want)
	}
	if cfg.PerUserMaxTokens != 0 {
		t.Fatalf("per-user default = %d, want 0 (disabled)", cfg.PerUserMaxTokens)
	}
	if cfg.PerFarmMaxTokens != 0 {
		t.Fatalf("per-farm default = %d, want 0 (disabled)", cfg.PerFarmMaxTokens)
	}
	if cfg.AnyEnabled() {
		t.Fatalf("AnyEnabled should be false when both caps are 0")
	}
}

func TestLoadCostGuardConfigFromEnv_ParsesAndClamps(t *testing.T) {
	t.Setenv("CHAT_COST_WINDOW_HOURS", "6")
	t.Setenv("CHAT_COST_MAX_TOKENS_PER_USER", "50000")
	t.Setenv("CHAT_COST_MAX_TOKENS_PER_FARM", "100000")

	cfg := LoadCostGuardConfigFromEnv()
	if got, want := cfg.Window, 6*time.Hour; got != want {
		t.Fatalf("window: got %s want %s", got, want)
	}
	if cfg.PerUserMaxTokens != 50000 {
		t.Fatalf("per-user: got %d want 50000", cfg.PerUserMaxTokens)
	}
	if cfg.PerFarmMaxTokens != 100000 {
		t.Fatalf("per-farm: got %d want 100000", cfg.PerFarmMaxTokens)
	}
	if !cfg.AnyEnabled() {
		t.Fatalf("AnyEnabled should be true once caps are set")
	}

	t.Setenv("CHAT_COST_WINDOW_HOURS", "9999")
	t.Setenv("CHAT_COST_MAX_TOKENS_PER_USER", "999999999999")
	cfg2 := LoadCostGuardConfigFromEnv()
	if cfg2.Window != time.Duration(MaxCostWindowHours)*time.Hour {
		t.Fatalf("window clamp: got %s want %s", cfg2.Window, time.Duration(MaxCostWindowHours)*time.Hour)
	}
	if cfg2.PerUserMaxTokens != MaxCostMaxTokens {
		t.Fatalf("per-user clamp: got %d want %d", cfg2.PerUserMaxTokens, MaxCostMaxTokens)
	}

	t.Setenv("CHAT_COST_WINDOW_HOURS", "not-a-number")
	t.Setenv("CHAT_COST_MAX_TOKENS_PER_USER", "-5")
	t.Setenv("CHAT_COST_MAX_TOKENS_PER_FARM", "garbage")
	cfg3 := LoadCostGuardConfigFromEnv()
	if cfg3.Window != time.Hour {
		t.Fatalf("garbage window should fall back to 1h, got %s", cfg3.Window)
	}
	if cfg3.PerUserMaxTokens != 0 || cfg3.PerFarmMaxTokens != 0 {
		t.Fatalf("garbage caps should fall back to 0, got %d / %d", cfg3.PerUserMaxTokens, cfg3.PerFarmMaxTokens)
	}
}

// fakeCostQuerier lets the unit tests stub out the two SUM queries without
// reaching into a real DB.
type fakeCostQuerier struct {
	userTotals  db.ChatTokenTotals
	farmTotals  db.ChatTokenTotals
	userErr     error
	farmErr     error
	userCalls   int
	farmCalls   int
	lastUserID  uuid.UUID
	lastFarmID  int64
	lastUserSin time.Time
	lastFarmSin time.Time
}

func (f *fakeCostQuerier) SumChatTokensSinceForUser(_ context.Context, userID uuid.UUID, since time.Time) (db.ChatTokenTotals, error) {
	f.userCalls++
	f.lastUserID = userID
	f.lastUserSin = since
	return f.userTotals, f.userErr
}

func (f *fakeCostQuerier) SumChatTokensSinceForFarm(_ context.Context, farmID int64, since time.Time) (db.ChatTokenTotals, error) {
	f.farmCalls++
	f.lastFarmID = farmID
	f.lastFarmSin = since
	return f.farmTotals, f.farmErr
}

func TestCheckBudget_DisabledSkipsDB(t *testing.T) {
	fq := &fakeCostQuerier{}
	d, err := CheckBudget(context.Background(), fq, CostGuardConfig{}, uuid.New(), 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !d.Allowed {
		t.Fatalf("expected Allowed when no caps configured")
	}
	if fq.userCalls != 0 || fq.farmCalls != 0 {
		t.Fatalf("DB should not be queried when disabled: user=%d farm=%d", fq.userCalls, fq.farmCalls)
	}
}

func TestCheckBudget_PerUserOver(t *testing.T) {
	fq := &fakeCostQuerier{
		userTotals: db.ChatTokenTotals{TotalTokens: 5_001},
	}
	cfg := CostGuardConfig{
		Window:           time.Hour,
		PerUserMaxTokens: 5_000,
	}
	userID := uuid.New()
	d, err := CheckBudget(context.Background(), fq, cfg, userID, 42)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if d.Allowed {
		t.Fatalf("expected NOT allowed (over per-user cap)")
	}
	if d.Reason != "per_user" {
		t.Fatalf("reason = %q want per_user", d.Reason)
	}
	if d.UsedTokens != 5_001 || d.MaxTokens != 5_000 {
		t.Fatalf("counters: used=%d max=%d", d.UsedTokens, d.MaxTokens)
	}
	if d.RetryAfter != time.Hour {
		t.Fatalf("retry-after = %s want 1h", d.RetryAfter)
	}
	if d.WindowSeconds != 3600 {
		t.Fatalf("window seconds = %d want 3600", d.WindowSeconds)
	}
	if fq.farmCalls != 0 {
		t.Fatalf("farm query should be skipped when user cap fires first")
	}
	if fq.lastUserID != userID {
		t.Fatalf("user id propagation: got %s want %s", fq.lastUserID, userID)
	}
	if time.Since(fq.lastUserSin) < 50*time.Minute || time.Since(fq.lastUserSin) > 70*time.Minute {
		t.Fatalf("since should be ~1h ago, got %s", time.Since(fq.lastUserSin))
	}
}

func TestCheckBudget_PerFarmOver(t *testing.T) {
	fq := &fakeCostQuerier{
		userTotals: db.ChatTokenTotals{TotalTokens: 1_000}, // under cap
		farmTotals: db.ChatTokenTotals{TotalTokens: 75_000},
	}
	cfg := CostGuardConfig{
		Window:           2 * time.Hour,
		PerUserMaxTokens: 5_000,
		PerFarmMaxTokens: 50_000,
	}
	d, err := CheckBudget(context.Background(), fq, cfg, uuid.New(), 99)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if d.Allowed {
		t.Fatalf("expected NOT allowed (over per-farm cap)")
	}
	if d.Reason != "per_farm" {
		t.Fatalf("reason = %q want per_farm", d.Reason)
	}
	if d.UsedTokens != 75_000 || d.MaxTokens != 50_000 {
		t.Fatalf("counters: used=%d max=%d", d.UsedTokens, d.MaxTokens)
	}
	if fq.lastFarmID != 99 {
		t.Fatalf("farm id propagation: got %d want 99", fq.lastFarmID)
	}
}

func TestCheckBudget_PerFarmSkippedWithoutFarm(t *testing.T) {
	fq := &fakeCostQuerier{
		farmTotals: db.ChatTokenTotals{TotalTokens: 999_999}, // would blow farm cap...
	}
	cfg := CostGuardConfig{
		Window:           time.Hour,
		PerFarmMaxTokens: 1_000,
	}
	d, err := CheckBudget(context.Background(), fq, cfg, uuid.New(), 0) // no farm
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !d.Allowed {
		t.Fatalf("farm cap must NOT apply when farmID is 0")
	}
	if fq.farmCalls != 0 {
		t.Fatalf("farm query should be skipped when farmID=0")
	}
}

func TestCheckBudget_AllowedBelowCap(t *testing.T) {
	fq := &fakeCostQuerier{
		userTotals: db.ChatTokenTotals{TotalTokens: 200},
		farmTotals: db.ChatTokenTotals{TotalTokens: 300},
	}
	cfg := CostGuardConfig{
		Window:           time.Hour,
		PerUserMaxTokens: 1_000,
		PerFarmMaxTokens: 1_000,
	}
	d, err := CheckBudget(context.Background(), fq, cfg, uuid.New(), 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !d.Allowed {
		t.Fatalf("expected Allowed; used user=200 farm=300, cap 1000")
	}
	if fq.userCalls != 1 || fq.farmCalls != 1 {
		t.Fatalf("both queries should run when both caps are configured: user=%d farm=%d", fq.userCalls, fq.farmCalls)
	}
}

func TestCheckBudget_PropagatesUserQueryError(t *testing.T) {
	wantErr := errors.New("boom")
	fq := &fakeCostQuerier{userErr: wantErr}
	cfg := CostGuardConfig{
		Window:           time.Hour,
		PerUserMaxTokens: 100,
	}
	_, err := CheckBudget(context.Background(), fq, cfg, uuid.New(), 0)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v want %v", err, wantErr)
	}
}
