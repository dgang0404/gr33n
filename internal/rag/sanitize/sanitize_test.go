package sanitize

import (
	"strings"
	"testing"
)

func TestPlainNotes(t *testing.T) {
	if got := PlainNotes("  hello  ", 10); got != "hello" {
		t.Fatalf("trim: %q", got)
	}
	long := strings.Repeat("é", 20)
	if got := PlainNotes(long, 5); len([]rune(got)) != 5 {
		t.Fatalf("cap runes: %#v", []rune(got))
	}
}

func TestAutomationDetailsJSON(t *testing.T) {
	raw := []byte(`{"reason":"low_ec","webhook_url":"https://example.com/hook?token=secret","step":3}`)
	got := AutomationDetailsJSON(raw)
	if strings.Contains(got, "example.com") || strings.Contains(got, "secret") || strings.Contains(got, "webhook_url") {
		t.Fatalf("leaked sensitive field: %q", got)
	}
	if !strings.Contains(got, "reason") || !strings.Contains(got, "step") {
		t.Fatalf("expected preserved keys: %q", got)
	}
}

func TestFertigationProgramMetadataForEmbed(t *testing.T) {
	raw := []byte(`{"tags":["veg","daily"],"steps":[{"type":"http_post","url":"https://evil.example/hook"}],"safe_note":"hello"}`)
	got := FertigationProgramMetadataForEmbed(raw)
	if strings.Contains(got, "evil.example") || strings.Contains(got, "http_post") || strings.Contains(got, "steps") {
		t.Fatalf("must drop steps / secrets: %q", got)
	}
	if !strings.Contains(got, "tags") || !strings.Contains(got, "veg") {
		t.Fatalf("expected tags: %q", got)
	}
	if !strings.Contains(got, "safe_note") || !strings.Contains(got, "hello") {
		t.Fatalf("expected safe_note: %q", got)
	}
}
