package farmguardian

import (
	"strings"
	"testing"

	db "gr33n-api/internal/db"
)

func TestNudgeContextBlock(t *testing.T) {
	block := NudgeContextBlock(ContextRef{
		NudgeCategory: "critical_alert",
		NudgeID:       "alert-99",
	})
	if !strings.Contains(block, "critical_alert") {
		t.Fatalf("missing category: %q", block)
	}
	if !strings.Contains(block, "alert-99") {
		t.Fatalf("missing nudge_id: %q", block)
	}
	if !strings.Contains(block, "Skip pleasantries") {
		t.Fatalf("missing framing: %q", block)
	}
}

func TestAlertSeverityWarnOrHigher(t *testing.T) {
	low := db.Gr33ncoreNotificationPriorityEnumLow
	if alertSeverityWarnOrHigher(&low) {
		t.Fatal("low should not qualify")
	}
	if !alertSeverityWarnOrHigher(nil) {
		t.Fatal("nil severity should qualify as warn+")
	}
}
