package farmguardian

import "testing"

func TestShouldRunLookupCropSymptomsIntent(t *testing.T) {
	cases := []struct {
		q    string
		want bool
	}{
		{"Yellow leaves on my tomato", true},
		{"What is wrong with my basil?", true},
		{"Tip burn on lettuce", true},
		{"Compare cannabis vs tomato EC targets", false},
		{"What did this zone cost?", false},
	}
	for _, c := range cases {
		got := shouldRunLookupCropSymptomsIntent(c.q, nil)
		if got != c.want {
			t.Fatalf("shouldRunLookupCropSymptomsIntent(%q) = %v want %v", c.q, got, c.want)
		}
	}
}

func TestReadToolIDsIncludesLookupCropSymptoms(t *testing.T) {
	ids := ReadToolIDs()
	found := false
	for _, id := range ids {
		if id == "lookup_crop_symptoms" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ReadToolIDs missing lookup_crop_symptoms: %v", ids)
	}
}
