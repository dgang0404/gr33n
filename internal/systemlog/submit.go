package systemlog

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/authctx"
	commontypes "gr33n-api/internal/platform/commontypes"
	db "gr33n-api/internal/db"
)

// Submit writes a system log row; failures are logged and do not affect callers.
func Submit(ctx context.Context, q db.Querier, farmID *int64, level commontypes.LogLevelEnum, source, message string, contextData map[string]any) {
	if err := SubmitErr(ctx, q, farmID, level, source, message, contextData); err != nil {
		log.Printf("system log: %v", err)
	}
}

func SubmitErr(ctx context.Context, q db.Querier, farmID *int64, level commontypes.LogLevelEnum, source, message string, contextData map[string]any) error {
	var uid pgtype.UUID
	if u, ok := authctx.UserID(ctx); ok {
		uid = pgtype.UUID{Bytes: u, Valid: true}
	}
	ctxJSON := []byte("{}")
	if contextData != nil {
		if b, err := json.Marshal(contextData); err == nil {
			ctxJSON = b
		}
	}
	return q.InsertSystemLog(ctx, db.InsertSystemLogParams{
		FarmID:          farmID,
		UserID:          uid,
		LogLevel:        level,
		EventType:       nil,
		Message:         message,
		SourceComponent: &source,
		Column7:         ctxJSON,
	})
}

func FarmIDPtr(id int64) *int64 {
	v := id
	return &v
}
