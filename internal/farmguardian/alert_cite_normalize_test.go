package farmguardian

import (
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

// Run #9 archive — item 2 has correct [2] plus stray "platform docs [3]".
const run9UnreadAlertsStrayRef = `Unread Alert Summary:

1. High Humidity - The humidity level inside the Flower Room has reached a high of 72.4% RH [1].

2. OHN Batch Below Minimum - An Oriental Herbal Nutrient (OHN) batch, specifically SEED-OHN-001 with 0.35 L remaining [2], is below the minimum threshold of 0.5 L required for your immunity drenches. You should either prepare to brew another fresh OHN batch or reconsider and potentially increase the inventory's reorder point as per platform docs [3].

3. Light Schedule Change - There is an upcoming change in your light schedule for Flower Room [3].`

func run9AlertChunksOnly() []db.SearchRagNearestNeighborsFilteredRow {
	return []db.SearchRagNearestNeighborsFilteredRow{
		{ID: 5, SourceType: SourceTypeAlertNotification, ContentText: "severity: high\nsubject: Humidity high — Flower Room"},
		{ID: 4, SourceType: SourceTypeAlertNotification, ContentText: "severity: medium\nsubject: OHN batch below minimum"},
		{ID: 3, SourceType: SourceTypeAlertNotification, ContentText: "severity: low\nsubject: Light schedule change"},
	}
}

func TestNormalizeAlertListCitations_run9StrayRef(t *testing.T) {
	t.Parallel()
	got, ok := NormalizeAlertListCitations(run9UnreadAlertsStrayRef, run9AlertChunksOnly())
	if !ok {
		t.Fatal("expected normalization")
	}
	if strings.Contains(got, "platform docs [3]") {
		t.Fatalf("stray [3] on item 2 should be removed: %s", got)
	}
	if !strings.Contains(got, "[2]") {
		t.Fatal("expected [2] preserved on item 2")
	}
	if strings.Count(got, "[3]") != 1 {
		t.Fatalf("expected exactly one [3] on item 3, got: %s", got)
	}
}

func TestMultipleCitationsPerListItemNote_run9(t *testing.T) {
	t.Parallel()
	note := MultipleCitationsPerListItemNote(run9UnreadAlertsStrayRef)
	if note != "multiple_citations_per_list_item: item 2" {
		t.Fatalf("note=%q", note)
	}
}

func TestMultipleCitationsPerListItemNote_normalizedPasses(t *testing.T) {
	t.Parallel()
	got, _ := NormalizeAlertListCitations(run9UnreadAlertsStrayRef, run9AlertChunksOnly())
	if note := MultipleCitationsPerListItemNote(got); note != "" {
		t.Fatalf("normalized answer should pass: %s", note)
	}
}

func TestCitationClaimMismatchNote_run9AfterNormalize(t *testing.T) {
	t.Parallel()
	got, _ := NormalizeAlertListCitations(run9UnreadAlertsStrayRef, run9AlertChunksOnly())
	cites := []CitationSummary{
		{Ref: 1, Excerpt: "severity: high\nsubject: Humidity high — Flower Room\nmessage: 72.4% RH"},
		{Ref: 2, Excerpt: "severity: medium\nsubject: OHN batch below minimum\nmessage: SEED-OHN-001 0.35 L"},
		{Ref: 3, Excerpt: "severity: low\nsubject: Light schedule change\nmessage: photoperiod transition"},
	}
	if note := CitationClaimMismatchNote(got, cites); note != "" {
		t.Fatalf("expected no mismatch after normalize: %s", note)
	}
}

func TestNormalizeAlertListCitations_mixedChunksNoOp(t *testing.T) {
	t.Parallel()
	chunks := []db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "platform_doc"},
		{SourceType: SourceTypeAlertNotification},
		{SourceType: SourceTypeAlertNotification},
	}
	_, ok := NormalizeAlertListCitations(run9UnreadAlertsStrayRef, chunks)
	if ok {
		t.Fatal("expected no-op when sources are not alert-only")
	}
}
