// Phase 210 WS1 — natural farming read tools (process catalog + farm inventory).

package farmguardian

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"gr33n-api/internal/croplibrary"
	"gr33n-api/internal/db"
	"gr33n-api/internal/naturalfarmingcatalog"
)

const ReadToolsMaxNFInventoryLines = 12

var (
	lookupProcessCatalogIntent = regexp.MustCompile(`(?i)\b(what is|how do i make|how to make|steps for|tell me about)\b.*\b(jms|jlf|fpj|ffj|lab|ohn|faa|jwa|js|jhs|wca|wcs|brv|compost tea|aact|microbial solution|liquid fertilizer|fermented plant|fermented fruit|lactic acid)\b|\b(jms|jlf|fpj|ffj|lab|ohn|faa|jwa|js|jhs|wca|wcs|brv)\b.*\b(what is|how|make|steps|prepare)\b`)
	suggestProcessMaterialIntent = regexp.MustCompile(`(?i)\b(goldenrod|comfrey|nettle|solidago|ferment|drench|foliar|plant juice|biomass)\b`)
	summarizeNFInventoryIntent    = regexp.MustCompile(`(?i)(\b(what|which)\b.{0,40}\b(ferments?|batches?|inputs?)\b.{0,40}\b(have|ready|on hand|stock)\b|\b(ready batches?|batches ready|ferments on hand|natural farming inventory)\b|\b(jms|jlf|fpj|ffj|wca|faa|lab|ohn|jwa)\b.{0,30}\b(have|ready|stock|on hand)\b)`)
)

var (
	nfCatalogOnce  sync.Once
	nfMaterialCat  map[string]any
	nfRecipeCanon  map[string]any
	nfCatalogErr   error
	nfCatalogRoot  string
)

func defaultNFCatalogs() (material map[string]any, canon map[string]any, root string, err error) {
	nfCatalogOnce.Do(func() {
		nfCatalogRoot, nfCatalogErr = croplibrary.FindRepoRoot()
		if nfCatalogErr != nil {
			return
		}
		nfMaterialCat, nfCatalogErr = naturalfarmingcatalog.LoadMaterialCatalog(nfCatalogRoot)
		if nfCatalogErr != nil {
			return
		}
		nfRecipeCanon, nfCatalogErr = naturalfarmingcatalog.LoadRecipeCanon(nfCatalogRoot)
	})
	return nfMaterialCat, nfRecipeCanon, nfCatalogRoot, nfCatalogErr
}

// LookupProcessCatalog runs lookup_process_catalog (exported for tests).
func LookupProcessCatalog(question string) (string, error) {
	return renderLookupProcessCatalog(question)
}

// SuggestProcessFromMaterial runs suggest_process_from_material (exported for tests).
func SuggestProcessFromMaterial(ctx context.Context, q db.Querier, farmID int64, question string) (string, error) {
	return renderSuggestProcessFromMaterial(ctx, q, farmID, question)
}

func shouldRunLookupProcessCatalogReadIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if shouldRunSuggestProcessFromMaterialReadIntent(question) {
		return false
	}
	if lookupCropTargetsIntent.MatchString(q) && !processMentionedInQuestion(q) {
		return false
	}
	return lookupProcessCatalogIntent.MatchString(q) || (processMentionedInQuestion(q) && strings.Contains(strings.ToLower(q), "how"))
}

func shouldRunSuggestProcessFromMaterialReadIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if !suggestProcessMaterialIntent.MatchString(q) {
		return false
	}
	mat, _, _, err := defaultNFCatalogs()
	if err != nil {
		return suggestProcessMaterialIntent.MatchString(q)
	}
	if len(naturalfarmingcatalog.MaterialsMatchingQuery(mat, q)) > 0 {
		return true
	}
	return suggestProcessMaterialIntent.MatchString(q)
}

func shouldRunSummarizeNaturalFarmingInventoryReadIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if shouldRunSummarizeFarmLowStockReadIntent(q) {
		return false
	}
	return summarizeNFInventoryIntent.MatchString(q)
}

func processMentionedInQuestion(q string) bool {
	lower := strings.ToLower(q)
	for _, tok := range []string{"jms", "jlf", "fpj", "ffj", "lab", "ohn", "faa", "jwa", "js", "jhs", "wca", "wcs", "brv", "compost tea", "aact"} {
		if strings.Contains(lower, tok) {
			return true
		}
	}
	return false
}

