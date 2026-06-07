package farmguardian

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})))
	t.Cleanup(func() { slog.SetDefault(slog.Default()) })
	return &buf
}

func TestLogMatcherProposalHit(t *testing.T) {
	buf := captureSlog(t)
	LogMatcherProposalHit(1, "ack_alert")
	out := buf.String()
	for _, want := range []string{
		"guardian_matcher_proposal_hit",
		"event=guardian_matcher_proposal_hit",
		"farm_id=1",
		"tool=ack_alert",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log missing %q:\n%s", want, out)
		}
	}
}

func TestLogLLMProposalSuggested(t *testing.T) {
	buf := captureSlog(t)
	LogLLMProposalSuggested(2, "patch_fertigation_program")
	out := buf.String()
	for _, want := range []string{
		"guardian_llm_proposal_suggested",
		"event=guardian_llm_proposal_suggested",
		"farm_id=2",
		"tool=patch_fertigation_program",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log missing %q:\n%s", want, out)
		}
	}
}

func TestLogLLMProposalRejected(t *testing.T) {
	buf := captureSlog(t)
	LogLLMProposalRejected(3, "patch_rule", "program_id not on farm")
	out := buf.String()
	for _, want := range []string{
		"guardian_llm_proposal_rejected",
		"event=guardian_llm_proposal_rejected",
		"farm_id=3",
		"tool=patch_rule",
		"reason=\"program_id not on farm\"",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log missing %q:\n%s", want, out)
		}
	}
}
