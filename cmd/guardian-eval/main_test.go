// Pure-logic coverage for -fail-on-regression and -check-pending-proposals.
// No live LLM needed: regressionFailures/passedProposalFixtures only inspect
// already-scored results, and reportPendingProposals is tested against a
// local httptest server standing in for GET /v1/chat/proposals.

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/farmguardian/eval"
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

func TestPassedProposalFixtures_countsOnlyPassedExpectProposal(t *testing.T) {
	fixtures := []eval.Question{
		{ID: "write-feed", ExpectProposal: true},
		{ID: "write-ack", ExpectProposal: true},
		{ID: "farm-alerts", ExpectProposal: false},
	}
	scores := []farmguardian.EvalQuestionScore{
		{ID: "write-feed", Passed: true},
		{ID: "write-ack", Passed: false},
		{ID: "farm-alerts", Passed: true},
	}
	if got := passedProposalFixtures(fixtures, scores); got != 1 {
		t.Fatalf("expected 1 passed proposal fixture, got %d", got)
	}
}

func TestReportPendingProposals_enoughRowsIsNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"proposals": []map[string]any{
				{"proposal_id": "p1", "tool": "update_fertigation_program", "summary": "Set feed to 0.3L", "risk_tier": "medium"},
			},
		})
	}))
	defer srv.Close()

	client := eval.NewAPIClient(srv.URL, "test-token", 1)
	if err := reportPendingProposals(t.Context(), client, 1, nil); err != nil {
		t.Fatalf("expected nil error with enough pending rows, got %v", err)
	}
}

func TestReportPendingProposals_tooFewRowsErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"proposals": []map[string]any{}})
	}))
	defer srv.Close()

	client := eval.NewAPIClient(srv.URL, "test-token", 1)
	err := reportPendingProposals(t.Context(), client, 2, nil)
	if err == nil {
		t.Fatal("expected error when fewer pending rows than expected")
	}
}

func TestPassedProposalIDs_collectsFromPassedWriteFixtures(t *testing.T) {
	fixtures := []eval.Question{
		{ID: "write-feed", ExpectProposal: true},
		{ID: "write-ack", ExpectProposal: true},
	}
	scores := []farmguardian.EvalQuestionScore{
		{ID: "write-feed", Passed: true, ProposalIDs: []string{"p-feed"}},
		{ID: "write-ack", Passed: false, ProposalIDs: []string{"p-ack"}},
	}
	got := passedProposalIDs(fixtures, scores)
	if len(got) != 1 || got[0] != "p-feed" {
		t.Fatalf("got %v want [p-feed]", got)
	}
}

func TestReportPendingProposals_requiredIDsMustMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"proposals": []map[string]any{
				{"proposal_id": "stale-old", "tool": "ack_alert", "summary": "old", "risk_tier": "low"},
			},
		})
	}))
	defer srv.Close()

	client := eval.NewAPIClient(srv.URL, "test-token", 1)
	err := reportPendingProposals(t.Context(), client, 1, []string{"fresh-new"})
	if err == nil || !strings.Contains(err.Error(), "fresh-new") {
		t.Fatalf("expected missing fresh-new error, got %v", err)
	}
}

func TestReportPendingProposals_requiredIDsPassWhenPresent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"proposals": []map[string]any{
				{"proposal_id": "p1", "tool": "ack_alert", "summary": "Ack", "risk_tier": "low"},
			},
		})
	}))
	defer srv.Close()

	client := eval.NewAPIClient(srv.URL, "test-token", 1)
	if err := reportPendingProposals(t.Context(), client, 0, []string{"p1"}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestReportPendingProposals_zeroExpectedNeverErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"proposals": []map[string]any{}})
	}))
	defer srv.Close()

	client := eval.NewAPIClient(srv.URL, "test-token", 1)
	if err := reportPendingProposals(t.Context(), client, 0, nil); err != nil {
		t.Fatalf("expected nil error when nothing was expected, got %v", err)
	}
}
