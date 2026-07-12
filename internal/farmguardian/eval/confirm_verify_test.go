package eval

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gr33n-api/internal/farmguardian"
)

func TestVerifyWriteAck_confirmResultAndList(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/farms/1/alerts" {
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 5, "is_acknowledged": true, "subject": "Humidity high"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewAPIClient(srv.URL, "tok", 1)
	err := VerifyConfirmSideEffect(context.Background(), client, ConfirmVerificationInput{
		FixtureID: "write-ack",
		Tool:      "ack_alert",
		Args:      map[string]any{"alert_id": float64(5)},
		Result:    map[string]any{"alert_id": float64(5), "is_acknowledged": true},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifyWriteFeed_programVolume(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/farms/1/fertigation/programs" {
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 12, "name": "Veg Tent", "total_volume_liters": 0.3},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewAPIClient(srv.URL, "tok", 1)
	err := VerifyConfirmSideEffect(context.Background(), client, ConfirmVerificationInput{
		FixtureID: "write-feed",
		Tool:      "patch_fertigation_program",
		Args: map[string]any{
			"program_id":          float64(12),
			"total_volume_liters": 0.3,
		},
		Result: map[string]any{"program_id": float64(12), "is_active": true},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifyWriteSchedule_paused(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/farms/1/schedules" {
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 7, "name": "Veg lights", "is_active": false},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewAPIClient(srv.URL, "tok", 1)
	err := VerifyConfirmSideEffect(context.Background(), client, ConfirmVerificationInput{
		FixtureID: "write-schedule",
		Tool:      "patch_schedule",
		Args: map[string]any{
			"schedule_id": float64(7),
			"is_active":   false,
		},
		Result: map[string]any{"schedule_id": float64(7), "is_active": false},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifyWriteTask_taskInList(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/farms/1/tasks" {
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 99, "title": "Refill calcium nitrate"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewAPIClient(srv.URL, "tok", 1)
	err := VerifyConfirmSideEffect(context.Background(), client, ConfirmVerificationInput{
		FixtureID: "write-task",
		Tool:      "create_task",
		Args:      map[string]any{"title": "Refill calcium nitrate"},
		Result:    map[string]any{"task_id": float64(99), "title": "Refill calcium nitrate"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestProposalConfirmTargets_passedOnly(t *testing.T) {
	t.Parallel()
	fixtures := []Question{
		{ID: "write-ack", ExpectProposal: true},
		{ID: "write-feed", ExpectProposal: true},
	}
	scores := []farmguardian.EvalQuestionScore{
		{ID: "write-ack", Passed: true, ProposalIDs: []string{"a1"}},
		{ID: "write-feed", Passed: false, ProposalIDs: []string{"f1"}},
	}
	got := ProposalConfirmTargets(fixtures, scores)
	if len(got) != 1 || got[0].ProposalID != "a1" || got[0].FixtureID != "write-ack" {
		t.Fatalf("got %#v", got)
	}
}

func TestConfirmProposal_success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v1/chat/confirm" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"summary": "Acknowledged alert",
				"result":  map[string]any{"alert_id": 5, "is_acknowledged": true},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewAPIClient(srv.URL, "tok", 1)
	cr, err := client.ConfirmProposal(context.Background(), "uuid-1")
	if err != nil {
		t.Fatal(err)
	}
	if !cr.Result["is_acknowledged"].(bool) {
		t.Fatalf("result %#v", cr.Result)
	}
}
