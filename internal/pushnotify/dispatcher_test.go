package pushnotify

import (
	"encoding/json"
	"testing"
)

func TestMergeDeliveryAttempt_AppendsChannel(t *testing.T) {
	existing := json.RawMessage(`{"push":[{"at":"2026-01-01T00:00:00Z","ok":true}]}`)
	merged := mergeDeliveryAttempt(existing, "push", false, "FCM not configured")
	var root map[string]any
	if err := json.Unmarshal(merged, &root); err != nil {
		t.Fatal(err)
	}
	list, ok := root["push"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("push list = %#v", root["push"])
	}
	last, ok := list[1].(map[string]any)
	if !ok {
		t.Fatalf("last entry type %T", list[1])
	}
	if last["ok"] != false {
		t.Fatalf("ok = %v", last["ok"])
	}
	if last["detail"] != "FCM not configured" {
		t.Fatalf("detail = %v", last["detail"])
	}
}

func TestMergeDeliveryAttempt_NewChannel(t *testing.T) {
	merged := mergeDeliveryAttempt(nil, "email", true, "")
	var root map[string]any
	if err := json.Unmarshal(merged, &root); err != nil {
		t.Fatal(err)
	}
	list, ok := root["email"].([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("email list = %#v", root["email"])
	}
}

func TestMergeDeliveryAttempt_ReplacesNonArrayChannel(t *testing.T) {
	bad := json.RawMessage(`{"push":"not-an-array"}`)
	merged := mergeDeliveryAttempt(bad, "push", true, "")
	var root map[string]any
	if err := json.Unmarshal(merged, &root); err != nil {
		t.Fatal(err)
	}
	list, ok := root["push"].([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("push list = %#v", root["push"])
	}
}
