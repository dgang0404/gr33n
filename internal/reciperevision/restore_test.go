package reciperevision

import (
	"encoding/json"
	"testing"

	db "gr33n-api/internal/db"
)

func TestRevisionSummary(t *testing.T) {
	ratio := "1:10"
	raw, err := json.Marshal(Snapshot{
		Recipe: RecipeSnapshot{DilutionRatio: &ratio},
		Components: []ComponentSnapshot{
			{InputDefinitionID: 1, InputName: "JMS", PartValue: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	rev := db.Gr33nnaturalfarmingApplicationRecipeRevision{
		Snapshot: json.RawMessage(raw),
	}
	dilution, count, err := RevisionSummary(rev)
	if err != nil {
		t.Fatal(err)
	}
	if dilution != "1:10" || count != 1 {
		t.Fatalf("got dilution=%q count=%d", dilution, count)
	}
}
