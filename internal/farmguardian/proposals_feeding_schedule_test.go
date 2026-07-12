package farmguardian

import "testing"

func TestPauseScheduleIntent_MatchesLightsSchedulePhrase(t *testing.T) {
	q := "Pause the lights schedule for Veg Tent until tomorrow."
	if !pauseScheduleIntent.MatchString(q) {
		t.Fatalf("pauseScheduleIntent should match %q", q)
	}
}

func TestZoneIntentMatchesNickname_VegTent(t *testing.T) {
	if !zoneIntentMatchesNickname("pause the lights schedule for veg tent", "Veg Room") {
		t.Fatal("veg tent should match Veg Room zone")
	}
}
