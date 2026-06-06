package farmguardian

import (
	"strings"
	"testing"
)

func TestSetupModeActive(t *testing.T) {
	snap := Snapshot{ZoneCount: 0}
	if !SetupModeActive(snap, false) {
		t.Fatal("expected active when zone_count is 0")
	}
	snap.ZoneCount = 2
	if SetupModeActive(snap, false) {
		t.Fatal("expected inactive when zones exist and no explicit flag")
	}
	if !SetupModeActive(snap, true) {
		t.Fatal("expected active when explicit setup flag is set")
	}
}

func TestSetupModePromptBlock_ZeroZones(t *testing.T) {
	got := SetupModePromptBlock(Snapshot{ZoneCount: 0})
	if got == "" {
		t.Fatal("expected non-empty block")
	}
	for _, want := range []string{
		"Farm setup mode",
		"no grow rooms yet",
		"add a grow room → connect edge device",
		"apply_grow_setup_pack",
		"apply_bootstrap_template",
		"wire-pi-relay-light",
		"Do not insert proposals",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestSetupModePromptBlock_WithZones(t *testing.T) {
	got := SetupModePromptBlock(Snapshot{ZoneCount: 3, ZoneNames: []string{"Veg", "Flower"}})
	if strings.Contains(got, "no grow rooms yet") {
		t.Fatal("zero-zone hint should not appear when zones exist")
	}
	if !strings.Contains(got, "Farm setup mode") {
		t.Fatal("expected setup header")
	}
}
