package synthesis

import (
	"strings"

	db "gr33n-api/internal/db"
)

const platformDocGrounding = `When numbered sources include type=platform_doc, treat them as authoritative gr33n operator documentation for how-to, troubleshooting, Guardian PR workflows, Pi setup, and UI navigation. Prefer citing those sources for procedural questions. For live farm state (current humidity, active cycles, unread alerts, plant catalog right now), rely on the LIVE FARM STATE block and read-tool results in the system prompt — never invent sensor values or row counts from documentation alone.`

// HasPlatformDocChunks reports whether retrieval returned platform_doc sources.
func HasPlatformDocChunks(chunks []db.SearchRagNearestNeighborsFilteredRow) bool {
	for _, ch := range chunks {
		if strings.EqualFold(strings.TrimSpace(ch.SourceType), "platform_doc") {
			return true
		}
	}
	return false
}

// GuardianRAGInstructions returns synthesis citation rules plus platform-doc guidance when relevant.
func GuardianRAGInstructions(chunks []db.SearchRagNearestNeighborsFilteredRow) string {
	out := SystemPrompt()
	if HasPlatformDocChunks(chunks) {
		out += "\n\n" + platformDocGrounding
	}
	return out
}
