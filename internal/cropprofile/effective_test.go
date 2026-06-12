package cropprofile

import "testing"

func TestSlugifyVariety(t *testing.T) {
	if got := SlugifyVariety("Blue Dream"); got != "blue_dream" {
		t.Fatalf("got %q want blue_dream", got)
	}
	if got := SlugifyVariety("  OG Kush  "); got != "og_kush" {
		t.Fatalf("got %q want og_kush", got)
	}
	if SlugifyVariety("") != "" {
		t.Fatal("empty slug")
	}
}

func TestGeneticsCropKey(t *testing.T) {
	want := "genetics:cannabis:blue_dream"
	if got := GeneticsCropKey("cannabis", "blue_dream"); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
