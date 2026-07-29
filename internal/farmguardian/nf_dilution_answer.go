// Phase 211.06 WS4 — medium NF dilution grounding: force catalog ratios into the answer.

package farmguardian

import (
	"fmt"
	"strings"

	"gr33n-api/internal/naturalfarmingcatalog"
)

// EnsureNFDilutionRatiosInAnswer makes dilution/application answers quote catalog
// ratios when lookup_process_catalog already has them. Soft footers were ignored
// by phi3; this is the medium template step.
//
// ponytail: post-hoc rewrite from catalog YAML — ceiling is prose quality / nuance;
// upgrade = constrained decode, or skip LLM entirely for pure dilution intents.
func EnsureNFDilutionRatiosInAnswer(answer, question string) string {
	if !nfDilutionOrApplicationIntent(question) || !processMentionedInQuestion(question) {
		return answer
	}
	pt := processTypeFromQuestion(question)
	if pt == "" {
		return answer
	}
	_, canon, _, err := defaultNFCatalogs()
	if err != nil {
		return answer
	}
	recipes := naturalfarmingcatalog.CanonApplicationRecipesForProcess(canon, pt)
	if len(recipes) == 0 {
		// Fall back to dilution_start / dilution_strong on the input definition.
		if inp, ok := naturalfarmingcatalog.CanonInputByProcessType(canon, pt); ok {
			recipes = dilutionsAsPseudoRecipes(inp)
		}
	}
	if len(recipes) == 0 {
		return answer
	}
	lower := strings.ToLower(answer)
	present, missing := 0, 0
	var missingLines []string
	for _, rec := range recipes {
		dil, _ := rec["dilution"].(string)
		dil = strings.TrimSpace(dil)
		if dil == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(dil)) {
			present++
			continue
		}
		missing++
		name, _ := rec["seed_name"].(string)
		appType, _ := rec["target_application_type"].(string)
		line := "- " + strings.TrimSpace(name)
		if appType != "" {
			line += " (" + appType + ")"
		}
		line += ": " + dil
		missingLines = append(missingLines, line)
	}
	if missing == 0 {
		return answer
	}
	block := "Catalog application dilutions (from farm process catalog):\n" + strings.Join(missingLines, "\n")
	if present == 0 {
		// Model inventing with zero quoted ratios — prefer catalog skeleton over soup.
		label := strings.ToUpper(pt)
		return fmt.Sprintf(
			"%s application dilutions from the farm catalog:\n%s\n\nQuote these ratios; do not invent others.",
			label, strings.Join(missingLines, "\n"),
		)
	}
	ans := strings.TrimRight(answer, " \t\r\n")
	return ans + "\n\n" + block
}

func dilutionsAsPseudoRecipes(inp map[string]any) []map[string]any {
	var out []map[string]any
	if d, _ := inp["dilution_start"].(string); strings.TrimSpace(d) != "" {
		out = append(out, map[string]any{"seed_name": "start", "dilution": strings.TrimSpace(d)})
	}
	if d, _ := inp["dilution_strong"].(string); strings.TrimSpace(d) != "" {
		out = append(out, map[string]any{"seed_name": "stronger", "dilution": strings.TrimSpace(d)})
	}
	return out
}
