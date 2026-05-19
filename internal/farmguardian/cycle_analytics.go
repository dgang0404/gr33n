// Phase 28 WS3 — pull active-cycle analytics into Farm Guardian's live
// snapshot. Reuses the same Postgres queries the Phase 28 WS1 analytics
// API uses (GetFertigationAggregatesByCropCycle + GetCostTotalsByCropCycle)
// so the numbers Guardian sees match the numbers the operator sees in
// CropCycleSummary.vue.
//
// Design constraints:
//   - Best-effort: a per-cycle query failure must never block the chat
//     turn. The handler logs at WARN and renders whatever it has.
//   - Bounded prompt cost: only the first N active cycles get analytics
//     attached (SnapshotMaxAnalyticsCycles, default 3). Older cycles still
//     render their name + stage line via the existing path.
//   - Single-currency cost guard: matches the WS1 rule — cost_per_gram is
//     only emitted when costs live in exactly one currency, otherwise the
//     ratio is misleading.

package farmguardian

import (
	"context"
	"strings"
	"time"

	db "gr33n-api/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

// SnapshotMaxAnalyticsCycles caps how many active cycles get the rich
// per-cycle analytics attached to the snapshot. 3 keeps the prompt block
// well under the 200-token budget we set in the Phase 28 plan for this
// section, regardless of how many cycles are actually active.
const SnapshotMaxAnalyticsCycles = 3

// CycleAnalytics is the per-cycle, prompt-ready rollup the snapshot
// renders into a single bullet line. All optional fields are pointers so
// the renderer can omit absent values cleanly.
type CycleAnalytics struct {
	DurationDays  int64    // days since started_at (or 0 if start missing)
	EventCount    int64    // fertigation_events count
	TotalLiters   float64  // total volume_applied_liters
	LitersPerDay  *float64 // total_liters / duration_days (when both > 0)
	AvgECmSCm     float64  // average EC after-feed
	MinECmSCm     float64
	MaxECmSCm     float64
	AvgPH         float64
	YieldGrams    float64  // 0 when not yet recorded
	GramsPerDay   *float64 // yield_grams / duration_days (when both > 0)
	GramsPerLiter *float64 // yield_grams / total_liters (when both > 0)
	TotalExpenses float64  // single-currency expenses sum (0 when mixed/unknown)
	Currency      string   // populated only when costs live in exactly one currency
	CostPerGram   *float64 // total_expenses / yield_grams (single-currency only, when both > 0)
}

// Empty reports true when every numeric field is zero and every optional
// is nil — i.e. we have no useful per-cycle data and the renderer should
// skip the analytics line for this cycle. The cycle still appears in the
// snapshot via its existing Name/Strain/Stage entry.
func (a CycleAnalytics) Empty() bool {
	return a.EventCount == 0 &&
		a.TotalLiters == 0 &&
		a.YieldGrams == 0 &&
		a.TotalExpenses == 0 &&
		a.AvgECmSCm == 0 &&
		a.AvgPH == 0
}

// renderLine emits the inline string fragment that gets appended to a
// cycle's snapshot bullet. Empty CycleAnalytics renders as "" so callers
// can `if line != ""` cheaply.
//
// Format: "feed: 142 events / 980L (14.7L/d); EC 1.62 (1.12–2.05); pH 6.10;
//
//	cost: 312 USD; yield: 412g (6.06g/d); cost/g: 0.76 USD".
//
// Numbers are rounded for readability — Guardian only needs orientation,
// not 6-decimal precision.
func (a CycleAnalytics) renderLine() string {
	if a.Empty() {
		return ""
	}
	parts := []string{}
	if a.EventCount > 0 || a.TotalLiters > 0 {
		feed := "feed: "
		if a.EventCount > 0 {
			feed += formatInt(a.EventCount) + " events"
		}
		if a.TotalLiters > 0 {
			if a.EventCount > 0 {
				feed += " / "
			}
			feed += formatLiters(a.TotalLiters)
			if a.LitersPerDay != nil {
				feed += " (" + formatLitersPerDay(*a.LitersPerDay) + ")"
			}
		}
		parts = append(parts, feed)
	}
	if a.AvgECmSCm > 0 || a.MaxECmSCm > 0 {
		ec := "EC " + formatEC(a.AvgECmSCm)
		if a.MinECmSCm > 0 || a.MaxECmSCm > 0 {
			ec += " (" + formatEC(a.MinECmSCm) + "–" + formatEC(a.MaxECmSCm) + ")"
		}
		parts = append(parts, ec)
	}
	if a.AvgPH > 0 {
		parts = append(parts, "pH "+formatPH(a.AvgPH))
	}
	if a.TotalExpenses > 0 && a.Currency != "" {
		parts = append(parts, "cost: "+formatMoney(a.TotalExpenses)+" "+a.Currency)
	}
	if a.YieldGrams > 0 {
		yld := "yield: " + formatGrams(a.YieldGrams) + "g"
		if a.GramsPerDay != nil {
			yld += " (" + formatGramsPerDay(*a.GramsPerDay) + ")"
		}
		parts = append(parts, yld)
	}
	if a.CostPerGram != nil && a.Currency != "" {
		parts = append(parts, "cost/g: "+formatMoney(*a.CostPerGram)+" "+a.Currency)
	}
	return strings.Join(parts, "; ")
}

// fetchCycleAnalytics runs the two SQL aggregate calls and assembles a
// CycleAnalytics for a single crop cycle. Errors are returned to the
// caller so it can decide whether to fall back to the basic line; the
// caller is expected to log at WARN and continue rather than fail the
// turn.
func fetchCycleAnalytics(ctx context.Context, q *db.Queries, cycle db.Gr33nfertigationCropCycle) (CycleAnalytics, error) {
	out := CycleAnalytics{}
	out.DurationDays = durationDaysSinceStart(cycle.StartedAt, cycle.HarvestedAt)
	out.YieldGrams = numericToFloat64(cycle.YieldGrams)

	fert, err := q.GetFertigationAggregatesByCropCycle(ctx, cycle.ID)
	if err != nil {
		return out, err
	}
	out.EventCount = fert.EventCount
	out.TotalLiters = numericToFloat64(fert.TotalLiters)
	out.AvgECmSCm = numericToFloat64(fert.AvgECmSCm)
	out.MinECmSCm = numericToFloat64(fert.MinECmSCm)
	out.MaxECmSCm = numericToFloat64(fert.MaxECmSCm)
	out.AvgPH = numericToFloat64(fert.AvgPH)

	if out.DurationDays > 0 && out.TotalLiters > 0 {
		v := out.TotalLiters / float64(out.DurationDays)
		out.LitersPerDay = &v
	}
	if out.DurationDays > 0 && out.YieldGrams > 0 {
		v := out.YieldGrams / float64(out.DurationDays)
		out.GramsPerDay = &v
	}
	if out.TotalLiters > 0 && out.YieldGrams > 0 {
		v := out.YieldGrams / out.TotalLiters
		out.GramsPerLiter = &v
	}

	cid := cycle.ID
	costRows, err := q.GetCostTotalsByCropCycle(ctx, &cid)
	if err != nil {
		// Fertigation half is still useful — return what we have plus
		// the error so the caller can decide whether to log.
		return out, err
	}
	currencyTotals := map[string]float64{}
	for _, row := range costRows {
		k := strings.TrimSpace(row.Currency)
		currencyTotals[k] += numericToFloat64(row.Expense)
	}
	if len(currencyTotals) == 1 {
		for k, v := range currencyTotals {
			out.Currency = k
			out.TotalExpenses = v
		}
		if out.TotalExpenses > 0 && out.YieldGrams > 0 {
			v := out.TotalExpenses / out.YieldGrams
			out.CostPerGram = &v
		}
	}
	return out, nil
}

// durationDaysSinceStart mirrors the cropcycle handler's helper but is
// duplicated here so farmguardian doesn't depend on the cropcycle package
// (which would invert the import direction — cropcycle uses farmguardian
// for nothing, but circular imports are still worth pre-empting).
func durationDaysSinceStart(started, harvested pgtype.Date) int64 {
	if !started.Valid {
		return 0
	}
	end := harvested
	if !end.Valid {
		now := nowFunc().UTC()
		end = pgtype.Date{
			Time:  time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
			Valid: true,
		}
	}
	diff := end.Time.Sub(started.Time).Hours() / 24
	if diff < 0 {
		return 0
	}
	return int64(diff + 0.5)
}

// numericToFloat64 is intentionally duplicated from the cropcycle handler
// for the same reason as durationDaysSinceStart — keeps farmguardian
// import-light.
func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

// nowFunc is the indirection seam tests use to freeze the clock. Held as
// a var (not a const) so the snapshot tests can stub it without reaching
// into time.Now globally.
var nowFunc = time.Now
