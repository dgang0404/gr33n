// Phase 152 WS2 — guard-clause coverage for ResolveCitationRoute that
// doesn't need a live DB (matches context_ref_test.go's own q==nil pattern
// for this package). DB-backed success paths live in
// citation_route_db_test.go.

package farmguardian

import "testing"

func TestResolveCitationRoute_guardClauses(t *testing.T) {
	tests := []struct {
		name       string
		farmID     int64
		sourceType string
		sourceID   int64
	}{
		{"nil queries always fails regardless of args", 1, "crop_cycle", 2},
		{"zero farm id", 0, "crop_cycle", 2},
		{"zero source id", 1, "crop_cycle", 0},
		{"unknown source type", 1, "schedule", 10},
		{"unmapped alert type (needs a follow-up hop)", 1, "alert_notification", 5},
		{"unmapped platform doc", 1, "platform_doc", 5},
		{"unmapped field guide", 1, "field_guide", 5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := ResolveCitationRoute(t.Context(), nil, tc.farmID, tc.sourceType, tc.sourceID); ok {
				t.Fatalf("expected no route for %q", tc.name)
			}
		})
	}
}
