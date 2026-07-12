package zone

import (
	"encoding/json"
	"testing"
)

func TestValidateZoneLayout(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "valid with defaults for w/h",
			input: `{"x":0.12,"y":0.4}`,
		},
		{
			name:  "valid explicit size",
			input: `{"x":0.1,"y":0.2,"w":0.22,"h":0.18}`,
		},
		{
			name:    "rejects tile past right edge",
			input:   `{"x":0.9,"y":0.1,"w":0.2,"h":0.1}`,
			wantErr: true,
		},
		{
			name:    "rejects tile past bottom edge",
			input:   `{"x":0.1,"y":0.9,"w":0.1,"h":0.2}`,
			wantErr: true,
		},
		{
			name:    "malformed json",
			input:   `{bad`,
			wantErr: true,
		},
		{
			name:  "empty is no-op",
			input: ``,
		},
		{
			name:  "clamps negative coordinates",
			input: `{"x":-0.5,"y":-0.1,"w":0.1,"h":0.1}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateZoneLayout(json.RawMessage(tc.input))
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateZoneLayoutDefaults(t *testing.T) {
	got, err := ValidateZoneLayout(json.RawMessage(`{"x":0.5,"y":0.5}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.W == nil || *got.W != defaultLayoutW {
		t.Fatalf("expected default w=%v, got %v", defaultLayoutW, got.W)
	}
	if got.H == nil || *got.H != defaultLayoutH {
		t.Fatalf("expected default h=%v, got %v", defaultLayoutH, got.H)
	}
}

func TestExtractZoneLayout(t *testing.T) {
	meta := json.RawMessage(`{"layout":{"x":0.1,"y":0.2},"greenhouse_climate":{"cover_type":"glass"}}`)
	raw, err := ExtractZoneLayout(meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(raw) == "" {
		t.Fatal("expected layout payload")
	}

	raw2, err := ExtractZoneLayout(json.RawMessage(`{"greenhouse_climate":{}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raw2) != 0 {
		t.Errorf("expected nil for absent key, got %s", raw2)
	}
}

func TestMergeZoneMetaData(t *testing.T) {
	existing := json.RawMessage(`{"greenhouse_climate":{"cover_type":"glass"},"photo_attachment_ids":[1]}`)
	incoming := json.RawMessage(`{"layout":{"x":0.12,"y":0.4,"w":0.2,"h":0.18}}`)

	merged, err := MergeZoneMetaData(existing, incoming)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(merged, &m); err != nil {
		t.Fatalf("decode merged: %v", err)
	}
	if _, ok := m["greenhouse_climate"]; !ok {
		t.Fatal("greenhouse_climate should be preserved")
	}
	if _, ok := m["photo_attachment_ids"]; !ok {
		t.Fatal("photo_attachment_ids should be preserved")
	}
	if _, ok := m["layout"]; !ok {
		t.Fatal("layout should be added")
	}

	// Empty incoming preserves existing.
	again, err := MergeZoneMetaData(existing, nil)
	if err != nil {
		t.Fatalf("merge empty incoming: %v", err)
	}
	if string(again) != string(existing) {
		t.Fatalf("expected existing meta unchanged, got %s", again)
	}

	// Greenhouse update must not erase layout.
	gcOnly := json.RawMessage(`{"greenhouse_climate":{"cover_type":"film","automation_policy":"manual"}}`)
	withLayout, err := MergeZoneMetaData(merged, gcOnly)
	if err != nil {
		t.Fatalf("merge greenhouse update: %v", err)
	}
	if err := json.Unmarshal(withLayout, &m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := m["layout"]; !ok {
		t.Fatal("layout should survive greenhouse_climate update")
	}
}
