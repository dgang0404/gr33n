// Phase 67 WS5 — crop-profile grounding for field photo diagnosis.

package farmguardian

import (
	"context"
	"fmt"
	"strings"

	db "gr33n-api/internal/db"
)

// FieldPhotoCropGroundingBlock injects crop targets when vision attachments are present.
func FieldPhotoCropGroundingBlock(ctx context.Context, q db.Querier, farmID int64, ref *ContextRef) string {
	if q == nil || farmID <= 0 || ref == nil {
		return ""
	}
	hasZone := strings.EqualFold(ref.Type, "zone") && ref.ID > 0
	if !hasZone && ref.CropCycleID <= 0 {
		return ""
	}
	block, err := renderLookupCropTargets(ctx, q, farmID, "field photo diagnosis crop targets", ref)
	if err != nil || block == "" {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf(`
Field photo diagnosis (Phase 67):
- Ground visual hypotheses on the crop profile below — cite EC/pH/VPD/DLI targets from this block only.
- Cross-check against live zone readings in the snapshot when discussing lockout or environment stress.
- Wiring or Pi questions: prefer summarize_device_health (Phase 65).

%s`, block))
}
