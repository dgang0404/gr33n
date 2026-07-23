//go:build dev

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"gr33n-api/internal/authctx"
)

func TestDevBypassContext_populatesUserIDFromBearer(t *testing.T) {
	prev := authMode
	authMode = "dev"
	t.Cleanup(func() { authMode = prev })
	jwtSecret = []byte("test-secret")
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tok, err := IssueToken("admin", time.Hour, map[string]any{
		"user_id": uid.String(),
		"email":   "dev@gr33n.local",
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/proposals", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	ctx := devBypassContext(req)
	got, ok := authctx.UserID(ctx)
	if !ok || got != uid {
		t.Fatalf("user_id missing in dev bypass context: ok=%v got=%v", ok, got)
	}
}
