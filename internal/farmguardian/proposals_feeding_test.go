package farmguardian

import (
	"context"
	"testing"
)

func TestMatchFeedingProgramIntent_Volume(t *testing.T) {
	ctx := context.Background()
	snap := Snapshot{
		ZoneNames: []string{"Flower Room"},
	}
	tool, args, summary, ok := matchFeedingProgramIntent(ctx, nil, 1, "Set feed volume to 0.3 L for Flower Room", snap)
	if ok {
		t.Fatalf("expected no match without querier, got %s", tool)
	}
	_ = args
	_ = summary
}

func TestMatchFeedingProgramIntent_IrrigationOnlyPhrase(t *testing.T) {
	q := "Switch Flower Room to plain water-only irrigation"
	if !irrigationOnlyIntent.MatchString(q) {
		t.Fatal("irrigation only intent should match")
	}
}

func TestMatchFeedingProgramIntent_VolumeRegex(t *testing.T) {
	m := feedVolumeIntent.FindStringSubmatch("Set feeding volume to 0.3 L for Flower Room")
	if len(m) < 2 || m[1] != "0.3" {
		t.Fatalf("volume regex: %#v", m)
	}
}

func TestMatchSummarizeZoneFertigationIntent_Phase47Phrases(t *testing.T) {
	for _, q := range []string{
		"When is the next feed for Flower Room?",
		"Is it safe to run feed now?",
		"Switch to water-only irrigation",
		"Does the reservoir need top-up?",
	} {
		if !matchSummarizeZoneFertigationIntent(q) {
			t.Fatalf("expected fertigation read intent for %q", q)
		}
	}
}
