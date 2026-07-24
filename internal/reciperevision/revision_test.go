package reciperevision

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

func TestBuildSnapshot_roundTrip(t *testing.T) {
	desc := "drench"
	ratio := "1:10"
	notes := "base"
	unitID := int64(3)
	inputID := int64(9)

	var part pgtype.Numeric
	if err := part.Scan("2.5"); err != nil {
		t.Fatal(err)
	}

	raw, err := BuildSnapshot(db.Gr33nnaturalfarmingApplicationRecipe{
		ID:                    1,
		FarmID:                2,
		Name:                  "JMS Soil Drench",
		Description:           &desc,
		TargetApplicationType: db.Gr33nnaturalfarmingApplicationTargetEnumSoilDrench,
		DilutionRatio:         &ratio,
	}, []db.ListRecipeComponentsRow{{
		InputDefinitionID: inputID,
		InputName:         "JMS",
		PartValue:         part,
		PartUnitID:        &unitID,
		Notes:             &notes,
	}})
	if err != nil {
		t.Fatal(err)
	}

	var snap Snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Recipe.Name != "JMS Soil Drench" {
		t.Fatalf("recipe name: %q", snap.Recipe.Name)
	}
	if len(snap.Components) != 1 || snap.Components[0].PartValue != 2.5 {
		t.Fatalf("components: %+v", snap.Components)
	}
}
