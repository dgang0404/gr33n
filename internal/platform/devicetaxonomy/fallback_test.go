package devicetaxonomy

import "testing"

func TestCurrent_TempFIsClimate(t *testing.T) {
	reg := Current()
	if got := reg.PlantNeed("sensor", "temp_f"); got != "air" {
		t.Fatalf("temp_f plant_need: got %q want air", got)
	}
	if lbl := reg.DisplayLabel("sensor", "temp_f"); lbl != "Temperature (°F)" {
		t.Fatalf("temp_f label: got %q", lbl)
	}
}

func TestCurrent_PulsePump(t *testing.T) {
	reg := Current()
	if !reg.SupportsPulse("pump") {
		t.Fatal("pump should support pulse")
	}
	if reg.GHRole("shade_screen") != "shade" {
		t.Fatal("shade_screen gh_role")
	}
}
