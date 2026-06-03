package automation

import (
	"testing"

	db "gr33n-api/internal/db"
)

func TestIsLuxPARSensorType(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"lux", true},
		{"par_umol", true},
		{"soil_moisture", false},
	} {
		if got := IsLuxPARSensorType(tc.in); got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.in, got, tc.want)
		}
	}
}

func TestGreenhouseRuleFamily(t *testing.T) {
	if GreenhouseRuleFamily("GH — High lux: deploy shade (zone 3)") != ghRuleFamilyHighLux {
		t.Fatal("expected high_lux family")
	}
	if GreenhouseRuleFamily("other rule") != "" {
		t.Fatal("expected empty family")
	}
}

func TestPlanGreenhouseTemplateSkips(t *testing.T) {
	shade := int64(1)
	_, err := planGreenhouseTemplateSkips(&shade, nil, nil, nil, false, false)
	if err == nil {
		t.Fatal("expected error without lux sensor or override")
	}
	skipped, err := planGreenhouseTemplateSkips(&shade, nil, nil, nil, true, false)
	if err != nil || len(skipped) != 1 || skipped[0] != ghRuleFamilyHighLux {
		t.Fatalf("got skipped=%v err=%v", skipped, err)
	}
}

func TestValidateGreenhouseRuleActivation_MissingSensor(t *testing.T) {
	rule := db.Gr33ncoreAutomationRule{
		FarmID: 1,
		Name:   "GH — High lux: deploy shade (zone 1)",
		TriggerConfiguration: []byte(`{"op":"gt","value":80000}`),
	}
	err := ValidateGreenhouseRuleActivation(t.Context(), nil, rule)
	if err == nil {
		t.Fatal("expected error for missing sensor_id")
	}
}
