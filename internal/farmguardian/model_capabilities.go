package farmguardian

import (
	"os"
	"strings"
)

// IsChatCapable reports whether a model may be used for Guardian chat.
// Embedding-only models (capabilities == ["embedding"]) are excluded.
// When capabilities are unknown (empty), the model is included for back-compat
// unless IsEmbeddingModel identifies it as the configured embed model.
func IsChatCapable(caps []string) bool {
	if IsEmbeddingOnlyCapabilities(caps) {
		return false
	}
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

// IsSelectableChatModel is the chat-dropdown filter: chat-capable and not an embed model.
func IsSelectableChatModel(m ModelInfo) bool {
	if IsEmbeddingModel(m.Name, m.Capabilities) {
		return false
	}
	return IsChatCapable(m.Capabilities)
}

func IsEmbeddingOnlyCapabilities(caps []string) bool {
	if len(caps) == 0 {
		return false
	}
	hasEmbed := false
	for _, c := range caps {
		switch c {
		case "embedding":
			hasEmbed = true
		case "completion", "vision":
			return false
		}
	}
	return hasEmbed
}

// IsEmbeddingModel reports whether name/capabilities identify a RAG embedding model
// (not for Guardian chat answers). Used to keep embed models out of the chat dropdown.
func IsEmbeddingModel(name string, caps []string) bool {
	if IsEmbeddingOnlyCapabilities(caps) {
		return true
	}
	lower := strings.ToLower(strings.TrimSpace(name))
	if strings.Contains(lower, "embed") {
		return true
	}
	env := strings.TrimSpace(os.Getenv("EMBEDDING_MODEL"))
	if env == "" {
		return false
	}
	for _, key := range modelLookupKeys(name) {
		for _, ek := range modelLookupKeys(env) {
			if key != "" && key == ek {
				return true
			}
		}
	}
	return false
}
