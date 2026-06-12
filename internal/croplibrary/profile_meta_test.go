package croplibrary

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseProfileCatalogMeta(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"substrate":         "coco/rockwool",
		"watering_style":    "pulse_dryback",
		"runoff_pct_target": "10–20%",
		"moisture_guidance": "Allow dryback between pulses",
		"catalog_version":   4,
	})
	m := ParseProfileCatalogMeta(raw)
	if m.Substrate != "coco/rockwool" || m.WateringStyle != "pulse_dryback" {
		t.Fatalf("unexpected meta: %+v", m)
	}
	block := m.FormatMoistureBlock()
	for _, want := range []string{"Substrate:", "pulse dryback", "Runoff target:"} {
		if !strings.Contains(block, want) {
			t.Fatalf("block missing %q: %q", want, block)
		}
	}
}

func TestProfileCatalogMetaEmpty(t *testing.T) {
	if ParseProfileCatalogMeta(nil).HasMoistureInfo() {
		t.Fatal("expected empty meta")
	}
}
