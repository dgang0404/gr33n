package ingest

import (
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

func TestTaskDocument(t *testing.T) {
	desc := "harvest basil"
	tk := db.Gr33ncoreTask{
		ID:          1,
		Title:       "Pick herbs",
		Description: &desc,
		Status:      commontypes.TaskStatusEnum("todo"),
	}
	out := TaskDocument(tk)
	if out == "" || len(out) < 10 {
		t.Fatalf("unexpected: %q", out)
	}
}

func TestCropCycleDocument(t *testing.T) {
	c := db.Gr33nfertigationCropCycle{
		ID:          9,
		FarmID:      1,
		ZoneID:      42,
		Name:        "Basil Block A",
		IsActive:    true,
		StartedAt:   pgtype.Date{Valid: true, Time: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)},
		CurrentStage: db.NullGr33nfertigationGrowthStageEnum{
			Valid:                           true,
			Gr33nfertigationGrowthStageEnum: db.Gr33nfertigationGrowthStageEnumLateVeg,
		},
	}
	out := CropCycleDocument(c)
	if out == "" || len(out) < 20 {
		t.Fatalf("unexpected: %q", out)
	}
	for _, sub := range []string{"crop_cycle:", "Basil Block A", "zone_id: 42", "late_veg", "active: yes"} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in: %q", sub, out)
		}
	}
}

func TestFertigationProgramDocument(t *testing.T) {
	desc := "Daily JLF foliar feed"
	meta := []byte(`{"tags":["veg","demo"]}`)
	p := db.Gr33nfertigationProgram{
		ID:         3,
		FarmID:     1,
		Name:       "Veg Daily JLF Program",
		Description: &desc,
		IsActive:   true,
		Metadata:   meta,
	}
	out := FertigationProgramDocument(p)
	if len(out) < 30 {
		t.Fatalf("unexpected: %q", out)
	}
	for _, sub := range []string{"fertigation_program:", "Veg Daily JLF", "active: yes", "tags", "veg"} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in: %q", sub, out)
		}
	}
}

func TestScheduleDocument(t *testing.T) {
	s := db.Gr33ncoreSchedule{
		ID:             1,
		FarmID:         1,
		Name:           "Morning lights",
		ScheduleType:   "cron",
		CronExpression: "0 8 * * *",
		Timezone:       "America/Los_Angeles",
		IsActive:       true,
	}
	out := ScheduleDocument(s)
	for _, sub := range []string{"schedule:", "Morning lights", "cron_expression:", "America/Los_Angeles"} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in: %q", sub, out)
		}
	}
}

func TestAutomationRuleDocument(t *testing.T) {
	r := db.Gr33ncoreAutomationRule{
		ID:            2,
		FarmID:        1,
		Name:          "EC low alert",
		IsActive:      true,
		TriggerSource: commontypes.AutomationTriggerSensor,
	}
	out := AutomationRuleDocument(r)
	for _, sub := range []string{"automation_rule:", "EC low alert", "sensor_threshold"} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in: %q", sub, out)
		}
	}
}

func TestExecutableActionDocument(t *testing.T) {
	sid := int64(10)
	params := []byte(`{"reason":"demo","webhook_url":"https://x.example/hook"}`)
	a := db.Gr33ncoreExecutableAction{
		ID:               3,
		ScheduleID:       &sid,
		ExecutionOrder:   1,
		ActionType:       commontypes.ExecutableActionNotification,
		ActionParameters: params,
	}
	out := ExecutableActionDocument(a)
	if strings.Contains(out, "example") || strings.Contains(out, "webhook_url") {
		t.Fatalf("must scrub sensitive params: %q", out)
	}
	if !strings.Contains(out, "schedule_id") || !strings.Contains(out, "send_notification") {
		t.Fatalf("unexpected: %q", out)
	}
}

func TestCostTransactionDocumentOmitsMoney(t *testing.T) {
	desc := "Organic potting soil"
	ref := "INV-2024-001"
	cp := "Local nursery"
	amt := pgtype.Numeric{}
	_ = amt.Scan("123.45")
	d := pgtype.Date{Valid: true, Time: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)}
	c := db.Gr33ncoreCostTransaction{
		ID:               100,
		FarmID:           1,
		TransactionDate:  d,
		Category:         commontypes.CostCategorySeedsPlants,
		Subcategory:      ptrStr("inputs"),
		Description:      &desc,
		IsIncome:         false,
		Amount:           amt,
		Currency:         "USD",
		DocumentReference: &ref,
		Counterparty:     &cp,
	}
	out := CostTransactionDocument(c)
	if strings.Contains(out, "123") || strings.Contains(out, "USD") || strings.Contains(out, "amount") {
		t.Fatalf("must not embed money or currency: %q", out)
	}
	for _, sub := range []string{"cost_transaction", "seeds_plants", "inputs", "expense", "Organic potting"} {
		if !strings.Contains(out, sub) {
			t.Fatalf("missing %q in: %q", sub, out)
		}
	}
	if !strings.Contains(out, ref) || !strings.Contains(out, cp) {
		t.Fatalf("expected memo fields: %q", out)
	}
}

func ptrStr(s string) *string { return &s }
