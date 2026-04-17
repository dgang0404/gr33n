// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestListTasks(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/tasks")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestTaskCreate(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/farms/1/tasks", map[string]any{
		"title": "smoke task",
	})
	expectStatus(t, resp, 201)
}

// Phase 20.95 WS1 — task labor log round-trip with running SUM semantics.

func TestTaskLaborLogRoundtrip(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/farms/1/tasks", map[string]any{
		"title": uniqueName("labor_task"),
	})
	expectStatus(t, resp, 201)
	task := decodeMap(t, resp)
	taskID := int64(task["id"].(float64))

	// Two inserts: 30 + 45 = 75.
	resp = authPost(t, tok, fmt.Sprintf("/tasks/%d/labor", taskID), map[string]any{
		"started_at": "2026-05-01T09:00:00Z",
		"ended_at":   "2026-05-01T09:30:00Z",
		"minutes":    30,
	})
	expectStatus(t, resp, 201)
	first := decodeMap(t, resp)
	firstID := int64(first["id"].(float64))

	resp = authPost(t, tok, fmt.Sprintf("/tasks/%d/labor", taskID), map[string]any{
		"started_at": "2026-05-01T13:00:00Z",
		"ended_at":   "2026-05-01T13:45:00Z",
		"minutes":    45,
	})
	expectStatus(t, resp, 201)

	resp = authGet(t, tok, fmt.Sprintf("/tasks/%d/labor", taskID))
	expectStatus(t, resp, 200)
	rows := decodeSlice(t, resp)
	if len(rows) != 2 {
		t.Fatalf("expected 2 labor rows, got %d", len(rows))
	}

	// Assert tasks.time_spent_minutes == 75 (direct DB read — no task-read endpoint).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var spent *int32
	if err := testPool.QueryRow(ctx,
		`SELECT time_spent_minutes FROM gr33ncore.tasks WHERE id = $1`, taskID,
	).Scan(&spent); err != nil {
		t.Fatalf("read time_spent_minutes: %v", err)
	}
	if spent == nil || *spent != 75 {
		got := "nil"
		if spent != nil {
			got = fmt.Sprintf("%d", *spent)
		}
		t.Fatalf("expected time_spent_minutes=75 after 30+45, got %s", got)
	}

	// Delete the 30-minute row → expect 45.
	resp = authDelete(t, tok, fmt.Sprintf("/labor/%d", firstID))
	expectStatus(t, resp, 204)

	resp = authGet(t, tok, fmt.Sprintf("/tasks/%d/labor", taskID))
	expectStatus(t, resp, 200)
	rows = decodeSlice(t, resp)
	if len(rows) != 1 {
		t.Fatalf("expected 1 labor row after delete, got %d", len(rows))
	}

	if err := testPool.QueryRow(ctx,
		`SELECT time_spent_minutes FROM gr33ncore.tasks WHERE id = $1`, taskID,
	).Scan(&spent); err != nil {
		t.Fatalf("read time_spent_minutes after delete: %v", err)
	}
	if spent == nil || *spent != 45 {
		got := "nil"
		if spent != nil {
			got = fmt.Sprintf("%d", *spent)
		}
		t.Fatalf("expected time_spent_minutes=45 after delete of 30m row, got %s", got)
	}
}

// TestPhase2095AnimalAquaponicsScope — Phase 20.95 WS4.
// Asserts that the new animal_groups scope columns and aquaponics.loops
// topology columns round-trip through the DB. Handlers / OpenAPI wiring
// come in Phase 20.6 / 20.8; this phase only confirms the columns exist
// with the documented defaults and accept writes.

