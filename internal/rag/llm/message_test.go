package llm

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMessage_MarshalText(t *testing.T) {
	raw, err := json.Marshal(Message{Role: "user", Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"content":"hello"`) {
		t.Fatalf("got %s", raw)
	}
}

func TestMessage_MarshalMultimodal(t *testing.T) {
	m := UserMessageWithImages("leaves?", []ImageAttachment{{
		DataURL: "data:image/png;base64,abc",
	}})
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, want := range []string{`"type":"text"`, `"type":"image_url"`, `"url":"data:image/png;base64,abc"`} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}

func TestMessage_TextContent(t *testing.T) {
	m := UserMessageWithImages("check RH", []ImageAttachment{{DataURL: "data:image/png;base64,x"}})
	if m.TextContent() != "check RH" {
		t.Fatalf("got %q", m.TextContent())
	}
}
