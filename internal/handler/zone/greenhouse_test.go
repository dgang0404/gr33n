package zone

import (
	"encoding/json"
	"testing"
)

func TestValidateGreenhouseClimate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		wantWarning bool
	}{
		{
			name:    "valid glass auto with shade actuator",
			input:   `{"cover_type":"glass","automation_policy":"auto","shade_actuator_id":5}`,
			wantErr: false,
		},
		{
			name:    "valid polycarbonate manual no actuators",
			input:   `{"cover_type":"polycarbonate","automation_policy":"manual"}`,
			wantErr: false,
		},
		{
			name:    "valid schedule_only film",
			input:   `{"cover_type":"film","automation_policy":"schedule_only","shade_actuator_id":7}`,
			wantErr: false,
		},
		{
			name:        "auto policy no actuators triggers warning",
			input:       `{"cover_type":"glass","automation_policy":"auto"}`,
			wantErr:     false,
			wantWarning: true,
		},
		{
			name:    "invalid cover_type",
			input:   `{"cover_type":"metal_sheet"}`,
			wantErr: true,
		},
		{
			name:    "invalid automation_policy",
			input:   `{"automation_policy":"hybrid"}`,
			wantErr: true,
		},
		{
			name:    "empty raw is a no-op",
			input:   ``,
			wantErr: false,
		},
		{
			name:    "malformed json",
			input:   `{not json`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := json.RawMessage(tc.input)
			_, warns, err := ValidateGreenhouseClimate(raw)
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.wantWarning && len(warns) == 0 {
				t.Errorf("expected at least one warning")
			}
		})
	}
}

func TestExtractGreenhouseClimate(t *testing.T) {
	meta := json.RawMessage(`{"photo_attachment_ids":[1],"greenhouse_climate":{"cover_type":"glass"}}`)
	raw, err := ExtractGreenhouseClimate(meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(raw) == "" {
		t.Fatal("expected non-empty greenhouse_climate")
	}

	// Meta without the key returns nil.
	raw2, err := ExtractGreenhouseClimate(json.RawMessage(`{"photo_attachment_ids":[1]}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raw2) != 0 {
		t.Errorf("expected nil for absent key, got %s", raw2)
	}
}
