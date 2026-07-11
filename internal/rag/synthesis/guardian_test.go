package synthesis

import (
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

func TestHasPlatformDocChunks(t *testing.T) {
	if HasPlatformDocChunks(nil) {
		t.Fatal("expected false for nil")
	}
	if HasPlatformDocChunks([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "task"},
	}) {
		t.Fatal("expected false for task only")
	}
	if !HasPlatformDocChunks([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "platform_doc"},
	}) {
		t.Fatal("expected true for platform_doc")
	}
}

func TestGuardianRAGInstructionsIncludesPlatformDocHint(t *testing.T) {
	base := GuardianRAGInstructions(nil)
	withDoc := GuardianRAGInstructions([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "platform_doc"},
	})
	if withDoc == base {
		t.Fatal("expected extra platform_doc guidance")
	}
	if len(withDoc) <= len(base) {
		t.Fatal("expected longer instructions with platform_doc chunks")
	}
}

func TestGuardianRAGInstructionsIncludesFieldGuideHint(t *testing.T) {
	base := GuardianRAGInstructions(nil)
	withField := GuardianRAGInstructions([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "field_guide"},
	})
	if withField == base {
		t.Fatal("expected extra field_guide guidance")
	}
	if !strings.Contains(withField, "field_guide") {
		t.Fatal("expected field_guide grounding text")
	}
}

func TestHasMultipleAlertChunks(t *testing.T) {
	if HasMultipleAlertChunks(nil) {
		t.Fatal("expected false for nil")
	}
	if HasMultipleAlertChunks([]db.SearchRagNearestNeighborsFilteredRow{{SourceType: "alert_notification"}}) {
		t.Fatal("expected false for single alert")
	}
	if !HasMultipleAlertChunks([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "alert_notification"},
		{SourceType: "alert_notification"},
	}) {
		t.Fatal("expected true for 2+ alerts")
	}
}

func TestGuardianRAGInstructionsIncludesAlertCitationDiscipline(t *testing.T) {
	base := GuardianRAGInstructions([]db.SearchRagNearestNeighborsFilteredRow{{SourceType: "platform_doc"}})
	withAlerts := GuardianRAGInstructions([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "alert_notification"},
		{SourceType: "alert_notification"},
		{SourceType: "platform_doc"},
	})
	if !strings.Contains(withAlerts, "exactly one [n] citation per numbered list item") {
		t.Fatal("expected one-cite-per-item instruction")
	}
	if strings.Contains(base, "most severe to least severe") {
		t.Fatal("did not expect alert discipline instruction without 2+ alerts")
	}
}

func TestGuardianRAGInstructionsIncludesAlertOnlyDiscipline(t *testing.T) {
	withAlertsOnly := GuardianRAGInstructions([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "alert_notification"},
		{SourceType: "alert_notification"},
		{SourceType: "alert_notification"},
	})
	if !strings.Contains(withAlertsOnly, "only alert_notification rows") {
		t.Fatal("expected alert-only discipline when all chunks are alerts")
	}
	withMixed := GuardianRAGInstructions([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "alert_notification"},
		{SourceType: "alert_notification"},
		{SourceType: "platform_doc"},
	})
	if strings.Contains(withMixed, "only alert_notification rows") {
		t.Fatal("did not expect alert-only discipline with mixed chunks")
	}
}

func TestHasOnlyAlertChunks(t *testing.T) {
	if HasOnlyAlertChunks(nil) {
		t.Fatal("expected false for nil")
	}
	if !HasOnlyAlertChunks([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "alert_notification"},
		{SourceType: "alert_notification"},
	}) {
		t.Fatal("expected true for alert-only")
	}
	if HasOnlyAlertChunks([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "alert_notification"},
		{SourceType: "platform_doc"},
	}) {
		t.Fatal("expected false for mixed")
	}
}

func TestGuardianRAGInstructionsIncludesAlertCitationDiscipline_tail(t *testing.T) {
	base := GuardianRAGInstructions([]db.SearchRagNearestNeighborsFilteredRow{{SourceType: "platform_doc"}})
	withAlerts := GuardianRAGInstructions([]db.SearchRagNearestNeighborsFilteredRow{
		{SourceType: "alert_notification"},
		{SourceType: "alert_notification"},
		{SourceType: "platform_doc"},
	})
	if !strings.Contains(withAlerts, "most severe to least severe") {
		t.Fatal("expected alert citation discipline instruction")
	}
	if !strings.Contains(withAlerts, "LIVE FARM STATE") {
		t.Fatal("expected LIVE FARM STATE context-only instruction")
	}
	if !strings.Contains(withAlerts, "Do not use markdown links") {
		t.Fatal("expected no-markdown-links instruction")
	}
	if strings.Contains(base, "most severe to least severe") {
		t.Fatal("did not expect alert discipline instruction without 2+ alerts")
	}
}

func TestZeroChunkGuardBlock(t *testing.T) {
	block := ZeroChunkGuardBlock()
	if !strings.Contains(block, "0 RAG chunks") || !strings.Contains(block, "Do NOT use [n]") {
		t.Fatalf("unexpected block: %q", block)
	}
}
