package tools

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/cropcycle/opstimeline"
)

func execListCropCycleOps(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	cycleID, err := int64FromArgs(args, "crop_cycle_id")
	if err != nil {
		return nil, err
	}
	cc, err := deps.Q.GetCropCycleByID(ctx, cycleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("crop cycle %d not found", cycleID)
		}
		return nil, err
	}
	if err := ensureFarmScope(cc.FarmID, deps.FarmID); err != nil {
		return nil, err
	}

	from, to := opstimeline.DefaultRange(cc, time.Now().UTC())
	if raw, err := optionalStringFromArgs(args, "from"); err != nil {
		return nil, err
	} else if raw != nil {
		t, err := opstimeline.ParseTimeQuery(*raw)
		if err != nil {
			return nil, fmt.Errorf("from must be RFC3339 or YYYY-MM-DD")
		}
		from = t
	}
	if raw, err := optionalStringFromArgs(args, "to"); err != nil {
		return nil, err
	} else if raw != nil {
		t, err := opstimeline.ParseTimeQuery(*raw)
		if err != nil {
			return nil, fmt.Errorf("to must be RFC3339 or YYYY-MM-DD")
		}
		to = t
		if len(*raw) == len("2006-01-02") {
			to = to.Add(24*time.Hour - time.Nanosecond)
		}
	}
	if to.Before(from) {
		return nil, errors.New("to must be on or after from")
	}

	return opstimeline.Build(ctx, deps.Q, cc, from, to)
}
