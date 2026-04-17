// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"net/http"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	resp := get(t, "/health")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body["status"])
	}
}

func TestAuthModeEndpoint(t *testing.T) {
	resp := get(t, "/auth/mode")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["mode"] != "auth_test" {
		t.Fatalf("expected mode=auth_test, got %v", body["mode"])
	}
}

func TestNotificationPreferencesAndPushTokens(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPatch(t, tok, "/profile/notification-preferences", map[string]any{
		"push_enabled": false,
		"min_priority": "medium",
	})
	expectStatus(t, resp, 200)
	resp = authGet(t, tok, "/profile/notification-preferences")
	expectStatus(t, resp, 200)
	m := decodeMap(t, resp)
	if pe, ok := m["push_enabled"].(bool); !ok || pe {
		t.Fatalf("expected push_enabled false after reset, got %+v", m)
	}
	resp = authPatch(t, tok, "/profile/notification-preferences", map[string]any{
		"push_enabled": true,
		"min_priority": "high",
	})
	expectStatus(t, resp, 200)
	m = decodeMap(t, resp)
	if m["push_enabled"] != true || m["min_priority"] != "high" {
		t.Fatalf("patch result %+v", m)
	}
	fakeTok := "smoke-fcm-" + uniqueName("tok")
	resp = authPost(t, tok, "/profile/push-tokens", map[string]any{
		"platform":  "android",
		"fcm_token": fakeTok,
	})
	expectStatus(t, resp, 200)
	resp = authGet(t, tok, "/profile/push-tokens")
	expectStatus(t, resp, 200)
	slice := decodeSlice(t, resp)
	if len(slice) != 1 {
		t.Fatalf("expected 1 push token, got %d", len(slice))
	}
	resp = authDeleteJSON(t, tok, "/profile/push-tokens", map[string]any{"fcm_token": fakeTok})
	expectStatus(t, resp, 204)
}

func TestJWTRequiredForDashboard(t *testing.T) {
	resp := get(t, "/farms/1")
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestLoginBadCredentials(t *testing.T) {
	resp := postNoAuth("/auth/login", map[string]any{
		"username": smokeDevEmail,
		"password": "not-the-password",
	})
	expectStatus(t, resp, http.StatusUnauthorized)
}

// ── Farm + Zone + Sensor Reads ──────────────────────────────────────────────

func TestProfileGetAndUpdate(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/profile")
	expectStatus(t, resp, http.StatusOK)
	profile := decodeMap(t, resp)
	if profile["email"] == nil && profile["user_id"] == nil {
		t.Fatalf("profile missing expected fields: %#v", profile)
	}

	resp = authPut(t, tok, "/profile", map[string]any{
		"full_name": "Smoke Test User",
		"timezone":  "America/New_York",
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["full_name"] != "Smoke Test User" {
		t.Fatalf("expected full_name update, got %v", updated["full_name"])
	}
}
