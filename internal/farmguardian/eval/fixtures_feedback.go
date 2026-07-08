package eval

// Phase 146 WS3 — operator-promoted feedback regression fixtures.
// Run scripts/guardian-feedback-to-fixture.sh to emit candidate JSON;
// promote rows here manually after triage (see guardian-feedback-review-runbook.md).

// FeedbackFixtureCandidates are not executed automatically — they document
// thumbs-down rows worth promoting into score_*_test.go archived answers.
var FeedbackFixtureCandidates = []Question{
	// Example shape after promotion:
	// {ID: "feedback-ec-ph-drift-001", Category: "field_guide", Prompt: "...", ExpectCitation: true, Grounded: true},
}
