package programrules_test

import (
	"errors"
	"testing"

	"gr33n-api/internal/fertigation/programrules"
)

func TestValidateIrrigationOnlyRejectsRecipe(t *testing.T) {
	rid := int64(1)
	err := programrules.ValidateCreateUpdate(true, &rid)
	if !errors.Is(err, programrules.ErrIrrigationOnlyNoRecipe) {
		t.Fatalf("expected ErrIrrigationOnlyNoRecipe, got %v", err)
	}
}

func TestNeedsMixBatch(t *testing.T) {
	if programrules.NeedsMixBatch(true, nil) {
		t.Fatal("irrigation_only should not need mix")
	}
	rid := int64(1)
	if !programrules.NeedsMixBatch(false, &rid) {
		t.Fatal("fertigation with recipe should need mix")
	}
}
