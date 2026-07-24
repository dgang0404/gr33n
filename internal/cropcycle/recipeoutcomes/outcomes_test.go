package recipeoutcomes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

type stubBuildQ struct {
	cycles  []db.ListHarvestedCyclesForRecipeOutcomesRow
	hits    map[int64][]db.ListRecipeAttributionHitsForCycleRow
	costs   map[int64][]db.GetCostTotalsByCropCycleRow
	recipes map[int64]db.Gr33nnaturalfarmingApplicationRecipe
}

func (s *stubBuildQ) ListHarvestedCyclesForRecipeOutcomes(_ context.Context, _ db.ListHarvestedCyclesForRecipeOutcomesParams) ([]db.ListHarvestedCyclesForRecipeOutcomesRow, error) {
	return s.cycles, nil
}

func (s *stubBuildQ) ListRecipeAttributionHitsForCycle(_ context.Context, _ db.ListRecipeAttributionHitsForCycleParams) ([]db.ListRecipeAttributionHitsForCycleRow, error) {
	return nil, nil
}

func (s *stubBuildQ) GetCostTotalsByCropCycle(_ context.Context, cropCycleID *int64) ([]db.GetCostTotalsByCropCycleRow, error) {
	if cropCycleID == nil {
		return nil, nil
	}
	return s.costs[*cropCycleID], nil
}

func (s *stubBuildQ) GetRecipeByID(_ context.Context, id int64) (db.Gr33nnaturalfarmingApplicationRecipe, error) {
	if rec, ok := s.recipes[id]; ok {
		return rec, nil
	}
	return db.Gr33nnaturalfarmingApplicationRecipe{}, fmt.Errorf("recipe %d not found", id)
}

func pgNumeric(v float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%f", v))
	return n
}

func pgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t.UTC(), Valid: true}
}

func hit(recipeID int64, rev int64) db.ListRecipeAttributionHitsForCycleRow {
	return db.ListRecipeAttributionHitsForCycleRow{
		ApplicationRecipeID:         recipeID,
		ApplicationRecipeRevisionID: rev,
	}
}

type seqHitsStub struct {
	stubBuildQ
	call int
}

func (s *seqHitsStub) ListRecipeAttributionHitsForCycle(_ context.Context, _ db.ListRecipeAttributionHitsForCycleParams) ([]db.ListRecipeAttributionHitsForCycleRow, error) {
	if s.call >= len(s.cycles) {
		return nil, nil
	}
	id := s.cycles[s.call].ID
	s.call++
	return s.hits[id], nil
}

func TestFillStats_mixedCurrencyOmitsAvgCost(t *testing.T) {
	t.Parallel()
	cpgUSD := 0.2
	cpgEUR := 0.3
	ro := RecipeOutcome{CycleCount: 2}
	fillStats(&ro, []attributedCycle{
		{yieldG: 100, costPerG: &cpgUSD, currency: "USD"},
		{yieldG: 200, costPerG: &cpgEUR, currency: "EUR"},
	}, true)
	if ro.AvgCostPerGram != nil {
		t.Fatalf("expected no avg cost on mixed currency, got %v", *ro.AvgCostPerGram)
	}
}

func TestFillStats_singleCurrencyIncludesAvgCost(t *testing.T) {
	t.Parallel()
	cpg1 := 0.2
	cpg2 := 0.4
	ro := RecipeOutcome{CycleCount: 2}
	fillStats(&ro, []attributedCycle{
		{yieldG: 100, costPerG: &cpg1, currency: "USD"},
		{yieldG: 200, costPerG: &cpg2, currency: "USD"},
	}, true)
	if ro.AvgCostPerGram == nil || ro.CostCurrency != "USD" {
		t.Fatalf("avg cost = %v currency = %q", ro.AvgCostPerGram, ro.CostCurrency)
	}
	if *ro.AvgCostPerGram < 0.29 || *ro.AvgCostPerGram > 0.31 {
		t.Fatalf("avg cost = %v want ~0.3", *ro.AvgCostPerGram)
	}
}

func TestCostPerGramForCycle_zeroYieldExcluded(t *testing.T) {
	t.Parallel()
	q := &stubBuildQ{
		costs: map[int64][]db.GetCostTotalsByCropCycleRow{
			1: {{Currency: "USD", Expense: pgNumeric(10)}},
		},
	}
	_, _, ok := costPerGramForCycle(context.Background(), q, 1, 0)
	if ok {
		t.Fatal("expected zero-yield cycle to skip cost per gram")
	}
}

