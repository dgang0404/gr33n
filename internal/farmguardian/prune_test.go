package farmguardian

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestLoadPruneConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv("CHAT_SESSION_TTL_DAYS", "")
	t.Setenv("CHAT_SESSION_PRUNE_INTERVAL_HOURS", "")
	t.Setenv("CHAT_SESSION_PRUNE_STARTUP_DELAY_SECONDS", "")
	cfg := LoadPruneConfigFromEnv()
	if cfg.TTLDays != DefaultPruneTTLDays {
		t.Fatalf("TTLDays default = %d, want %d", cfg.TTLDays, DefaultPruneTTLDays)
	}
	if cfg.Interval != time.Duration(DefaultPruneIntervalHours)*time.Hour {
		t.Fatalf("Interval default = %v, want %dh", cfg.Interval, DefaultPruneIntervalHours)
	}
	if cfg.StartupDelay != time.Duration(DefaultPruneStartupDelaySec)*time.Second {
		t.Fatalf("StartupDelay default = %v, want %ds", cfg.StartupDelay, DefaultPruneStartupDelaySec)
	}
	if !cfg.Enabled() {
		t.Fatalf("default config should be Enabled")
	}
}

func TestLoadPruneConfigFromEnv_DisabledByTTLZero(t *testing.T) {
	t.Setenv("CHAT_SESSION_TTL_DAYS", "0")
	cfg := LoadPruneConfigFromEnv()
	if cfg.TTLDays != 0 {
		t.Fatalf("TTLDays = %d, want 0", cfg.TTLDays)
	}
	if cfg.Enabled() {
		t.Fatalf("Enabled() must be false when TTLDays=0")
	}
}

func TestLoadPruneConfigFromEnv_Clamps(t *testing.T) {
	t.Setenv("CHAT_SESSION_TTL_DAYS", "99999")
	t.Setenv("CHAT_SESSION_PRUNE_INTERVAL_HOURS", "9999")
	t.Setenv("CHAT_SESSION_PRUNE_STARTUP_DELAY_SECONDS", "99999")
	cfg := LoadPruneConfigFromEnv()
	if cfg.TTLDays != MaxPruneTTLDays {
		t.Fatalf("TTLDays clamp = %d, want %d", cfg.TTLDays, MaxPruneTTLDays)
	}
	if cfg.Interval != time.Duration(MaxPruneIntervalHours)*time.Hour {
		t.Fatalf("Interval clamp = %v, want %dh", cfg.Interval, MaxPruneIntervalHours)
	}
	if cfg.StartupDelay != 600*time.Second {
		t.Fatalf("StartupDelay clamp = %v, want 600s", cfg.StartupDelay)
	}
}

func TestLoadPruneConfigFromEnv_GarbageFallsBack(t *testing.T) {
	t.Setenv("CHAT_SESSION_TTL_DAYS", "abc")
	t.Setenv("CHAT_SESSION_PRUNE_INTERVAL_HOURS", "-1")
	t.Setenv("CHAT_SESSION_PRUNE_STARTUP_DELAY_SECONDS", "nope")
	cfg := LoadPruneConfigFromEnv()
	if cfg.TTLDays != DefaultPruneTTLDays {
		t.Fatalf("TTLDays garbage -> default; got %d", cfg.TTLDays)
	}
	if cfg.Interval != time.Duration(DefaultPruneIntervalHours)*time.Hour {
		t.Fatalf("Interval garbage -> default; got %v", cfg.Interval)
	}
}

// fakeQ is a prunableQuerier double for the PruneOnce / StartPruneLoop tests.
type fakeQ struct {
	turnsByCall    []int64
	sessionsByCall []int64
	turnsErr       error
	sessionsErr    error
	calls          int
	lastCutoff     time.Time
}

func (f *fakeQ) DeleteStaleConversationTurns(_ context.Context, cutoff time.Time) (int64, error) {
	f.lastCutoff = cutoff
	if f.turnsErr != nil {
		return 0, f.turnsErr
	}
	if f.calls < len(f.turnsByCall) {
		v := f.turnsByCall[f.calls]
		// don't advance calls here; sessions side does that
		return v, nil
	}
	return 0, nil
}

func (f *fakeQ) DeleteStaleConversationSessions(_ context.Context, cutoff time.Time) (int64, error) {
	if f.sessionsErr != nil {
		return 0, f.sessionsErr
	}
	v := int64(0)
	if f.calls < len(f.sessionsByCall) {
		v = f.sessionsByCall[f.calls]
	}
	f.calls++
	return v, nil
}

func TestPruneOnce_HappyPath(t *testing.T) {
	q := &fakeQ{turnsByCall: []int64{7}, sessionsByCall: []int64{3}}
	ttl := 30 * 24 * time.Hour
	before := time.Now()
	res, err := PruneOnce(context.Background(), q, ttl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TurnsDeleted != 7 || res.SessionsDeleted != 3 {
		t.Fatalf("counts: %+v", res)
	}
	// Cutoff must be ttl behind now (±1s for the timestamp captured inside PruneOnce).
	want := before.Add(-ttl)
	if res.Cutoff.Before(want.Add(-time.Second)) || res.Cutoff.After(want.Add(time.Second)) {
		t.Fatalf("cutoff drift: got %v, want ~%v", res.Cutoff, want)
	}
}

func TestPruneOnce_TurnsErrShortCircuits(t *testing.T) {
	q := &fakeQ{turnsErr: errors.New("db down"), sessionsByCall: []int64{99}}
	_, err := PruneOnce(context.Background(), q, 1*time.Hour)
	if err == nil {
		t.Fatalf("expected error from turns pass")
	}
	if q.calls != 0 {
		t.Fatalf("sessions pass must be skipped when turns fail, calls=%d", q.calls)
	}
}

func TestStartPruneLoop_Disabled(t *testing.T) {
	q := &fakeQ{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// TTLDays=0 means disabled; should return immediately.
	done := make(chan struct{})
	go func() {
		StartPruneLoop(context.Background(), q, PruneConfig{TTLDays: 0, Interval: time.Hour, StartupDelay: 0}, logger)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("StartPruneLoop with TTLDays=0 should return immediately")
	}
	if q.calls != 0 {
		t.Fatalf("disabled loop must not query DB; calls=%d", q.calls)
	}
}

func TestStartPruneLoop_RespectsContextDuringStartupDelay(t *testing.T) {
	q := &fakeQ{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(context.Background())
	cfg := PruneConfig{TTLDays: 30, Interval: time.Hour, StartupDelay: 5 * time.Second}
	done := make(chan struct{})
	go func() {
		StartPruneLoop(ctx, q, cfg, logger)
		close(done)
	}()
	cancel() // cancel before the 5s startup delay elapses
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("StartPruneLoop should exit promptly when ctx cancels during startup delay")
	}
	if q.calls != 0 {
		t.Fatalf("loop must not call DB after cancellation; calls=%d", q.calls)
	}
}

func TestStartPruneLoop_RunsOpeningPruneThenStops(t *testing.T) {
	q := &fakeQ{turnsByCall: []int64{2, 0}, sessionsByCall: []int64{1, 0}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(context.Background())
	cfg := PruneConfig{TTLDays: 30, Interval: time.Hour, StartupDelay: 0}
	done := make(chan struct{})
	go func() {
		StartPruneLoop(ctx, q, cfg, logger)
		close(done)
	}()
	// Give the opening prune a moment to run, then cancel before the ticker would fire.
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("StartPruneLoop should exit after ctx cancel")
	}
	if q.calls != 1 {
		t.Fatalf("expected 1 prune call (opening), got %d", q.calls)
	}
}
