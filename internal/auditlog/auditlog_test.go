package auditlog

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
)

type mockQuerier struct {
	db.Querier
	insertFn func(ctx context.Context, arg db.InsertUserActivityLogParams) error
}

func (m *mockQuerier) InsertUserActivityLog(ctx context.Context, arg db.InsertUserActivityLogParams) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, arg)
	}
	return nil
}

func TestSubmitErr_PersistsEvent(t *testing.T) {
	uid := uuid.New()
	farmID := int64(1)
	var got db.InsertUserActivityLogParams
	mq := &mockQuerier{
		insertFn: func(_ context.Context, arg db.InsertUserActivityLogParams) error {
			got = arg
			return nil
		},
	}
	ctx := authctx.WithUserID(context.Background(), uid)
	req := httptest.NewRequest("POST", "/farms/1/tasks", nil)
	req.Header.Set("User-Agent", "phase117-test/1.0")
	err := SubmitErr(ctx, mq, req, Event{
		FarmID: FarmIDPtr(farmID),
		Action: db.Gr33ncoreUserActionTypeEnumLoginSuccess,
		Status: "success",
		Details: map[string]any{"tool_id": "create_task"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !got.UserID.Valid || got.UserID.Bytes != uid {
		t.Fatalf("user id not recorded: %+v", got.UserID)
	}
	if got.FarmID == nil || *got.FarmID != farmID {
		t.Fatalf("farm id = %v", got.FarmID)
	}
	if got.UserAgent == nil || *got.UserAgent != "phase117-test/1.0" {
		t.Fatalf("user agent = %v", got.UserAgent)
	}
	var details map[string]any
	if err := json.Unmarshal(got.Details, &details); err != nil {
		t.Fatal(err)
	}
	if details["tool_id"] != "create_task" {
		t.Fatalf("details = %#v", details)
	}
}

func TestSubmitErr_DefaultStatusSuccess(t *testing.T) {
	var gotStatus *string
	mq := &mockQuerier{
		insertFn: func(_ context.Context, arg db.InsertUserActivityLogParams) error {
			gotStatus = arg.Status
			return nil
		},
	}
	err := SubmitErr(context.Background(), mq, nil, Event{
		Action: db.Gr33ncoreUserActionTypeEnumLogout,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotStatus == nil || *gotStatus != "success" {
		t.Fatalf("status = %v", gotStatus)
	}
}

func TestSubmitErr_NoUserInContext(t *testing.T) {
	var gotUID pgtype.UUID
	mq := &mockQuerier{
		insertFn: func(_ context.Context, arg db.InsertUserActivityLogParams) error {
			gotUID = arg.UserID
			return nil
		},
	}
	if err := SubmitErr(context.Background(), mq, nil, Event{
		Action: db.Gr33ncoreUserActionTypeEnumLoginFailure,
		Status: "failure",
	}); err != nil {
		t.Fatal(err)
	}
	if gotUID.Valid {
		t.Fatal("expected invalid user id when context has no user")
	}
}
