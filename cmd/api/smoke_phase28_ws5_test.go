// Phase 28 WS5 — token-usage endpoint + budget-warning hook smoke
// tests. Drives `GET /v1/chat/usage` against the live server and
// exercises `MaybeFireBudgetWarning` against real Postgres so the
// SQL bindings (SumChatTokensSinceForUser, SumChatTokensSinceForFarm,
// GetRecentChatBudgetWarningForUser, CreateAlert) are validated as a
// single surface.

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

// ─── /v1/chat/usage endpoint ─────────────────────────────────────────────

func TestPhase28WS5_GetUsage_UserDimension(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/v1/chat/usage")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)

	if got, ok := body["ai_enabled"].(bool); !ok || !got {
		t.Fatalf("expected ai_enabled=true, got %v", body["ai_enabled"])
	}
	if _, ok := body["window_hours"].(float64); !ok {
		t.Fatalf("expected window_hours numeric, got %v", body["window_hours"])
	}
	user, ok := body["user"].(map[string]any)
	if !ok {
		t.Fatalf("expected user object, got %v", body["user"])
	}
	for _, k := range []string{"used_tokens", "max_tokens", "remaining_tokens", "pct_used"} {
		if _, ok := user[k]; !ok {
			t.Fatalf("user dimension missing %q in %v", k, user)
		}
	}
	// farm dimension must be absent when no farm_id query param.
	if _, present := body["farm"]; present {
		t.Fatalf("farm dimension should be absent without ?farm_id")
	}
}

func TestPhase28WS5_GetUsage_FarmDimension(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/chat/usage?farm_id=1")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	farm, ok := body["farm"].(map[string]any)
	if !ok {
		t.Fatalf("expected farm object with farm_id=1, got %v", body["farm"])
	}
	if id, _ := farm["farm_id"].(float64); int64(id) != 1 {
		t.Fatalf("expected farm_id=1, got %v", farm["farm_id"])
	}
	for _, k := range []string{"used_tokens", "max_tokens", "remaining_tokens", "pct_used"} {
		if _, ok := farm[k]; !ok {
			t.Fatalf("farm dimension missing %q in %v", k, farm)
		}
	}
}

func TestPhase28WS5_GetUsage_RejectsInvalidFarmID(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/chat/usage?farm_id=not-a-number")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusBadRequest)
}

func TestPhase28WS5_GetUsage_RejectsForeignFarm(t *testing.T) {
	// farm_id 99999 doesn't exist — the farm-member check returns a
	// 403/404 depending on the gate; we just assert it isn't 200.
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/chat/usage?farm_id=99999")
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected non-200 for non-existent farm, got %d", resp.StatusCode)
	}
}

func TestPhase28WS5_GetUsage_RequiresAuth(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/v1/chat/usage")
	if err != nil {
		t.Fatalf("GET /v1/chat/usage (no auth): %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without JWT, got %d", resp.StatusCode)
	}
}

