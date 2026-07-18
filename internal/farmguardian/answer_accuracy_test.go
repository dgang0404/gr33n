// Phase 148 WS1 — regression fixtures from smoke run #6 (2026-07-08) archive
// data/guardian_qa_runs/20260708T153829_smoke_phi3-mini.json.

package farmguardian

import (
	"strings"
	"testing"
)

func TestGarbledTokenNote_detectsRunSixOHNTypo(t *testing.T) {
	t.Parallel()
	answer := "the current amount stands at 0.35 L against a threshold of 0sourced from FIELD GUIDE [1]."
	if note := GarbledTokenNote(answer); note == "" {
		t.Fatal("expected garbled token detection")
	}
}

func TestGarbledTokenNote_ignoresLegitimateUnits(t *testing.T) {
	t.Parallel()
	answer := "Target 1.2–2.0 mS/cm and Air Humidity Indoor read 72.4% RH at 15–20 °C with a 1:1000 dilution."
	if note := GarbledTokenNote(answer); note != "" {
		t.Fatalf("unexpected garbled note %q", note)
	}
}

func TestDuplicateListItemNote_detectsRepeatedOHNAlert(t *testing.T) {
	t.Parallel()
	answer := `1. High humidity alert in the Flower Room [3] - address ventilation.
2. Low OHN batch below minimum [4] - reorder or brew soon for immunity drenches.
3. Light schedule change alert [5] - confirm timers.
4. OHN batch below minimum - ensure OHN replenishment for immunity drenches ongoing.`
	note := DuplicateListItemNote(answer)
	if note == "" {
		t.Fatal("expected duplicate list item detection between items 2 and 4")
	}
}

func TestDuplicateListItemNote_passesDistinctItems(t *testing.T) {
	t.Parallel()
	answer := `1. High humidity alert in the Flower Room - address ventilation.
2. Low OHN batch below minimum - reorder or brew soon.
3. Light schedule change alert - confirm timers before the flip.`
	if note := DuplicateListItemNote(answer); note != "" {
		t.Fatalf("unexpected duplicate note %q", note)
	}
}

func TestCitationClaimMismatchNote_detectsRunSixAlertMislabel(t *testing.T) {
	t.Parallel()
	answer := "1. High humidity alert in the Flower Room [3] - address ventilation before mildew risk."
	cites := []CitationSummary{
		{Ref: 3, Excerpt: "severity: low\nsubject: Light schedule change in 48 hours — Flower Room\nmessage: Photoperiod transition reminder."},
		{Ref: 5, Excerpt: "severity: high\nsubject: Humidity high — Flower Room\nmessage: Air Humidity Indoor read 72.4% RH (alert threshold 65% for late flower)."},
	}
	note := CitationClaimMismatchNote(answer, cites)
	if note == "" {
		t.Fatal("expected citation_number_mismatch — humidity claim points at the photoperiod chunk")
	}
}

func TestCitationClaimMismatchNote_passesCorrectRef(t *testing.T) {
	t.Parallel()
	answer := "Humidity is high in the Flower Room [5] and requires dehumidification."
	cites := []CitationSummary{
		{Ref: 3, Excerpt: "severity: low\nsubject: Light schedule change in 48 hours — Flower Room."},
		{Ref: 5, Excerpt: "severity: high\nsubject: Humidity high — Flower Room\nmessage: Air Humidity Indoor read 72.4% RH."},
	}
	if note := CitationClaimMismatchNote(answer, cites); note != "" {
		t.Fatalf("unexpected mismatch note %q", note)
	}
}

func TestECPHUnitConfusionNote_detectsRunSixBlueberryPHAsEC(t *testing.T) {
	t.Parallel()
	answer := "For pH levels, the target is to maintain a slightly acidic environment with **4.5–5.5 mS/cm** EC specifically tailored for kale."
	cites := []CitationSummary{
		{Ref: 5, Excerpt: "Blueberries require an acidic root zone — pH 4.5–5.5 — not the 5.5–6.0 band most fruiting hydro crops use."},
	}
	note := ECPHUnitConfusionNote(answer, cites)
	if note == "" {
		t.Fatal("expected ph_ec_unit_confusion — 4.5–5.5 is a pH value in the cited excerpt, not mS/cm")
	}
}

