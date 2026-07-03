// Phase 115 — schema utilization smokes
package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestPhase115_FarmModulesAndGating(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/farms/1/modules")
	expectStatus(t, resp, http.StatusOK)
	modules := decodeSlice(t, resp)
	if len(modules) < 4 {
		t.Fatalf("expected seeded modules, got %d", len(modules))
	}

	patch := authPatch(t, tok, "/farms/1/modules/gr33nanimals", map[string]any{"is_enabled": false})
	expectStatus(t, patch, http.StatusOK)

	blocked := authGet(t, tok, "/farms/1/animal-groups")
	expectStatus(t, blocked, http.StatusForbidden)

	restore := authPatch(t, tok, "/farms/1/modules/gr33nanimals", map[string]any{"is_enabled": true})
	expectStatus(t, restore, http.StatusOK)

	ok := authGet(t, tok, "/farms/1/animal-groups")
	expectStatus(t, ok, http.StatusOK)
}

func TestPhase115_NotificationTemplates(t *testing.T) {
	tok := smokeJWT(t)
	key := uniqueName("phase115_tpl")

	create := authPost(t, tok, "/farms/1/notification-templates", map[string]any{
		"template_key":              key,
		"subject_template":          "Alert: {{rule_name}}",
		"body_template_text":        "Rule {{rule_id}} fired at {{triggered_at}}",
		"default_delivery_channels": []string{"in_app", "email"},
	})
	expectStatus(t, create, http.StatusCreated)
	created := decodeMap(t, create)
	id := int64(created["id"].(float64))

	list := authGet(t, tok, "/farms/1/notification-templates")
	expectStatus(t, list, http.StatusOK)
	rows := decodeSlice(t, list)
	found := false
	for _, row := range rows {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		if int64(m["id"].(float64)) == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created template %d not in list", id)
	}

	patch := authPatch(t, tok, fmt.Sprintf("/notification-templates/%d", id), map[string]any{
		"description": "phase115 smoke",
	})
	expectStatus(t, patch, http.StatusOK)
	updated := decodeMap(t, patch)
	if updated["description"] != "phase115 smoke" {
		t.Fatalf("description not updated: %#v", updated["description"])
	}
}

func TestPhase115_SystemLogsEndpoint(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/system-logs")
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	logs, ok := body["logs"].([]any)
	if !ok {
		t.Fatalf("logs missing: %#v", body["logs"])
	}
	_ = logs // empty is fine on fresh DB
}

func TestPhase115_AgronomySymptoms(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/commons/agronomy-symptoms")
	expectStatus(t, resp, http.StatusOK)
	body := decodeMap(t, resp)
	if _, ok := body["symptoms"]; !ok {
		t.Fatalf("symptoms key missing: %#v", body)
	}
}

func TestPhase115_TaskDurationAndComplete(t *testing.T) {
	tok := smokeJWT(t)

	create := authPost(t, tok, "/farms/1/tasks", map[string]any{
		"title":                       "phase115 duration task",
		"estimated_duration_minutes": 45,
	})
	expectStatus(t, create, http.StatusCreated)
	task := decodeMap(t, create)
	id := int64(task["id"].(float64))
	if int(task["estimated_duration_minutes"].(float64)) != 45 {
		t.Fatalf("estimated_duration_minutes: %#v", task["estimated_duration_minutes"])
	}

	complete := authPatch(t, tok, fmt.Sprintf("/tasks/%d/complete", id), map[string]any{
		"actual_start_time": "2026-07-01T10:00:00Z",
		"actual_end_time":   "2026-07-01T10:45:00Z",
	})
	expectStatus(t, complete, http.StatusOK)
	done := decodeMap(t, complete)
	if done["status"] != "completed" {
		t.Fatalf("expected completed, got %#v", done["status"])
	}
	if done["actual_start_time"] == nil || done["actual_end_time"] == nil {
		t.Fatalf("actual times not set: %#v", done)
	}
}

func TestPhase115_ModuleCatalog(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farm-modules/catalog")
	expectStatus(t, resp, http.StatusOK)
	catalog := decodeSlice(t, resp)
	if len(catalog) < 4 {
		t.Fatalf("expected module catalog entries, got %d", len(catalog))
	}
}