func TestPhase28WS5_GetUsage_ResponseShapeWithConfiguredCaps(t *testing.T) {
	// The smoke harness loads cost-guard config from env at boot, so
	// `pct_used` will be 0 unless caps were exported before TestMain
	// ran. We don't try to mutate config here — that would require
	// rebooting the harness — and instead assert the shape contract.
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/v1/chat/usage")
	defer resp.Body.Close()
	expectStatus(t, resp, http.StatusOK)
	var typed struct {
		WindowHours int  `json:"window_hours"`
		AIEnabled   bool `json:"ai_enabled"`
		User        struct {
			UsedTokens          int64    `json:"used_tokens"`
			MaxTokens           int64    `json:"max_tokens"`
			RemainingTokens     int64    `json:"remaining_tokens"`
			PctUsed             float64  `json:"pct_used"`
			WarningThresholdPct *float64 `json:"warning_threshold_pct,omitempty"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&typed); err != nil {
		t.Fatalf("decode typed response: %v", err)
	}
	if typed.WindowHours < 1 {
		t.Errorf("expected window_hours >= 1, got %d", typed.WindowHours)
	}
	if typed.User.PctUsed < 0 || typed.User.PctUsed > 1 {
		t.Errorf("pct_used out of [0,1]: %f", typed.User.PctUsed)
	}
	if typed.User.RemainingTokens < 0 {
		t.Errorf("remaining_tokens must be non-negative, got %d", typed.User.RemainingTokens)
	}
}

// ─── MaybeFireBudgetWarning (real DB) ───────────────────────────────────

func TestPhase28WS5_BudgetWarning_FiresAndDebounces(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}

	// Use a fresh user so we don't conflict with the shared smoke
	// user's accumulated turns/alerts. Insert a profile row so the
	// FK on alerts_notifications.recipient_user_id resolves.
	userID := uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := testPool.Exec(ctx, `
INSERT INTO auth.users (id, email, password_hash, created_at)
VALUES ($1, $2, 'x', NOW())`, userID, "ws5_"+userID.String()+"@test.local"); err != nil {
		t.Fatalf("seed auth.users: %v", err)
	}
	email := "ws5_" + userID.String() + "@test.local"
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.profiles (user_id, full_name, email, hourly_rate, created_at, updated_at)
VALUES ($1, 'WS5 Tester', $2, 0, NOW(), NOW())`, userID, email); err != nil {
		t.Fatalf("seed profiles: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(ctx,
			`DELETE FROM gr33ncore.alerts_notifications WHERE recipient_user_id = $1`, userID)
		_, _ = testPool.Exec(ctx,
			`DELETE FROM gr33ncore.conversation_turns WHERE user_id = $1`, userID)
		_, _ = testPool.Exec(ctx,
			`DELETE FROM gr33ncore.conversation_sessions WHERE user_id = $1`, userID)
		_, _ = testPool.Exec(ctx,
			`DELETE FROM gr33ncore.profiles WHERE user_id = $1`, userID)
		_, _ = testPool.Exec(ctx,
			`DELETE FROM auth.users WHERE id = $1`, userID)
	})

	// Seed a conversation_turn with token usage that pushes the user
	// to 95% of a 1000-token cap. Schema requires a valid session row
	// + session_id (uuid) for FK / sidebar tracking.
	sessionID := uuid.New()
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_sessions
    (id, user_id, title, created_at, updated_at)
VALUES ($1, $2, 'ws5-session', NOW(), NOW())`, sessionID, userID); err != nil {
		t.Fatalf("seed conversation_sessions: %v", err)
	}
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.conversation_turns
    (session_id, user_id, farm_id, turn_index,
     user_message, assistant_message, llm_model,
     grounded, context_count, citations,
     prompt_tokens, completion_tokens,
     created_at)
VALUES ($1, $2, 1, 0,
        'why is it humid', 'because the dehumidifier is off', 'llama3.1:70b',
        true, 0, '[]',
        $3, $4,
        NOW())`,
		sessionID, userID, 600, 350); err != nil {
		t.Fatalf("seed conversation_turns: %v", err)
	}

	queries := db.New(testPool)
	cfg := farmguardian.CostGuardConfig{
		Window:           time.Hour,
		PerUserMaxTokens: 1000, // user just burned 950 → 95%
	}

	// First call: must fire.
	first, err := farmguardian.MaybeFireBudgetWarning(ctx, queries, cfg, userID, 1)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if !first.Fired {
		t.Fatalf("first call must fire at 95%%; got %+v", first)
	}
	if first.AlertID == 0 {
		t.Errorf("expected non-zero AlertID, got %d", first.AlertID)
	}
	if first.PctUsed < 0.94 || first.PctUsed > 0.96 {
		t.Errorf("expected PctUsed ~0.95, got %v", first.PctUsed)
	}

	// Second call within the same window: debounce hit, must not fire.
	second, err := farmguardian.MaybeFireBudgetWarning(ctx, queries, cfg, userID, 1)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if second.Fired {
		t.Fatalf("second call should debounce, got %+v", second)
	}

	// Verify exactly one alert row exists for this user with the right
	// source_type + recipient_user_id.
	var count int
	if err := testPool.QueryRow(ctx, `
SELECT COUNT(*) FROM gr33ncore.alerts_notifications
WHERE recipient_user_id = $1
  AND triggering_event_source_type = $2`,
		userID, farmguardian.ChatBudgetWarningSourceType,
	).Scan(&count); err != nil {
		t.Fatalf("count alerts: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 warning alert, got %d", count)
	}

	// Verify shape of the inserted row.
	var subject, srcType string
	var severity string
	if err := testPool.QueryRow(ctx, `
SELECT subject_rendered, triggering_event_source_type, severity::text
FROM gr33ncore.alerts_notifications
WHERE recipient_user_id = $1
  AND triggering_event_source_type = $2`,
		userID, farmguardian.ChatBudgetWarningSourceType,
	).Scan(&subject, &srcType, &severity); err != nil {
		t.Fatalf("read alert: %v", err)
	}
	if subject != "Chat token budget at 95%" {
		t.Errorf("unexpected subject %q", subject)
	}
	if srcType != "chat_budget_warning" {
		t.Errorf("unexpected source type %q", srcType)
	}
	if severity != "medium" {
		t.Errorf("expected severity=medium, got %q", severity)
	}
}

func TestPhase28WS5_BudgetWarning_BelowThresholdNoAlertRow(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	userID := uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := testPool.Exec(ctx, `
INSERT INTO auth.users (id, email, password_hash, created_at)
VALUES ($1, $2, 'x', NOW())`, userID, "ws5lo_"+userID.String()+"@test.local"); err != nil {
		t.Fatalf("seed auth.users: %v", err)
	}
	loEmail := "ws5lo_" + userID.String() + "@test.local"
	if _, err := testPool.Exec(ctx, `
INSERT INTO gr33ncore.profiles (user_id, full_name, email, hourly_rate, created_at, updated_at)
VALUES ($1, 'WS5 Lo', $2, 0, NOW(), NOW())`, userID, loEmail); err != nil {
		t.Fatalf("seed profiles: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = testPool.Exec(ctx, `DELETE FROM gr33ncore.alerts_notifications WHERE recipient_user_id = $1`, userID)
		_, _ = testPool.Exec(ctx, `DELETE FROM gr33ncore.profiles WHERE user_id = $1`, userID)
		_, _ = testPool.Exec(ctx, `DELETE FROM auth.users WHERE id = $1`, userID)
	})

	queries := db.New(testPool)
	cfg := farmguardian.CostGuardConfig{Window: time.Hour, PerUserMaxTokens: 1000}

	res, err := farmguardian.MaybeFireBudgetWarning(ctx, queries, cfg, userID, 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Fired {
		t.Fatalf("must not fire when user has zero usage")
	}

	var count int
	_ = testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.alerts_notifications WHERE recipient_user_id = $1`, userID,
	).Scan(&count)
	if count != 0 {
		t.Fatalf("expected 0 alert rows when below threshold, got %d", count)
	}
}
