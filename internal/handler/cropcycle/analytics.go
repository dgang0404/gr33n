// Phase 28 WS1 — crop cycle analytics: per-cycle summary + multi-cycle
// compare. Reporting-only (read), one SQL call per sub-object, JWT + farm
// member authorisation. CSV variants ride the same builder by switching on
// the `.csv` suffix on the URL path.

package cropcycle

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// MaxCompareCycles caps how many cycles can be compared at once. More than
// five columns side-by-side becomes a dashboard, not a comparison.
const MaxCompareCycles = 5

// summaryFertigation is the JSON-friendly view of FertigationAggregates.
type summaryFertigation struct {
	EventCount  int64   `json:"event_count"`
	TotalLiters float64 `json:"total_liters"`
	AvgECmSCm   float64 `json:"avg_ec_mscm"`
	MinECmSCm   float64 `json:"min_ec_mscm"`
	MaxECmSCm   float64 `json:"max_ec_mscm"`
	AvgPH       float64 `json:"avg_ph"`
}

type summaryCostCategory struct {
	Category string  `json:"category"`
	Currency string  `json:"currency"`
	Income   float64 `json:"income"`
	Expense  float64 `json:"expense"`
	Net      float64 `json:"net"`
	TxCount  int64   `json:"tx_count"`
}

type summaryCostTotal struct {
	Currency      string  `json:"currency"`
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
	Net           float64 `json:"net"`
}

type summaryCost struct {
	Totals     []summaryCostTotal    `json:"totals"`
	ByCategory []summaryCostCategory `json:"by_category"`
}

type summaryYield struct {
	Grams         float64  `json:"grams"`
	GramsPerLiter *float64 `json:"grams_per_liter"`
	GramsPerDay   *float64 `json:"grams_per_day"`
	CostPerGram   *float64 `json:"cost_per_gram"`
}

type summaryStage struct {
	Stage     string `json:"stage"`
	EnteredAt string `json:"entered_at,omitempty"`
}

// cycleSummary is the response shape for both summary and compare. Kept
// flat so the UI can render any sub-block independently and the CSV
// exporter only has to know about leaf fields.
type cycleSummary struct {
	Cycle                 db.Gr33nfertigationCropCycle `json:"cycle"`
	DurationDays          int64                        `json:"duration_days"`
	Fertigation           summaryFertigation           `json:"fertigation"`
	Cost                  summaryCost                  `json:"cost"`
	Yield                 summaryYield                 `json:"yield"`
	Stages                []summaryStage               `json:"stages"`
	StageHistorySupported bool                         `json:"stage_history_supported"`
}

// Summary — GET /crop-cycles/{id}/summary
//
// Returns the full per-cycle story: fertigation aggregates, cost totals
// (per-currency + per-category), yield metrics, and stage history. Auth:
// JWT + farm member (farm_id resolved from the cycle row). Stage history
// is currently a single-row "current stage entered at started_at" stand-in
// because the schema only stores current_stage — flagged by
// stage_history_supported reflects whether crop_cycle_stage_events has rows.
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	asCSV := strings.HasSuffix(r.URL.Path, ".csv")
	rawID := r.PathValue("id")
	rawID = strings.TrimSuffix(rawID, ".csv")
	cycleID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	cycle, err := h.q.GetCropCycleByID(r.Context(), cycleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, cycle.FarmID) {
		return
	}
	summary, err := h.buildSummary(r, cycle)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if asCSV {
		writeSummaryCSV(w, []cycleSummary{summary}, fmt.Sprintf("crop-cycle-%d-summary.csv", cycleID))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, summary)
}

// Compare — GET /farms/{id}/crop-cycles/compare?ids=1,2,3
//
// Returns a parallel array of cycleSummary objects for every id supplied,
// in the order supplied. All ids must belong to the URL farm; capped at
// MaxCompareCycles. The UI does the side-by-side rendering. JWT + farm
// member.
func (h *Handler) Compare(w http.ResponseWriter, r *http.Request) {
	asCSV := strings.HasSuffix(r.URL.Path, ".csv")
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	idsRaw := strings.TrimSpace(r.URL.Query().Get("ids"))
	if idsRaw == "" {
		httputil.WriteError(w, http.StatusBadRequest, "ids query parameter required (comma-separated crop cycle ids)")
		return
	}
	parts := strings.Split(idsRaw, ",")
	if len(parts) < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "at least one id is required")
		return
	}
	if len(parts) > MaxCompareCycles {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Sprintf("at most %d cycles can be compared per call", MaxCompareCycles))
		return
	}
	seen := make(map[int64]struct{}, len(parts))
	summaries := make([]cycleSummary, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid id %q", strings.TrimSpace(part)))
			return
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		cycle, err := h.q.GetCropCycleByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, fmt.Sprintf("crop cycle %d not found", id))
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if cycle.FarmID != farmID {
			httputil.WriteError(w, http.StatusBadRequest, fmt.Sprintf("crop cycle %d does not belong to this farm", id))
			return
		}
		summary, err := h.buildSummary(r, cycle)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		summaries = append(summaries, summary)
	}
	if asCSV {
		writeSummaryCSV(w, summaries, fmt.Sprintf("farm-%d-crop-cycles-compare.csv", farmID))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"cycles": summaries})
}

