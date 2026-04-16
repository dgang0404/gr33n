package auditlog

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
)

// Event describes one append-only audit row for a farm-scoped action.
type Event struct {
	FarmID int64
	Action db.Gr33ncoreUserActionTypeEnum
	// Optional row context (module.table / record id).
	TargetSchema, TargetTable, TargetRecordID, TargetDesc *string
	Status                                                string // success | failure | pending
	FailureReason                                         *string
	Details                                               map[string]any
}

// Submit writes an audit event; failures are logged and do not affect the caller's flow.
func Submit(ctx context.Context, q db.Querier, r *http.Request, ev Event) {
	if err := SubmitErr(ctx, q, r, ev); err != nil {
		log.Printf("audit log: %v", err)
	}
}

// SubmitErr returns an error if persistence fails (for tests or strict callers).
func SubmitErr(ctx context.Context, q db.Querier, r *http.Request, ev Event) error {
	var uid pgtype.UUID
	if u, ok := authctx.UserID(ctx); ok {
		uid = pgtype.UUID{Bytes: u, Valid: true}
	}
	detailsJSON := []byte("{}")
	if ev.Details != nil {
		b, err := json.Marshal(ev.Details)
		if err != nil {
			detailsJSON = []byte(`{"error":"audit_details_marshal_failed"}`)
		} else {
			detailsJSON = b
		}
	}
	st := strings.TrimSpace(ev.Status)
	if st == "" {
		st = "success"
	}
	var ua *string
	if r != nil {
		if s := strings.TrimSpace(r.UserAgent()); s != "" {
			ua = &s
		}
	}
	return q.InsertUserActivityLog(ctx, db.InsertUserActivityLogParams{
		UserID:                  uid,
		FarmID:                  &ev.FarmID,
		ActionType:              ev.Action,
		TargetModuleSchema:      ev.TargetSchema,
		TargetTableName:         ev.TargetTable,
		TargetRecordID:          ev.TargetRecordID,
		TargetRecordDescription: ev.TargetDesc,
		UserAgent:               ua,
		Status:                  &st,
		FailureReason:           ev.FailureReason,
		Details:                 detailsJSON,
	})
}
