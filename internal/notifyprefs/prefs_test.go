package notifyprefs

import (
	"encoding/json"
	"testing"

	db "gr33n-api/internal/db"
)

func TestFromPreferencesJSON_nested(t *testing.T) {
	raw := []byte(`{"theme":"dark","notify":{"push_enabled":true,"min_priority":"high"}}`)
	n := FromPreferencesJSON(raw)
	if !n.PushEnabled || n.MinPriority != "high" {
		t.Fatalf("got %+v", n)
	}
}

func TestSetNotify_preservesOtherKeys(t *testing.T) {
	raw := []byte(`{"theme":"dark"}`)
	out, err := SetNotify(raw, Notify{PushEnabled: true, MinPriority: "low"})
	if err != nil {
		t.Fatal(err)
	}
	n := FromPreferencesJSON(out)
	if !n.PushEnabled || n.MinPriority != "low" {
		t.Fatalf("notify %+v", n)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(out, &root); err != nil {
		t.Fatal(err)
	}
	if string(root["theme"]) != `"dark"` {
		t.Fatalf("theme lost: %s", root["theme"])
	}
}

func TestAlertMeetsMinPriority(t *testing.T) {
	a := db.Gr33ncoreAlertsNotification{
		Severity: db.NullGr33ncoreNotificationPriorityEnum{
			Gr33ncoreNotificationPriorityEnum: db.Gr33ncoreNotificationPriorityEnumHigh,
			Valid:                             true,
		},
	}
	if !AlertMeetsMinPriority(a, "medium") {
		t.Fatal("high should meet medium")
	}
	if AlertMeetsMinPriority(a, "critical") {
		t.Fatal("high should not meet critical")
	}
	b := db.Gr33ncoreAlertsNotification{Severity: db.NullGr33ncoreNotificationPriorityEnum{Valid: false}}
	if !AlertMeetsMinPriority(b, "low") {
		t.Fatal("default medium meets low")
	}
}
