package actuator

import (
	"encoding/json"
	"testing"
)

func TestCommandAllowed(t *testing.T) {
	tests := []struct {
		actuatorType string
		command      string
		want         bool
	}{
		{"shade_screen", "deploy", true},
		{"shade_screen", "DEPLOY", true},
		{"shade_screen", "dispense", false},
		{"ridge_vent", "open", true},
		{"exhaust_fan", "retract", false},
		{"feeder_hopper", "dispense", true},
	}
	for _, tc := range tests {
		got := CommandAllowed(tc.actuatorType, tc.command)
		if got != tc.want {
			t.Errorf("CommandAllowed(%q, %q) = %v, want %v", tc.actuatorType, tc.command, got, tc.want)
		}
	}
}

func TestBuildPendingCommandJSON(t *testing.T) {
	raw, err := BuildPendingCommandJSON(12, "deploy", "operator", "manual shade")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["command"] != "deploy" || m["actuator_id"].(float64) != 12 {
		t.Fatalf("payload: %#v", m)
	}
	if m["source"] != "operator" || m["reason"] != "manual shade" {
		t.Fatalf("source/reason: %#v", m)
	}
}

func TestBuildPendingCommandJSON_duration(t *testing.T) {
	d := 2
	raw, err := BuildPendingCommandJSONFull(PendingCommandInput{
		ActuatorID: 5, Command: "on", Source: "operator", DurationSeconds: &d,
	})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["duration_seconds"].(float64) != 2 {
		t.Fatalf("duration: %#v", m)
	}
}

func TestValidatePulseDuration(t *testing.T) {
	d := 3
	if err := ValidatePulseDuration("pump", &d); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePulseDuration("grow_light", &d); err == nil {
		t.Fatal("expected error for grow_light pulse")
	}
}
