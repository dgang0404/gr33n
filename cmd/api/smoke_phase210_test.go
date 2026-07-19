// Phase 210 — dedicated animal automation.
//
//  1. `animal_lifecycle_event` trigger_source + the `animal_event`
//     predicate type: a rule can react to "the flock's most recent
//     lifecycle event is X" (e.g. released_to_pasture -> open gate).
//  2. Write-time validation for both the trigger_configuration and the
//     conditions_jsonb shape (cross-farm animal_group_id rejected, missing
//     event_type rejected).
//  3. `duration_seconds` on a control_actuator action's action_parameters
//     lets a single scheduled/rule action run a timed pulse (feeder hopper),
//     reusing the same validation the manual "Run pulse" button already
//     enforces — and confirms a non-pulseable actuator_type (gate) is
//     rejected.
package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func seedAnimalGroupForRule(t *testing.T, tok string, zoneID int64) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/animal-groups", map[string]any{
		"label":           uniqueName("ws210_flock"),
		"species":         "chicken",
		"count":           8,
		"primary_zone_id": zoneID,
	})
	expectStatus(t, resp, http.StatusCreated)
	group := decodeMap(t, resp)
	t.Cleanup(func() {
		resp := authDelete(t, tok, fmt.Sprintf("/animal-groups/%d", int64(group["id"].(float64))))
		resp.Body.Close()
	})
	return int64(group["id"].(float64))
}

func postLifecycleEvent(t *testing.T, tok string, groupID int64, eventType string) {
	t.Helper()
	resp := authPost(t, tok, fmt.Sprintf("/animal-groups/%d/lifecycle-events", groupID), map[string]any{
		"event_type": eventType,
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()
}

// TestPhase210AnimalEventPredicateGateRule proves the full flock-event ->
// gate loop: a rule with trigger_source=animal_lifecycle_event and an
// animal_event predicate only fires the gate actuator once the flock's
// latest lifecycle event matches, and flips accordingly when a later,
// different event lands — without any polling window/timer, purely by
// checking "what's the latest event now".
func TestPhase210AnimalEventPredicateGateRule(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	zoneID := seedZoneForAnimal(t, tok)
	groupID := seedAnimalGroupForRule(t, tok, zoneID)
	gateID := seedRuleActuator(t, uniqueName("ws210_gate"))

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws210_open_gate"),
		"is_active":       true,
		"trigger_source":  "animal_lifecycle_event",
		"trigger_configuration": map[string]any{"animal_group_id": groupID},
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"type": "animal_event", "animal_group_id": groupID, "event_type": "released_to_pasture"},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order":    0,
		"action_type":        "control_actuator",
		"target_actuator_id": gateID,
		"action_command":     "on",
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	// No matching event yet — rule must skip, not fire.
	testWorker.TickRules(ctx)
	var status, msg string
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, '') FROM gr33ncore.automation_runs
		WHERE rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg); err != nil {
		t.Fatalf("read latest rule run: %v", err)
	}
	if status != "skipped" || msg != "no_animal_event_yet" {
		t.Fatalf("expected skipped/no_animal_event_yet before any lifecycle event, got status=%s msg=%s", status, msg)
	}

	// Flock is released to pasture -> predicate now matches -> gate opens.
	postLifecycleEvent(t, tok, groupID, "released_to_pasture")
	testWorker.TickRules(ctx)

	var eventCount int
	var commandSent string
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(MAX(command_sent), '')
		FROM gr33ncore.actuator_events
		WHERE triggered_by_rule_id = $1 AND actuator_id = $2`, ruleID, gateID,
	).Scan(&eventCount, &commandSent); err != nil {
		t.Fatalf("count actuator events: %v", err)
	}
	if eventCount != 1 || commandSent != "on" {
		t.Fatalf("expected exactly 1 'on' actuator event after released_to_pasture, got count=%d command=%s", eventCount, commandSent)
	}

	// A later, different event flips the predicate back to failing — the
	// gate rule must not fire again (cooldown-independent: this is a
	// distinct condition_not_met run, not a skip).
	postLifecycleEvent(t, tok, groupID, "penned_for_night")
	testWorker.TickRules(ctx)
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.actuator_events
		WHERE triggered_by_rule_id = $1 AND actuator_id = $2`, ruleID, gateID,
	).Scan(&eventCount); err != nil {
		t.Fatalf("recount actuator events: %v", err)
	}
	if eventCount != 1 {
		t.Fatalf("expected no additional actuator event after penned_for_night, still got %d", eventCount)
	}
}