func TestPhase2095AnimalAquaponicsScope(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Pick a zone on farm 1 so we can exercise the FK references.
	var zoneID int64
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM gr33ncore.zones WHERE farm_id = 1 ORDER BY id LIMIT 1`,
	).Scan(&zoneID); err != nil {
		t.Fatalf("seed zone lookup: %v", err)
	}

	// ── animal_groups: count / primary_zone_id / active / archived_* ──
	var agID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33nanimals.animal_groups
		    (farm_id, label, species, count, primary_zone_id, active, archived_at, archived_reason)
		VALUES (1, 'ws4_smoke_flock', 'chicken', 12, $1, TRUE, NULL, NULL)
		RETURNING id`, zoneID).Scan(&agID); err != nil {
		t.Fatalf("insert animal group: %v", err)
	}
	var (
		count    *int32
		pz       *int64
		active   bool
		archived *time.Time
		reason   *string
	)
	if err := testPool.QueryRow(ctx, `
		SELECT count, primary_zone_id, active, archived_at, archived_reason
		  FROM gr33nanimals.animal_groups WHERE id = $1`, agID,
	).Scan(&count, &pz, &active, &archived, &reason); err != nil {
		t.Fatalf("read animal group: %v", err)
	}
	if count == nil || *count != 12 {
		t.Fatalf("expected count=12, got %v", count)
	}
	if pz == nil || *pz != zoneID {
		t.Fatalf("expected primary_zone_id=%d, got %v", zoneID, pz)
	}
	if !active {
		t.Fatalf("expected active=true (default), got false")
	}

	// Archive round-trip.
	if _, err := testPool.Exec(ctx, `
		UPDATE gr33nanimals.animal_groups
		   SET active = FALSE, archived_at = NOW(), archived_reason = 'ws4 smoke archive'
		 WHERE id = $1`, agID); err != nil {
		t.Fatalf("archive animal group: %v", err)
	}

	// ── aquaponics.loops: fish_tank_zone_id / grow_bed_zone_id ──
	var loopID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33naquaponics.loops
		    (farm_id, label, fish_tank_zone_id, grow_bed_zone_id)
		VALUES (1, 'ws4_smoke_loop', $1, $1)
		RETURNING id`, zoneID).Scan(&loopID); err != nil {
		t.Fatalf("insert loop: %v", err)
	}
	var fishZone, growZone *int64
	if err := testPool.QueryRow(ctx, `
		SELECT fish_tank_zone_id, grow_bed_zone_id
		  FROM gr33naquaponics.loops WHERE id = $1`, loopID,
	).Scan(&fishZone, &growZone); err != nil {
		t.Fatalf("read loop: %v", err)
	}
	if fishZone == nil || *fishZone != zoneID {
		t.Fatalf("expected fish_tank_zone_id=%d, got %v", zoneID, fishZone)
	}
	if growZone == nil || *growZone != zoneID {
		t.Fatalf("expected grow_bed_zone_id=%d, got %v", zoneID, growZone)
	}
}

// TestPhase2095ExecutableActionsProgramID — Phase 20.95 WS3.
// Asserts:
//   1. program_id column round-trips on gr33ncore.executable_actions.
//   2. The new num_nonnulls(...) = 1 CHECK rejects two-source rows at the DB.
//   3. The rules_handler rejects two-source writes with a 400 before reaching
//      the DB CHECK.

func TestTaskUpdateAndDelete(t *testing.T) {
	tok := smokeJWT(t)

	resp := authPost(t, tok, "/farms/1/tasks", map[string]any{
		"title":    "smoke update task",
		"priority": 1,
	})
	expectStatus(t, resp, 201)
	created := decodeMap(t, resp)
	taskID := int64(created["id"].(float64))

	updatedTitle := uniqueName("smoke_task_upd")
	resp = authPut(t, tok, fmt.Sprintf("/tasks/%d", taskID), map[string]any{
		"title":    updatedTitle,
		"priority": 2,
	})
	expectStatus(t, resp, 200)
	updated := decodeMap(t, resp)
	if updated["title"] != updatedTitle {
		t.Fatalf("expected title=%s, got %v", updatedTitle, updated["title"])
	}
	prio, _ := updated["priority"].(float64)
	if int(prio) != 2 {
		t.Fatalf("expected priority=2, got %v", updated["priority"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/tasks/%d", taskID))
	expectStatus(t, resp, 204)

	resp = authGet(t, tok, "/farms/1/tasks")
	expectStatus(t, resp, 200)
	taskList := decodeSlice(t, resp)
	for _, item := range taskList {
		if m, ok := item.(map[string]any); ok {
			if int64(m["id"].(float64)) == taskID {
				t.Fatal("soft-deleted task still appears in list")
			}
		}
	}
}
