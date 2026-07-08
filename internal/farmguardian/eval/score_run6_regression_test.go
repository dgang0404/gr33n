// Phase 148 — regression against the actual smoke run #6 archive
// (data/guardian_qa_runs/20260708T153829_smoke_phi3-mini.json) showing that
// the new accuracy detectors now catch the unread-alerts citation mismatch
// that Phase 145's topical drift scorer missed (it only checked field_guide).

package eval

import (
	"testing"

	"gr33n-api/internal/farmguardian"
)

func TestSmokeTopicDrift_runSixUnreadAlertsCitationMismatchNowCaught(t *testing.T) {
	t.Parallel()
	answer := `1. **High humidity alert in the Flower Room** [3] - The current sensor reading indicates a high level of humidity at 72.4% RH, which is above your threshold for late flower growth stages (65%). To address this:
   - Consider increasing airflow or using dehumidifiers to lower moisture levels in the Flower Room before it affects plant health adversely.

2. **Low OHN batch below minimum** [4] - Your Oriental Herbal Nutrient (OHN) seed stock is running low, as the current amount stands at 0.35 L against a threshold of 0sourced from FIELD GUIDE [1].

3. **Light schedule change alert** [5] - A photoperiod transition is due within 48 hours from the current time in your Flower Room as per Alert Notification source_id=7 and severity of low.

4. OHN batch below minimum - As per FIELD GUIDE [1], ensure that if not already confirmed or acted upon by acknowledging the alert and then confirm any necessary changes in Inventory settings as mentioned above for OHN replenishment.`

	cites := []farmguardian.CitationSummary{
		{Ref: 1, SourceType: "platform_doc", Excerpt: "workflow-guide.md — Alerts are automatically generated when a sensor reading crosses a configured threshold."},
		{Ref: 2, SourceType: "platform_doc", Excerpt: "local-operator-bootstrap.md — acknowledge the humidity alert proposal card Confirm."},
		{Ref: 3, SourceType: "alert_notification", Excerpt: "severity: low\nsubject: Light schedule change in 48 hours — Flower Room\nmessage: Photoperiod transition reminder."},
		{Ref: 4, SourceType: "alert_notification", Excerpt: "severity: medium\nsubject: OHN batch below minimum — reorder or brew soon\nmessage: Batch SEED-OHN-001 has 0.35 L remaining (threshold 0.5 L)."},
		{Ref: 5, SourceType: "alert_notification", Excerpt: "severity: high\nsubject: Humidity high — Flower Room\nmessage: Air Humidity Indoor read 72.4% RH (alert threshold 65% for late flower)."},
	}

	note := farmguardian.SmokeTopicDriftNote(farmguardian.SmokeTopicDriftInput{
		QuestionID: "smoke-unread-alerts",
		Category:   "farm_state",
		Prompt:     "Summarize my unread alerts and what I should do about each one.",
		Answer:     answer,
		Citations:  cites,
	})
	if note == "" {
		t.Fatal("expected Phase 148 accuracy detector to flag the run #6 alert answer (garbled token, duplicate item, or citation mismatch)")
	}
	t.Logf("caught: %s", note)
}
