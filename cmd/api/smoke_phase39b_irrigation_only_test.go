// Phase 39b — irrigation_only programs: no recipe, no mix enqueue, mix-preview skip.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPhase39bIrrigationOnlyValidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uid := fmt.Sprintf("ph39b_%d", time.Now().UnixNano())
	jwt := smokeJWT(t)

	// Create irrigation_only program via API.
	createResp := authPost(t, jwt, "/farms/1/fertigation/programs", map[string]any{
		"name":                  uid + " Well Pulse",
		"irrigation_only":       true,
		"total_volume_liters":   40,
		"run_duration_seconds":  60,
		"ec_trigger_low":        0,
		"ph_trigger_low":        6,
		"ph_trigger_high":       8,
		"is_active":             true,
	})
	defer createResp.Body.Close()
	expectStatus(t, createResp, http.StatusCreated)

	var prog map[string]any
	if err := json.NewDecoder(createResp.Body).Decode(&prog); err != nil {
		t.Fatalf("decode program: %v", err)
	}
	progID := int64(prog["id"].(float64))
	if prog["irrigation_only"] != true {
		t.Fatalf("expected irrigation_only=true; got %v", prog["irrigation_only"])
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(),
			`UPDATE gr33nfertigation.programs SET deleted_at = NOW() WHERE id = $1`, progID)
	})

	// Reject recipe on irrigation_only create.
	badResp := authPost(t, jwt, "/farms/1/fertigation/programs", map[string]any{
		"name":                    uid + " Bad",
		"irrigation_only":         true,
		"application_recipe_id":   1,
		"total_volume_liters":     40,
		"ec_trigger_low":          0,
		"ph_trigger_low":          6,
		"ph_trigger_high":         8,
		"is_active":               true,
	})
	defer badResp.Body.Close()
	expectStatus(t, badResp, http.StatusBadRequest)

	// mix-preview skips mix.
	prev := authGet(t, jwt, fmt.Sprintf("/fertigation/programs/%d/mix-preview", progID))
	defer prev.Body.Close()
	expectStatus(t, prev, http.StatusOK)
	var prevOut map[string]any
	if err := json.NewDecoder(prev.Body).Decode(&prevOut); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if prevOut["mix_required"] != false {
		t.Fatalf("expected mix_required=false; got %v", prevOut)
	}

	// mix-jobs enqueue refused.
	mixResp := authPost(t, jwt, "/farms/1/fertigation/mix-jobs", map[string]any{
		"program_id": progID,
	})
	defer mixResp.Body.Close()
	expectStatus(t, mixResp, http.StatusBadRequest)

	// No mix_batch rows for this program's device (none linked — just ensure zero mix_batch in DB for program)
	var mixCount int
	_ = testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.device_commands
		WHERE program_id = $1 AND command_type = 'mix_batch'
	`, progID).Scan(&mixCount)
	if mixCount != 0 {
		t.Fatalf("expected no mix_batch commands for irrigation_only program; got %d", mixCount)
	}
}
