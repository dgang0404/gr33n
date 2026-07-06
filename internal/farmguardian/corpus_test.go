package farmguardian

import (
	"testing"
	"time"
)

func TestTierFreshness(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-2 * time.Hour)
	aging := now.Add(-3 * 24 * time.Hour)
	stale := now.Add(-10 * 24 * time.Hour)

	cases := []struct {
		count int64
		at    *time.Time
		want  string
	}{
		{0, nil, FreshnessEmpty},
		{5, nil, FreshnessStale},
		{5, &fresh, FreshnessFresh},
		{5, &aging, FreshnessAging},
		{5, &stale, FreshnessStale},
	}
	for _, tc := range cases {
		if got := TierFreshness(tc.count, tc.at, now); got != tc.want {
			t.Fatalf("TierFreshness(%d,%v)=%q want %q", tc.count, tc.at, got, tc.want)
		}
	}
}

func TestBuildCorpusHealth_StalenessFlags(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	staleOp := now.Add(-14 * 24 * time.Hour)

	empty := BuildCorpusHealth(CorpusStatsInput{}, now)
	if empty.Staleness != StalenessFieldGuideEmpty {
		t.Fatalf("empty corpus staleness=%q", empty.Staleness)
	}

	opStale := BuildCorpusHealth(CorpusStatsInput{
		FieldGuideChunks:          10,
		FieldGuideLastIngestedAt:  ptrTime(now.Add(-time.Hour)),
		OperationalChunks:         40,
		OperationalLastIngestedAt: &staleOp,
	}, now)
	if opStale.Staleness != StalenessOperationalStale {
		t.Fatalf("operational stale=%q", opStale.Staleness)
	}
}

func TestCorpusWarningMessages_FarmCounsel(t *testing.T) {
	msgs := CorpusWarningMessages(CorpusHealth{
		FieldGuideChunks: 0,
		Staleness:        StalenessFieldGuideEmpty,
	}, WarmupModeFarmCounsel)
	if len(msgs) == 0 {
		t.Fatal("expected warnings for empty corpus")
	}
	if len(CorpusWarningMessages(CorpusHealth{}, WarmupModeQuick)) != 0 {
		t.Fatal("quick mode should not warn")
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
