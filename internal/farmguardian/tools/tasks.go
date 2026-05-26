package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

func execCreateTaskFromAlert(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	alertID, err := alertIDFromArgs(args)
	if err != nil {
		return nil, err
	}
	overrides := taskFromAlertOverrides(args)
	return createTaskFromAlertRow(ctx, deps, alertID, overrides)
}

func execCreateTask(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	title, err := stringFromArgs(args, "title")
	if err != nil {
		return nil, err
	}
	desc, err := optionalStringFromArgs(args, "description")
	if err != nil {
		return nil, err
	}
	zoneID, err := optionalInt64FromArgs(args, "zone_id")
	if err != nil {
		return nil, err
	}
	if zoneID != nil {
		z, err := deps.Q.GetZoneByID(ctx, *zoneID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("zone %d not found", *zoneID)
			}
			return nil, err
		}
		if err := ensureFarmScope(z.FarmID, deps.FarmID); err != nil {
			return nil, err
		}
	}
	priority := int32(1)
	if p, err := optionalInt32FromArgs(args, "priority"); err != nil {
		return nil, err
	} else if p != nil {
		priority = *p
		if priority < 0 || priority > 3 {
			return nil, errors.New("priority must be 0–3")
		}
	}
	taskType, err := optionalStringFromArgs(args, "task_type")
	if err != nil {
		return nil, err
	}
	tt := "general"
	if taskType != nil {
		tt = *taskType
	}
	var createdBy pgtype.UUID
	if deps.HasUser {
		createdBy = pgtype.UUID{Bytes: deps.UserID, Valid: true}
	}
	row, err := deps.Q.CreateTask(ctx, db.CreateTaskParams{
		FarmID:           deps.FarmID,
		ZoneID:           zoneID,
		Title:            title,
		Description:      desc,
		TaskType:         &tt,
		Status:           commontypes.TaskStatusEnum("todo"),
		Priority:         &priority,
		CreatedByUserID:  createdBy,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"task_id": row.ID, "title": row.Title}, nil
}

type taskFromAlertOpts struct {
	Title       *string
	Description *string
	ZoneID      *int64
	Priority    *int32
}

func taskFromAlertOverrides(args map[string]any) taskFromAlertOpts {
	var o taskFromAlertOpts
	if t, err := optionalStringFromArgs(args, "title"); err == nil {
		o.Title = t
	}
	if d, err := optionalStringFromArgs(args, "description"); err == nil {
		o.Description = d
	}
	if z, err := optionalInt64FromArgs(args, "zone_id"); err == nil {
		o.ZoneID = z
	}
	if p, err := optionalInt32FromArgs(args, "priority"); err == nil {
		o.Priority = p
	}
	return o
}

func createTaskFromAlertRow(ctx context.Context, deps ExecutorDeps, alertID int64, overrides taskFromAlertOpts) (any, error) {
	alertRow, err := deps.Q.GetAlertNotificationByID(ctx, alertID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("alert %d not found", alertID)
		}
		return nil, err
	}
	if err := ensureFarmScope(alertRow.FarmID, deps.FarmID); err != nil {
		return nil, err
	}

	title := ""
	if overrides.Title != nil {
		title = strings.TrimSpace(*overrides.Title)
	}
	if title == "" && alertRow.SubjectRendered != nil {
		title = strings.TrimSpace(*alertRow.SubjectRendered)
	}
	if title == "" {
		title = fmt.Sprintf("Follow up on alert #%d", alertRow.ID)
	}

	var description *string
	if overrides.Description != nil {
		description = overrides.Description
	} else if alertRow.MessageTextRendered != nil && strings.TrimSpace(*alertRow.MessageTextRendered) != "" {
		msg := *alertRow.MessageTextRendered
		description = &msg
	}

	priority := int32(1)
	if overrides.Priority != nil {
		priority = *overrides.Priority
	} else if alertRow.Severity.Valid {
		switch alertRow.Severity.Gr33ncoreNotificationPriorityEnum {
		case "critical":
			priority = 3
		case "high":
			priority = 2
		case "medium":
			priority = 1
		case "low":
			priority = 0
		}
	}

	zoneID := overrides.ZoneID
	if zoneID == nil &&
		alertRow.TriggeringEventSourceType != nil &&
		*alertRow.TriggeringEventSourceType == "sensor_reading" &&
		alertRow.TriggeringEventSourceID != nil {
		if sensor, err := deps.Q.GetSensorByID(ctx, *alertRow.TriggeringEventSourceID); err == nil {
			if sensor.FarmID == alertRow.FarmID && sensor.ZoneID != nil {
				zid := *sensor.ZoneID
				zoneID = &zid
			}
		}
	}

	tt := "alert_follow_up"
	var createdBy pgtype.UUID
	if deps.HasUser {
		createdBy = pgtype.UUID{Bytes: deps.UserID, Valid: true}
	}
	aid := alertRow.ID
	row, err := deps.Q.CreateTask(ctx, db.CreateTaskParams{
		FarmID:          alertRow.FarmID,
		ZoneID:          zoneID,
		Title:           title,
		Description:     description,
		TaskType:        &tt,
		Status:          commontypes.TaskStatusEnum("todo"),
		Priority:        &priority,
		SourceAlertID:   &aid,
		CreatedByUserID: createdBy,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"task_id": row.ID, "alert_id": alertID, "title": row.Title}, nil
}

func optionalInt32FromArgs(args map[string]any, key string) (*int32, error) {
	n, err := optionalInt64FromArgs(args, key)
	if err != nil || n == nil {
		return nil, err
	}
	v := int32(*n)
	return &v, nil
}
