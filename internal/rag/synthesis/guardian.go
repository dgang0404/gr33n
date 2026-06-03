package synthesis

import (
	"strings"

	db "gr33n-api/internal/db"
)

const platformDocGrounding = `When numbered sources include type=platform_doc, treat them as authoritative gr33n operator documentation for how-to, troubleshooting, Guardian PR workflows, Pi setup, and UI navigation. Prefer citing those sources for procedural questions. For live farm state (current humidity, active cycles, unread alerts, plant catalog right now), rely on the LIVE FARM STATE block and read-tool results in the system prompt — never invent sensor values or row counts from documentation alone.`

const fieldGuideGrounding = `When numbered sources include type=field_guide, treat them as authoritative physical install and trades guidance (Pi GPIO/relay/sensor wiring, irrigation basics, electrical safety boundaries). Prefer field_guide + platform_doc for wiring, plumbing, and field troubleshooting questions. Respect safety_tier in source metadata: never give step-by-step mains AC or pressurized/potable water instructions — stop and tell the operator to use a qualified electrician or plumber. Offer to start a guided procedure when the operator needs hands-on steps. Guardian cannot see the wiring; ask the worker to confirm what they observe (operator-stated facts are labeled, not measured).`

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

// GuardianRAGInstructions returns synthesis citation rules plus corpus-specific guidance when relevant.
func GuardianRAGInstructions(chunks []db.SearchRagNearestNeighborsFilteredRow) string {
	out := SystemPrompt()
	if HasFieldGuideChunks(chunks) {
		out += "\n\n" + fieldGuideGrounding
	}
	if HasPlatformDocChunks(chunks) {
		out += "\n\n" + platformDocGrounding
	}
	return out
}
