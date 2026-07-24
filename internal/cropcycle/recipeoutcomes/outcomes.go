package recipeoutcomes

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/cropcycle"
	db "gr33n-api/internal/db"
)

// MinSampleSize is the minimum harvested cycles before surfacing aggregate stats.
const MinSampleSize = 2

const maxSampleCycleIDs = 8

// Options filters farm-level recipe outcome aggregation.
type Options struct {
	CropKey             *string
	ApplicationRecipeID *int64
	IncludeCosts        bool
}

// RecipeOutcome is one recipe/revision rollup for a crop_key.
type RecipeOutcome struct {
	CropKey               string   `json:"crop_key"`
	CatalogDisplayName    string   `json:"catalog_display_name"`
	ApplicationRecipeID   int64    `json:"application_recipe_id"`
	ApplicationRecipeRevisionID *int64 `json:"application_recipe_revision_id,omitempty"`
	RecipeName            string   `json:"recipe_name"`
	CycleCount            int      `json:"cycle_count"`
	AvgYieldGrams         *float64 `json:"avg_yield_grams,omitempty"`
	MedianYieldGrams      *float64 `json:"median_yield_grams,omitempty"`
	MinYieldGrams         *float64 `json:"min_yield_grams,omitempty"`
	MaxYieldGrams         *float64 `json:"max_yield_grams,omitempty"`
	AvgCostPerGram        *float64 `json:"avg_cost_per_gram,omitempty"`
	CostCurrency          string   `json:"cost_currency,omitempty"`
	AvgDurationDays       *float64 `json:"avg_duration_days,omitempty"`
	SampleCycleIDs        []int64  `json:"sample_cycle_ids,omitempty"`
}

// Result is the API / Guardian payload.
type Result struct {
	FarmID                 int64           `json:"farm_id"`
	MinSampleSize          int             `json:"min_sample_size"`
	Outcomes               []RecipeOutcome `json:"outcomes"`
	MixedCycleCount        int             `json:"mixed_cycle_count"`
	UnattributedCycleCount int             `json:"unattributed_cycle_count"`
	InsufficientHistory    []RecipeOutcome `json:"insufficient_history,omitempty"`
}

type attributedCycle struct {
	cycleID   int64
	cropKey   string
	recipeID  int64
	revision  *int64
	yieldG    float64
	duration  int64
	costPerG  *float64
	currency  string
}

type buildQuerier interface {
	ListHarvestedCyclesForRecipeOutcomes(ctx context.Context, arg db.ListHarvestedCyclesForRecipeOutcomesParams) ([]db.ListHarvestedCyclesForRecipeOutcomesRow, error)
	ListRecipeAttributionHitsForCycle(ctx context.Context, arg db.ListRecipeAttributionHitsForCycleParams) ([]db.ListRecipeAttributionHitsForCycleRow, error)
	GetCostTotalsByCropCycle(ctx context.Context, cropCycleID *int64) ([]db.GetCostTotalsByCropCycleRow, error)
	GetRecipeByID(ctx context.Context, id int64) (db.Gr33nnaturalfarmingApplicationRecipe, error)
}

// Build aggregates harvested cycles with clear recipe attribution.
func Build(ctx context.Context, q buildQuerier, farmID int64, opt Options) (Result, error) {
	out := Result{
		FarmID:        farmID,
		MinSampleSize: MinSampleSize,
		Outcomes:      []RecipeOutcome{},
	}
	var cropKey *string
	if opt.CropKey != nil && *opt.CropKey != "" {
		ck := *opt.CropKey
		cropKey = &ck
	}
	rows, err := q.ListHarvestedCyclesForRecipeOutcomes(ctx, db.ListHarvestedCyclesForRecipeOutcomesParams{
		FarmID:  farmID,
		CropKey: cropKey,
	})
	if err != nil {
		return out, err
	}

	attributed := make([]attributedCycle, 0, len(rows))
	for _, row := range rows {
		from, to := cycleWindow(row.StartedAt, row.HarvestedAt)
		zoneID := row.ZoneID
		hits, err := q.ListRecipeAttributionHitsForCycle(ctx, db.ListRecipeAttributionHitsForCycleParams{
			FarmID: farmID,
			ZoneID: &zoneID,
			FromTs: from,
			ToTs:   to,
		})
		if err != nil {
			return out, err
		}
		key, mixed, total := AttributeCycle(sqlHitsToAttribution(hits))
		if total == 0 {
			out.UnattributedCycleCount++
			continue
		}
		if mixed {
			out.MixedCycleCount++
			continue
		}
		if opt.ApplicationRecipeID != nil && key.RecipeID != *opt.ApplicationRecipeID {
			continue
		}
		yieldG := numericToFloat64(row.YieldGrams)
		dur := durationDays(row.StartedAt, row.HarvestedAt)
		ac := attributedCycle{
			cycleID:  row.ID,
			cropKey:  derefString(row.CropKey),
			recipeID: key.RecipeID,
			revision: revisionPtr(key.RevisionID),
			yieldG:   yieldG,
			duration: dur,
		}
		if opt.IncludeCosts {
			cpg, cur, ok := costPerGramForCycle(ctx, q, row.ID, yieldG)
			if ok {
				ac.costPerG = &cpg
				ac.currency = cur
			}
		}
		attributed = append(attributed, ac)
	}

	groups := map[string][]attributedCycle{}
	for _, ac := range attributed {
		gk := groupKey(ac.cropKey, ac.recipeID, ac.revision)
		groups[gk] = append(groups[gk], ac)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, gk := range keys {
		cycles := groups[gk]
		if len(cycles) == 0 {
			continue
		}
		first := cycles[0]
		ro := RecipeOutcome{
			CropKey:                     first.cropKey,
			CatalogDisplayName:          cropcycle.CatalogDisplayName(first.cropKey),
			ApplicationRecipeID:         first.recipeID,
			ApplicationRecipeRevisionID: first.revision,
			CycleCount:                  len(cycles),
		}
		if rec, err := q.GetRecipeByID(ctx, first.recipeID); err == nil {
			ro.RecipeName = rec.Name
		}
		ids := make([]int64, 0, len(cycles))
		for _, c := range cycles {
			ids = append(ids, c.cycleID)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] > ids[j] })
		if len(ids) > maxSampleCycleIDs {
			ids = ids[:maxSampleCycleIDs]
		}
		ro.SampleCycleIDs = ids

		if len(cycles) < MinSampleSize {
			out.InsufficientHistory = append(out.InsufficientHistory, ro)
			continue
		}
		fillStats(&ro, cycles, opt.IncludeCosts)
		out.Outcomes = append(out.Outcomes, ro)
	}
	return out, nil
}

