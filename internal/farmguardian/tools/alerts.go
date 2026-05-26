package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

func alertIDFromArgs(args map[string]any) (int64, error) {
	raw, ok := args["alert_id"]
	if !ok {
		return 0, errors.New("alert_id required")
	}
	switch v := raw.(type) {
	case float64:
		id := int64(v)
		if id <= 0 {
			return 0, errors.New("invalid alert_id")
		}
		return id, nil
	case int64:
		if v <= 0 {
			return 0, errors.New("invalid alert_id")
		}
		return v, nil
	case int:
		if v <= 0 {
			return 0, errors.New("invalid alert_id")
		}
		return int64(v), nil
	default:
		return 0, errors.New("invalid alert_id type")
	}
}

func execMarkAlertRead(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	alertID, err := alertIDFromArgs(args)
	if err != nil {
		return nil, err
	}
	a0, err := deps.Q.GetAlertNotificationByID(ctx, alertID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("alert %d not found", alertID)
		}
		return nil, err
	}
	if deps.FarmID > 0 && a0.FarmID != deps.FarmID {
		return nil, errors.New("alert is outside proposal farm scope")
	}
	row, err := deps.Q.MarkAlertRead(ctx, alertID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"alert_id": row.ID, "is_read": row.IsRead}, nil
}

func execAckAlert(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	alertID, err := alertIDFromArgs(args)
	if err != nil {
		return nil, err
	}
	a0, err := deps.Q.GetAlertNotificationByID(ctx, alertID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("alert %d not found", alertID)
		}
		return nil, err
	}
	if deps.FarmID > 0 && a0.FarmID != deps.FarmID {
		return nil, errors.New("alert is outside proposal farm scope")
	}
	uid := pgtype.UUID{Bytes: deps.UserID, Valid: true}
	row, err := deps.Q.MarkAlertAcknowledged(ctx, db.MarkAlertAcknowledgedParams{
		ID:                   alertID,
		AcknowledgedByUserID: uid,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"alert_id":         row.ID,
		"is_acknowledged":  row.IsAcknowledged,
		"acknowledged_at":  row.AcknowledgedAt,
	}, nil
}
