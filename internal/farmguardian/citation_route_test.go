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
		{"unknown source type", 1, "symptom_guide", 5},
		{"nil queries for schedule", 1, "schedule", 10},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := ResolveCitationRoute(t.Context(), nil, tc.farmID, tc.sourceType, tc.sourceID); ok {
				t.Fatalf("expected no route for %q", tc.name)
			}
		})
	}
}

func TestLandingDocRoute(t *testing.T) {
	if route, ok := landingDocRoute("platform_doc", ""); !ok || route != "/operator-guide?tab=library&section=guide" {
		t.Fatalf("platform_doc landing = %q,%v", route, ok)
	}
	if route, ok := landingDocRoute("field_guide", "basil"); !ok || route != "/operator-guide?tab=symptoms&crop_key=basil" {
		t.Fatalf("field_guide crop landing = %q,%v", route, ok)
	}
}

func TestZonePath(t *testing.T) {
	got := zonePath(3, "ops", "alerts")
	if got != "/zones/3?ops=alerts&tab=ops" {
		t.Fatalf("zonePath = %q", got)
	}
	if got := zonePath(2, "light", ""); got != "/zones/2?tab=light" {
		t.Fatalf("zonePath light = %q", got)
	}
}
