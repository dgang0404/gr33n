package farmguardian

import "testing"

func TestExtractPlantDisplayName(t *testing.T) {
	cases := []struct {
		q    string
		want string
		ok   bool
	}{
		{"add my philodendron to Living Room with a light fertigation program", "Philodendron", true},
		{"create pothos in Bedroom", "Pothos", true},
		{"add plant to Flower Room", "", false},
		{"create a task to check humidity", "", false},
	}
	for _, c := range cases {
		got, ok := extractPlantDisplayName(c.q)
		if ok != c.ok || (c.ok && got != c.want) {
			t.Fatalf("extractPlantDisplayName(%q) = %q,%v want %q,%v", c.q, got, ok, c.want, c.ok)
		}
	}
}

func TestResolveZoneNameForSetupPack(t *testing.T) {
	snap := Snapshot{
		ZoneNames: []string{"Living Room", "Flower Room"},
	}
	name, ok := resolveZoneNameForSetupPack("add philodendron to Living Room", snap)
	if !ok || name != "Living Room" {
		t.Fatalf("got %q ok=%v", name, ok)
	}
	if _, ok := resolveZoneNameForSetupPack("add philodendron to Narnia", snap); ok {
		t.Fatal("expected no zone match for nonsense name")
	}
}

func TestBuildSetupPackArgs_Profiles(t *testing.T) {
	house := buildSetupPackArgs("house_plant", 12, "Living Room", "Philodendron")
	if house["profile"] != "house_plant" {
		t.Fatalf("profile %#v", house["profile"])
	}
	prog, _ := house["program"].(map[string]any)
	if prog["total_volume_liters"].(float64) != 0.5 {
		t.Fatalf("house volume %#v", prog["total_volume_liters"])
	}

	comm := buildSetupPackArgs("commercial_zone", 3, "Veg Room", "Tomato")
	prog2, _ := comm["program"].(map[string]any)
	if prog2["total_volume_liters"].(float64) != 95.0 {
		t.Fatalf("commercial volume %#v", prog2["total_volume_liters"])
	}
	cycle, _ := comm["cycle"].(map[string]any)
	if cycle["current_stage"] != "late_veg" {
		t.Fatalf("stage %#v", cycle["current_stage"])
	}
}

func TestPlantAlreadyOnFarm(t *testing.T) {
	snap := Snapshot{PlantNames: []string{"Philodendron (heartleaf)"}}
	if !plantAlreadyOnFarm("Philodendron", snap) {
		t.Fatal("expected duplicate plant detection")
	}
	if plantAlreadyOnFarm("Pothos", snap) {
		t.Fatal("pothos not on farm")
	}
}

func TestMatchSetupPackIntent_SkipsBusyZone(t *testing.T) {
	snap := Snapshot{
		ZoneNames:    []string{"Living Room"},
		ActiveCycles: []ActiveCycle{{ZoneName: "Living Room", Name: "Existing"}},
	}
	_, _, ok := matchSetupPackIntent(t.Context(), nil, 1,
		"add my philodendron to Living Room with fertigation program", snap)
	if ok {
		t.Fatal("expected skip when zone has active cycle")
	}
}

// Phase 44 WS8 — starter chip phrase must pass setup-pack intent patterns.
func TestMatchSetupPackIntent_StarterPhrase(t *testing.T) {
	msg := "Add my philodendron to Flower Room with a light fertigation program"
	if !setupPackVerbIntent.MatchString(msg) || !setupPackGrowIntent.MatchString(msg) {
		t.Fatal("starter phrase should match verb + grow intent patterns")
	}
	name, ok := extractPlantDisplayName(msg)
	if !ok || name != "Philodendron" {
		t.Fatalf("extractPlantDisplayName(%q) = %q,%v", msg, name, ok)
	}
	zone, ok := resolveZoneNameForSetupPack(msg, Snapshot{ZoneNames: []string{"Flower Room"}})
	if !ok || zone != "Flower Room" {
		t.Fatalf("resolveZoneNameForSetupPack = %q,%v", zone, ok)
	}
}

func TestInferSetupProfile(t *testing.T) {
	if inferSetupProfile("Living Room", "") != "house_plant" {
		t.Fatal("living room should be house_plant")
	}
	if inferSetupProfile("Veg Room", "") != "commercial_zone" {
		t.Fatal("veg room should be commercial_zone")
	}
}
