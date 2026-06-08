// Phase 55 WS1 — ops / grow / money read enrichments (Confirm N/A).

package farmguardian

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

var (
	cycleIDPattern = regexp.MustCompile(`(?i)\b(?:crop\s*)?cycle\s*#?\s*(\d+)\b`)
	summarizeCycleCostIntent = regexp.MustCompile(`(?i)\b(cost|spent|spending|expense|receipt|dollar|money)\b.*\b(cycle|grow|room|harvest|flower|veg)\b|\b(cycle|grow|room)\b.*\b(cost|spent|spending)\b|\bcost\s+so\s+far\b|\bcost\s+per\s+gram\b`)
	summarizeFarmSpendingIntent = regexp.MustCompile(`(?i)\b(spending|spent|expenses?|receipts?|money)\b.*\b(month|category|categories)\b|\b(month|this\s+month)\b.*\b(spend|spent|expenses?)\b|\bspending\s+by\s+category\b|\bbiggest\s+spend|\bwhat did i spend\b|\bspent this month\b|\bexplain this month`)
	restockPriorityIntent = regexp.MustCompile(`(?i)\brestock\s+first\b|\bwhat\s+should\s+i\s+restock\b|\brestock\s+priority\b|\bpriority\s+restock\b|\breorder\s+first\b`)
	summarizeActiveGrowsIntent = regexp.MustCompile(`(?i)\bwhat(?:'s|s|\s+is)\s+growing\b|\bactive\s+grows?\b|\bgrowing\s+where\b|\bwhat\s+am\s+i\s+growing\b|\bactive\s+cycles?\b`)
)

const readToolsMaxLowStockLines = 12
const readToolsMaxSpendingCategories = 8
const readToolsMaxActiveGrows = 12

func shouldRunSummarizeCycleCostReadIntent(question string, ref *ContextRef) bool {
	q := strings.TrimSpace(question)
	if q == "" && (ref == nil || ref.CropCycleID <= 0) {
		return false
	}
	if ref != nil && (ref.CropCycleID > 0 || strings.EqualFold(ref.Type, "crop_cycle") || strings.EqualFold(ref.Type, "cycle")) {
		if q == "" || summarizeCycleCostIntent.MatchString(q) || strings.Contains(strings.ToLower(q), "cost") {
			return true
		}
	}
	return summarizeCycleCostIntent.MatchString(q)
}

func shouldRunSummarizeFarmSpendingReadIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if shouldRunSummarizeCycleCostReadIntent(q, nil) && !summarizeFarmSpendingIntent.MatchString(q) {
		return false
	}
	return summarizeFarmSpendingIntent.MatchString(q)
}

func shouldRunRestockPriorityReadIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if restockPriorityIntent.MatchString(q) {
		return true
	}
	lower := strings.ToLower(q)
	return strings.Contains(lower, "restock first") || strings.Contains(lower, "what should i restock")
}

func shouldRunSummarizeActiveGrowsReadIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if matchListPlantsIntent(q) {
		return false
	}
	return summarizeActiveGrowsIntent.MatchString(q)
}

