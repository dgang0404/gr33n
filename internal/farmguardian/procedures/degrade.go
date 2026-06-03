package procedures

import "strings"

// DegradeBanner prefixes static answers when the LLM is unavailable (Phase 37 WS1).
const DegradeBanner = "**Field assistant (LLM offline)** — using authored checklists and procedures. No AI inference on this turn.\n\n"

// IsFieldRelatedQuestion reports whether we should offer field degrade instead of a hard 503.
func IsFieldRelatedQuestion(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}
	keywords := []string{
		"wire", "wiring", "gpio", "relay", "sensor", "pi ", "raspberry",
		"actuator", "install", "heartbeat", "offline queue", "procedure",
		"plumb", "irrigation", "pump", "field guide", "pinout", "grow light",
		"help me", "won't turn", "wont turn", "no reading", "not reading",
	}
	for _, k := range keywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	if SuggestID(message) != "" {
		return true
	}
	if strings.Contains(lower, "start procedure") || strings.Contains(lower, "list procedure") {
		return true
	}
	return false
}

// TryFieldDegrade returns a static field-assistant answer when the LLM cannot run.
func TryFieldDegrade(repoRoot, message string, meta SessionMeta) (answer string, newMeta SessionMeta, payload *TurnPayload, ok bool) {
	if handled, ans, nm, pl := HandleTurn(repoRoot, message, meta); handled {
		if !strings.HasPrefix(ans, "**Field assistant") {
			ans = DegradeBanner + ans
		}
		return ans, nm, pl, true
	}
	if id := SuggestID(message); id != "" {
		handled, ans, nm, pl := HandleTurn(repoRoot, "start procedure "+id, meta)
		if handled {
			ans = DegradeBanner + ans
			return ans, nm, pl, true
		}
	}
	if !IsFieldRelatedQuestion(message) {
		return "", meta, nil, false
	}
	summary, err := ListSummary(repoRoot)
	if err != nil {
		return DegradeBanner + "Guided field procedures are not available on this server. Use **GET /v1/field-guides/procedures/{id}/print** for static checklists, or restore the local LLM at LLM_BASE_URL.", meta, nil, true
	}
	return DegradeBanner + summary + "\n\n_Print any checklist (no LLM):_ open `/v1/field-guides/procedures/{id}/print` in the API or use the **Print checklist** link on a procedure card.", meta, nil, true
}
