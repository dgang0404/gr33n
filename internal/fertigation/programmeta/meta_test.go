package programmeta

import "testing"

func TestCheckFit_VegProgramFlowerStage(t *testing.T) {
	m := Meta{
		RecommendedCropKeys: []string{"cannabis", "tomato"},
		RecommendedStages:   []string{"early_veg", "late_veg"},
		ECBandMSCM:          &ECBand{Min: 1.4, Max: 2.2},
	}
	fit := m.CheckFit("cannabis", "early_flower")
	if fit.OK {
		t.Fatal("expected stage mismatch")
	}
	if len(fit.Warnings) == 0 {
		t.Fatal("expected warnings")
	}
}

func TestCheckFit_MatchingVegGrow(t *testing.T) {
	m := Meta{
		RecommendedCropKeys: []string{"cannabis"},
		RecommendedStages:   []string{"late_veg"},
	}
	fit := m.CheckFit("cannabis", "late_veg")
	if !fit.OK || len(fit.Warnings) != 0 {
		t.Fatalf("unexpected fit: %#v", fit)
	}
}

func TestCheckFit_UntaggedProgram(t *testing.T) {
	fit := Meta{}.CheckFit("cannabis", "early_flower")
	if !fit.OK {
		t.Fatalf("untagged program should not warn: %#v", fit)
	}
}

func TestParse_Empty(t *testing.T) {
	if Parse(nil).HasCatalogTags() {
		t.Fatal("empty metadata")
	}
}