func resolveCycleForCost(ctx context.Context, q db.Querier, farmID int64, question string, snap Snapshot, ref *ContextRef) (db.Gr33nfertigationCropCycle, bool) {
	if ref != nil {
		if id := ref.CropCycleID; id > 0 {
			if c, err := q.GetCropCycleByID(ctx, id); err == nil && c.FarmID == farmID {
				return c, true
			}
		}
		refType := strings.ToLower(strings.TrimSpace(ref.Type))
		if (refType == "crop_cycle" || refType == "cycle") && ref.ID > 0 {
			if c, err := q.GetCropCycleByID(ctx, ref.ID); err == nil && c.FarmID == farmID {
				return c, true
			}
		}
		if refType == "zone" && ref.ID > 0 {
			if c, ok := activeCycleForZoneID(ctx, q, farmID, ref.ID); ok {
				return c, true
			}
		}
	}

	if m := cycleIDPattern.FindStringSubmatch(question); len(m) > 1 {
		if id, err := parseInt64(m[1]); err == nil && id > 0 {
			if c, err := q.GetCropCycleByID(ctx, id); err == nil && c.FarmID == farmID {
				return c, true
			}
		}
	}

	if zone, ok := resolveZoneForSummary(ctx, q, farmID, question, snap); ok {
		if c, found := activeCycleForZoneID(ctx, q, farmID, zone.ID); found {
			return c, true
		}
	}

	for _, ac := range snap.ActiveCycles {
		if ac.ID <= 0 {
			continue
		}
		lowerQ := strings.ToLower(question)
		if ac.ZoneName != "" && strings.Contains(lowerQ, strings.ToLower(ac.ZoneName)) {
			if c, err := q.GetCropCycleByID(ctx, ac.ID); err == nil && c.FarmID == farmID {
				return c, true
			}
		}
		if ac.Name != "" && strings.Contains(lowerQ, strings.ToLower(ac.Name)) {
			if c, err := q.GetCropCycleByID(ctx, ac.ID); err == nil && c.FarmID == farmID {
				return c, true
			}
		}
	}

	if len(snap.ActiveCycles) == 1 && snap.ActiveCycles[0].ID > 0 {
		if c, err := q.GetCropCycleByID(ctx, snap.ActiveCycles[0].ID); err == nil && c.FarmID == farmID {
			return c, true
		}
	}

	return db.Gr33nfertigationCropCycle{}, false
}

func activeCycleForZoneID(ctx context.Context, q db.Querier, farmID, zoneID int64) (db.Gr33nfertigationCropCycle, bool) {
	c, err := q.GetActiveCropCycleForZone(ctx, zoneID)
	if err != nil || c.FarmID != farmID {
		return db.Gr33nfertigationCropCycle{}, false
	}
	return c, true
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

func renderSummarizeCycleCost(ctx context.Context, q db.Querier, farmID int64, question string, snap Snapshot, ref *ContextRef) (string, error) {
	cycle, ok := resolveCycleForCost(ctx, q, farmID, question, snap, ref)
	if !ok {
		return "summarize_cycle_cost: no matching crop cycle — ask which room or cycle #.", nil
	}

	zoneLabel := ""
	if z, err := q.GetZoneByID(ctx, cycle.ZoneID); err == nil {
		zoneLabel = strings.TrimSpace(z.Name)
	}

	cid := cycle.ID
	rows, err := q.GetCostTotalsByCropCycle(ctx, &cid)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("summarize_cycle_cost — cycle #%d %s", cycle.ID, strings.TrimSpace(cycle.Name)))
	if zoneLabel != "" {
		b.WriteString(fmt.Sprintf(" (%s)", zoneLabel))
	}
	if cycle.CurrentStage != nil {
		b.WriteString(fmt.Sprintf("; stage %s", string(*cycle.CurrentStage)))
	}
	if !cycle.IsActive {
		b.WriteString("; status harvested/inactive")
	} else {
		b.WriteString("; status active")
	}

	var expenseTotal float64
	currency := ""
	type catLine struct {
		cat string
		exp float64
	}
	var cats []catLine
	for _, row := range rows {
		exp := numericToFloat64(row.Expense)
		expenseTotal += exp
		if currency == "" {
			currency = strings.TrimSpace(row.Currency)
		}
		if exp > 0 {
			cats = append(cats, catLine{cat: humanizeCostCategory(string(row.Category)), exp: exp})
		}
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].exp > cats[j].exp })

	if expenseTotal <= 0 && len(cats) == 0 {
		b.WriteString("\nNo tagged expenses for this cycle yet.")
		return b.String(), nil
	}

	b.WriteString(fmt.Sprintf("\nTotal spent: %.2f %s", expenseTotal, currency))
	for i, c := range cats {
		if i >= 6 {
			b.WriteString(fmt.Sprintf("\n(+ %d more categories)", len(cats)-6))
			break
		}
		b.WriteString(fmt.Sprintf("\n- %s: %.2f", c.cat, c.exp))
	}

	yieldG := numericToFloat64(cycle.YieldGrams)
	if cycle.YieldGrams.Valid && yieldG > 0 && expenseTotal > 0 {
		b.WriteString(fmt.Sprintf("\nCost per gram: %.4f %s/g", expenseTotal/yieldG, currency))
	}
	return b.String(), nil
}