func sqlHitsToAttribution(rows []db.ListRecipeAttributionHitsForCycleRow) []Hit {
	out := make([]Hit, 0, len(rows))
	for _, r := range rows {
		if r.ApplicationRecipeID <= 0 {
			continue
		}
		var rev *int64
		if r.ApplicationRecipeRevisionID > 0 {
			rev = &r.ApplicationRecipeRevisionID
		}
		out = append(out, Hit{
			ApplicationRecipeID:           r.ApplicationRecipeID,
			ApplicationRecipeRevisionID: rev,
		})
	}
	return out
}

func fillStats(ro *RecipeOutcome, cycles []attributedCycle, includeCosts bool) {
	yields := make([]float64, 0, len(cycles))
	durations := make([]float64, 0, len(cycles))
	costs := make([]float64, 0, len(cycles))
	currency := ""
	for _, c := range cycles {
		yields = append(yields, c.yieldG)
		if c.duration > 0 {
			durations = append(durations, float64(c.duration))
		}
		if includeCosts && c.costPerG != nil {
			if currency == "" {
				currency = c.currency
			}
			if currency == c.currency {
				costs = append(costs, *c.costPerG)
			}
		}
	}
	sort.Float64s(yields)
	avgY := avgFloat64(yields)
	medY := medianFloat64(yields)
	ro.AvgYieldGrams = &avgY
	ro.MedianYieldGrams = &medY
	ro.MinYieldGrams = ptrFloat64(yields[0])
	ro.MaxYieldGrams = ptrFloat64(yields[len(yields)-1])
	if len(durations) > 0 {
		avgD := avgFloat64(durations)
		ro.AvgDurationDays = &avgD
	}
	if includeCosts && len(costs) == len(cycles) && len(costs) > 0 {
		avgC := avgFloat64(costs)
		ro.AvgCostPerGram = &avgC
		ro.CostCurrency = currency
	}
}

func costPerGramForCycle(ctx context.Context, q buildQuerier, cycleID int64, yieldG float64) (float64, string, bool) {
	if yieldG <= 0 {
		return 0, "", false
	}
	cycleIDArg := cycleID
	rows, err := q.GetCostTotalsByCropCycle(ctx, &cycleIDArg)
	if err != nil {
		return 0, "", false
	}
	currency := ""
	var expense float64
	for _, row := range rows {
		cur := trimString(row.Currency)
		if cur == "" {
			continue
		}
		if currency == "" {
			currency = cur
		}
		if currency != cur {
			return 0, "", false
		}
		expense += numericToFloat64(row.Expense)
	}
	if currency == "" || expense <= 0 {
		return 0, "", false
	}
	return expense / yieldG, currency, true
}

func cycleWindow(started, harvested pgtype.Date) (time.Time, time.Time) {
	from := time.Now().UTC().AddDate(0, -3, 0)
	if started.Valid {
		from = started.Time.UTC()
	}
	to := time.Now().UTC()
	if harvested.Valid {
		end := harvested.Time.UTC().Add(24 * time.Hour)
		if end.After(to) {
			to = end
		}
	}
	return from, to
}

func durationDays(started, harvested pgtype.Date) int64 {
	if !started.Valid {
		return 0
	}
	start := started.Time.UTC()
	end := time.Now().UTC()
	if harvested.Valid {
		end = harvested.Time.UTC()
	}
	days := int64(end.Sub(start).Hours()/24) + 1
	if days < 0 {
		return 0
	}
	return days
}

func revisionPtr(rev int64) *int64 {
	if rev == 0 {
		return nil
	}
	return &rev
}

func groupKey(cropKey string, recipeID int64, revision *int64) string {
	rev := "0"
	if revision != nil {
		rev = formatInt64(*revision)
	}
	return cropKey + "|" + formatInt64(recipeID) + "|" + rev
}

func formatInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func trimString(s string) string {
	return strings.TrimSpace(s)
}

func numericToFloat64(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

func avgFloat64(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func medianFloat64(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	n := len(vals)
	if n%2 == 1 {
		return vals[n/2]
	}
	return (vals[n/2-1] + vals[n/2]) / 2
}

func ptrFloat64(v float64) *float64 {
	return &v
}
