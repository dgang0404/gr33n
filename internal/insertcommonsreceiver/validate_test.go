package insertcommonsreceiver

import (
	"testing"
	"time"
)

func TestValidatePayload_minimalValid(t *testing.T) {
	raw := []byte(`{
		"schema_version": "gr33n.insert_commons.v1",
		"generated_at": "2026-04-16T12:00:00Z",
		"farm_pseudonym": "abc123",
		"farm_profile": {
			"scale_tier": "small",
			"timezone_bucket": "UTC",
			"currency": "USD",
			"operational_status": "active"
		},
		"aggregates": {
			"costs": {},
			"tasks": {},
			"devices": {}
		},
		"privacy": { "includes_pii": false }
	}`)
	fp, gen, err := validatePayload(raw)
	if err != nil {
		t.Fatal(err)
	}
	if fp != "abc123" {
		t.Fatalf("farm pseudonym: %q", fp)
	}
	if !gen.Equal(time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("time: %v", gen)
	}
}

func TestValidatePayload_wrongSchema(t *testing.T) {
	raw := []byte(`{"schema_version":"v0","generated_at":"2026-04-16T12:00:00Z","farm_pseudonym":"x","farm_profile":{"scale_tier":"","timezone_bucket":"","currency":"","operational_status":""},"aggregates":{"costs":{},"tasks":{},"devices":{}},"privacy":{"includes_pii":false}}`)
	_, _, err := validatePayload(raw)
	if err == nil {
		t.Fatal("expected error")
	}
}
