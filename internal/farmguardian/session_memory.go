// Phase 63 — Guardian session memory: topic tags + prior-session context injection.

package farmguardian

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/llm"
)

// Session memory topic tags (tag-based overlap, no embeddings v1).
var sessionMemoryTopics = []string{
	"alerts", "feeding", "comfort", "grow", "stock", "money", "setup",
}

var sessionTopicKeywords = map[string][]string{
	"alerts":  {"alert", "alarm", "notification", "unread", "acknowledge", "severity"},
	"feeding": {"feed", "feeding", "fertigation", "nutrient", "ec", "ph", "reservoir", "irrigation", "schedule"},
	"comfort": {"comfort", "setpoint", "target", "band", "temperature", "humidity", "vpd", "co2", "climate"},
	"grow":    {"grow", "crop", "cycle", "harvest", "stage", "flower", "veg", "dli", "vpd", "strain", "plant"},
	"stock":   {"stock", "supply", "supplies", "inventory", "batch", "restock", "low stock", "refill"},
	"money":   {"cost", "money", "spend", "receipt", "budget", "finance"},
	"setup":   {"setup", "wizard", "first run", "add zone", "connect device", "pi", "wiring"},
}

var routeTopicHints = map[string][]string{
	"/alerts":              {"alerts"},
	"/feeding":             {"feeding"},
	"/operations/feeding":  {"feeding"},
	"/fertigation":         {"feeding"},
	"/comfort-targets":     {"comfort"},
	"/setpoints":           {"comfort"},
	"/operations/supplies": {"stock"},
	"/inventory":           {"stock"},
	"/operations/money":    {"money"},
	"/costs":               {"money"},
	"/plants":              {"grow"},
	"/pi-setup":            {"setup"},
}

// InferSessionTopics scans conversation text for known topic tags.
func InferSessionTopics(parts ...string) []string {
	joined := strings.ToLower(strings.Join(parts, "\n"))
	var out []string
	for topic, kws := range sessionTopicKeywords {
		for _, kw := range kws {
			if strings.Contains(joined, kw) {
				out = append(out, topic)
				break
			}
		}
	}
	return dedupeStrings(out)
}

// TopicsForRoute maps the operator's current UI path to memory topic tags.
func TopicsForRoute(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return nil
	}
	if topics, ok := routeTopicHints[path]; ok {
		return append([]string(nil), topics...)
	}
	if strings.HasPrefix(path, "/zones/") {
		return []string{"grow", "comfort"}
	}
	if strings.HasPrefix(path, "/farms/") && strings.Contains(path, "/setup") {
		return []string{"setup"}
	}
	if strings.HasPrefix(path, "/farms/") && strings.Contains(path, "/devices/new") {
		return []string{"setup"}
	}
	if strings.HasPrefix(path, "/farms/") && strings.Contains(path, "/zones/new") {
		return []string{"setup"}
	}
	return nil
}

// TopicsOverlap reports whether two topic slices share at least one tag.
func TopicsOverlap(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(a))
	for _, t := range a {
		seen[t] = struct{}{}
	}
	for _, t := range b {
		if _, ok := seen[t]; ok {
			return true
		}
	}
	return false
}

// BuildSessionSummaryText produces a short recap from session turns.
func BuildSessionSummaryText(turns []db.ListConversationTurnsBySessionRow, llmClient llm.ChatCompleter) (string, []string, error) {
	if len(turns) == 0 {
		return "", nil, nil
	}
	var transcript strings.Builder
	for _, t := range turns {
		transcript.WriteString("User: ")
		transcript.WriteString(strings.TrimSpace(t.UserMessage))
		transcript.WriteByte('\n')
		transcript.WriteString("Assistant: ")
		transcript.WriteString(strings.TrimSpace(t.AssistantMessage))
		transcript.WriteByte('\n')
	}
	textBlob := transcript.String()
	topics := InferSessionTopics(textBlob)

	if llmClient != nil {
		system := "Summarize this Farm Guardian chat in 2–3 short sentences for the operator's future sessions. " +
			"State what they asked about and what Guardian suggested. Do not invent outcomes. Plain language."
		user := "Transcript:\n" + trimForSummary(textBlob, 6000)
		if summary, err := llmClient.ChatCompletion(context.Background(), system, user); err == nil {
			summary = strings.TrimSpace(summary)
			if summary != "" {
				if len(topics) == 0 {
					topics = InferSessionTopics(summary)
				}
				return summary, topics, nil
			}
		}
	}

	first := strings.TrimSpace(turns[0].UserMessage)
	last := strings.TrimSpace(turns[len(turns)-1].AssistantMessage)
	summary := fallbackSessionSummary(first, last)
	if len(topics) == 0 {
		topics = InferSessionTopics(first, last)
	}
	if len(topics) == 0 {
		topics = []string{"grow"}
	}
	return summary, topics, nil
}

