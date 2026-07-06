package eval

import "testing"

func TestPhase127Fixtures_countAndPrompts(t *testing.T) {
	fixtures := Phase127Fixtures()
	if len(fixtures) != 4 {
		t.Fatalf("expected 4 phase127 fixtures, got %d", len(fixtures))
	}
	if fixtures[0].ID != "p128-devices" || fixtures[0].ExpectTool != "summarize_device_health" {
		t.Fatalf("devices fixture: %+v", fixtures[0])
	}
	if fixtures[3].ID != "p128-fert-triage" || !fixtures[3].ExpectCitation {
		t.Fatalf("triage fixture: %+v", fixtures[3])
	}
}

func TestFixturesForSuite_phase127(t *testing.T) {
	for _, suite := range []string{"phase127", "phase128", "p128"} {
		if len(FixturesForSuite(suite)) != 4 {
			t.Fatalf("suite %q", suite)
		}
	}
}

func TestScore_phase127Devices(t *testing.T) {
	pass := Score(ScoreInput{
		Question: Question{ID: "p128-devices"},
		Answer:   "One edge device on the demo farm shows offline in the snapshot — check the Pi client heartbeat.",
	})
	if !pass.Passed {
		t.Fatalf("expected pass: %s", pass.Notes)
	}
}

func TestScore_phase127FertManual(t *testing.T) {
	pass := Score(ScoreInput{
		Question: Question{ID: "p128-fert-manual"},
		Answer:   "Outdoor JLF Soil Drench is manual-only with no cron schedule on this farm.",
	})
	if !pass.Passed {
		t.Fatalf("expected pass: %s", pass.Notes)
	}
}

func TestScore_phase127RejectsInvention(t *testing.T) {
	fail := Score(ScoreInput{
		Question: Question{ID: "p128-devices"},
		Answer:   "The Secret Mars Dome GPIO bank is offline.",
	})
	if fail.Passed {
		t.Fatal("expected fail on invented zone")
	}
}
