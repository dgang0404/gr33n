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

func TestAll_ZoneVocabulary(t *testing.T) {
	p := All()
	if len(p.ZoneTypes) != 8 {
		t.Fatalf("zone_types: want 8, got %d", len(p.ZoneTypes))
	}
	wizardCount := 0
	for _, zt := range p.ZoneTypes {
		if zt.WizardVisible {
			wizardCount++
		}
		if zt.Value == "veg" && zt.Label != "Veg room (legacy)" {
			t.Fatalf("veg label: %#v", zt)
		}
	}
	if wizardCount != 3 {
		t.Fatalf("wizard_visible zone types: want 3, got %d", wizardCount)
	}
	if len(p.GreenhouseCoverTypes) != 3 {
		t.Fatalf("greenhouse_cover_types: want 3, got %d", len(p.GreenhouseCoverTypes))
	}
	if !IsValidGreenhouseCoverType("film") || IsValidGreenhouseCoverType("canvas") {
		t.Fatal("cover type validation mismatch")
	}
	if len(p.GreenhouseAutomationPolicies) != 3 {
		t.Fatalf("greenhouse_automation_policies: want 3, got %d", len(p.GreenhouseAutomationPolicies))
	}
	if GreenhouseCoverTypeLabel("film") != "Film / poly" {
		t.Fatalf("cover label: %q", GreenhouseCoverTypeLabel("film"))
	}
	if GreenhouseAutomationPolicyLabel("auto") != "Auto (sensor rules)" {
		t.Fatalf("policy label: %q", GreenhouseAutomationPolicyLabel("auto"))
	}
}
