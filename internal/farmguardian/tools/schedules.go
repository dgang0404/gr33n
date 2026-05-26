package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

func execPatchSchedule(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	scheduleID, err := int64FromArgs(args, "schedule_id")
	if err != nil {
		return nil, err
	}
	if len(args) <= 1 {
		return nil, errors.New("at least one patch field required")
	}
	sch, err := deps.Q.GetScheduleByID(ctx, scheduleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("schedule %d not found", scheduleID)
		}
		return nil, err
	}
	if err := ensureFarmScope(sch.FarmID, deps.FarmID); err != nil {
		return nil, err
	}

	name := sch.Name
	if v, err := optionalStringFromArgs(args, "name"); err != nil {
		return nil, err
	} else if v != nil {
		name = *v
	}
	cron := sch.CronExpression
	if v, err := optionalStringFromArgs(args, "cron_expression"); err != nil {
		return nil, err
	} else if v != nil {
		cron = *v
	}
	isActive := sch.IsActive
	if v, err := optionalBoolFromArgs(args, "is_active"); err != nil {
		return nil, err
	} else if v != nil {
		isActive = *v
	}
	if name == "" || cron == "" {
		return nil, errors.New("name and cron_expression cannot be empty")
	}
	preconds := sch.Preconditions
	if len(preconds) == 0 {
		preconds = []byte("[]")
	}
	meta := sch.MetaData
	if len(meta) == 0 {
		meta = []byte("{}")
	}
	row, err := deps.Q.UpdateSchedule(ctx, db.UpdateScheduleParams{
		ID:             scheduleID,
		Name:           name,
		Description:    sch.Description,
		ScheduleType:   sch.ScheduleType,
		CronExpression: cron,
		Timezone:       sch.Timezone,
		IsActive:       isActive,
		MetaData:       meta,
		Preconditions:  preconds,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"schedule_id":     row.ID,
		"name":            row.Name,
		"cron_expression": row.CronExpression,
		"is_active":       row.IsActive,
	}, nil
}
