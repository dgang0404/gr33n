package farmguardian

import (
	"context"
	"strings"
	"testing"
)

func TestReadToolIDs_IncludesPhase210NaturalFarmingTools(t *testing.T) {
	want := map[string]bool{
		"lookup_process_catalog":              false,
		"suggest_process_from_material":       false,
		"summarize_natural_farming_inventory": false,
	}
	for _, id := range ReadToolIDs() {
		if _, ok := want[id]; ok {
			want[id] = true
		}
	}
	for id, ok := range want {
		if !ok {
			t.Fatalf("%s missing from ReadToolIDs", id)
		}
	}
}

func TestShouldRunLookupProcessCatalogReadIntent(t *testing.T) {
	cases := []struct {
		q    string
		want bool
	}{
		{"How do I make JMS?", true},
		{"What is JLF?", true},
		{"What EC should my tomato be?", false},
		{"goldenrod JLF method", false},
	}
	for _, tc := range cases {
		if got := shouldRunLookupProcessCatalogReadIntent(tc.q); got != tc.want {
			t.Fatalf("q=%q got %v want %v", tc.q, got, tc.want)
		}
	}
}

func TestShouldRunSuggestProcessFromMaterialReadIntent(t *testing.T) {
	if !shouldRunSuggestProcessFromMaterialReadIntent("Can I ferment goldenrod into JLF?") {
		t.Fatal("expected goldenrod material intent")
	}
	if shouldRunSuggestProcessFromMaterialReadIntent("list unread alerts") {
		t.Fatal("should not match unrelated question")
	}
}

func TestShouldRunSummarizeNaturalFarmingInventoryReadIntent(t *testing.T) {
	if !shouldRunSummarizeNaturalFarmingInventoryReadIntent("What ferments do I have ready?") {
		t.Fatal("expected inventory intent")
	}
	if shouldRunSummarizeNaturalFarmingInventoryReadIntent("What supplies are running low?") {
		t.Fatal("low stock uses summarize_farm_low_stock")
	}
}

func TestSuggestProcessFromMaterial_Goldenrod(t *testing.T) {
	block, err := SuggestProcessFromMaterial(context.Background(), nil, 0, "goldenrod biomass for cherry understory")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(block, "goldenrod") {
		t.Fatalf("expected goldenrod in block: %q", block)
	}
	if !strings.Contains(strings.ToLower(block), "jlf") {
		t.Fatalf("expected jlf process: %q", block)
	}
	if !strings.Contains(block, "1:100") {
		t.Fatalf("expected dilution band: %q", block)
	}
	if !strings.Contains(block, "extension_method") {
		t.Fatalf("expected extension_method tier: %q", block)
	}
}

func TestLookupProcessCatalog_JMS(t *testing.T) {
	block, err := LookupProcessCatalog("How do I make JMS?")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(block, "lookup_process_catalog") {
		t.Fatalf("missing header: %q", block)
	}
	if !strings.Contains(strings.ToLower(block), "jms") {
		t.Fatalf("expected JMS: %q", block)
	}
	if !strings.Contains(block, "natural-farming-jms.md") {
		t.Fatalf("expected guide path: %q", block)
	}
}
