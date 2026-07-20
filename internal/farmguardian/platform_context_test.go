package farmguardian

import (
	"strings"
	"testing"
	"unicode/utf8"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian/tools"
)

func TestPlatformContextBlock_ContainsRequiredFacts(t *testing.T) {
	block := PlatformContextBlock(ai.Config{Enabled: true}, true, tools.IDs())
	for _, want := range []string{
		"on-prem",
		"Confirm",
		"not autonomous",
		"ack_alert",
		"list_unread_alerts",
		"summarize_zone",
		"list_plants",
		"summarize_zone_fertigation",
		"summarize_farm_low_stock",
		"summarize_cycle_cost",
		"summarize_farm_spending",
		"restock_priority",
		"summarize_active_grows",
		"Phase 55",
		"Water / Light / Climate",
		"pending_command",
		"Phase 39",
		"Reads (live lookup, no Confirm)",
		"lookup_crop_symptoms",
		"Symptom catalog",
		"subscription",
		"LAN/intranet",
	} {
		if !strings.Contains(block, want) {
			t.Fatalf("platform block missing %q:\n%s", want, block)
		}
	}
}

func TestPlatformContextBlock_LiteMode(t *testing.T) {
	block := PlatformContextBlock(ai.Config{Enabled: false}, false, nil)
	if !strings.Contains(block, "Lite mode") {
		t.Fatal("expected Lite mode wording")
	}
}

func TestPlatformContextBlock_TruncatesLongToolList(t *testing.T) {
	many := make([]string, 25)
	for i := range many {
		many[i] = "tool_" + strings.Repeat("x", 2)
	}
	block := PlatformContextBlock(ai.Config{Enabled: true}, true, many)
	if !strings.Contains(block, "(+") {
		t.Fatal("expected truncation marker for long tool list")
	}
}

func TestChatSystemPrompt_IncludesPersonaAndPlatform(t *testing.T) {
	full := ChatSystemPrompt(ai.Config{Enabled: true}, true)
	if !strings.Contains(full, "Farm Guardian") {
		t.Fatal("missing persona")
	}
	if !strings.Contains(full, "Platform context") {
		t.Fatal("missing platform block")
	}
	// Regression guard — platform context grew with Phases 55–152; keep bounded but
	// don't fail on every new read-tool line. Revisit if this exceeds ~8k runes.
	n := utf8.RuneCountInString(PlatformContextBlock(ai.Config{Enabled: true}, true, tools.IDs()))
	if n > 7000 {
		t.Fatalf("platform block grew too large for token budget (%d runes)", n)
	}
}
