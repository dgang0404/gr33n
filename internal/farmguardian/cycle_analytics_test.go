// Phase 28 WS3 — unit coverage for the CycleAnalytics renderer + the
// snapshot integration. SQL-touching paths live in
// cmd/api/smoke_phase28_ws3_test.go so the unit tests can stay fast and
// pure (no Postgres, no docker).

package farmguardian

import (
	"strings"
	"testing"
)

func TestCycleAnalytics_EmptyRendersEmptyLine(t *testing.T) {
	if got := (CycleAnalytics{}).renderLine(); got != "" {
		t.Fatalf("empty analytics should render empty, got %q", got)
	}
}

func TestCycleAnalytics_Empty_Helper(t *testing.T) {
	if !(CycleAnalytics{}).Empty() {
		t.Fatal("zero CycleAnalytics must report Empty")
	}
	if (CycleAnalytics{EventCount: 1}).Empty() {
		t.Fatal("analytics with events is not Empty")
	}
	if (CycleAnalytics{YieldGrams: 100}).Empty() {
		t.Fatal("analytics with yield is not Empty")
	}
	if (CycleAnalytics{TotalExpenses: 12.5, Currency: "USD"}).Empty() {
		t.Fatal("analytics with cost is not Empty")
	}
}

func TestCycleAnalytics_RenderFullLine(t *testing.T) {
	lpd := 14.5
	gpd := 6.06
	cpg := 0.76
	a := CycleAnalytics{
		DurationDays:  68,
		EventCount:    142,
		TotalLiters:   980,
		LitersPerDay:  &lpd,
		AvgECmSCm:     1.62,
		MinECmSCm:     1.12,
		MaxECmSCm:     2.05,
		AvgPH:         6.10,
		YieldGrams:    412,
		GramsPerDay:   &gpd,
		TotalExpenses: 312.40,
		Currency:      "USD",
		CostPerGram:   &cpg,
	}
	got := a.renderLine()
	for _, want := range []string{
		"feed: 142 events / 980L (14.5L/d)",
		"EC 1.62 (1.12–2.05)",
		"pH 6.10",
		"cost: 312.40 USD",
		"yield: 412g (6.06g/d)",
		"cost/g: 0.76 USD",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in:\n%s", want, got)
		}
	}
}

func TestCycleAnalytics_OmitsAbsentSections(t *testing.T) {
	// Cycle has only fertigation data: no yield, no costs.
	a := CycleAnalytics{
		EventCount:  10,
		TotalLiters: 50,
		AvgECmSCm:   1.4,
		AvgPH:       6.05,
	}
	got := a.renderLine()
	for _, missing := range []string{"yield:", "cost:", "cost/g:"} {
		if strings.Contains(got, missing) {
			t.Fatalf("did not expect %q in:\n%s", missing, got)
		}
	}
	for _, want := range []string{"feed:", "EC 1.40", "pH 6.05"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in:\n%s", want, got)
		}
	}
}

func TestCycleAnalytics_NoCostPerGramWhenMixedCurrency(t *testing.T) {
	// Empty currency means the helper detected multi-currency and
	// declined to commit to a TotalExpenses figure. CostPerGram must
	// also be nil and the renderer must not emit either section.
	a := CycleAnalytics{
		EventCount:  1,
		TotalLiters: 10,
		YieldGrams:  100,
		// Currency intentionally left blank.
	}
	got := a.renderLine()
	if strings.Contains(got, "cost:") || strings.Contains(got, "cost/g:") {
		t.Fatalf("multi-currency cycles should not emit cost lines, got:\n%s", got)
	}
}

func TestSnapshot_RendersAnalyticsAttachedToCycle(t *testing.T) {
	lpd := 14.5
	s := Snapshot{
		ActiveCycles: []ActiveCycle{
			{
				Name:     "FlowerRun3",
				ZoneName: "B",
				BatchLabel: "OG",
				Stage:    "late_flower",
				Analytics: CycleAnalytics{
					EventCount:   142,
					TotalLiters:  980,
					LitersPerDay: &lpd,
					AvgECmSCm:    1.62,
					MinECmSCm:    1.12,
					MaxECmSCm:    2.05,
				},
			},
			{Name: "BasilWinter", ZoneName: "A"}, // no analytics — older cycle
		},
	}
	got := s.Render()
	for _, want := range []string{
		"FlowerRun3 — zone B (OG; stage: late_flower)",
		"metrics: feed: 142 events / 980L (14.5L/d); EC 1.62 (1.12–2.05)",
		"BasilWinter — zone A",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in:\n%s", want, got)
		}
	}
	// BasilWinter must NOT have a metrics line beneath it.
	idx := strings.Index(got, "BasilWinter")
	if idx >= 0 && strings.Contains(got[idx:], "metrics:") {
		t.Fatalf("BasilWinter must not have a metrics line:\n%s", got[idx:])
	}
}

func TestFormatHelpers_RoundingAndUnits(t *testing.T) {
	cases := []struct {
		got, want string
	}{
		{formatLiters(980), "980L"},
		{formatLiters(980.45), "980.5L"},
		{formatLiters(980.04), "980L"},
		{formatLitersPerDay(14.499), "14.5L/d"},
		{formatEC(1.6249), "1.62"},
		{formatPH(6.104), "6.10"},
		{formatPH(6.106), "6.11"}, // round-half-up via math.Round
		{formatMoney(312.404), "312.40"},
		{formatGrams(411.4), "411"},
		{formatGrams(411.6), "412"},
		{formatGramsPerDay(6.059), "6.06g/d"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q want %q", c.got, c.want)
		}
	}
}
