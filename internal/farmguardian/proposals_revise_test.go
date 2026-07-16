package farmguardian

import (
	"strings"
	"testing"
	"time"
)

func setupPackArgsFixture() map[string]any {
	return map[string]any{
		"profile":   "house_plant",
		"zone_id":   int64(12),
		"zone_name": "Tent A",
		"plant":     map[string]any{"crop_key": "basil"},
		"cycle":     map[string]any{"name": "Philodendron — Tent A", "current_stage": "early_veg"},
		"program": map[string]any{
			"name":                "Philodendron light feed",
			"total_volume_liters": 0.5,
			"ec_trigger_low":      0.8,
			"ph_trigger_low":      5.8,
			"ph_trigger_high":     6.5,
		},
	}
}

func TestApplyRevisionDeltas_SetupPackVolume(t *testing.T) {
	prior := setupPackArgsFixture()
	next, changed := applyRevisionDeltas("apply_grow_setup_pack", prior, "no, use 0.3 L not 0.5")
	if !changed {
		t.Fatal("expected changed=true for volume correction")
	}
	prog := next["program"].(map[string]any)
	if prog["total_volume_liters"].(float64) != 0.3 {
		t.Fatalf("volume = %#v want 0.3", prog["total_volume_liters"])
	}
	// Prior must be untouched (frozen args rebuilt, not mutated in place).
	if setupPackArgsFixture()["program"].(map[string]any)["total_volume_liters"].(float64) != 0.5 {
		t.Fatal("fixture sanity")
	}
	if prior["program"].(map[string]any)["total_volume_liters"].(float64) != 0.5 {
		t.Fatal("prior args were mutated by applyRevisionDeltas")
	}
}

func TestApplyRevisionDeltas_SetupPackStageAndPH(t *testing.T) {
	next, changed := applyRevisionDeltas("apply_grow_setup_pack", setupPackArgsFixture(),
		"make the cycle flower and use pH 5.5-6.2")
	if !changed {
		t.Fatal("expected changed=true")
	}
	cycle := next["cycle"].(map[string]any)
	if cycle["current_stage"] != "early_flower" {
		t.Fatalf("stage = %v want early_flower", cycle["current_stage"])
	}
	prog := next["program"].(map[string]any)
	if prog["ph_trigger_low"].(float64) != 5.5 || prog["ph_trigger_high"].(float64) != 6.2 {
		t.Fatalf("pH = %v/%v want 5.5/6.2", prog["ph_trigger_low"], prog["ph_trigger_high"])
	}
}

func TestApplyRevisionDeltas_NoActionableChange(t *testing.T) {
	_, changed := applyRevisionDeltas("apply_grow_setup_pack", setupPackArgsFixture(),
		"what does this do again?")
	if changed {
		t.Fatal("expected changed=false for a non-correction turn")
	}
}

func TestApplyRevisionDeltas_CreateTaskTitleCallIt(t *testing.T) {
	prior := map[string]any{"title": "Check humidity in grow room", "zone_id": float64(3)}
	next, changed := applyRevisionDeltas("create_task", prior, "call it Inspect tent RH instead")
	if !changed {
		t.Fatal("expected changed=true")
	}
	if next["title"] != "Inspect tent RH" {
		t.Fatalf("title = %#v want Inspect tent RH", next["title"])
	}
}

func TestApplyRevisionDeltas_CreateTaskInsteadOf(t *testing.T) {
	prior := map[string]any{"title": "Check humidity in grow room"}
	next, changed := applyRevisionDeltas("create_task", prior, "Inspect tent RH instead of Check humidity in grow room")
	if !changed {
		t.Fatal("expected changed=true")
	}
	if next["title"] != "Inspect tent RH" {
		t.Fatalf("title = %#v", next["title"])
	}
}

func TestApplyRevisionDeltas_CreateTaskDescription(t *testing.T) {
	prior := map[string]any{"title": "Refill OHN", "description": "Low stock alert"}
	next, changed := applyRevisionDeltas("create_task_from_alert", prior, "description should be Restock OHN from Supplies")
	if !changed {
		t.Fatal("expected changed=true")
	}
	if next["description"] != "Restock OHN from Supplies" {
		t.Fatalf("description = %#v", next["description"])
	}
}

// Phase 191 — live turn: "Please revise this change request — Create task:
// Follow up from Guardian chat. Correction: Should this task mention
// checking stock in Veg Tent?" This question-phrased correction previously
// matched no revise pattern, so the turn fell through to open-ended chat and
// the pending create_task proposal was never actually revised.
func TestApplyRevisionDeltas_CreateTaskDescriptionAppend_questionPhrased(t *testing.T) {
	prior := map[string]any{"title": "Follow up from Guardian chat"}
	next, changed := applyRevisionDeltas("create_task", prior, "Should this task mention checking stock in Veg Tent?")
	if !changed {
		t.Fatal("expected changed=true")
	}
	if next["description"] != "Checking stock in Veg Tent." {
		t.Fatalf("description = %#v", next["description"])
	}
}

