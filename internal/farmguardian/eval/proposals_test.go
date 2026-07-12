package eval

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerifyPendingProposalIDs_allPresent(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"proposals": []map[string]any{
				{"proposal_id": "p1", "tool": "ack_alert"},
				{"proposal_id": "p2", "tool": "create_task"},
			},
		})
	}))
	defer srv.Close()

	client := NewAPIClient(srv.URL, "tok", 1)
	if err := VerifyPendingProposalIDs(context.Background(), client, []string{"p1", "p2"}); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyPendingProposalIDs_missingErrors(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"proposals": []map[string]any{
				{"proposal_id": "p1", "tool": "ack_alert"},
			},
		})
	}))
	defer srv.Close()

	client := NewAPIClient(srv.URL, "tok", 1)
	err := VerifyPendingProposalIDs(context.Background(), client, []string{"p1", "expired"})
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected missing expired error, got %v", err)
	}
}
