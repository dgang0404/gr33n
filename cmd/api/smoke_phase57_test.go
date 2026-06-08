// Phase 57 — per-device Pi API keys (issue → edge auth → revoke → 401).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPhase57_DeviceAPIKeyIssueAuthRevoke(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
	issueResp := authPost(t, tok, fmt.Sprintf("/devices/%d/api-keys", deviceID), map[string]any{
		"label": "phase57 smoke",
	})
	expectStatus(t, issueResp, http.StatusCreated)
	issueBody := decodeMap(t, issueResp)
	issueResp.Body.Close()

	apiKey, _ := issueBody["api_key"].(string)
	if apiKey == "" || !bytes.HasPrefix([]byte(apiKey), []byte("gdev_")) {
		t.Fatalf("api_key = %q want gdev_* prefix", apiKey)
	}
	keyID := int64(issueBody["key"].(map[string]any)["id"].(float64))

	patchBody, _ := json.Marshal(map[string]string{"status": "online"})
	patchOK := piPatchWithDeviceKey(t, deviceID, patchBody, apiKey)
	expectStatus(t, patchOK, http.StatusOK)
	patchOK.Body.Close()

	revokeResp := authPost(t, tok, fmt.Sprintf("/devices/%d/api-keys/%d/revoke", deviceID, keyID), nil)
	expectStatus(t, revokeResp, http.StatusOK)
	revokeResp.Body.Close()

	patchDenied := piPatchWithDeviceKey(t, deviceID, patchBody, apiKey)
	expectStatus(t, patchDenied, http.StatusForbidden)
	patchDenied.Body.Close()
}

func piPatchWithDeviceKey(t *testing.T, deviceID int64, body []byte, deviceKey string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPatch, testServer.URL+fmt.Sprintf("/devices/%d/status", deviceID), bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-Key", deviceKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /devices/%d/status: %v", deviceID, err)
	}
	return resp
}
