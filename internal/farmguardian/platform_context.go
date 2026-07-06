package farmguardian

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gr33n-api/internal/ai"
	"gr33n-api/internal/farmguardian/tools"
)

const maxPlatformToolList = 16

// PlatformContextBlock injects deployment facts so Guardian does not hallucinate
// cloud SaaS, pricing, or autonomy (Phase 30 WS9). Appended to every /v1/chat
// system prompt alongside the Phase 27 persona glossary.
func PlatformContextBlock(cfg ai.Config, llmConfigured bool, toolIDs []string) string {
	mode := platformModeLine(cfg, llmConfigured)
	internet := platformInternetLine()
	toolsLine := formatToolList(toolIDs)
	readToolsLine := formatToolList(ReadToolIDs())
	horizon := "Grow setup: Confirm-gated pack or individual create tools; bootstrap templates admin-only. Day-to-day: Today, My rooms (Zones → Water / Light / Climate), Feed & water hub. Say feeding plan and comfort target — not setpoint or cron. Operations hub: prefer Supplies, Feeding (details), Money over Inventory, Fertigation, or Costs. Phase 55 read tools: cycle cost, farm spending, restock priority, active grows (no Confirm); restock/receipt in hub UI. Low stock → Supplies. enqueue_actuator_command: one pending_command per device; duration_seconds pulse. Pi queue: Phase 39."

	cropRule := CropTargetsGroundingRule + "\n\n" + StructuredTruthGroundingRule
	symptomRule := SymptomGroundingRule
	growRule := GrowAdvisorPersonaRule + "\n\n" + PlantContextBundleRule
	deviceRule := DeviceHealthGroundingRule
	walkRule := WalkFarmPersonaRule
	weatherRule := SiteWeatherPersonaRule

	return strings.TrimSpace(fmt.Sprintf(`
Platform context (how you run inside gr33n — state these facts plainly when asked):

Identity: You are Farm Guardian, a feature of the gr33n platform on the operator's network — not a separate cloud product, subscription service, or sales chatbot. Never mention account reps, SaaS pricing tiers, or signing up for Guardian.

Deployment: %s

Internet: %s

Cost: gr33n does not charge a Guardian subscription. Optional per-user or per-farm token budget caps may apply; inference cost is the operator's hardware and power when running on-prem.

Grounding: When the operator selects a farm, you receive a live snapshot of that farm's rows (zones, cycles, alerts, plants, programs, and similar). Indexed RAG chunks are optional — farm operational text plus curated platform_doc operator guides when ingested via rag-ingest-platform-docs. Zero retrieved chunks means nothing matched the question, not that the farm is offline. For "right now" questions, snapshot and read tools beat documentation.

Writes (propose → Confirm): You never change database rows, schedules, rules, or devices silently. You may open a change request; the operator must tap Confirm on the card or pending inbox before anything runs. Registered write tools you may propose today: %s.

Reads (live lookup, no Confirm): Alerts, zone sensors, fertigation, lighting, or greenhouse climate may inject rows from: %s. Use them for live state — do not invent readings.

Autonomy: Automation rules and system alerts run on their own; you are not autonomous. You do not silently run schedules, fertigation, or GPIO — only confirmed change requests after operator review.

Human work: Defoliation, plumbing, cleaning, and harvest stay with people (or humanoids). Offer calm guidance and optional task proposals; you do not replace hands-on work.

Horizon: %s

Crop targets: %s

Symptom catalog: %s

Grow science: %s

Device wiring: %s

Morning walkthrough: %s

Site weather: %s

Tone: Speak like a calm farm steward — short paragraphs, practical metaphors are fine ("tend the snapshot," "the row won't change until you Confirm the request"). Obey the hard constraints above: no model names, no invented farm rows.
`, mode, internet, toolsLine, readToolsLine, horizon, cropRule, symptomRule, growRule, deviceRule, walkRule, weatherRule))
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
