package farmguardian

import (
	"fmt"
	"sort"
	"testing"

	db "gr33n-api/internal/db"
)

// Audit every distinct RAG source on farm 1 — coverage table in -v output.
// Run: go test ./internal/farmguardian/ -run TestCitationRouteAudit -v
func TestCitationRouteAudit(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := t.Context()
	q := db.New(pool)

	rows, err := pool.Query(ctx, `
		SELECT DISTINCT source_type, source_id
		FROM gr33ncore.rag_embedding_chunks
		WHERE farm_id = 1
		ORDER BY source_type, source_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	type pair struct {
		t  string
		id int64
	}
	var pairs []pair
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.t, &p.id); err != nil {
			t.Fatal(err)
		}
		pairs = append(pairs, p)
	}

	okBy, failBy := map[string]int{}, map[string]int{}
	var sample []string
	for _, p := range pairs {
		if _, ok := ResolveCitationRoute(ctx, q, 1, p.t, p.id); ok {
			okBy[p.t]++
		} else {
			failBy[p.t]++
			if len(sample) < 40 {
				sample = append(sample, fmt.Sprintf("%s #%d", p.t, p.id))
			}
		}
	}

	seen := map[string]bool{}
	var types []string
	for _, p := range pairs {
		if !seen[p.t] {
			seen[p.t] = true
			types = append(types, p.t)
		}
	}
	sort.Strings(types)

	t.Logf("Audited %d distinct pairs", len(pairs))
	var failedTypes []string
	for _, typ := range types {
		t.Logf("%-22s total=%d ok=%d fail=%d", typ, okBy[typ]+failBy[typ], okBy[typ], failBy[typ])
		if failBy[typ] > 0 {
			failedTypes = append(failedTypes, typ)
		}
	}
	for _, s := range sample {
		t.Logf("fail sample: %s", s)
	}
	if len(failedTypes) > 0 {
		t.Fatalf("citation routes unresolved for: %v (see fail samples)", failedTypes)
	}
}
