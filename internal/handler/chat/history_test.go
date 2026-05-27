package chat

import (
	"strings"
	"testing"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/llm"
)

func TestParseOrNewSession_GeneratesWhenEmpty(t *testing.T) {
	id, err := parseOrNewSession("")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id.String() == "" {
		t.Fatal("expected non-zero UUID")
	}
}

func TestParseOrNewSession_AcceptsValidUUID(t *testing.T) {
	want := "33333333-3333-4333-8333-333333333333"
	id, err := parseOrNewSession(want)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id.String() != want {
		t.Fatalf("got %q want %q", id.String(), want)
	}
}

func TestParseOrNewSession_RejectsBadUUID(t *testing.T) {
	if _, err := parseOrNewSession("not-a-uuid"); err == nil {
		t.Fatal("expected error for invalid uuid")
	}
}

func TestReplayHistory_EmptyReturnsNil(t *testing.T) {
	if got := replayHistory(nil, 10); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestReplayHistory_InterleavesUserAssistant(t *testing.T) {
	rows := []db.ListConversationTurnsBySessionRow{
		{TurnIndex: 0, UserMessage: "u1", AssistantMessage: "a1"},
		{TurnIndex: 1, UserMessage: "u2", AssistantMessage: "a2"},
	}
	got := replayHistory(rows, 10)
	want := []llm.Message{
		{Role: "user", Content: "u1"},
		{Role: "assistant", Content: "a1"},
		{Role: "user", Content: "u2"},
		{Role: "assistant", Content: "a2"},
	}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Role != want[i].Role || got[i].TextContent() != want[i].TextContent() {
			t.Fatalf("msg %d: got %+v want %+v", i, got[i], want[i])
		}
	}
}

func TestReplayHistory_DropsOldestWhenOverCap(t *testing.T) {
	rows := make([]db.ListConversationTurnsBySessionRow, 5)
	for i := range rows {
		rows[i] = db.ListConversationTurnsBySessionRow{
			TurnIndex:        int32(i),
			UserMessage:      "u" + string(rune('0'+i)),
			AssistantMessage: "a" + string(rune('0'+i)),
		}
	}
	got := replayHistory(rows, 2)
	// 2 turns × (user + assistant) = 4 messages
	if len(got) != 4 {
		t.Fatalf("expected 4 messages after cap, got %d", len(got))
	}
	if got[0].Content != "u3" || got[1].Content != "a3" || got[2].Content != "u4" || got[3].Content != "a4" {
		t.Fatalf("expected oldest dropped (kept turns 3 and 4), got %+v", got)
	}
}

func TestBuildMessages_NoHistory(t *testing.T) {
	got := buildMessages("system-x", nil, llm.Message{Role: "user", Content: "user-x"})
	if len(got) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(got))
	}
	if got[0].Role != "system" || got[0].Content != "system-x" {
		t.Fatalf("system slot wrong: %+v", got[0])
	}
	if got[1].Role != "user" || got[1].Content != "user-x" {
		t.Fatalf("user slot wrong: %+v", got[1])
	}
}

func TestNormaliseTitle(t *testing.T) {
	if got := normaliseTitle(nil); got != nil {
		t.Fatalf("nil in → nil out, got %v", got)
	}
	empty := "   "
	if got := normaliseTitle(&empty); got != nil {
		t.Fatalf("whitespace → nil, got %v", got)
	}
	plain := "  My Sunday morning chat  "
	got := normaliseTitle(&plain)
	if got == nil || *got != "My Sunday morning chat" {
		t.Fatalf("expected trimmed, got %v", got)
	}
	long := strings.Repeat("a", 200)
	got = normaliseTitle(&long)
	if got == nil {
		t.Fatal("expected non-nil for long input")
	}
	// 120 a's + the ellipsis rune.
	if !strings.HasSuffix(*got, "…") {
		t.Fatalf("expected ellipsis suffix, got %q", *got)
	}
}

func TestBuildMessages_WithHistory(t *testing.T) {
	hist := []llm.Message{
		{Role: "user", Content: "prev-u"},
		{Role: "assistant", Content: "prev-a"},
	}
	got := buildMessages("sys", hist, llm.Message{Role: "user", Content: "now"})
	if len(got) != 4 {
		t.Fatalf("expected 4 (system + 2 history + current user), got %d", len(got))
	}
	if got[0].Content != "sys" || got[1].Content != "prev-u" || got[2].Content != "prev-a" || got[3].Content != "now" {
		t.Fatalf("ordering wrong: %+v", got)
	}
}
