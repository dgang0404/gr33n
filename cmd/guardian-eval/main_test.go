// Phase 153 WS1 — pure-logic coverage for -fail-on-regression. No LLM/DB
// needed: regressionFailures only inspects already-scored results.

package main

import (
	"testing"

	"gr33n-api/internal/farmguardian"
)

func TestRegressionFailures_allPassedIsEmpty(t *testing.T) {
	details := map[string][]farmguardian.EvalQuestionScore{
		"phi3-mini": {
			{ID: "smoke-ec-ph", Passed: true},
			{ID: "smoke-unread-alerts", Passed: true},
		},
	}
	if got := regressionFailures(details); len(got) != 0 {
		t.Fatalf("expected no failures, got %v", got)
	}
}

func TestRegressionFailures_reportsFailedFixtures(t *testing.T) {
	details := map[string][]farmguardian.EvalQuestionScore{
		"phi3-mini": {
			{ID: "smoke-ec-ph", Passed: true},
			{ID: "smoke-unread-alerts", Passed: false, Notes: "citation_number_mismatch"},
		},
	}
	got := regressionFailures(details)
	if len(got) != 1 {
		t.Fatalf("expected 1 failure, got %v", got)
	}
	if got[0] != "phi3-mini/smoke-unread-alerts: citation_number_mismatch" {
		t.Fatalf("unexpected failure line: %q", got[0])
	}
}

func TestRegressionFailures_sortedAcrossModels(t *testing.T) {
	details := map[string][]farmguardian.EvalQuestionScore{
		"tinyllama": {{ID: "smoke-ec-ph", Passed: false, Notes: "low_relevance"}},
		"phi3-mini": {{ID: "smoke-ec-ph", Passed: false, Notes: "topic_drift"}},
	}
	got := regressionFailures(details)
	if len(got) != 2 {
		t.Fatalf("expected 2 failures, got %v", got)
	}
	if got[0] != "phi3-mini/smoke-ec-ph: topic_drift" || got[1] != "tinyllama/smoke-ec-ph: low_relevance" {
		t.Fatalf("expected alphabetically sorted failures, got %v", got)
	}
}
