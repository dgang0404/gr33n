// Phase 105 — crop profile override audit trail smoke.
package main

import (
	"context"
	"testing"
	"time"
)

func TestPhase105_CropOverrideAuditEvents(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	const cropKey = "cannabis"

	del := authDelete(t, tok, "/farms/1/crop-profiles/"+cropKey)
	del.Body.Close()

	get0 := authGet(t, tok, "/farms/1/crop-profiles/"+cropKey)
	expectStatus(t, get0, 200)
	before := decodeMap(t, get0)
	get0.Body.Close()
	stagesBefore, _ := before["stages"].([]any)
	if len(stagesBefore) == 0 {
		t.Fatal("expected cannabis builtin stages")
	}
	first, _ := stagesBefore[0].(map[string]any)
	stageName, _ := first["stage"].(string)

	put := authPut(t, tok, "/farms/1/crop-profiles/"+cropKey, map[string]any{
		"display_name": before["display_name"],
		"stages": []map[string]any{{
			"stage":     stageName,
			"ec_min":    first["ec_min"],
			"ec_max":    8.88,
			"ec_target": first["ec_target"],
		}},
	})
	expectStatus(t, put, 200)
	put.Body.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var kind string
	var catalogVersion int32
	err := testPool.QueryRow(ctx, `
SELECT details->>'kind', (details->>'catalog_version')::int
FROM gr33ncore.user_activity_log
WHERE farm_id = 1
  AND details->>'kind' = 'crop_profile_override_upsert'
  AND details->>'crop_key' = $1
ORDER BY activity_time DESC
LIMIT 1`, cropKey).Scan(&kind, &catalogVersion)
	if err != nil {
		t.Fatalf("upsert audit row: %v", err)
	}
	if catalogVersion < 1 {
		t.Fatalf("catalog_version want >= 1, got %d", catalogVersion)
	}

	resp := authGet(t, tok, "/farms/1/audit-events?limit=20")
	expectStatus(t, resp, 200)
	events := decodeSlice(t, resp)
	resp.Body.Close()
	foundUpsert := false
	for _, raw := range events {
		ev, _ := raw.(map[string]any)
		details, _ := ev["details"].(map[string]any)
		if details["kind"] == "crop_profile_override_upsert" && details["crop_key"] == cropKey {
			foundUpsert = true
			break
		}
	}
	if !foundUpsert {
		t.Fatal("farm audit-events feed missing crop_profile_override_upsert")
	}

	del2 := authDelete(t, tok, "/farms/1/crop-profiles/"+cropKey)
	expectStatus(t, del2, 204)
	del2.Body.Close()

	err = testPool.QueryRow(ctx, `
SELECT details->>'kind'
FROM gr33ncore.user_activity_log
WHERE farm_id = 1
  AND details->>'kind' = 'crop_profile_override_deleted'
  AND details->>'crop_key' = $1
ORDER BY activity_time DESC
LIMIT 1`, cropKey).Scan(&kind)
	if err != nil {
		t.Fatalf("delete audit row: %v", err)
	}
}
