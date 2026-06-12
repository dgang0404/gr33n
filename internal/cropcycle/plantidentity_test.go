package cropcycle

import (
	"testing"

	db "gr33n-api/internal/db"
)

func TestResolveCycleCropIdentity(t *testing.T) {
	ck := "tomato"
	plant := db.Gr33ncropsPlant{CropKey: &ck, DisplayName: "Cherry Tomatoes"}
	cycle := db.Gr33nfertigationCropCycle{Name: "Run 1"}
	id := ResolveCycleCropIdentity(cycle, &plant)
	if id.CropKey == nil || *id.CropKey != "tomato" {
		t.Fatalf("crop_key: %#v", id.CropKey)
	}
	if id.CatalogDisplayName == nil || *id.CatalogDisplayName == "" {
		t.Fatalf("catalog display name missing")
	}
}

func TestCatalogDisplayNameKnownCrop(t *testing.T) {
	dn := CatalogDisplayName("cannabis")
	if dn == "" {
		t.Fatal("expected display name for cannabis")
	}
}
