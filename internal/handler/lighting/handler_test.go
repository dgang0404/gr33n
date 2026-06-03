package lighting

import (
	"testing"
)

func TestParseHHMM(t *testing.T) {
	tests := []struct {
		input   string
		wantH   int
		wantM   int
		wantErr bool
	}{
		{"06:00", 6, 0, false},
		{"00:00", 0, 0, false},
		{"23:59", 23, 59, false},
		{"18:30", 18, 30, false},
		{"6:00", 6, 0, false},
		{"", 0, 0, true},
		{"25:00", 0, 0, true},
		{"12:60", 0, 0, true},
		{"abc", 0, 0, true},
	}
	for _, tt := range tests {
		h, m, err := parseHHMM(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseHHMM(%q): expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseHHMM(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if h != tt.wantH || m != tt.wantM {
			t.Errorf("parseHHMM(%q) = (%d, %d), want (%d, %d)", tt.input, h, m, tt.wantH, tt.wantM)
		}
	}
}

func TestBuildCronExpressions(t *testing.T) {
	tests := []struct {
		lightsOnAt string
		onHours    int32
		wantOn     string
		wantOff    string
	}{
		{"06:00", 18, "0 6 * * *", "0 0 * * *"},   // 18/6: on at 06, off at 00
		{"06:00", 12, "0 6 * * *", "0 18 * * *"},  // 12/12: on at 06, off at 18
		{"06:00", 22, "0 6 * * *", "0 4 * * *"},   // 22/2: on at 06, off at 04 next day
		{"00:00", 18, "0 0 * * *", "0 18 * * *"},  // on at midnight, off at 18
		{"06:30", 18, "30 6 * * *", "30 0 * * *"}, // half-hour anchor
	}
	for _, tt := range tests {
		on, off, err := buildCronExpressions(tt.lightsOnAt, tt.onHours)
		if err != nil {
			t.Errorf("buildCronExpressions(%q, %d): unexpected error: %v", tt.lightsOnAt, tt.onHours, err)
			continue
		}
		if on != tt.wantOn {
			t.Errorf("buildCronExpressions(%q, %d) ON = %q, want %q", tt.lightsOnAt, tt.onHours, on, tt.wantOn)
		}
		if off != tt.wantOff {
			t.Errorf("buildCronExpressions(%q, %d) OFF = %q, want %q", tt.lightsOnAt, tt.onHours, off, tt.wantOff)
		}
	}
}

func TestCreateProgramRequestValidate(t *testing.T) {
	valid := createProgramRequest{
		Name:       "Test",
		ZoneID:     1,
		ActuatorID: 2,
		OnHours:    18,
		OffHours:   6,
		LightsOnAt: "06:00",
		Timezone:   "UTC",
	}
	if err := valid.validate(); err != nil {
		t.Errorf("expected valid request to pass, got: %v", err)
	}

	// Wrong sum.
	bad := valid
	bad.OnHours = 10
	bad.OffHours = 5
	if err := bad.validate(); err == nil {
		t.Error("expected error when on+off != 24")
	}

	// Missing name.
	bad2 := valid
	bad2.Name = ""
	if err := bad2.validate(); err == nil {
		t.Error("expected error for empty name")
	}

	// Bad timezone.
	bad3 := valid
	bad3.Timezone = "Not/AReal/Zone"
	if err := bad3.validate(); err == nil {
		t.Error("expected error for invalid timezone")
	}
}

func TestPresetList(t *testing.T) {
	list := PresetList()
	if len(list) == 0 {
		t.Fatal("expected at least one preset")
	}
	keys := map[string]bool{}
	for _, p := range list {
		k, _ := p["key"].(string)
		if k == "" {
			t.Errorf("preset missing key: %v", p)
		}
		if keys[k] {
			t.Errorf("duplicate preset key: %s", k)
		}
		keys[k] = true
	}
	// Verify required presets exist.
	for _, required := range []string{"veg_18_6", "flower_12_12", "peas_22_2"} {
		if !keys[required] {
			t.Errorf("missing required preset %q", required)
		}
	}
}
