// Phase 152 WS2 — citation deep links. Mirrors ContextRef's page→Guardian
// mapping (context_ref.go) in the opposite direction: given a citation's
// source_type + source_id, resolve the UI route an operator can click
// through to (their zone, crop cycle, etc.) so the sidebar becomes a way to
// navigate from an answer, not just read it.

package farmguardian

import (
	"context"
	"strconv"

	db "gr33n-api/internal/db"
)

// ResolveCitationRoute maps a citation to a UI path, or ok=false when the
// source type has no route yet, the row can't be found, or it belongs to a
// different farm (defense in depth beyond the already farm-scoped RAG
// retrieval — a citation should never be able to route a click into another
// farm's data).
//
// Scoped to source types with a single, always-reliable FK to a zone or a
// dedicated summary page. Source types without a direct FK (schedule has no
// zone_id of its own; alert_notification only has an indirect
// triggering_event_source_type/id) are left unresolved for now — see Phase
// 152 plan WS2b.
func ResolveCitationRoute(ctx context.Context, q *db.Queries, farmID int64, sourceType string, sourceID int64) (string, bool) {
	if q == nil || farmID <= 0 || sourceID <= 0 {
		return "", false
	}
	switch sourceType {
	case "crop_cycle":
		c, err := q.GetCropCycleByID(ctx, sourceID)
		if err != nil || c.FarmID != farmID {
			return "", false
		}
		return "/crop-cycles/" + strconv.FormatInt(sourceID, 10) + "/summary", true
	case "fertigation_program":
		p, err := q.GetFertigationProgramByID(ctx, sourceID)
		if err != nil || p.FarmID != farmID || p.TargetZoneID == nil || *p.TargetZoneID <= 0 {
			return "", false
		}
		return "/zones/" + strconv.FormatInt(*p.TargetZoneID, 10) + "?tab=water", true
	case "task":
		t, err := q.GetTaskByID(ctx, sourceID)
		if err != nil || t.FarmID != farmID || t.ZoneID == nil || *t.ZoneID <= 0 {
			return "", false
		}
		return "/zones/" + strconv.FormatInt(*t.ZoneID, 10), true
	default:
		return "", false
	}
}
