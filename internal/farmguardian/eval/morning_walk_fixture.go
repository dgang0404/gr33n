package eval

import "gr33n-api/internal/farmguardian"

// MorningWalkContextRef matches Today → Morning check (dashboard_morning).
func MorningWalkContextRef() *farmguardian.ContextRef {
	return &farmguardian.ContextRef{
		Type:         "route",
		Path:         "/",
		Name:         "Today",
		GuardianMode: "morning_walkthrough",
		Surface:      "dashboard_morning",
	}
}

// MorningWalkPrompt matches buildMorningWalkStarters() in ui/src/lib/guardianStarters.js.
func MorningWalkPrompt() string {
	return "Run my morning farm walkthrough. Use walk_farm — check unacknowledged alerts, today's feeds, offline devices, comfort bands, and low stock. Skip categories with nothing to flag. Plain language only."
}
