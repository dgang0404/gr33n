// Phase 28 WS1 — crop cycle analytics. Hand-written Go bindings to avoid a
// repository-wide sqlc regen (kept the same pattern as Phase 27's
// conversation_turns additions). When the next routine sqlc pass happens,
// these queries will fold back into crop_cycles.sql.go cleanly because
// the SQL definitions live alongside the rest in db/queries/crop_cycles.sql.

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getFertigationAggregatesByCropCycle = `-- name: GetFertigationAggregatesByCropCycle :one
SELECT
    COUNT(*)::bigint                                                  AS event_count,
    COALESCE(SUM(volume_applied_liters), 0)::numeric                  AS total_liters,
    COALESCE(AVG(ec_after_mscm), 0)::numeric                          AS avg_ec_mscm,
    COALESCE(MIN(ec_after_mscm), 0)::numeric                          AS min_ec_mscm,
    COALESCE(MAX(ec_after_mscm), 0)::numeric                          AS max_ec_mscm,
    COALESCE(AVG((COALESCE(ph_before,0) + COALESCE(ph_after,0)) / NULLIF(
        ((CASE WHEN ph_before IS NULL THEN 0 ELSE 1 END) +
         (CASE WHEN ph_after  IS NULL THEN 0 ELSE 1 END)), 0
    )), 0)::numeric                                                   AS avg_ph
FROM gr33nfertigation.fertigation_events
WHERE crop_cycle_id = $1
`

// FertigationAggregates rolls up every fertigation_event linked to a single
// crop_cycle_id into the shape the cycle-summary endpoint returns. All
// numeric fields are COALESCEd to zero so consumers never have to handle
// SQL NULLs — when a cycle has zero events, every field is 0 and
// event_count makes the empty case obvious to the UI.
type FertigationAggregates struct {
	EventCount  int64          `db:"event_count" json:"event_count"`
	TotalLiters pgtype.Numeric `db:"total_liters" json:"total_liters"`
	AvgECmSCm   pgtype.Numeric `db:"avg_ec_mscm" json:"avg_ec_mscm"`
	MinECmSCm   pgtype.Numeric `db:"min_ec_mscm" json:"min_ec_mscm"`
	MaxECmSCm   pgtype.Numeric `db:"max_ec_mscm" json:"max_ec_mscm"`
	AvgPH       pgtype.Numeric `db:"avg_ph" json:"avg_ph"`
}

// GetFertigationAggregatesByCropCycle returns rolling fertigation stats for
// the cycle-summary endpoint. EC after-feed is the canonical "what the
// plants actually experienced" reading; pH average blends pre + post so
// the number reflects the working solution, not just the freshly-mixed
// batch.
func (q *Queries) GetFertigationAggregatesByCropCycle(ctx context.Context, cropCycleID int64) (FertigationAggregates, error) {
	var a FertigationAggregates
	err := q.db.QueryRow(ctx, getFertigationAggregatesByCropCycle, cropCycleID).Scan(
		&a.EventCount,
		&a.TotalLiters,
		&a.AvgECmSCm,
		&a.MinECmSCm,
		&a.MaxECmSCm,
		&a.AvgPH,
	)
	return a, err
}
