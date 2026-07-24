package recipeoutcomes

// DominantThreshold is the minimum share of attribution hits for one recipe/revision pair.
const DominantThreshold = 0.6

type attributionKey struct {
	RecipeID   int64
	RevisionID int64 // 0 when unset
}

// Hit is one mixing or program-run row carrying recipe metadata.
type Hit struct {
	ApplicationRecipeID         int64
	ApplicationRecipeRevisionID *int64
}

func revisionIDVal(rev *int64) int64 {
	if rev == nil {
		return 0
	}
	return *rev
}

// AttributeCycle picks the dominant recipe/revision from attribution hits.
// mixed is true when no pair reaches DominantThreshold.
func AttributeCycle(hits []Hit) (key attributionKey, mixed bool, total int) {
	counts := map[attributionKey]int{}
	for _, h := range hits {
		if h.ApplicationRecipeID <= 0 {
			continue
		}
		k := attributionKey{
			RecipeID:   h.ApplicationRecipeID,
			RevisionID: revisionIDVal(h.ApplicationRecipeRevisionID),
		}
		counts[k]++
		total++
	}
	if total == 0 {
		return attributionKey{}, false, 0
	}
	best := attributionKey{}
	bestCount := 0
	for k, c := range counts {
		if c > bestCount {
			bestCount = c
			best = k
		}
	}
	if float64(bestCount)/float64(total) < DominantThreshold {
		return best, true, total
	}
	return best, false, total
}
