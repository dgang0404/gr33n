package commonscatalog

import (
	"encoding/json"
	"testing"
)

func TestParsePackBodyKinds(t *testing.T) {
	raw := json.RawMessage(`{"catalog_version":"gr33n.commons_catalog.v1","kind":"documentation_pack","readme_md":"# Hi"}`)
	b, err := ParsePackBody(raw)
	if err != nil {
		t.Fatal(err)
	}
	if b.Kind != KindDocumentationPack {
		t.Fatalf("kind %q", b.Kind)
	}
}

func TestValidatePublishRecipePack(t *testing.T) {
	if err := ValidatePublishBody(PackBody{Kind: KindFertigationRecipePack, Programs: nil}); err == nil {
		t.Fatal("expected error for empty programs")
	}
	err := ValidatePublishBody(PackBody{
		Kind: KindFertigationRecipePack,
		Programs: []RecipeProgram{{
			Name:              "Test",
			TotalVolumeLiters: 1,
			EcTriggerLow:      1,
			PhTriggerLow:      5.8,
			PhTriggerHigh:     6.2,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeSlug(t *testing.T) {
	s, err := NormalizeSlug("My-Farm-Pack-01")
	if err != nil || s != "my-farm-pack-01" {
		t.Fatalf("got %q err=%v", s, err)
	}
	if _, err := NormalizeSlug("bad slug!"); err == nil {
		t.Fatal("expected invalid slug error")
	}
}