func processTypeFromQuestion(q string) string {
	lower := strings.ToLower(q)
	checks := []struct{ phrase, typ string }{
		{"microbial solution", "jms"},
		{"jadam liquid fertilizer", "jlf"},
		{"fermented plant juice", "fpj"},
		{"fermented fruit juice", "ffj"},
		{"lactic acid bacteria", "lab"},
		{"oriental herbal nutrient", "ohn"},
		{"fish amino acid", "faa"},
		{"compost tea", "compost_tea_aact"},
		{"actively aerated", "compost_tea_aact"},
		{"jms", "jms"},
		{"jlf", "jlf"},
		{"fpj", "fpj"},
		{"ffj", "ffj"},
		{"lab", "lab"},
		{"ohn", "ohn"},
		{"faa", "faa"},
		{"jwa", "jwa"},
		{"jhs", "jhs"},
		{"js", "js"},
		{"wca", "wca"},
		{"wcs", "wcs"},
		{"brv", "brv"},
	}
	for _, c := range checks {
		if strings.Contains(lower, c.phrase) {
			return c.typ
		}
	}
	return ""
}

func renderLookupProcessCatalog(question string) (string, error) {
	_, canon, root, err := defaultNFCatalogs()
	if err != nil {
		return "", err
	}
	pt := processTypeFromQuestion(question)
	if pt == "" {
		return "lookup_process_catalog: name a process (JMS, JLF, FPJ, FFJ, LAB, OHN, FAA, …) to load steps and dilution from the process catalog.", nil
	}
	inp, ok := naturalfarmingcatalog.CanonInputByProcessType(canon, pt)
	if !ok {
		return fmt.Sprintf("lookup_process_catalog: no catalog entry for process_type %q.", pt), nil
	}
	seedName, _ := inp["seed_name"].(string)
	guide, _ := inp["guide"].(string)
	tradition, _ := inp["tradition"].(string)
	tier, _ := inp["source_tier"].(string)

	var b strings.Builder
	b.WriteString("lookup_process_catalog — ")
	b.WriteString(strings.TrimSpace(seedName))
	if pt != "" {
		b.WriteString(fmt.Sprintf(" (process_type: %s)", pt))
	}
	if tradition != "" {
		b.WriteString("\nTradition: " + tradition)
	}
	if tier != "" {
		b.WriteString("\nSource tier: " + tier)
	}
	if guide != "" {
		b.WriteString("\nField guide: field-guides/" + guide)
		if excerpt := nfGuideExcerpt(root, guide); excerpt != "" {
			b.WriteString("\nSteps (excerpt): " + excerpt)
		}
	}
	return b.String(), nil
}

func renderSuggestProcessFromMaterial(ctx context.Context, q db.Querier, farmID int64, question string) (string, error) {
	mat, canon, root, err := defaultNFCatalogs()
	if err != nil {
		return "", err
	}
	materials := naturalfarmingcatalog.MaterialsMatchingQuery(mat, question)
	materials = append(materials, materialsFromPlants(ctx, q, farmID, mat)...)

	seen := map[string]bool{}
	var uniq []map[string]any
	for _, m := range materials {
		id, _ := m["id"].(string)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		uniq = append(uniq, m)
	}
	if len(uniq) == 0 {
		return "suggest_process_from_material: no catalog material matched — try goldenrod, nettle, or comfrey, or name the biomass.", nil
	}

	var b strings.Builder
	b.WriteString("suggest_process_from_material — matched materials from Phase 208 catalog:")
	for _, m := range uniq {
		id, _ := m["id"].(string)
		tier, _ := m["source_tier"].(string)
		b.WriteString("\n- " + id)
		if tier != "" {
			b.WriteString(" (source_tier: " + tier + ")")
		}
		procs, _ := m["processes"].([]any)
		for _, p := range procs {
			proc, ok := p.(map[string]any)
			if !ok {
				continue
			}
			pt, _ := proc["type"].(string)
			guide, _ := proc["guide"].(string)
			dStart, _ := proc["dilution_start"].(string)
			dStrong, _ := proc["dilution_strong"].(string)
			line := fmt.Sprintf("\n  • process: %s", pt)
			if guide != "" {
				line += "; guide: field-guides/" + guide
			}
			if dStart != "" {
				line += "; dilution: " + dStart
				if dStrong != "" && dStrong != dStart {
					line += " / stronger " + dStrong
				}
			}
			b.WriteString(line)
			if pt != "" {
				if inp, ok := naturalfarmingcatalog.CanonInputByProcessType(canon, pt); ok {
					if seed, _ := inp["seed_name"].(string); seed != "" {
						b.WriteString("\n    linked input: " + seed)
					}
				}
			}
			if guide != "" {
				if excerpt := nfGuideExcerpt(root, guide); excerpt != "" {
					b.WriteString("\n    excerpt: " + excerpt)
				}
			}
		}
	}
	return b.String(), nil
}