func TestApplyRevisionDeltas_CreateTaskDescriptionAppend_appendsToExisting(t *testing.T) {
	prior := map[string]any{"title": "Refill calcium nitrate", "description": "Refill when stock is low."}
	next, changed := applyRevisionDeltas("create_task", prior, "Should it also mention checking stock in Veg Tent?")
	if !changed {
		t.Fatal("expected changed=true")
	}
	want := "Refill when stock is low. Also checking stock in Veg Tent."
	if next["description"] != want {
		t.Fatalf("description = %#v, want %q", next["description"], want)
	}
}

func TestApplyRevisionDeltas_CreateTaskDescriptionAppend_explicitReplaceStillWins(t *testing.T) {
	prior := map[string]any{"title": "Refill OHN", "description": "Old description"}
	next, changed := applyRevisionDeltas("create_task", prior, "description should be Check stock levels first")
	if !changed {
		t.Fatal("expected changed=true")
	}
	if next["description"] != "Check stock levels first" {
		t.Fatalf("description = %#v", next["description"])
	}
}

func TestApplyRevisionDeltas_CreateTaskDescriptionAppend_unrelatedQuestionNoMatch(t *testing.T) {
	prior := map[string]any{"title": "Refill OHN"}
	_, changed := applyRevisionDeltas("create_task", prior, "Before I confirm — which zone should this task refer to?")
	if changed {
		t.Fatal("expected changed=false — this is a zone clarification, not a description addition")
	}
}

func TestApplyRevisionDeltas_CreateTaskNoSpuriousChange(t *testing.T) {
	prior := map[string]any{"title": "Check humidity"}
	_, changed := applyRevisionDeltas("create_task", prior, "when should I run this?")
	if changed {
		t.Fatal("expected changed=false for clarifying question")
	}
}

func TestApplyRevisionDeltas_CreateTaskZoneIDNumeric(t *testing.T) {
	prior := map[string]any{"title": "Refill calcium nitrate"}
	next, changed := applyRevisionDeltas("create_task", prior, "assign it to zone 3 for now")
	if !changed {
		t.Fatal("expected changed=true")
	}
	if next["zone_id"].(float64) != 3 {
		t.Fatalf("zone_id = %#v want 3", next["zone_id"])
	}
}

func TestTaskZoneRevisionCue(t *testing.T) {
	if !taskZoneRevisionCue("Put it in Veg Room — that is the zone for this task.") {
		t.Fatal("expected zone revision cue for assignment turn")
	}
	if taskZoneRevisionCue("Before I confirm — which zone should this task refer to?") {
		t.Fatal("clarifying question should not trigger zone revise")
	}
}

func TestParseTaskZoneIDNumeric(t *testing.T) {
	if zid, ok := parseTaskZoneIDNumeric("use zone id 12"); !ok || zid != 12 {
		t.Fatalf("zone id 12: got %d ok=%v", zid, ok)
	}
	if _, ok := parseTaskZoneIDNumeric("which zone should this refer to?"); ok {
		t.Fatal("expected no numeric zone match")
	}
}

func TestParseTaskDueDateRevision(t *testing.T) {
	fixed := time.Date(2026, 7, 14, 15, 30, 0, 0, time.UTC)
	if due, ok := parseTaskDueDateRevisionAt("set the due date to 2026-07-20", fixed); !ok || due != "2026-07-20" {
		t.Fatalf("set due date: got %q ok=%v", due, ok)
	}
	if due, ok := parseTaskDueDateRevisionAt("due date should be 2026-08-01", fixed); !ok || due != "2026-08-01" {
		t.Fatalf("due date should be: got %q ok=%v", due, ok)
	}
	if due, ok := parseTaskDueDateRevisionAt("make it due tomorrow", fixed); !ok || due != "2026-07-15" {
		t.Fatalf("due tomorrow: got %q ok=%v", due, ok)
	}
	if due, ok := parseTaskDueDateRevisionAt("due in 3 days", fixed); !ok || due != "2026-07-17" {
		t.Fatalf("due in 3 days: got %q ok=%v", due, ok)
	}
	if _, ok := parseTaskDueDateRevisionAt("when should I run this?", fixed); ok {
		t.Fatal("clarifying question should not match due date")
	}
}

func TestApplyRevisionDeltas_CreateTaskDueDate(t *testing.T) {
	prior := map[string]any{"title": "Refill calcium nitrate"}
	next, changed := applyRevisionDeltas("create_task", prior, "make it due tomorrow")
	if !changed {
		t.Fatal("expected changed=true")
	}
	if next["title"] != "Refill calcium nitrate" {
		t.Fatalf("title = %#v want preserved Refill calcium nitrate", next["title"])
	}
	want := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	if next["due_date"] != want {
		t.Fatalf("due_date = %#v want %s", next["due_date"], want)
	}
}

