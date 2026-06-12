package domainenums

import "testing"

func TestAll_GrowthStages(t *testing.T) {
	p := All()
	if len(p.GrowthStages) != 11 {
		t.Fatalf("growth_stages: want 11, got %d", len(p.GrowthStages))
	}
	if p.GrowthStages[4].Value != "transition" || p.GrowthStages[8].Value != "flush" {
		t.Fatalf("unexpected growth stage order: %#v", p.GrowthStages)
	}
	if p.GrowthStages[0].Label != "clone" {
		t.Fatalf("label humanize: %#v", p.GrowthStages[0])
	}
}
