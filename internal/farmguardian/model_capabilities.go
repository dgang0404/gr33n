package farmguardian

// IsChatCapable reports whether a model may be used for Guardian chat.
// Embedding-only models (capabilities == ["embedding"]) are excluded.
// When capabilities are unknown (empty), the model is included for back-compat.
func IsChatCapable(caps []string) bool {
	if len(caps) == 0 {
		return true
	}
	for _, c := range caps {
		switch c {
		case "completion", "vision":
			return true
		}
	}
	return false
}
