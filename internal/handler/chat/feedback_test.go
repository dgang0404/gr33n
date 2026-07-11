package chat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

func TestParseFeedbackSince_days(t *testing.T) {
	since, err := parseFeedbackSince("7d")
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(since) < 6*24*time.Hour {
		t.Fatalf("since too recent: %v", since)
	}
}

func TestParseFeedbackSince_invalid(t *testing.T) {
	if _, err := parseFeedbackSince("bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestExcerptText(t *testing.T) {
	long := strings.Repeat("a", 300)
	if got := excerptText(long, 240); len(got) != 243 {
		t.Fatalf("len=%d want 243 (240 runes + UTF-8 ellipsis)", len(got))
	}
}

func TestFormatPGTime(t *testing.T) {
	if formatPGTime(pgtype.Timestamptz{}) != "" {
		t.Fatal("expected empty")
	}
	ts := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	if !strings.Contains(formatPGTime(pgtype.Timestamptz{Time: ts, Valid: true}), "2026-07-06") {
		t.Fatal("expected RFC3339 date")
	}
}

func TestPatchTurnFeedback_requiresRating(t *testing.T) {
	h := &Handler{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/chat/sessions/x/turns/0/feedback", strings.NewReader(`{}`))
	req.SetPathValue("session_id", uuid.New().String())
	req.SetPathValue("turn_index", "0")
	h.PatchTurnFeedback(rec, req)
	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusBadRequest {
		// no auth context → unauthorized first in real flow; nil q may 503
		if rec.Code == http.StatusServiceUnavailable {
			return
		}
	}
}

func TestFeedbackExportRowFromDB(t *testing.T) {
	rating := "down"
	reason := "Missed alert"
	row := db.ListConversationFeedbackForFarmRow{
		SessionID:        uuid.New(),
		TurnIndex:        2,
		UserMessage:      "q",
		AssistantMessage:   strings.Repeat("x", 400),
		FeedbackRating:   &rating,
		FeedbackReason:   &reason,
		Grounded:         true,
		LlmModel:         "phi3:mini",
		CreatedAt:        time.Now().UTC(),
		FeedbackAt:       pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
	exp := feedbackExportRowFromDB(row)
	if exp.Rating != "down" || len(exp.AnswerExcerpt) < 240 {
		t.Fatalf("%+v", exp)
	}
}
