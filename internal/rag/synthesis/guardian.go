package synthesis

import (
	"strings"

	db "gr33n-api/internal/db"
)

const platformDocGrounding = `When numbered sources include type=platform_doc, treat them as authoritative gr33n operator documentation for how-to, troubleshooting, Guardian PR workflows, Pi setup, and UI navigation. Prefer citing those sources for procedural questions. For live farm state (current humidity, active cycles, unread alerts, plant catalog right now), rely on the LIVE FARM STATE block and read-tool results in the system prompt — never invent sensor values or row counts from documentation alone.`

const fieldGuideGrounding = `When numbered sources include type=field_guide, treat them as authoritative physical install and trades guidance (Pi GPIO/relay/sensor wiring, irrigation basics, electrical safety boundaries). Prefer field_guide + platform_doc for wiring, plumbing, and field troubleshooting questions. Respect safety_tier in source metadata: never give step-by-step mains AC or pressurized/potable water instructions — stop and tell the operator to use a qualified electrician or plumber. Offer to start a guided procedure when the operator needs hands-on steps. Guardian can look up platform wiring (GPIO pin, relay channel, device assignment, reading freshness) via summarize_device_health. When wiring looks correct in the platform but the operator reports wrong behaviour, ask them to verify the physical connection matches the platform record.`

// HasPlatformDocChunks reports whether retrieval returned platform_doc sources.
func HasPlatformDocChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) bool {
	return hasSourceType(chunks, "platform_doc")
}

// HasFieldGuideChunks reports whether retrieval returned field_guide sources.
func HasFieldGuideChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) bool {
	return hasSourceType(chunks, "field_guide")
}

func hasSourceType(chunks []db.SearchRagNearestNeighborsFilteredRow, want string) bool {
	for _, ch := range chunks {
		if strings.EqualFold(strings.TrimSpace(ch.SourceType), want) {
			return true
		}
	}
	return false
}

const operationalNoteGrounding = `The following numbered sources marked as farm notes or operational rows may be outdated — prefer LIVE FARM DATA (snapshot and read tools) for current sensor values, alert counts, and zone state.`

// alertCitationDiscipline is Phase 149/151's fix for run #6's mislabeled alert
// citations and run #8's markdown-link skip: sources are pre-sorted
// most-severe-first, and LIVE FARM STATE is context only — each list item gets
// exactly one [n] matching its position (run #9: stray [3] on item 2).
const alertCitationDiscipline = `Multiple alert sources are listed below, ordered most severe to least severe. LIVE FARM STATE and list_unread_alerts blocks in the system prompt are context only — do not treat them as citations and do not invent markdown links or URLs for alerts.

When you list alerts:
- Use exactly the source order: list item 1 must cite [1] (same number as its position in the Sources list), item 2 must cite [2], and so on.
- Use exactly one [n] citation per numbered list item — item N must have only [N], never a second [n] on the same item.
- Place that single [n] at the end of the alert summary for that item.
- Do not cite platform_doc, field_guide, or documentation that is not in the Sources list.
- Do not use markdown links, invented URLs, or uncited alert summaries.
- Do not repeat the same alert under a second number, and do not renumber or reorder alerts.`

// alertOnlyCitationDiscipline applies when numbered Sources are exclusively
// alert_notification rows (Phase 151 WS4 filter) — removes run #9's "platform
// docs [3]" hallucination on an alert-only source list.
const alertOnlyCitationDiscipline = `The Sources list below contains only alert_notification rows. Every [n] you write must refer to one of those alert sources only. Do not mention or cite platform documentation, field guides, or operator bootstrap material with a [n] bracket — those sources are not in the list.`

// HasOperationalChunks reports retrieval rows that are not curated guides/docs.
func HasOperationalChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) bool {
	for _, ch := range chunks {
		st := strings.ToLower(strings.TrimSpace(ch.SourceType))
		switch st {
		case "field_guide", "platform_doc", "symptom_guide":
			continue
		default:
			if st != "" {
				return true
			}
		}
	}
	return false
}

// HasMultipleAlertChunks reports whether retrieval returned 2+ alert_notification sources.
func HasMultipleAlertChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) bool {
	n := 0
	for _, ch := range chunks {
		if strings.EqualFold(strings.TrimSpace(ch.SourceType), "alert_notification") {
			n++
			if n >= 2 {
				return true
			}
		}
	}
	return false
}

// HasOnlyAlertChunks reports whether every numbered source is alert_notification.
func HasOnlyAlertChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) bool {
	if len(chunks) < 2 {
		return false
	}
	for _, ch := range chunks {
		if !strings.EqualFold(strings.TrimSpace(ch.SourceType), "alert_notification") {
			return false
		}
	}
	return true
}

// GuardianRAGInstructions returns synthesis citation rules plus corpus-specific guidance when relevant.
// Call only when len(chunks) > 0 — use ZeroChunkGuardBlock for empty retrieval.
func GuardianRAGInstructions(chunks []db.SearchRagNearestNeighborsFilteredRow) string {
	out := SystemPrompt()
	if HasFieldGuideChunks(chunks) {
		out += "\n\n" + fieldGuideGrounding
	}
	if HasPlatformDocChunks(chunks) {
		out += "\n\n" + platformDocGrounding
	}
	if HasOperationalChunks(chunks) {
		out += "\n\n" + operationalNoteGrounding
	}
	if HasMultipleAlertChunks(chunks) {
		out += "\n\n" + alertCitationDiscipline
	}
	if HasOnlyAlertChunks(chunks) {
		out += "\n\n" + alertOnlyCitationDiscipline
	}
	return out
}

const zeroChunkGuardBlock = `No indexed documentation matched this question (0 RAG chunks).
- Do NOT use [n] citation brackets.
- Do NOT state EC, pH, VPD, DLI, or photoperiod numbers unless lookup_crop_targets results appear above in this system prompt.
- For each crop mentioned: if lookup_crop_targets returned a profile, use those mS/cm values; if not, say you have no built-in profile and offer Start grow / Plants.
- For crops outside gr33n support (e.g. woodland ephemerals), say so plainly.`

// ZeroChunkGuardBlock is injected when farm-grounded chat retrieved zero RAG chunks (Phase 82 WS1).
func ZeroChunkGuardBlock() string { return zeroChunkGuardBlock }
