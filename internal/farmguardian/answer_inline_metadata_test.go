package farmguardian

import (
	"strings"
	"testing"
)

// Phase 189 — live turn: "Acknowledge the highest severity unread alert."
const liveInlineSourceIDLeak = `The highest severity unread alert is a "High Humidity Alert". According to our field guide [5] (field_guide source id=8, chunk id=66), powdery mildew can be a likely cause. Improve airflow as recommended by our field guide [5] (field_guide source id=10).`

func TestRedactInlineSourceMetadata_liveHighHumidityAlert(t *testing.T) {
	t.Parallel()
	got, meta := RedactInlineSourceMetadata(liveInlineSourceIDLeak)
	if !meta.Redacted {
		t.Fatal("expected inline metadata redaction")
	}
	if meta.Occurrences != 2 {
		t.Fatalf("occurrences = %d, want 2", meta.Occurrences)
	}
	if strings.Contains(got, "source id") || strings.Contains(got, "chunk id") {
		t.Fatalf("inline metadata still present: %q", got)
	}
	if !strings.Contains(got, "High Humidity Alert") || !strings.Contains(got, "field guide [5]") {
		t.Fatalf("expected farm content preserved: %q", got)
	}
	if AnswerContainsInlineSourceMetadata(got) {
		t.Fatal("AnswerContainsInlineSourceMetadata should be false after redaction")
	}
}

// Phase 189 — live turn: "Pause the lights schedule for Veg Tent until tomorrow."
const liveInlineSourceUnderscoreLeak = `To pause the lights, note the cucumbers' cool-season nature from field_guide source_id=17 chunk_id=18. Avoid guessing EC values for unsupported crops mentioned in source_id=53.`

func TestRedactInlineSourceMetadata_liveVegTentPause(t *testing.T) {
	t.Parallel()
	got, meta := RedactInlineSourceMetadata(liveInlineSourceUnderscoreLeak)
	if !meta.Redacted {
		t.Fatal("expected inline metadata redaction")
	}
	if strings.Contains(got, "source_id") || strings.Contains(got, "chunk_id") {
		t.Fatalf("inline metadata still present: %q", got)
	}
	if !strings.Contains(got, "cool-season nature") {
		t.Fatalf("expected farm content preserved: %q", got)
	}
}

// Phase 189 — live turn: "Set the feed volume to 0.5 liters for the Veg Tent program."
const liveDocPathLeak = `This is likely growing microgreens based on their EC and substrate requirements as outlined [1] (field_guide: doc_path=field-guides/crop-microgreens-nutrition.md), so keep the volume low.`

func TestRedactInlineSourceMetadata_liveDocPathLeak(t *testing.T) {
	t.Parallel()
	got, meta := RedactInlineSourceMetadata(liveDocPathLeak)
	if !meta.Redacted {
		t.Fatal("expected inline metadata redaction")
	}
	if strings.Contains(got, "doc_path") {
		t.Fatalf("doc_path still present: %q", got)
	}
	if !strings.Contains(got, "microgreens") || !strings.Contains(got, "keep the volume low") {
		t.Fatalf("expected farm content preserved: %q", got)
	}
}

func TestRedactInlineSourceMetadata_noLeakUnchanged(t *testing.T) {
	t.Parallel()
	answer := "Check EC in the Veg Tent per field guide [1] before refilling calcium nitrate."
	got, meta := RedactInlineSourceMetadata(answer)
	if meta.Redacted || got != answer {
		t.Fatalf("unexpected redaction: meta=%+v got=%q", meta, got)
	}
}

// Phase 189 — live turn: "Summarize my unread alerts and what I should do about each one."
const liveLiteralPlaceholderLeak = `As per the field troubleshooting guide (source[n] citing source[3]), ensure the sensor wires are correctly connected to their respective pins.`

func TestRedactPlaceholderCitationMarkers_literalN(t *testing.T) {
	t.Parallel()
	got, meta := RedactPlaceholderCitationMarkers(liveLiteralPlaceholderLeak)
	if !meta.Redacted {
		t.Fatal("expected placeholder citation redaction")
	}
	if strings.Contains(strings.ToLower(got), "[n]") {
		t.Fatalf("literal [n] placeholder still present: %q", got)
	}
	if !strings.Contains(got, "[3]") {
		t.Fatalf("expected real citation [3] preserved: %q", got)
	}
	if !strings.Contains(got, "sensor wires are correctly connected") {
		t.Fatalf("expected farm content preserved: %q", got)
	}
}

// Phase 189 — live turn: "Create a task to refill calcium nitrate when stock is low."
const liveSourceColonDigitLeak = `Refer to source [1], which advises maintaining cooler temperatures for lettuce nutrition (source:[5]).`

func TestRedactPlaceholderCitationMarkers_sourceColonDigitNormalized(t *testing.T) {
	t.Parallel()
	got, meta := RedactPlaceholderCitationMarkers(liveSourceColonDigitLeak)
	if !meta.Redacted {
		t.Fatal("expected placeholder citation redaction")
	}
	if strings.Contains(got, "source:") || strings.Contains(got, "source [") {
		t.Fatalf("source-prefixed citation still present: %q", got)
	}
	if !strings.Contains(got, "[1]") || !strings.Contains(got, "[5]") {
		t.Fatalf("expected both citation numbers preserved: %q", got)
	}
}

func TestRedactPlaceholderCitationMarkers_noLeakUnchanged(t *testing.T) {
	t.Parallel()
	answer := "Lettuce runs low EC (~0.8–1.3 mS/cm) per [1]."
	got, meta := RedactPlaceholderCitationMarkers(answer)
	if meta.Redacted || got != answer {
		t.Fatalf("unexpected redaction: meta=%+v got=%q", meta, got)
	}
}
