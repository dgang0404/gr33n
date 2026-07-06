// Phase 142 — Virtual Pi config export smoke (demo-veg-relay-01).

package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestPhase142_VirtualPiConfigExport(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var deviceID int64
	err := testPool.QueryRow(ctx, `
SELECT id FROM gr33ncore.devices
WHERE farm_id = 1 AND device_uid = 'demo-veg-relay-01' AND deleted_at IS NULL
LIMIT 1`).Scan(&deviceID)
	if err != nil || deviceID == 0 {
		t.Skip("demo-veg-relay-01 missing — run migrations + master_seed.sql")
	}

	tok := smokeJWT(t)
	path := fmt.Sprintf("/devices/%d/pi-config?base_url=http://127.0.0.1:8080", deviceID)
	resp := authGet(t, tok, path)
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	resp.Body.Close()

	yaml, _ := body["yaml"].(string)
	if strings.TrimSpace(yaml) == "" {
		t.Fatalf("expected non-empty yaml, got %q", yaml)
	}
	if !strings.Contains(yaml, "device:") && !strings.Contains(yaml, "api:") {
		t.Fatalf("yaml missing expected keys: %s", truncateForTest(yaml, 200))
	}
	sha, _ := body["config_sha256"].(string)
	if strings.TrimSpace(sha) == "" {
		t.Fatal("expected config_sha256")
	}
}

func truncateForTest(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
