package recipeoutcomes

import "testing"

func TestAttributeCycle_dominantAtThreshold(t *testing.T) {
	t.Parallel()
	rev3 := int64(3)
	hits := []Hit{
		{ApplicationRecipeID: 10, ApplicationRecipeRevisionID: &rev3},
		{ApplicationRecipeID: 10, ApplicationRecipeRevisionID: &rev3},
		{ApplicationRecipeID: 10, ApplicationRecipeRevisionID: &rev3},
		{ApplicationRecipeID: 11, ApplicationRecipeRevisionID: nil},
	}
	key, mixed, total := AttributeCycle(hits)
	if total != 4 || mixed {
		t.Fatalf("expected dominant recipe 10, got mixed=%v total=%d", mixed, total)
	}
	if key.RecipeID != 10 || key.RevisionID == nil || *key.RevisionID != 3 {
		t.Fatalf("key = %+v", key)
	}
}

func TestAttributeCycle_mixedBelowThreshold(t *testing.T) {
	t.Parallel()
	hits := []Hit{
		{ApplicationRecipeID: 10},
		{ApplicationRecipeID: 11},
	}
	_, mixed, total := AttributeCycle(hits)
	if total != 2 || !mixed {
		t.Fatalf("expected mixed=true, got mixed=%v total=%d", mixed, total)
	}
}

func TestAttributeCycle_noHits(t *testing.T) {
	t.Parallel()
	_, mixed, total := AttributeCycle(nil)
	if total != 0 || mixed {
		t.Fatalf("expected empty attribution, got mixed=%v total=%d", mixed, total)
	}
}

func TestMedianFloat64(t *testing.T) {
	t.Parallel()
	if got := medianFloat64([]float64{100, 200, 300}); got != 200 {
		t.Fatalf("median = %v", got)
	}
	if got := medianFloat64([]float64{100, 200}); got != 150 {
		t.Fatalf("median = %v", got)
	}
}
