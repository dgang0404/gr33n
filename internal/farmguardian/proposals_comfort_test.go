package farmguardian

import (
	"context"
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

type fakeComfortQuerier struct {
	rules     []db.Gr33ncoreAutomationRule
	schedules []db.Gr33ncoreSchedule
	programs  []db.Gr33nfertigationProgram
	zones     []db.Gr33ncoreZone
	sensors   []db.Gr33ncoreSensor
}

func (f *fakeComfortQuerier) ListAutomationRulesByFarm(context.Context, int64) ([]db.Gr33ncoreAutomationRule, error) {
	return f.rules, nil
}

func (f *fakeComfortQuerier) ListSchedulesByFarm(context.Context, int64) ([]db.Gr33ncoreSchedule, error) {
	return f.schedules, nil
}

func (f *fakeComfortQuerier) ListProgramsByFarm(context.Context, int64) ([]db.Gr33nfertigationProgram, error) {
	return f.programs, nil
}

func (f *fakeComfortQuerier) ListZonesByFarm(context.Context, int64) ([]db.Gr33ncoreZone, error) {
	return f.zones, nil
}

func (f *fakeComfortQuerier) ListSensorsByFarm(context.Context, int64) ([]db.Gr33ncoreSensor, error) {
	return f.sensors, nil
}

func TestMatchComfortAutomationIntent_DisableShadeRule(t *testing.T) {
	ctx := context.Background()
	zoneID := int64(3)
	fq := &fakeComfortQuerier{
		zones: []db.Gr33ncoreZone{{ID: zoneID, Name: "Flower Room"}},
		rules: []db.Gr33ncoreAutomationRule{{
			ID:                   12,
			Name:                 "GH — deploy shade when hot",
			IsActive:             true,
			TriggerConfiguration: []byte(`{"zone_id": 3}`),
		}},
	}
	snap := Snapshot{ZoneNames: []string{"Flower Room"}}

	tool, args, summary, ok := matchComfortAutomationIntent(ctx, fq, 1,
		"Disable the greenhouse shade rule for Flower Room until I turn it back on", snap)
	if !ok {
		t.Fatal("expected match")
	}
	if tool != "patch_rule" {
		t.Fatalf("tool=%q want patch_rule", tool)
	}
	if args["rule_id"] != int64(12) {
		t.Fatalf("rule_id=%v want 12", args["rule_id"])
	}
	if active, _ := args["is_active"].(bool); active {
		t.Fatalf("is_active=%v want false", active)
	}
	if !strings.Contains(strings.ToLower(summary), "shade") {
		t.Fatalf("summary=%q", summary)
	}
}

func TestMatchComfortAutomationIntent_PauseLightsSchedule(t *testing.T) {
	ctx := context.Background()
	fq := &fakeComfortQuerier{
		schedules: []db.Gr33ncoreSchedule{{ID: 7, Name: "Flower lights ON", IsActive: true}},
	}
	tool, args, _, ok := matchComfortAutomationIntent(ctx, fq, 1, "Pause the lights schedule for tonight", Snapshot{})
	if !ok || tool != "patch_schedule" {
		t.Fatalf("got tool=%q ok=%v want patch_schedule", tool, ok)
	}
	if args["schedule_id"] != int64(7) {
		t.Fatalf("schedule_id=%v", args["schedule_id"])
	}
}

func TestMatchComfortAutomationIntent_SetEC(t *testing.T) {
	ctx := context.Background()
	zoneID := int64(3)
	fq := &fakeComfortQuerier{
		zones: []db.Gr33ncoreZone{{ID: zoneID, Name: "Flower Room"}},
		programs: []db.Gr33nfertigationProgram{{
			ID:           4,
			Name:         "Flower feed",
			IsActive:     true,
			TargetZoneID: &zoneID,
		}},
	}
	tool, args, summary, ok := matchComfortAutomationIntent(ctx, fq, 1,
		"Set EC target to 1.8 for Flower Room", Snapshot{ZoneNames: []string{"Flower Room"}})
	if !ok || tool != "patch_fertigation_program" {
		t.Fatalf("got tool=%q ok=%v", tool, ok)
	}
	if args["ec_trigger_low"] != 1.8 {
		t.Fatalf("ec=%v", args["ec_trigger_low"])
	}
	if !strings.Contains(summary, "1.8") {
		t.Fatalf("summary=%q", summary)
	}
}

func TestMatchComfortAutomationIntent_NoMatchOnQA(t *testing.T) {
	ctx := context.Background()
	fq := &fakeComfortQuerier{
		rules: []db.Gr33ncoreAutomationRule{{ID: 1, Name: "GH shade", IsActive: true}},
	}
	_, _, _, ok := matchComfortAutomationIntent(ctx, fq, 1, "What is EC?", Snapshot{})
	if ok {
		t.Fatal("expected no proposal for pure Q&A")
	}
}

func TestPickRuleForIntent_ByID(t *testing.T) {
	rules := []db.Gr33ncoreAutomationRule{
		{ID: 5, Name: "Vent when hot"},
		{ID: 12, Name: "GH shade"},
	}
	got, ok := pickRuleForIntent(rules, "pause rule #12", 0, nil)
	if !ok || got.ID != 12 {
		t.Fatalf("got id=%d ok=%v", got.ID, ok)
	}
}

func TestDisableRuleIntent_MatchesShadePhrase(t *testing.T) {
	q := "Turn off the shade automation for Flower Room"
	if !disableRuleIntent.MatchString(q) {
		t.Fatalf("expected disable intent match for %q", q)
	}
}
