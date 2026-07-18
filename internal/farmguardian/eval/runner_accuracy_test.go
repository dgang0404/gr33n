package eval

import (
	"encoding/json"
	"testing"
)

func TestAccuracyNoteFromChatResponse_topLevel(t *testing.T) {
	raw := `{"answer":"x","accuracy_note":"citation_number_mismatch"}`
	var parsed chatResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatal(err)
	}
	if got := accuracyNoteFromChatResponse(parsed); got != "citation_number_mismatch" {
		t.Fatalf("got %q", got)
	}
}

func TestAccuracyNoteFromChatResponse_debugFallback(t *testing.T) {
	raw := `{"answer":"x","debug":{"accuracy_note":"dangling_list_intro"}}`
	var parsed chatResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatal(err)
	}
	if got := accuracyNoteFromChatResponse(parsed); got != "dangling_list_intro" {
		t.Fatalf("got %q", got)
	}
}

func TestToEvalQuestionScores_includesAccuracyNote(t *testing.T) {
	scores := []ScoreResult{{
		ID: "fg-apple", Category: "field_guide", Passed: true,
		AccuracyNote: "citation_number_mismatch",
	}}
	out := ToEvalQuestionScores(scores)
	if len(out) != 1 || out[0].AccuracyNote != "citation_number_mismatch" {
		t.Fatalf("got %+v", out)
	}
}
