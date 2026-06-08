package farmguardian

import "testing"

func TestShouldRunSummarizeCycleCostReadIntent(t *testing.T) {
	ref := &ContextRef{Type: "zone", ID: 3, CropCycleID: 12}
	cases := []struct {
		q    string
		ref  *ContextRef
		want bool
	}{
		{"What did Flower Room cost so far?", nil, true},
		{"cost per gram for this grow", ref, true},
		{"list my plants", nil, false},
		{"", ref, true},
	}
	for _, c := range cases {
		got := shouldRunSummarizeCycleCostReadIntent(c.q, c.ref)
		if got != c.want {
			t.Fatalf("shouldRunSummarizeCycleCostReadIntent(%q) = %v want %v", c.q, got, c.want)
		}
	}
}

func TestShouldRunSummarizeFarmSpendingReadIntent(t *testing.T) {
	cases := []struct {
		q    string
		want bool
	}{
		{"Summarize spending this month by category", true},
		{"What did I spend this month?", true},
		{"What did Flower Room cost?", false},
	}
	for _, c := range cases {
		got := shouldRunSummarizeFarmSpendingReadIntent(c.q)
		if got != c.want {
			t.Fatalf("shouldRunSummarizeFarmSpendingReadIntent(%q) = %v want %v", c.q, got, c.want)
		}
	}
}

func TestShouldRunRestockPriorityReadIntent(t *testing.T) {
	for _, q := range []string{
		"What should I restock first?",
		"restock priority on this farm",
	} {
		if !shouldRunRestockPriorityReadIntent(q) {
			t.Fatalf("expected restock priority intent for %q", q)
		}
	}
	if shouldRunRestockPriorityReadIntent("list my plants") {
		t.Fatal("plants should not match restock priority")
	}
}

func TestShouldRunSummarizeActiveGrowsReadIntent(t *testing.T) {
	for _, q := range []string{
		"What's growing where?",
		"active grows on this farm",
	} {
		if !shouldRunSummarizeActiveGrowsReadIntent(q) {
			t.Fatalf("expected active grows intent for %q", q)
		}
	}
	if shouldRunSummarizeActiveGrowsReadIntent("list my plants") {
		t.Fatal("plants list should not match active grows")
	}
}

func TestReadToolIDs_IncludesPhase55OpsTools(t *testing.T) {
	ids := ReadToolIDs()
	for _, want := range []string{
		"summarize_cycle_cost",
		"summarize_farm_spending",
		"restock_priority",
		"summarize_active_grows",
	} {
		found := false
		for _, id := range ids {
			if id == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("ReadToolIDs missing %q: %v", want, ids)
		}
	}
}

func TestHumanizeCostCategory(t *testing.T) {
	if got := humanizeCostCategory("fertilizers_soil_amendments"); got != "fertilizers soil amendments" {
		t.Fatalf("got %q", got)
	}
}
