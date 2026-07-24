// Phase 211.05 WS4 — Guardian read tool for recipe historical outcomes.

package farmguardian

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gr33n-api/internal/authctx"
	"gr33n-api/internal/cropcycle/recipeoutcomes"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
)

// RecipeOutcomeGroundingRule is injected into every grounded chat system prompt.
const RecipeOutcomeGroundingRule = `Recipe outcome grounding (Phase 211.05): summarize_recipe_outcomes numbers are historical averages over named past cycles, not predictions. Always state N (cycle count). Never say a recipe "is better" or "will produce X" — say cycles "averaged X". Below minimum sample size, say insufficient history instead of citing a single-cycle number as a trend. Correlation is not causation — stage timing, zone, and season differ between cycles.`

var summarizeRecipeOutcomesIntent = regexp.MustCompile(`(?i)\b(which recipe|recipe worked|switching recipes|recipe help|based on history|historical|track record|cost per gram|yield by recipe|did .{0,40} recipe|compare recipes|predict my yield|forecast yield|recipe outcome|recipe performance)\b`)

func shouldRunSummarizeRecipeOutcomesReadIntent(question string, ref *ContextRef) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if summarizeRecipeOutcomesIntent.MatchString(q) {
		return true
	}
	if ref != nil && ref.CropCycleID > 0 {
		return regexp.MustCompile(`(?i)\b(this recipe|formula|revision|compared to|vs last|average|history)\b`).MatchString(q)
	}
	return false
}

func renderSummarizeRecipeOutcomes(ctx context.Context, q db.Querier, farmID int64, question string, ref *ContextRef) (string, error) {
	opt := recipeoutcomes.Options{IncludeCosts: true}
	if uid, ok := authctx.UserID(ctx); ok {
		has, err := farmauthz.HasFarmScope(ctx, q, uid, farmID, farmauthz.ScopeMoneyCostsRead)
		if err != nil {
			return "", err
		}
		opt.IncludeCosts = has
	} else if !authctx.FarmAuthzSkip(ctx) {
		opt.IncludeCosts = false
	}

	if keys := cropKeysMentionedInQuestion(question); len(keys) == 1 {
		ck := keys[0]
		opt.CropKey = &ck
	}

	result, err := recipeoutcomes.Build(ctx, q, farmID, opt)
	if err != nil {
		return "", err
	}

	farmLabel := fmt.Sprintf("farm #%d", farmID)
	if farm, err := q.GetFarmByID(ctx, farmID); err == nil {
		if name := strings.TrimSpace(farm.Name); name != "" {
			farmLabel = name
		}
	}

	var b strings.Builder
	b.WriteString("summarize_recipe_outcomes — " + farmLabel)
	if len(result.Outcomes) == 0 && len(result.InsufficientHistory) == 0 {
		b.WriteString("\nNo harvested cycles with clear recipe attribution yet — need finished grows with yield, catalog plant, and mixing/program runs stamped with application_recipe_id.")
		if result.MixedCycleCount > 0 {
			b.WriteString(fmt.Sprintf("\n%d cycle(s) used mixed/unclear recipes — excluded from recipe-specific averages.", result.MixedCycleCount))
		}
		if result.UnattributedCycleCount > 0 {
			b.WriteString(fmt.Sprintf("\n%d cycle(s) had no recipe-tagged mix/program events in-window.", result.UnattributedCycleCount))
		}
		b.WriteString("\nCorrelational only — not a prediction.")
		return b.String(), nil
	}

	outcomes := append([]recipeoutcomes.RecipeOutcome{}, result.Outcomes...)
	sort.Slice(outcomes, func(i, j int) bool {
		if outcomes[i].CropKey != outcomes[j].CropKey {
			return outcomes[i].CropKey < outcomes[j].CropKey
		}
		return outcomes[i].RecipeName < outcomes[j].RecipeName
	})

	for _, row := range outcomes {
		b.WriteString("\n")
		b.WriteString(formatRecipeOutcomeLine(row, opt.IncludeCosts))
	}

	for _, row := range result.InsufficientHistory {
		b.WriteString(fmt.Sprintf("\n%s rev %s: only %d harvested cycle — insufficient history for an average (need %d+).",
			row.RecipeName,
			revisionLabel(row.ApplicationRecipeRevisionID),
			row.CycleCount,
			result.MinSampleSize,
		))
	}
	if result.MixedCycleCount > 0 {
		b.WriteString(fmt.Sprintf("\n%d cycle(s) used mixed/unclear recipes — excluded from recipe-specific averages.", result.MixedCycleCount))
	}
	if result.UnattributedCycleCount > 0 {
		b.WriteString(fmt.Sprintf("\n%d cycle(s) had no recipe-tagged mix/program events in-window.", result.UnattributedCycleCount))
	}
	b.WriteString("\nCorrelational only — stage timing, zone, and season differ between cycles; not a controlled comparison or forecast.")
	return b.String(), nil
}

func formatRecipeOutcomeLine(row recipeoutcomes.RecipeOutcome, includeCosts bool) string {
	name := strings.TrimSpace(row.RecipeName)
	if name == "" {
		name = fmt.Sprintf("recipe #%d", row.ApplicationRecipeID)
	}
	line := fmt.Sprintf("%s (%s) rev %s: %d harvested cycles",
		name,
		row.CropKey,
		revisionLabel(row.ApplicationRecipeRevisionID),
		row.CycleCount,
	)
	if row.AvgYieldGrams != nil {
		line += fmt.Sprintf(" — avg yield %.0fg", *row.AvgYieldGrams)
		if row.MinYieldGrams != nil && row.MaxYieldGrams != nil {
			line += fmt.Sprintf(" (range %.0f–%.0fg)", *row.MinYieldGrams, *row.MaxYieldGrams)
		}
	}
	if includeCosts && row.AvgCostPerGram != nil && row.CostCurrency != "" {
		line += fmt.Sprintf(", avg %.2f %s/g", *row.AvgCostPerGram, row.CostCurrency)
	}
	if row.AvgDurationDays != nil {
		line += fmt.Sprintf(", avg %.0f days", *row.AvgDurationDays)
	}
	return line
}

func revisionLabel(rev *int64) string {
	if rev == nil {
		return "?"
	}
	return fmt.Sprintf("#%d", *rev)
}
