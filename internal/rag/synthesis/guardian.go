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
	return out
}

const zeroChunkGuardBlock = `No indexed documentation matched this question (0 RAG chunks).
- Do NOT use [n] citation brackets.
- Do NOT state EC, pH, VPD, DLI, or photoperiod numbers unless lookup_crop_targets results appear above in this system prompt.
- For each crop mentioned: if lookup_crop_targets returned a profile, use those mS/cm values; if not, say you have no built-in profile and offer Start grow / Plants.
- For crops outside gr33n support (e.g. woodland ephemerals), say so plainly.`

// ZeroChunkGuardBlock is injected when farm-grounded chat retrieved zero RAG chunks (Phase 82 WS1).
func ZeroChunkGuardBlock() string { return zeroChunkGuardBlock }