func TestParseTaskTitleRevision_rejectsDueTomorrowAsTitle(t *testing.T) {
	prior := map[string]any{"title": "Refill calcium nitrate"}
	if title, ok := parseTaskTitleRevision("make it due tomorrow", prior); ok {
		t.Fatalf("expected no title revision, got %q", title)
	}
}

func TestParseTaskTitleRevision_callItStillWorks(t *testing.T) {
	prior := map[string]any{"title": "Follow up from Guardian chat"}
	if title, ok := parseTaskTitleRevision("call it Refill calcium nitrate instead", prior); !ok || title != "Refill calcium nitrate" {
		t.Fatalf("got title=%q ok=%v", title, ok)
	}
}

func TestLooksLikeDueDatePhrase(t *testing.T) {
	for _, s := range []string{"due tomorrow", "tomorrow", "due in 3 days", "2026-07-20"} {
		if !looksLikeDueDatePhrase(s) {
			t.Fatalf("want due-date phrase: %q", s)
		}
	}
	for _, s := range []string{"Refill calcium nitrate", "Inspect tent RH", ""} {
		if looksLikeDueDatePhrase(s) {
			t.Fatalf("want not due-date phrase: %q", s)
		}
	}
}

func TestExtractOperatorFacts_RH(t *testing.T) {
	facts := extractOperatorFacts("there's no humidity sensor in Tent A — assume RH around 60%")
	if len(facts) != 1 {
		t.Fatalf("got %d facts want 1: %#v", len(facts), facts)
	}
	f := facts[0]
	if f.Field != "rh_pct" || f.Basis != "operator_stated" {
		t.Fatalf("fact = %#v", f)
	}
	if f.Value.(int) != 60 {
		t.Fatalf("value = %#v want 60", f.Value)
	}
	if f.Label == "" || !contains(f.Label, "operator-stated") {
		t.Fatalf("label must mark operator-stated: %q", f.Label)
	}
}

func TestExtractOperatorFacts_WaterSource(t *testing.T) {
	facts := extractOperatorFacts("water source is well water on this line")
	if len(facts) != 1 || facts[0].Field != "water_source" || facts[0].Value != "well" {
		t.Fatalf("facts = %#v", facts)
	}
}

func TestExtractOperatorFacts_NoneWithoutCue(t *testing.T) {
	if facts := extractOperatorFacts("set EC to 1.0"); len(facts) != 0 {
		t.Fatalf("expected no facts, got %#v", facts)
	}
}

func TestMergeOperatorFacts_LaterOverrides(t *testing.T) {
	prior := []OperatorFact{{Field: "rh_pct", Value: 55, Basis: "operator_stated"}}
	next := []OperatorFact{{Field: "rh_pct", Value: 60, Basis: "operator_stated"}}
	merged := mergeOperatorFacts(prior, next)
	if len(merged) != 1 || merged[0].Value != 60 {
		t.Fatalf("merged = %#v", merged)
	}
}

func TestImpactSummary_PatchFertigation(t *testing.T) {
	lines := ImpactSummary("patch_fertigation_program", map[string]any{
		"program_id":          float64(7),
		"total_volume_liters": 0.3,
	}, nil)
	if len(lines) == 0 || !contains(lines[0], "0.3") || !contains(lines[0], "no run triggered now") {
		t.Fatalf("impact = %#v", lines)
	}
}

func TestImpactSummary_CreateTaskFromAlert_LowStock(t *testing.T) {
	lines := ImpactSummary("create_task_from_alert", map[string]any{
		"title":             "Refill OHN",
		"alert_subject":     "Inventory low: OHN at 1.00 (threshold 3.00)",
		"alert_source_type": "inventory_low_stock",
	}, nil)
	if len(lines) != 1 {
		t.Fatalf("lines=%v", lines)
	}
	if !strings.Contains(lines[0], "refill task") || !strings.Contains(lines[0], "OHN") {
		t.Fatalf("unexpected impact: %q", lines[0])
	}
	if !strings.Contains(lines[0], "Supplies hub") {
		t.Fatalf("expected Supplies hub hint: %q", lines[0])
	}
}

func TestImpactSummary_AppendsOperatorFacts(t *testing.T) {
	lines := ImpactSummary("create_plant", map[string]any{"crop_key": "basil"}, []OperatorFact{
		{Field: "rh_pct", Value: 60, Basis: "operator_stated", Label: "RH 60% (operator-stated, not measured)"},
	})
	joined := ""
	for _, l := range lines {
		joined += l + "\n"
	}
	if !contains(joined, "crop_key=basil") || !contains(joined, "operator-stated") {
		t.Fatalf("impact = %#v", lines)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