func renderSummarizeFarmSpending(ctx context.Context, q db.Querier, farmID int64) (string, error) {
	farmLabel := fmt.Sprintf("farm #%d", farmID)
	if farm, err := q.GetFarmByID(ctx, farmID); err == nil {
		if name := strings.TrimSpace(farm.Name); name != "" {
			farmLabel = name
		}
	}

	now := nowFunc().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	rows, err := q.GetCostCategoryTotalsByFarmForYear(ctx, db.GetCostCategoryTotalsByFarmForYearParams{
		FarmID: farmID,
		Column2: pgtype.Date{
			Time:  monthStart,
			Valid: true,
		},
		Column3: pgtype.Date{
			Time:  monthEnd,
			Valid: true,
		},
	})
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("summarize_farm_spending — %s (%s)", farmLabel, monthStart.Format("January 2006")))

	type catRollup struct {
		label    string
		expense  float64
		income   float64
		txCount  int64
		currency string
	}
	byCat := map[string]*catRollup{}
	var totalExpense, totalIncome float64
	for _, row := range rows {
		label := humanizeCostCategory(string(row.Category))
		r := byCat[label]
		if r == nil {
			r = &catRollup{label: label, currency: strings.TrimSpace(row.Currency)}
			byCat[label] = r
		}
		r.expense += numericToFloat64(row.Expense)
		r.income += numericToFloat64(row.Income)
		r.txCount += row.TxCount
		totalExpense += numericToFloat64(row.Expense)
		totalIncome += numericToFloat64(row.Income)
	}

	if len(byCat) == 0 {
		b.WriteString("\nNo receipts logged this month yet.")
		return b.String(), nil
	}

	b.WriteString(fmt.Sprintf("\nMonth total — spent: %.2f; received: %.2f; net: %.2f", totalExpense, totalIncome, totalIncome-totalExpense))

	rollups := make([]*catRollup, 0, len(byCat))
	for _, r := range byCat {
		rollups = append(rollups, r)
	}
	sort.Slice(rollups, func(i, j int) bool { return rollups[i].expense > rollups[j].expense })

	for i, r := range rollups {
		if i >= readToolsMaxSpendingCategories {
			b.WriteString(fmt.Sprintf("\n(+ %d more categories)", len(rollups)-readToolsMaxSpendingCategories))
			break
		}
		if r.expense > 0 {
			b.WriteString(fmt.Sprintf("\n- %s: %.2f spent (%d receipt%s)", r.label, r.expense, r.txCount, pluralS(r.txCount)))
		}
	}
	return b.String(), nil
}

type lowStockPriorityRow struct {
	row      db.ListLowStockBatchesByFarmRow
	remaining float64
	threshold float64
	ratio    float64
}

