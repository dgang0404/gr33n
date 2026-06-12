package lighting

import (
	"sort"
	"testing"
)

// canonicalPresetKeys mirrors var presets in handler.go — Phase 99 drift guard.
var canonicalPresetKeys = []string{
	"flower_12_12",
	"peas_22_2",
	"seedling_16_8",
	"veg_18_6",
}

func TestParity_PresetListKeys(t *testing.T) {
	list := PresetList()
	if len(list) != len(canonicalPresetKeys) {
		t.Fatalf("preset count: got %d want %d", len(list), len(canonicalPresetKeys))
	}
	got := make([]string, 0, len(list))
	for _, row := range list {
		key, _ := row["key"].(string)
		if key == "" {
			t.Fatalf("preset missing key: %#v", row)
		}
		got = append(got, key)
	}
	sort.Strings(got)
	want := append([]string(nil), canonicalPresetKeys...)
	sort.Strings(want)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("preset keys mismatch: got %v want %v", got, want)
		}
	}
}
