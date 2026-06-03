package actuator

import (
	"encoding/json"
	"testing"
)

func TestValidCommands(t *testing.T) {
	tests := []struct {
		actuatorType string
		wantContains string
	}{
		{"shade_screen", "deploy"},
		{"ridge_vent", "open"},
		{"exhaust_fan", "on"},
		{"circulation_fan", "off"},
		{"glazing_panel", "open"},
		{"light", "on"},
		{"relay", "off"},
		{"unknown_type", "on"},
	}
	for _, tc := range tests {
		cmds := ValidCommands(tc.actuatorType)
		found := false
		for _, c := range cmds {
			if c == tc.wantContains {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidCommands(%q) missing %q; got %v", tc.actuatorType, tc.wantContains, cmds)
		}
	}
}

func TestValidateActuatorConfig(t *testing.T) {
	tests := []struct {
		name         string
		actuatorType string
		cfg          string
		wantErr      bool
	}{
		{
			name:         "shade_screen valid config",
			actuatorType: "shade_screen",
			cfg:          `{"channel":0,"max_run_seconds":30}`,
		},
		{
			name:         "shade_screen zero max_run_seconds rejected",
			actuatorType: "shade_screen",
			cfg:          `{"max_run_seconds":0}`,
			wantErr:      true,
		},
		{
			name:         "ridge_vent negative max_run_seconds rejected",
			actuatorType: "ridge_vent",
			cfg:          `{"max_run_seconds":-5}`,
			wantErr:      true,
		},
		{
			name:         "exhaust_fan any config ok",
			actuatorType: "exhaust_fan",
			cfg:          `{"channel":2}`,
		},
		{
			name:         "empty config is ok",
			actuatorType: "shade_screen",
			cfg:          ``,
		},
		{
			name:         "malformed config rejected",
			actuatorType: "exhaust_fan",
			cfg:          `{bad`,
			wantErr:      true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateActuatorConfig(tc.actuatorType, json.RawMessage(tc.cfg))
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