func TestBuild_excludesMixedRecipeCycles(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	q := &seqHitsStub{
		stubBuildQ: stubBuildQ{
			cycles: []db.ListHarvestedCyclesForRecipeOutcomesRow{
				{ID: 1, ZoneID: 10, StartedAt: pgDate(start), HarvestedAt: pgDate(start.AddDate(0, 2, 0)), YieldGrams: pgNumeric(100), CropKey: strPtr("tomato")},
			},
			hits: map[int64][]db.ListRecipeAttributionHitsForCycleRow{
				1: {hit(10, 1), hit(11, 1)},
			},
			recipes: map[int64]db.Gr33nnaturalfarmingApplicationRecipe{
				10: {ID: 10, Name: "JMS"},
				11: {ID: 11, Name: "FPJ"},
			},
		},
	}
	result, err := Build(context.Background(), q, 1, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if result.MixedCycleCount != 1 || len(result.Outcomes) != 0 {
		t.Fatalf("mixed=%d outcomes=%d", result.MixedCycleCount, len(result.Outcomes))
	}
}

func TestBuild_insufficientHistoryBelowMinSample(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rev := int64(3)
	q := &seqHitsStub{
		stubBuildQ: stubBuildQ{
			cycles: []db.ListHarvestedCyclesForRecipeOutcomesRow{
				{ID: 1, ZoneID: 10, StartedAt: pgDate(start), HarvestedAt: pgDate(start.AddDate(0, 2, 0)), YieldGrams: pgNumeric(180), CropKey: strPtr("tomato")},
			},
			hits: map[int64][]db.ListRecipeAttributionHitsForCycleRow{
				1: {hit(10, rev), hit(10, rev), hit(10, rev)},
			},
			recipes: map[int64]db.Gr33nnaturalfarmingApplicationRecipe{
				10: {ID: 10, Name: "JMS Foliar"},
			},
		},
	}
	result, err := Build(context.Background(), q, 1, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Outcomes) != 0 || len(result.InsufficientHistory) != 1 {
		t.Fatalf("outcomes=%d insufficient=%d", len(result.Outcomes), len(result.InsufficientHistory))
	}
	if result.InsufficientHistory[0].AvgYieldGrams != nil {
		t.Fatal("single-cycle insufficient history must not surface avg yield")
	}
}

func TestBuild_surfacesStatsAtMinSample(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rev := int64(3)
	makeCycle := func(id int64, yield float64) db.ListHarvestedCyclesForRecipeOutcomesRow {
		return db.ListHarvestedCyclesForRecipeOutcomesRow{
			ID: id, ZoneID: 10, StartedAt: pgDate(start), HarvestedAt: pgDate(start.AddDate(0, 2, 0)),
			YieldGrams: pgNumeric(yield), CropKey: strPtr("tomato"),
		}
	}
	q := &seqHitsStub{
		stubBuildQ: stubBuildQ{
			cycles: []db.ListHarvestedCyclesForRecipeOutcomesRow{
				makeCycle(1, 140),
				makeCycle(2, 220),
			},
			hits: map[int64][]db.ListRecipeAttributionHitsForCycleRow{
				1: {hit(10, rev), hit(10, rev), hit(10, rev)},
				2: {hit(10, rev), hit(10, rev), hit(10, rev)},
			},
			costs: map[int64][]db.GetCostTotalsByCropCycleRow{
				1: {{Currency: "USD", Expense: pgNumeric(28)}},
				2: {{Currency: "USD", Expense: pgNumeric(44)}},
			},
			recipes: map[int64]db.Gr33nnaturalfarmingApplicationRecipe{
				10: {ID: 10, Name: "JMS Foliar"},
			},
		},
	}
	result, err := Build(context.Background(), q, 1, Options{IncludeCosts: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Outcomes) != 1 {
		t.Fatalf("outcomes=%d", len(result.Outcomes))
	}
	ro := result.Outcomes[0]
	if ro.CycleCount != MinSampleSize || ro.AvgYieldGrams == nil || ro.AvgCostPerGram == nil {
		t.Fatalf("ro=%+v", ro)
	}
	if *ro.AvgYieldGrams != 180 {
		t.Fatalf("avg yield = %v want 180", *ro.AvgYieldGrams)
	}
}

func strPtr(s string) *string {
	return &s
}