// buildSummary fans out one query per sub-block. Cheap to read; the
// alternative (one giant LEFT JOIN) is harder to maintain and only slightly
// faster on small farms.
func (h *Handler) buildSummary(r *http.Request, cycle db.Gr33nfertigationCropCycle) (cycleSummary, error) {
	out := cycleSummary{Cycle: cycle}
	out.DurationDays = durationDays(cycle.StartedAt, cycle.HarvestedAt)

	fert, err := h.q.GetFertigationAggregatesByCropCycle(r.Context(), &cycle.ID)
	if err != nil {
		return out, err
	}
	out.Fertigation = summaryFertigation{
		EventCount:  fert.EventCount,
		TotalLiters: numericToFloat64(fert.TotalLiters),
		AvgECmSCm:   numericToFloat64(fert.AvgEcMscm),
		MinECmSCm:   numericToFloat64(fert.MinEcMscm),
		MaxECmSCm:   numericToFloat64(fert.MaxEcMscm),
		AvgPH:       numericToFloat64(fert.AvgPh),
	}

	cycleIDForCosts := cycle.ID
	costRows, err := h.q.GetCostTotalsByCropCycle(r.Context(), &cycleIDForCosts)
	if err != nil {
		return out, err
	}
	currencyTotals := map[string]*summaryCostTotal{}
	for _, row := range costRows {
		income := numericToFloat64(row.Income)
		expense := numericToFloat64(row.Expense)
		net := numericToFloat64(row.Net)
		out.Cost.ByCategory = append(out.Cost.ByCategory, summaryCostCategory{
			Category: string(row.Category),
			Currency: strings.TrimSpace(row.Currency),
			Income:   income,
			Expense:  expense,
			Net:      net,
			TxCount:  row.TxCount,
		})
		k := strings.TrimSpace(row.Currency)
		if _, ok := currencyTotals[k]; !ok {
			currencyTotals[k] = &summaryCostTotal{Currency: k}
		}
		currencyTotals[k].TotalIncome += income
		currencyTotals[k].TotalExpenses += expense
		currencyTotals[k].Net += net
	}
	for _, t := range currencyTotals {
		out.Cost.Totals = append(out.Cost.Totals, *t)
	}
	sort.Slice(out.Cost.Totals, func(i, j int) bool {
		return out.Cost.Totals[i].Currency < out.Cost.Totals[j].Currency
	})

	out.Yield = buildYield(cycle, out.Fertigation, out.Cost, out.DurationDays)

	events, err := h.q.ListCropCycleStageEventsByCycle(r.Context(), cycle.ID)
	if err != nil {
		return out, err
	}
	if len(events) > 0 {
		out.Stages = make([]summaryStage, 0, len(events))
		for _, ev := range events {
			out.Stages = append(out.Stages, summaryStage{
				Stage:     formatStageLabelFarmer(string(ev.GrowthStage)),
				EnteredAt: ev.EnteredAt.Format("2006-01-02"),
			})
		}
		out.StageHistorySupported = true
	} else if cycle.CurrentStage != nil {
		entered := ""
		if cycle.StartedAt.Valid {
			entered = cycle.StartedAt.Time.Format("2006-01-02")
		}
		out.Stages = append(out.Stages, summaryStage{
			Stage:     formatStageLabelFarmer(string(*cycle.CurrentStage)),
			EnteredAt: entered,
		})
		out.StageHistorySupported = false
	}
	return out, nil
}

// formatStageLabelFarmer turns enum slugs into readable timeline copy.
func formatStageLabelFarmer(stage string) string {
	stage = strings.TrimSpace(stage)
	if stage == "" {
		return "—"
	}
	return strings.ReplaceAll(stage, "_", " ")
}

func buildYield(cycle db.Gr33nfertigationCropCycle, fert summaryFertigation, cost summaryCost, days int64) summaryYield {
	grams := numericToFloat64(cycle.YieldGrams)
	out := summaryYield{Grams: grams}
	if grams <= 0 {
		return out
	}
	if fert.TotalLiters > 0 {
		v := grams / fert.TotalLiters
		out.GramsPerLiter = &v
	}
	if days > 0 {
		v := grams / float64(days)
		out.GramsPerDay = &v
	}
	// Cost-per-gram only makes sense when the cycle has costs in exactly
	// one currency; mixing currencies blindly is misleading.
	if len(cost.Totals) == 1 {
		spend := cost.Totals[0].TotalExpenses
		if spend > 0 {
			v := spend / grams
			out.CostPerGram = &v
		}
	}
	return out
}

