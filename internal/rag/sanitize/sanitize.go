// Package sanitize prepares text and JSON from operational tables for RAG indexing
// without leaking webhook URLs, tokens, or similar fields (Phase 24 threat model).
package sanitize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const defaultMaxNoteRunes = 8000

// PlainNotes trims and caps length for free-text notes (labor, cycle notes, etc.).
// It does not remove semantic content; use for fields that are already approved for indexing.
func PlainNotes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = defaultMaxNoteRunes
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	if len(runes) > maxRunes {
		runes = runes[:maxRunes]
	}
	return strings.TrimSpace(string(runes))
}

var sensitiveKeySubstrings = []string{
	"url", "token", "secret", "password", "webhook", "authorization",
	"api_key", "apikey", "bearer", "credential", "private_key",
}

func keyLooksSensitive(key string) bool {
	kl := strings.ToLower(strings.TrimSpace(key))
	if kl == "" {
		return true
	}
	for _, sub := range sensitiveKeySubstrings {
		if strings.Contains(kl, sub) {
			return true
		}
	}
	return false
}

// AutomationDetailsJSON strips sensitive keys from automation_runs.details JSONB and
// turns the remainder into compact "key: value" lines suitable for embedding.
// Invalid JSON or empty after filtering yields an empty string.
func AutomationDetailsJSON(raw []byte) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "{}" {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil || len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		if keyLooksSensitive(k) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		v := m[k]
		if v == nil {
			continue
		}
		s := valueStringForEmbed(v)
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(s)
	}
	return b.String()
}

func valueStringForEmbed(v any) string {
	switch t := v.(type) {
	case string:
		if looksLikeCredentialOrURL(t) {
			return ""
		}
		return t
	case float64:
		return strings.TrimSpace(fmt.Sprint(t))
	case bool:
		return fmt.Sprint(t)
	case map[string]any, []any:
		raw, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		// Nested JSON — only embed if recursive filter is needed; v1 keeps structure without re-walking keys.
		if len(raw) > 4000 {
			raw = raw[:4000]
		}
		return string(raw)
	default:
		raw, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(raw)
	}
}

func looksLikeCredentialOrURL(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 8 {
		return false
	}
	ls := strings.ToLower(s)
	if strings.HasPrefix(ls, "http://") || strings.HasPrefix(ls, "https://") {
		return true
	}
	if strings.Contains(ls, "://") && strings.Contains(ls, "@") {
		return true
	}
	return false
}
