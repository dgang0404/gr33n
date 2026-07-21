package eval

import "strings"

func normalizeSuiteName(suite string) string {
	return strings.ToLower(strings.TrimSpace(suite))
}

// SuiteNeedsWarmup is true when eval should POST /guardian/warmup before the first grounded prompt.
func SuiteNeedsWarmup(suite string) bool {
	switch normalizeSuiteName(suite) {
	case "smoke", "smoke-natural-farming", "smoke_natural_farming", "smoke-nf",
		"smoke-full", "smoke_full",
		"phase127", "phase128", "p128",
		"change-requests", "change_requests", "proposals", "pr":
		return true
	default:
		return IsScenarioSuite(suite)
	}
}

// SuiteRequiresWarmup fails the run when warmup does not reach ready (smoke / QA suites).
func SuiteRequiresWarmup(suite string) bool {
	switch normalizeSuiteName(suite) {
	case "smoke", "smoke-natural-farming", "smoke_natural_farming", "smoke-nf",
		"smoke-full", "smoke_full",
		"phase127", "phase128", "p128",
		"change-requests", "change_requests", "proposals", "pr":
		return true
	default:
		return false
	}
}
