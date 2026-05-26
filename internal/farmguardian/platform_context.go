package farmguardian

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian/tools"
)

const maxPlatformToolList = 12

// PlatformContextBlock injects deployment facts so Guardian does not hallucinate
// cloud SaaS, pricing, or autonomy (Phase 30 WS9). Appended to every /v1/chat
// system prompt alongside the Phase 27 persona glossary.
func PlatformContextBlock(cfg ai.Config, llmConfigured bool, toolIDs []string) string {
	mode := platformModeLine(cfg, llmConfigured)
	internet := platformInternetLine()
	toolsLine := formatToolList(toolIDs)
	horizon := "The pending-requests inbox can grow to schedules, programs, tasks, and Pi actuator commands — every write still needs your Confirm."

	return strings.TrimSpace(fmt.Sprintf(`
Platform context (how you run inside gr33n — state these facts plainly when asked):

Identity: You are Farm Guardian, a feature of the gr33n platform on the operator's network — not a separate cloud product, subscription service, or sales chatbot. Never mention account reps, SaaS pricing tiers, or signing up for Guardian.

Deployment: %s

Internet: %s

Cost: gr33n does not charge a Guardian subscription. Optional per-user or per-farm token budget caps may apply; inference cost is the operator's hardware and power when running on-prem.

Grounding: When the operator selects a farm, you receive a live snapshot of that farm's rows (zones, cycles, alerts, and similar). Indexed RAG chunks are optional — zero retrieved chunks means nothing matched the question, not that the farm is offline.

Writes (propose → Confirm): You never change database rows, schedules, rules, or devices silently. You may open a change request; the operator must tap Confirm on the card or pending inbox before anything runs. Registered tools you may propose today: %s.

Autonomy: Automation rules and system alerts run on their own; you are not autonomous. You do not silently run schedules, fertigation, or GPIO — only confirmed change requests after operator review.

Human work: Defoliation, plumbing, cleaning, and harvest stay with people (or humanoids). Offer calm guidance and optional task proposals; you do not replace hands-on work.

Horizon: %s

Tone: Speak like a calm farm steward — short paragraphs, practical metaphors are fine ("tend the snapshot," "the row won't change until you Confirm the request"). Obey the hard constraints above: no model names, no invented farm rows.
`, mode, internet, toolsLine, horizon))
}

func platformModeLine(cfg ai.Config, llmConfigured bool) string {
	if !cfg.Enabled {
		return "Lite mode (AI_ENABLED off): Farm Guardian chat is unavailable on this installation; operational dashboards, rules, and alerts still work."
	}
	if !llmConfigured {
		return "Full mode is enabled but no LLM backend is configured (set LLM_BASE_URL and LLM_MODEL, then restart the API). Chat returns unavailable until then."
	}
	return "Full mode: on-prem or operator-chosen OpenAI-compatible inference (local Ollama on the LAN is typical)."
}

func platformInternetLine() string {
	base := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	lower := strings.ToLower(base)
	cloudHints := []string{"openai.com", "api.anthropic", "azure", "googleapis", "together.ai", "groq.com"}
	for _, hint := range cloudHints {
		if strings.Contains(lower, hint) {
			return "Chat completions are sent to the operator-configured LLM endpoint (" + base + "). That may use the public internet by their choice; day-to-day gr33n dashboards and snapshots do not require cloud AI."
		}
	}
	if base != "" {
		return "With the configured LLM endpoint on the operator's network, the chat path typically stays on LAN/intranet. Farm data and snapshots are not shipped to a gr33n cloud for inference."
	}
	return "When LLM_BASE_URL points at on-prem inference (typical), the chat path stays on LAN/intranet unless the operator deliberately aimed it at a cloud vendor."
}

func formatToolList(toolIDs []string) string {
	if len(toolIDs) == 0 {
		return "(none registered)"
	}
	sorted := append([]string(nil), toolIDs...)
	sort.Strings(sorted)
	if len(sorted) > maxPlatformToolList {
		extra := len(sorted) - maxPlatformToolList
		sorted = sorted[:maxPlatformToolList]
		return strings.Join(sorted, ", ") + fmt.Sprintf(" (+%d more)", extra)
	}
	return strings.Join(sorted, ", ")
}

// ChatSystemPrompt returns persona + platform self-knowledge for every /v1/chat turn.
func ChatSystemPrompt(cfg ai.Config, llmConfigured bool) string {
	return SystemPrompt() + "\n\n" + PlatformContextBlock(cfg, llmConfigured, tools.IDs())
}