func fallbackSessionSummary(firstUser, lastAssistant string) string {
	q := truncateRunes(firstUser, 120)
	if q == "" {
		q = "a farm question"
	}
	a := truncateRunes(lastAssistant, 160)
	if a == "" {
		return fmt.Sprintf("Earlier you asked about %s.", q)
	}
	return fmt.Sprintf("Earlier you asked about %s. Guardian last suggested: %s", q, a)
}

func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || s == "" {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

func trimForSummary(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + "\n…(truncated)"
}

// PriorSessionContextBlock injects a matching prior summary into the system prompt.
func PriorSessionContextBlock(summary db.Gr33ncoreSessionSummary, now time.Time) string {
	if strings.TrimSpace(summary.SummaryText) == "" {
		return ""
	}
	age := humanizeAge(now.Sub(summary.CreatedAt))
	var b strings.Builder
	b.WriteString("[Prior session context: ")
	b.WriteString(age)
	b.WriteString(" ago you had a related Guardian conversation. ")
	b.WriteString(strings.TrimSpace(summary.SummaryText))
	b.WriteString(" Outcome: unknown.")
	if len(summary.Topics) > 0 {
		b.WriteString(" Topics: ")
		b.WriteString(strings.Join(summary.Topics, ", "))
	}
	b.WriteString("]\nAddress the current question with this in mind if relevant. Do not repeat this note to the user.")
	return b.String()
}

// MatchingTopicsForTurn merges question text and optional route context into topic tags.
func MatchingTopicsForTurn(question string, ref *ContextRef) []string {
	parts := []string{question}
	if ref != nil {
		if ref.Path != "" {
			parts = append(parts, ref.Path)
		}
		if ref.GuardianMode != "" {
			parts = append(parts, ref.GuardianMode)
		}
		if ref.NudgeCategory != "" {
			parts = append(parts, ref.NudgeCategory)
		}
	}
	topics := InferSessionTopics(parts...)
	if ref != nil && ref.Path != "" {
		for _, t := range TopicsForRoute(ref.Path) {
			topics = append(topics, t)
		}
	}
	return dedupeStrings(topics)
}

// FormatSessionSummariesExport renders plain-text export for operator download.
func FormatSessionSummariesExport(rows []db.Gr33ncoreSessionSummary, farmName string) string {
	var b strings.Builder
	b.WriteString("Farm Guardian session memory export\n")
	if farmName != "" {
		b.WriteString("Farm: ")
		b.WriteString(farmName)
		b.WriteByte('\n')
	}
	b.WriteString(fmt.Sprintf("Sessions: %d\n\n", len(rows)))
	for i, row := range rows {
		b.WriteString(fmt.Sprintf("--- Session %d ---\n", i+1))
		b.WriteString("session_id: ")
		b.WriteString(row.SessionID.String())
		b.WriteByte('\n')
		b.WriteString("created_at: ")
		b.WriteString(row.CreatedAt.UTC().Format(time.RFC3339))
		b.WriteByte('\n')
		if len(row.Topics) > 0 {
			b.WriteString("topics: ")
			b.WriteString(strings.Join(row.Topics, ", "))
			b.WriteByte('\n')
		}
		b.WriteString(strings.TrimSpace(row.SummaryText))
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// ParseSessionID is a thin helper for handlers.
func ParseSessionID(s string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(s))
}

// TopicLabel returns operator-facing chip text for a topic tag.
func TopicLabel(topic string) string {
	switch topic {
	case "alerts":
		return "Alerts"
	case "feeding":
		return "Feeding"
	case "comfort":
		return "Comfort"
	case "grow":
		return "Grow"
	case "stock":
		return "Stock"
	case "money":
		return "Money"
	case "setup":
		return "Setup"
	default:
		return topic
	}
}

// RecentTopicPrompt builds the continue-chip label from a summary row.
func RecentTopicPrompt(summary db.Gr33ncoreSessionSummary) string {
	label := "this topic"
	if len(summary.Topics) > 0 {
		label = TopicLabel(summary.Topics[0])
		if strings.EqualFold(label, "Grow") && strings.Contains(strings.ToLower(summary.SummaryText), "vpd") {
			label = "VPD"
		}
	}
	return fmt.Sprintf("You recently asked about %s — continue?", label)
}
