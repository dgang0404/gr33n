package farmguardian

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian/tools"
)

// LLMProposalPolicy gates Phase 46 hybrid-C (matchers first, LLM proposal on miss).
type LLMProposalPolicy struct {
	Enabled bool
}

// LoadLLMProposalPolicyFromEnv reads GUARDIAN_LLM_PROPOSALS (true/1 enables).
func LoadLLMProposalPolicyFromEnv() LLMProposalPolicy {
	raw := strings.TrimSpace(os.Getenv("GUARDIAN_LLM_PROPOSALS"))
	return LLMProposalPolicy{
		Enabled: raw == "1" || strings.EqualFold(raw, "true") || strings.EqualFold(raw, "yes"),
	}
}

// LLMProposalAttemptInput carries the hybrid-C gate checks from phase 46 §4.1.
type LLMProposalAttemptInput struct {
	Question       string
	MatcherMatched bool
	HasOperate     bool
	InProcedure    bool
	Policy         LLMProposalPolicy
}

// ShouldAttemptLLMProposal returns true when the LLM proposal path may run.
func ShouldAttemptLLMProposal(in LLMProposalAttemptInput) bool {
	if !in.Policy.Enabled || !in.HasOperate || in.MatcherMatched || in.InProcedure {
		return false
	}
	return HasWriteIntent(in.Question)
}

// llmToolAllowlist is the Phase 46 §5 v1 narrow set. Read tools and bundle writes stay off.
var llmToolAllowlist = map[string]bool{
	"patch_fertigation_program": true,
	"patch_schedule":            true,
	"patch_rule":                  true,
	"ack_alert":                   true,
	"create_task":                 true,
	"create_task_from_alert":      true,
	"update_cycle_stage":          true,
}

var (
	writeIntentVerb = regexp.MustCompile(`(?i)\b(set|change|update|adjust|pause|disable|enable|resume|stop|turn\s+off|turn\s+on|acknowledge|ack|create|patch|switch)\b`)
	readOnlyIntent  = regexp.MustCompile(`(?i)^\s*(why|what\s+is|what's|explain|how\s+does|how\s+do|tell\s+me\s+about|describe|summarize|list|show)\b`)
	procedureIntent = regexp.MustCompile(`(?i)\bstart\s+procedure\b`)
)

// HasWriteIntent is a lightweight gate — imperative edit verbs without pure Q&A.
func HasWriteIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	if readOnlyIntent.MatchString(q) {
		return false
	}
	if procedureIntent.MatchString(q) {
		return false
	}
	return writeIntentVerb.MatchString(q)
}

// IsLLMToolAllowed reports whether a tool may be suggested by the LLM path.
func IsLLMToolAllowed(toolID string) bool {
	return llmToolAllowlist[strings.TrimSpace(toolID)]
}

// LLMProposalDraft is parsed structured output from assistant text (§4.2).
type LLMProposalDraft struct {
	Tool       string
	Args       map[string]any
	Summary    string
	Confidence string
}

// ParseLLMProposalFromAssistant extracts a tool proposal JSON block from LLM text.
func ParseLLMProposalFromAssistant(text string) (LLMProposalDraft, bool) {
	for _, block := range extractJSONBlocks(text) {
		var raw struct {
			Tool       string         `json:"tool"`
			Args       map[string]any `json:"args"`
			Summary    string         `json:"summary"`
			Confidence string         `json:"confidence"`
		}
		if err := json.Unmarshal([]byte(block), &raw); err != nil {
			continue
		}
		tool := strings.TrimSpace(raw.Tool)
		if tool == "" {
			continue
		}
		args := raw.Args
		if args == nil {
			args = map[string]any{}
		}
		return LLMProposalDraft{
			Tool:       tool,
			Args:       args,
			Summary:    strings.TrimSpace(raw.Summary),
			Confidence: strings.TrimSpace(raw.Confidence),
		}, true
	}
	return LLMProposalDraft{}, false
}

func extractJSONBlocks(text string) []string {
	var out []string
	rest := text
	for {
		start := strings.Index(rest, "```")
		if start < 0 {
			break
		}
		rest = rest[start+3:]
		if strings.HasPrefix(strings.ToLower(rest), "json") {
			rest = rest[4:]
		}
		end := strings.Index(rest, "```")
		if end < 0 {
			break
		}
		block := strings.TrimSpace(rest[:end])
		if block != "" {
			out = append(out, block)
		}
		rest = rest[end+3:]
	}
	trim := strings.TrimSpace(text)
	if strings.HasPrefix(trim, "{") && strings.Contains(trim, `"tool"`) {
		out = append(out, trim)
	}
	return out
}

// ValidateLLMProposalDraft applies allowlist + registry checks (WS1); ID binding expands in WS2.
func ValidateLLMProposalDraft(draft LLMProposalDraft, hasAdmin bool) (rejectReason string) {
	tool := strings.TrimSpace(draft.Tool)
	if tool == "" {
		return "missing tool"
	}
	if !IsLLMToolAllowed(tool) {
		return "tool not on LLM allowlist"
	}
	t, err := tools.Lookup(tool)
	if err != nil {
		return "unknown tool"
	}
	if t.RequiresAdmin && !hasAdmin {
		return "tool requires admin"
	}
	if draft.Summary == "" {
		return "missing summary"
	}
	if len(draft.Args) == 0 {
		return "missing args"
	}
	if strings.EqualFold(draft.Confidence, "low") && tools.RiskTierForTool(tool, draft.Args) == "high" {
		return "low confidence on high-tier tool"
	}
	return ""
}

// TryBuildLLMProposalsFromAssistant inserts a validated LLM-sourced proposal when policy allows.
// Called after rule-assisted matchers miss (Phase 46 WS3 wires this into chat handler).
func TryBuildLLMProposalsFromAssistant(
	ctx context.Context,
	q db.Querier,
	userID uuid.UUID,
	farmID int64,
	sessionID uuid.UUID,
	question string,
	assistantText string,
	policy LLMProposalPolicy,
	hasOperate bool,
	hasAdmin bool,
	inProcedure bool,
	matcherMatched bool,
) ([]ActionProposal, error) {
	if !ShouldAttemptLLMProposal(LLMProposalAttemptInput{
		Question:       question,
		MatcherMatched: matcherMatched,
		HasOperate:     hasOperate,
		InProcedure:    inProcedure,
		Policy:         policy,
	}) {
		return nil, nil
	}
	draft, ok := ParseLLMProposalFromAssistant(assistantText)
	if !ok {
		return nil, nil
	}
	if reason := ValidateLLMProposalDraft(draft, hasAdmin); reason != "" {
		return nil, nil
	}
	row, err := insertProposal(ctx, q, insertProposalInput{
		userID:     userID,
		farmID:     farmID,
		sessionID:  sessionID,
		toolID:     draft.Tool,
		args:       draft.Args,
		summary:    draft.Summary,
		revision:   1,
		llmSourced: true,
	})
	if err != nil {
		return nil, err
	}
	return []ActionProposal{ActionProposalFromRow(row)}, nil
}
