package croplibrary_test

import (
	"testing"

	"gr33n-api/internal/croplibrary"
)

func TestRegistry_ResolveAlias(t *testing.T) {
	cat, err := croplibrary.LoadCatalog(repoRoot(t), croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	reg := croplibrary.NewRegistry(cat)
	m, ok := reg.ResolveTerm("aubergine")
	if !ok || m.Key != "eggplant" || m.Kind != croplibrary.MentionCrop {
		t.Fatalf("aubergine: %+v ok=%v", m, ok)
	}
	m, ok = reg.ResolveTerm("wild_leek")
	if !ok || m.Key != "ramps" || m.Kind != croplibrary.MentionUnsupported {
		t.Fatalf("wild_leek: %+v", m)
	}
}

func TestRegistry_FindMentionsMultiCrop(t *testing.T) {
	cat, err := croplibrary.LoadCatalog(repoRoot(t), croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	reg := croplibrary.NewRegistry(cat)
	mentions := reg.FindMentions("Compare cucumber vs tomato feed targets")
	var keys []string
	for _, m := range mentions {
		if m.Kind == croplibrary.MentionCrop {
			keys = append(keys, m.Key)
		}
	}
	if len(keys) < 2 {
		t.Fatalf("want cucumber+tomato, got %v", keys)
	}
}

func TestRegistry_FruitTreeApple(t *testing.T) {
	cat, err := croplibrary.LoadCatalog(repoRoot(t), croplibrary.DefaultCatalogPath)
	if err != nil {
		t.Fatal(err)
	}
	reg := croplibrary.NewRegistry(cat)
	m, ok := reg.ResolveTerm("apple")
	if !ok || m.Key != "apple" {
		t.Fatalf("apple resolve: %+v ok=%v", m, ok)
	}
}