func renderRestockPriority(ctx context.Context, q db.Querier, farmID int64) (string, error) {
	farmLabel := fmt.Sprintf("farm #%d", farmID)
	if farm, err := q.GetFarmByID(ctx, farmID); err == nil {
		if name := strings.TrimSpace(farm.Name); name != "" {
			farmLabel = name
		}
	}

	rows, err := q.ListLowStockBatchesByFarm(ctx, farmID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("restock_priority — " + farmLabel)
	if len(rows) == 0 {
		b.WriteString("\nNo batches below their low-stock threshold right now.")
		return b.String(), nil
	}

	prioritized := make([]lowStockPriorityRow, 0, len(rows))
	for _, row := range rows {
		rem := numericToFloat64(row.CurrentQuantityRemaining)
		thr := numericToFloat64(row.LowStockThreshold)
		ratio := 1.0
		if thr > 0 {
			ratio = rem / thr
		}
		prioritized = append(prioritized, lowStockPriorityRow{row: row, remaining: rem, threshold: thr, ratio: ratio})
	}
	sort.Slice(prioritized, func(i, j int) bool {
		if prioritized[i].ratio == prioritized[j].ratio {
			return prioritized[i].remaining < prioritized[j].remaining
		}
		return prioritized[i].ratio < prioritized[j].ratio
	})

	for i, item := range prioritized {
		if i >= readToolsMaxLowStockLines {
			b.WriteString(fmt.Sprintf("\n(+ %d more low-stock batches)", len(prioritized)-readToolsMaxLowStockLines))
			break
		}
		unit := batchQuantityUnitLabel(item.row.QuantityUnitID)
		line := fmt.Sprintf("\n%d. %s — %.2f / threshold %.2f", i+1, strings.TrimSpace(item.row.InputName), item.remaining, item.threshold)
		if unit != "" {
			line += " " + unit
		}
		line += fmt.Sprintf("; batch #%d", item.row.ID)
		b.WriteString(line)
	}
	b.WriteString("\nRestock in Supplies hub (+ Add qty); Guardian cannot change stock without operator action in UI.")
	return b.String(), nil
}

func renderSummarizeActiveGrows(ctx context.Context, q db.Querier, farmID int64) (string, error) {
	farmLabel := fmt.Sprintf("farm #%d", farmID)
	if farm, err := q.GetFarmByID(ctx, farmID); err == nil {
		if name := strings.TrimSpace(farm.Name); name != "" {
			farmLabel = name
		}
	}

	cycles, err := q.ListCropCyclesByFarm(ctx, farmID)
	if err != nil {
		return "", err
	}

	zones, _ := q.ListZonesByFarm(ctx, farmID)
	zoneName := map[int64]string{}
	for _, z := range zones {
		zoneName[z.ID] = strings.TrimSpace(z.Name)
	}

	var active []db.Gr33nfertigationCropCycle
	for _, c := range cycles {
		if c.IsActive {
			active = append(active, c)
		}
	}

	var b strings.Builder
	b.WriteString("summarize_active_grows — " + farmLabel)
	if len(active) == 0 {
		b.WriteString("\nNo active grows right now — start one from a zone Overview or Plants.")
		return b.String(), nil
	}

	sort.Slice(active, func(i, j int) bool {
		zi, zj := zoneName[active[i].ZoneID], zoneName[active[j].ZoneID]
		if zi == zj {
			return active[i].Name < active[j].Name
		}
		return zi < zj
	})

	for i, c := range active {
		if i >= readToolsMaxActiveGrows {
			b.WriteString(fmt.Sprintf("\n(+ %d more active grows)", len(active)-readToolsMaxActiveGrows))
			break
		}
		zn := zoneName[c.ZoneID]
		if zn == "" {
			zn = fmt.Sprintf("zone #%d", c.ZoneID)
		}
		line := fmt.Sprintf("\n- %s: %s (cycle #%d)", zn, strings.TrimSpace(c.Name), c.ID)
		if c.CurrentStage != nil {
			line += fmt.Sprintf("; stage %s", string(*c.CurrentStage))
		}
		if c.StrainOrVariety != nil && strings.TrimSpace(*c.StrainOrVariety) != "" {
			line += "; strain " + strings.TrimSpace(*c.StrainOrVariety)
		}
		b.WriteString(line)
	}
	return b.String(), nil
}

func humanizeCostCategory(cat string) string {
	cat = strings.TrimSpace(cat)
	if cat == "" {
		return "Other"
	}
	return strings.ReplaceAll(cat, "_", " ")
}

func pluralS(n int64) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func cycleCostSummaryLine(ctx context.Context, q db.Querier, cycleID int64) string {
	cid := cycleID
	rows, err := q.GetCostTotalsByCropCycle(ctx, &cid)
	if err != nil || len(rows) == 0 {
		return ""
	}
	var total float64
	currency := ""
	for _, row := range rows {
		total += numericToFloat64(row.Expense)
		if currency == "" {
			currency = strings.TrimSpace(row.Currency)
		}
	}
	if total <= 0 {
		return ""
	}
	return fmt.Sprintf("Tagged spend so far: %.2f %s (use summarize_cycle_cost for category breakdown).", total, currency)
}