func TestECPHUnitConfusionNote_passesWhenRangeIsGenuinelyEC(t *testing.T) {
	t.Parallel()
	answer := "Kale runs a slightly higher EC than lettuce at **1.0–1.5 mS/cm**."
	cites := []CitationSummary{
		{Ref: 3, Excerpt: "Kale sits slightly above lettuce EC (~1.0–1.5 mS/cm) and tolerates cooler root-zone temps."},
	}
	if note := ECPHUnitConfusionNote(answer, cites); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

func TestAnswerAccuracyNote_cleanAnswerPasses(t *testing.T) {
	t.Parallel()
	answer := "Humidity is high in the Flower Room [1] and OHN batch is low [2]; reorder soon."
	cites := []CitationSummary{
		{Ref: 1, Excerpt: "severity: high\nsubject: Humidity high — Flower Room."},
		{Ref: 2, Excerpt: "OHN batch below minimum — reorder or brew soon."},
	}
	if note := AnswerAccuracyNote(answer, cites); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

// TestTruncatedAnswerTailNote_liveUIRun152 reproduces the "...consistent and
// ade0:" cutoff seen in a live Farm Counsel run (Phase 152).
func TestTruncatedAnswerTailNote_liveUIRun152(t *testing.T) {
	t.Parallel()
	answer := "This suggests that your plants are receiving consistent and ade0:"
	if note := TruncatedAnswerTailNote(answer); note == "" {
		t.Fatal("expected truncated_answer_tail note")
	}
}

func TestTruncatedAnswerTailNote_completeSentencePasses(t *testing.T) {
	t.Parallel()
	answer := "Lights turn on at 06:00 daily during the first two weeks [4]."
	if note := TruncatedAnswerTailNote(answer); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

func TestTruncatedAnswerTailNote_allowlistedChemistryPasses(t *testing.T) {
	t.Parallel()
	answer := "Aquaponics biofilter off-gasses excess CO2"
	if note := TruncatedAnswerTailNote(answer); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

// Phase 190 — live turn: "Create a task to refill calcium nitrate when stock
// is low." Completion hit exactly 1024/1024 tokens; the answer stops right
// after promising a step-by-step procedure.
func TestDanglingListIntroNote_liveCalciumNitrateTask(t *testing.T) {
	t.Parallel()
	answer := "Objective: To ensure the lettuce crop's health, refill calcium nitrate in your hydroponic system when stock levels are low and maintain appropriate nutrient concentrations as per Guardian's field guide. Here is a step-by-step procedure for you to follow based on our sources [1] (Lettuce Nutrition), which will help manage the EC, pH balance, and prevent tip burn while refilling calcium nitrate:"
	note := DanglingListIntroNote(answer)
	if note == "" {
		t.Fatal("expected dangling_list_intro note")
	}
	if !strings.HasPrefix(note, "dangling_list_intro:") {
		t.Fatalf("got %q", note)
	}
}

// Phase 190 — live turn: "Set the feed volume to 0.3 liters for the Veg Tent
// program." The answer stops right after promising to walk through setup.
func TestDanglingListIntroNote_liveFeedVolumeSetup(t *testing.T) {
	t.Parallel()
	answer := "Firstly, ensure to have assigned the appropriate lettuce nutrition profile using `lookup_crop_targets` or by starting a new grow in Start grow and selecting 'Lettuce' from Plants for structured EC mS/cm targeting as suggested [5]. Once you confirm this setup:"
	if note := DanglingListIntroNote(answer); note == "" {
		t.Fatal("expected dangling_list_intro note")
	}
}

func TestDanglingListIntroNote_completeSentencePasses(t *testing.T) {
	t.Parallel()
	answer := "Set the feed volume to 0.3 liters per cycle for the Veg Tent program per [1]."
	if note := DanglingListIntroNote(answer); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

func TestDanglingListIntroNote_shortAnswerNotFlagged(t *testing.T) {
	t.Parallel()
	// Short trailing-colon fragments (e.g. a label) are below the length
	// floor so they don't false-positive on tiny, legitimate answers.
	answer := "Zone:"
	if note := DanglingListIntroNote(answer); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

func TestUncitedTimelineClaimNote_liveUIRun152(t *testing.T) {
	t.Parallel()
	answer := "As the cycle started on June 20th with no prior harvest tasks noted and considering it's now Week 9, you should be observing well-developed trichomes."
	if note := UncitedTimelineClaimNote(answer); note == "" {
		t.Fatal("expected uncited_timeline_claim note")
	}
}

func TestUncitedTimelineClaimNote_citedNearbyPasses(t *testing.T) {
	t.Parallel()
	answer := "The prior task in that room reports Week 9 [5], with the flush already complete."
	if note := UncitedTimelineClaimNote(answer); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

func TestInventedAssumptionMathNote_liveUIRun152(t *testing.T) {
	t.Parallel()
	answer := "That translates into about ~1.2 mL per plant if we assume an average yield density for your cultivar in this stage."
	if note := InventedAssumptionMathNote(answer); note == "" {
		t.Fatal("expected invented_assumption_math note")
	}
}

func TestInventedAssumptionMathNote_noNumberPasses(t *testing.T) {
	t.Parallel()
	answer := "Assuming the alert severity is accurate, escalate to the operator immediately."
	if note := InventedAssumptionMathNote(answer); note != "" {
		t.Fatalf("unexpected note %q", note)
	}
}

// TestAnswerAccuracyNote_liveUIBloomRunRun152 replays a Farm Counsel
// "Bloom run (12/12)" answer shape (Phase 152 / 164 demo farm) against the
// full detector chain.
func TestAnswerAccuracyNote_liveUIBloomRunRun152(t *testing.T) {
	t.Parallel()
	answer := `The "Bloom run (12/12)" in the Flower Room is currently at stage [1] early_flower, as indicated by both farm notes and fertigation programs active for this cycle on your farm today ([5]). The fertigation program is scheduled to run daily... and aims to deliver approximately 95 liters total volume, which translates into about [4] ~1.2 mL per plant if we assume an average yield density for your cultivar in this stage ([2]).

As the cycle started on June 20th of last year with no prior harvest tasks noted and considering it's now Week 9, you should be checking bloom openness and stem length on your chrysanthemum run. This suggests that your plants are receiving consistent and ade0:`
	cites := []CitationSummary{
		{Ref: 1, Excerpt: "crop_cycle: Bloom run (12/12) stage: early_flower active: yes started_at: 2026-06-20"},
		{Ref: 2, Excerpt: "fertigation_program: Flower Daily FFJ+WCA Program total_volume_liters: 95 ec_trigger_low: 1.4 ph_trigger_low: 5.8"},
		{Ref: 3, Excerpt: "schedule: Water Early Flower Daily ~900mL per plant daily."},
		{Ref: 4, Excerpt: "schedule: Light ON 12/12 Flower Lights on at 06:00. active: no"},
		{Ref: 5, Excerpt: "task: Harvest Flower Room A status: completed Week 9 short-day crop. Flush complete."},
	}
	if note := AnswerAccuracyNote(answer, cites); note == "" {
		t.Fatal("expected an accuracy note for the live-UI bloom-run answer")
	}
}
