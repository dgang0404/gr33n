package chat

import (
	"context"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func farmGroundedTimeout(ctx context.Context, q *db.Queries, farmID int64, grounded bool) time.Duration {
	if !grounded || farmID <= 0 || q == nil {
		return 0
	}
	farm, err := q.GetFarmByID(ctx, farmID)
	if err != nil {
		return 0
	}
	return farmguardian.FarmGroundedChatTimeout(&farm)
}