// durationDays returns the integer day count between started_at and
// harvested_at (or today if the cycle is still active). Returns 0 when
// started_at is missing.
func durationDays(started, harvested pgtype.Date) int64 {
	if !started.Valid {
		return 0
	}
	end := harvested
	if !end.Valid {
		// Active cycle — use today rounded to UTC midnight.
		now := nowDate()
		end = pgtype.Date{Time: now.Time, Valid: true}
	}
	diff := end.Time.Sub(started.Time).Hours() / 24
	if diff < 0 {
		return 0
	}
	return int64(diff + 0.5) // round half up
}

// numericToFloat64 is the same helper as cost.handler — duplicated locally
// rather than exported to keep packages decoupled.
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

// writeSummaryCSV emits a single wide CSV with one row per cycle. Columns
// are the leaf metrics in cycleSummary so anyone pasting into a spreadsheet
// gets a complete picture without nested JSON.
func writeSummaryCSV(w http.ResponseWriter, summaries []cycleSummary, filename string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	cw := csv.NewWriter(w)
	defer cw.Flush()

	_ = cw.Write([]string{
		"cycle_id", "cycle_name", "strain", "farm_id", "zone_id",
		"started_at", "harvested_at", "duration_days", "current_stage",
		"event_count", "total_liters",
		"avg_ec_mscm", "min_ec_mscm", "max_ec_mscm", "avg_ph",
		"yield_grams", "grams_per_liter", "grams_per_day", "cost_per_gram",
		"total_expenses", "total_income", "net", "currency",
	})
	for _, s := range summaries {
		started := ""
		if s.Cycle.StartedAt.Valid {
			started = s.Cycle.StartedAt.Time.Format("2006-01-02")
		}
		harvested := ""
		if s.Cycle.HarvestedAt.Valid {
			harvested = s.Cycle.HarvestedAt.Time.Format("2006-01-02")
		}
		strain := ""
		if s.Cycle.BatchLabel != nil {
			strain = *s.Cycle.BatchLabel
		}
		stage := ""
		if s.Cycle.CurrentStage != nil {
			stage = string(*s.Cycle.CurrentStage)
		}
		expenses, income, net, currency := costSummaryForCSV(s.Cost)
		_ = cw.Write([]string{
			strconv.FormatInt(s.Cycle.ID, 10),
			s.Cycle.Name,
			strain,
			strconv.FormatInt(s.Cycle.FarmID, 10),
			strconv.FormatInt(s.Cycle.ZoneID, 10),
			started, harvested,
			strconv.FormatInt(s.DurationDays, 10),
			stage,
			strconv.FormatInt(s.Fertigation.EventCount, 10),
			formatFloat(s.Fertigation.TotalLiters),
			formatFloat(s.Fertigation.AvgECmSCm),
			formatFloat(s.Fertigation.MinECmSCm),
			formatFloat(s.Fertigation.MaxECmSCm),
			formatFloat(s.Fertigation.AvgPH),
			formatFloat(s.Yield.Grams),
			formatOptFloat(s.Yield.GramsPerLiter),
			formatOptFloat(s.Yield.GramsPerDay),
			formatOptFloat(s.Yield.CostPerGram),
			formatFloat(expenses),
			formatFloat(income),
			formatFloat(net),
			currency,
		})
	}
}

func costSummaryForCSV(c summaryCost) (expenses, income, net float64, currency string) {
	if len(c.Totals) == 0 {
		return 0, 0, 0, ""
	}
	// Single-currency case is the common one: emit the real numbers. Multi
	// currency: emit zeros + a sentinel currency string so the operator
	// pulls the JSON for the breakdown.
	if len(c.Totals) == 1 {
		t := c.Totals[0]
		return t.TotalExpenses, t.TotalIncome, t.Net, t.Currency
	}
	currencies := make([]string, 0, len(c.Totals))
	for _, t := range c.Totals {
		currencies = append(currencies, t.Currency)
	}
	sort.Strings(currencies)
	return 0, 0, 0, "MIXED:" + strings.Join(currencies, "|")
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatOptFloat(p *float64) string {
	if p == nil {
		return ""
	}
	return formatFloat(*p)
}

// jsonResponse is a tiny convenience used in unit tests for asserting body
// shape. Avoids importing the cost package's helpers.
type jsonResponse = json.RawMessage

// nowDate returns "today" as a UTC-midnight pgtype.Date. Held as a var
// (not a const) so tests can stub it for deterministic duration math.
var nowDate = func() pgtype.Date {
	t := nowFunc().UTC()
	return pgtype.Date{
		Time:  time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}

// nowFunc is the indirection seam tests use to freeze the clock.
var nowFunc = time.Now
