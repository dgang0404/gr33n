// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestGetFarm(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["name"] == nil {
		t.Fatal("expected farm to have a name")
	}
}

func TestOrganizationCreateListUsageAndFarmLink(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/organizations", map[string]any{"name": "Smoke Tenant Org"})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	orgID := int64(created["id"].(float64))

	resp = authGet(t, tok, "/organizations")
	expectStatus(t, resp, http.StatusOK)
	list := decodeSlice(t, resp)
	if len(list) == 0 {
		t.Fatal("expected at least one organization")
	}

	resp = authGet(t, tok, fmt.Sprintf("/organizations/%d/usage-summary", orgID))
	expectStatus(t, resp, http.StatusOK)
	summary := decodeMap(t, resp)
	if _, ok := summary["farm_count"]; !ok {
		t.Fatalf("usage summary missing farm_count: %#v", summary)
	}

	resp = authGet(t, tok, fmt.Sprintf("/organizations/%d/audit-events?limit=10", orgID))
	expectStatus(t, resp, http.StatusOK)
	_ = decodeSlice(t, resp)

	resp = authPatch(t, tok, "/farms/1/organization", map[string]any{"organization_id": orgID})
	expectStatus(t, resp, http.StatusOK)
	farm := decodeMap(t, resp)
	if int64(farm["organization_id"].(float64)) != orgID {
		t.Fatalf("expected farm linked to org %d, got %#v", orgID, farm["organization_id"])
	}

	resp = authPatch(t, tok, "/farms/1/organization", map[string]any{"organization_id": nil})
	expectStatus(t, resp, http.StatusOK)
}

func TestOrgDefaultBootstrapOnFarmCreate(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/organizations", map[string]any{"name": uniqueName("org_bootstrap_default")})
	expectStatus(t, resp, http.StatusCreated)
	org := decodeMap(t, resp)
	orgID := int64(org["id"].(float64))

	resp = authPatch(t, tok, fmt.Sprintf("/organizations/%d", orgID), map[string]any{
		"default_bootstrap_template": "jadam_indoor_photoperiod_v1",
	})
	expectStatus(t, resp, http.StatusOK)

	name := uniqueName("org_default_farm")
	resp = authPost(t, tok, "/farms", map[string]any{
		"name":               name,
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"organization_id":    orgID,
	})
	expectStatus(t, resp, http.StatusCreated)
	payload := decodeMap(t, resp)
	farmObj, ok := payload["farm"].(map[string]any)
	if !ok {
		t.Fatalf("expected farm + bootstrap from org default, got %#v", payload)
	}
	if _, ok := payload["bootstrap"]; !ok {
		t.Fatal("expected bootstrap in response when org default applies")
	}
	fid := int64(farmObj["id"].(float64))
	zones := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/zones", fid)))
	if len(zones) < 4 {
		t.Fatalf("expected org default bootstrap zones, got %d", len(zones))
	}

	name2 := uniqueName("org_default_farm_explicit_none")
	resp = authPost(t, tok, "/farms", map[string]any{
		"name":               name2,
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"organization_id":    orgID,
		"bootstrap_template": "none",
	})
	expectStatus(t, resp, http.StatusCreated)
	payload2 := decodeMap(t, resp)
	farm2 := payload2["farm"].(map[string]any)
	fid2 := int64(farm2["id"].(float64))
	zones2 := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/zones", fid2)))
	if len(zones2) != 0 {
		t.Fatalf("bootstrap_template none should skip org default, got %d zones", len(zones2))
	}
}

func TestCrossFarmWriteForbidden(t *testing.T) {
	email := fmt.Sprintf("norole_%d@smoke.test", rand.Int())
	resp := postNoAuth("/auth/register", map[string]any{
		"email":     email,
		"password":  "longpassword1",
		"full_name": "No Farm",
	})
	expectStatus(t, resp, http.StatusCreated)
	reg := decodeMap(t, resp)
	otherTok, _ := reg["token"].(string)
	if otherTok == "" {
		t.Fatal("expected token from register")
	}

	resp = authPost(t, otherTok, "/farms/1/tasks", map[string]any{"title": "should fail"})
	expectStatus(t, resp, http.StatusForbidden)
}

// ── Fertigation CRUD ────────────────────────────────────────────────────────

