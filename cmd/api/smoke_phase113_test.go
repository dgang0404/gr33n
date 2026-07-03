// Phase 113 — security hardening smokes
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestPhase113_SecurityHeaders(t *testing.T) {
	resp := get(t, "/health")
	defer resp.Body.Close()
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("missing X-Content-Type-Options")
	}
	if resp.Header.Get("X-Frame-Options") != "DENY" {
		t.Fatalf("missing X-Frame-Options")
	}
	if resp.Header.Get("Referrer-Policy") == "" {
		t.Fatalf("missing Referrer-Policy")
	}
}

func TestPhase113_QueryTokenRejectedOnDashboard(t *testing.T) {
	tok := smokeJWT(t)
	resp := get(t, "/farms/1?token="+tok)
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestPhase113_LoginRateLimit(t *testing.T) {
	const attempts = 12
	var last *http.Response
	for i := 0; i < attempts; i++ {
		last = postNoAuth("/auth/login", map[string]any{
			"username": "rate-limit-phase113@smoke.test",
			"password": "wrong-password-phase113",
		})
	}
	if last.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("attempt %d: want 429, got %d", attempts, last.StatusCode)
	}
	if last.Header.Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header on 429")
	}
	last.Body.Close()
}

func TestPhase113_PasswordChangeDBUser(t *testing.T) {
	tok := smokeJWT(t)
	newPass := fmt.Sprintf("phase113-%d", time.Now().UnixNano())
	resp := authPatch(t, tok, "/auth/password", map[string]any{
		"current_password": smokeDevPass,
		"new_password":     newPass,
	})
	expectStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	loginResp := postNoAuth("/auth/login", map[string]any{
		"username": smokeDevEmail,
		"password": newPass,
	})
	expectStatus(t, loginResp, http.StatusOK)
	loginResp.Body.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(smokeDevPass), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testPool.Exec(context.Background(), `UPDATE auth.users SET password_hash = $1 WHERE id = $2`, hash, smokeDevUserUUID); err != nil {
		t.Fatalf("restore password: %v", err)
	}
}

func TestPhase113_InviteCreateAndList(t *testing.T) {
	tok := smokeJWT(t)
	create := authPost(t, tok, "/auth/invites", map[string]any{"ttl_hours": 24})
	expectStatus(t, create, http.StatusCreated)
	body := decodeMap(t, create)
	create.Body.Close()
	code, _ := body["code"].(string)
	if code == "" {
		t.Fatal("missing invite code")
	}

	list := authGet(t, tok, "/auth/invites")
	expectStatus(t, list, http.StatusOK)
	listBody := decodeMap(t, list)
	list.Body.Close()
	invites, _ := listBody["invites"].([]any)
	if len(invites) == 0 {
		t.Fatal("expected at least one invite")
	}
}

func TestPhase113_RegistrationModeOpenInAuthTest(t *testing.T) {
	resp := get(t, "/auth/registration-mode")
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	resp.Body.Close()
	if body["mode"] != "open" {
		t.Fatalf("auth_test default registration mode want open, got %v", body["mode"])
	}
}

func TestPhase113_LegacyPiKeyDisabled(t *testing.T) {
	t.Setenv("PI_LEGACY_KEY_DISABLED", "true")
	resp := piGet(t, "/devices/by-uid/smoke-device/config/version")
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("legacy key disabled want 401/403, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	t.Setenv("PI_LEGACY_KEY_DISABLED", "false")
}

func TestPhase113_UploadSniffRejectsMismatch(t *testing.T) {
	tok := smokeJWT(t)
	resp := authMultipartPost(t, tok, "/farms/1/cost-receipts", "file", "fake.pdf", "application/pdf", []byte("not-a-valid-receipt"), nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400 for unknown content, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