func materialsFromPlants(ctx context.Context, q db.Querier, farmID int64, catalog map[string]any) []map[string]any {
	if q == nil || farmID <= 0 {
		return nil
	}
	plants, err := q.ListPlantsByFarm(ctx, farmID)
	if err != nil {
		return nil
	}
	var out []map[string]any
	for _, p := range plants {
		label := strings.TrimSpace(p.DisplayName)
		if label == "" {
			continue
		}
		out = append(out, naturalfarmingcatalog.MaterialsMatchingQuery(catalog, label)...)
	}
	return out
}

func renderSummarizeNaturalFarmingInventory(ctx context.Context, q db.Querier, farmID int64) (string, error) {
	farmLabel := fmt.Sprintf("farm #%d", farmID)
	if farm, err := q.GetFarmByID(ctx, farmID); err == nil {
		if name := strings.TrimSpace(farm.Name); name != "" {
			farmLabel = name
		}
	}

	inputs, err := q.ListInputDefinitionsByFarm(ctx, farmID)
	if err != nil {
		return "", err
	}
	inputName := map[int64]string{}
	for _, in := range inputs {
		inputName[in.ID] = strings.TrimSpace(in.Name)
	}

	batches, err := q.ListInputBatchesByFarm(ctx, farmID)
	if err != nil {
		return "", err
	}

	type bucket struct {
		status string
		lines  []string
	}
	byStatus := map[string]*bucket{}
	addLine := func(status, line string) {
		if byStatus[status] == nil {
			byStatus[status] = &bucket{status: status}
		}
		byStatus[status].lines = append(byStatus[status].lines, line)
	}

	for _, b := range batches {
		st := string(b.Status)
		name := inputName[b.InputDefinitionID]
		if name == "" {
			name = fmt.Sprintf("input #%d", b.InputDefinitionID)
		}
		label := name
		if b.BatchIdentifier != nil && strings.TrimSpace(*b.BatchIdentifier) != "" {
			label = fmt.Sprintf("%s (%s)", name, strings.TrimSpace(*b.BatchIdentifier))
		}
		qty := formatBatchQuantity(b.CurrentQuantityRemaining)
		line := fmt.Sprintf("%s — %s remaining; batch #%d", label, qty, b.ID)
		addLine(st, line)
	}

	var b strings.Builder
	b.WriteString("summarize_natural_farming_inventory — " + farmLabel)
	if len(batches) == 0 {
		b.WriteString("\nNo input batches on this farm yet.")
		return b.String(), nil
	}

	order := []string{"ready_for_use", "partially_used", "fermenting_brewing", "maturing_aging", "planning", "ingredients_gathered", "mixing_in_progress", "fully_used", "expired_discarded"}
	listed := 0
	for _, st := range order {
		bkt := byStatus[st]
		if bkt == nil || len(bkt.lines) == 0 {
			continue
		}
		sort.Strings(bkt.lines)
		b.WriteString("\n" + strings.ReplaceAll(st, "_", " ") + ":")
		for _, line := range bkt.lines {
			if listed >= ReadToolsMaxNFInventoryLines {
				b.WriteString(fmt.Sprintf("\n(+ %d more batches not listed)", len(batches)-listed))
				goto done
			}
			b.WriteString("\n- " + line)
			listed++
		}
	}
done:
	lowRows, err := q.ListLowStockBatchesByFarm(ctx, farmID)
	if err == nil && len(lowRows) > 0 {
		b.WriteString(fmt.Sprintf("\nLow stock (ready batches below threshold): %d — use summarize_farm_low_stock for restock detail.", len(lowRows)))
	}
	return b.String(), nil
}

func nfGuideExcerpt(repoRoot, guideFile string) string {
	guideFile = strings.TrimSpace(guideFile)
	if guideFile == "" {
		return ""
	}
	path := guideFile
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, "docs", "field-guides", guideFile)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	text := string(data)
	if idx := strings.Index(text, "---"); idx >= 0 {
		if idx2 := strings.Index(text[idx+3:], "---"); idx2 >= 0 {
			text = text[idx+3+idx2+3:]
		}
	}
	start := strings.Index(text, "## Step-by-step preparation")
	if start < 0 {
		start = strings.Index(text, "## Step-by-step")
	}
	if start < 0 {
		return truncateRunes(strings.TrimSpace(text), 280)
	}
	chunk := text[start:]
	if nl := strings.Index(chunk, "\n## "); nl > 0 {
		chunk = chunk[:nl]
	}
	chunk = strings.ReplaceAll(chunk, "## Step-by-step preparation", "")
	chunk = strings.ReplaceAll(chunk, "## Step-by-step", "")
	chunk = strings.TrimSpace(chunk)
	return truncateRunes(chunk, 320)
}