func TestFarmBootstrapOnCreate(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("bootstrap_farm")
	resp := authPost(t, tok, "/farms", map[string]any{
		"name":               name,
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"bootstrap_template": "jadam_indoor_photoperiod_v1",
	})
	expectStatus(t, resp, http.StatusCreated)
	payload := decodeMap(t, resp)
	farmObj, ok := payload["farm"].(map[string]any)
	if !ok {
		t.Fatalf("expected farm in response, got %v", payload)
	}
	fid := int64(farmObj["id"].(float64))
	zones := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/zones", fid)))
	if len(zones) < 4 {
		t.Fatalf("expected at least 4 zones from bootstrap, got %d", len(zones))
	}
	resp2 := authPost(t, tok, fmt.Sprintf("/farms/%d/bootstrap-template", fid), map[string]any{
		"template": "jadam_indoor_photoperiod_v1",
	})
	expectStatus(t, resp2, http.StatusOK)
	again := decodeMap(t, resp2)
	boot := again["bootstrap"].(map[string]any)
	if applied, _ := boot["already_applied"].(bool); !applied {
		t.Fatalf("expected already_applied on second template apply, got %#v", boot)
	}
}

// TestPhase205BootstrapTemplates verifies Phase 20.5 WS2 farm bootstrap keys
// (chicken_coop_v1, greenhouse_climate_v1, drying_room_v1, small_aquaponics_v1)
// each land the expected zones, sensors, automation rules, and (for aquaponics)
// a gr33naquaponics.loops row.

func TestPhase205BootstrapTemplates(t *testing.T) {
	tok := smokeJWT(t)
	cases := []struct {
		key           string
		wantZones     []string
		minRules      int
		wantLoopLabel string
	}{
		{
			key:       "chicken_coop_v1",
			wantZones: []string{"Chicken Coop"},
			minRules:  4,
		},
		{
			key:       "greenhouse_climate_v1",
			wantZones: []string{"Greenhouse"},
			minRules:  4,
		},
		{
			key:       "drying_room_v1",
			wantZones: []string{"Drying Room"},
			minRules:  3,
		},
		{
			key:           "small_aquaponics_v1",
			wantZones:     []string{"Fish Tank", "Grow Bed"},
			minRules:      2,
			wantLoopLabel: "Main aquaponics loop",
		},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			name := uniqueName("farm_" + tc.key)
			resp := authPost(t, tok, "/farms", map[string]any{
				"name":               name,
				"owner_user_id":      smokeDevUserUUID,
				"timezone":           "UTC",
				"currency":           "USD",
				"operational_status": "active",
				"scale_tier":         "small",
				"bootstrap_template": tc.key,
			})
			expectStatus(t, resp, http.StatusCreated)
			payload := decodeMap(t, resp)
			farmObj := payload["farm"].(map[string]any)
			fid := int64(farmObj["id"].(float64))
			boot := payload["bootstrap"].(map[string]any)
			if errStr, _ := boot["error"].(string); errStr != "" {
				t.Fatalf("bootstrap error for %s: %s — %#v", tc.key, errStr, boot)
			}
			if applied, _ := boot["applied"].(bool); !applied {
				t.Fatalf("expected applied=true for %s, got %#v", tc.key, boot)
			}

			zones := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/zones", fid)))
			zoneNames := map[string]struct{}{}
			for _, z := range zones {
				if m, ok := z.(map[string]any); ok {
					if n, ok := m["name"].(string); ok {
						zoneNames[n] = struct{}{}
					}
				}
			}
			for _, wz := range tc.wantZones {
				if _, ok := zoneNames[wz]; !ok {
					t.Fatalf("farm %d template %s: missing zone %q (have %v)", fid, tc.key, wz, zoneNames)
				}
			}

			rules := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/automation/rules", fid)))
			if len(rules) < tc.minRules {
				t.Fatalf("farm %d template %s: expected at least %d rules, got %d", fid, tc.key, tc.minRules, len(rules))
			}

			if tc.wantLoopLabel != "" && testPool != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var cnt int
				err := testPool.QueryRow(ctx,
					`SELECT COUNT(*) FROM gr33naquaponics.loops WHERE farm_id = $1 AND label = $2 AND deleted_at IS NULL`,
					fid, tc.wantLoopLabel,
				).Scan(&cnt)
				if err != nil {
					t.Fatalf("count loops: %v", err)
				}
				if cnt != 1 {
					t.Fatalf("expected 1 aquaponics loop %q for farm %d, got %d", tc.wantLoopLabel, fid, cnt)
				}
			}
		})
	}
}
