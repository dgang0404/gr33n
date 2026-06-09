package farmguardian

import "testing"

func TestShouldRunSiteWeatherReadIntent(t *testing.T) {
	if !shouldRunSiteWeatherReadIntent("Do I need supplemental light today?") {
		t.Fatal("expected supplemental light intent")
	}
	if shouldRunSiteWeatherReadIntent("list unread alerts") {
		t.Fatal("should not match unrelated question")
	}
}

func TestReadToolIDs_IncludesSiteWeather(t *testing.T) {
	for _, id := range ReadToolIDs() {
		if id == "site_weather" {
			return
		}
	}
	t.Fatal("site_weather missing from ReadToolIDs")
}