// TestPhase210AnimalEventValidation exercises the write-time guards added
// alongside the animal_event predicate and animal_lifecycle_event trigger.
func TestPhase210AnimalEventValidation(t *testing.T) {
	tok := smokeJWT(t)

	// Cross-farm animal_group_id in trigger_configuration -> 400.
	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":                  uniqueName("rule_ws210_bad_trigger"),
		"trigger_source":        "animal_lifecycle_event",
		"trigger_configuration": map[string]any{"animal_group_id": 999999},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown animal_group_id in trigger_configuration, got %d", resp.StatusCode)
	}

	// Missing event_type on an animal_event predicate -> 400.
	zoneID := seedZoneForAnimal(t, tok)
	groupID := seedAnimalGroupForRule(t, tok, zoneID)
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws210_missing_event_type"),
		"trigger_source":  "animal_lifecycle_event",
		"trigger_configuration": map[string]any{"animal_group_id": groupID},
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"type": "animal_event", "animal_group_id": groupID},
		},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for animal_event predicate missing event_type, got %d", resp.StatusCode)
	}

	// Unknown trigger_source string is still rejected (enum widened, not opened up).
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":           uniqueName("rule_ws210_bad_source"),
		"trigger_source": "flock_migration_arc",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for made-up trigger_source, got %d", resp.StatusCode)
	}
}

// TestPhase210ScheduledFeedingPulseDuration proves the "scheduled feeding"
// mechanism: a control_actuator action's action_parameters.duration_seconds
// is accepted for a pulse-capable actuator type (feeder_hopper family —
// reusing the 'relay' fixture, which shares the same PulseDurationAllowed
// branch) and rejected for a non-pulseable one (gate).
func TestPhase210ScheduledFeedingPulseDuration(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)

	hopperID := seedRuleActuator(t, uniqueName("ws210_hopper")) // actuator_type='relay', pulse-capable
	resp := authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            uniqueName("smoke_feed_schedule"),
		"schedule_type":   "cron",
		"cron_expression": "0 7 * * *",
		"timezone":        "UTC",
		"is_active":       true,
	})
	expectStatus(t, resp, http.StatusCreated)
	schedID := int64(decodeMap(t, resp)["id"].(float64))

	resp = authPost(t, tok, fmt.Sprintf("/schedules/%d/actions", schedID), map[string]any{
		"execution_order":     0,
		"action_type":         "control_actuator",
		"target_actuator_id":  hopperID,
		"action_command":      "on",
		"action_parameters":   map[string]any{"duration_seconds": 5},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	params, ok := created["action_parameters"].(map[string]any)
	if !ok || int(params["duration_seconds"].(float64)) != 5 {
		t.Fatalf("expected action_parameters.duration_seconds=5 to round-trip, got %v", created["action_parameters"])
	}

	// Gate is a toggle, not a timed-run device — duration_seconds must 400.
	gateID := seedRuleActuator(t, uniqueName("ws210_gate2"))
	if _, err := testPool.Exec(context.Background(),
		`UPDATE gr33ncore.actuators SET actuator_type = 'gate' WHERE id = $1`, gateID); err != nil {
		t.Fatalf("retype actuator to gate: %v", err)
	}
	resp = authPost(t, tok, fmt.Sprintf("/schedules/%d/actions", schedID), map[string]any{
		"execution_order":    1,
		"action_type":        "control_actuator",
		"target_actuator_id": gateID,
		"action_command":     "on",
		"action_parameters":  map[string]any{"duration_seconds": 5},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for duration_seconds on a gate actuator, got %d", resp.StatusCode)
	}
}
