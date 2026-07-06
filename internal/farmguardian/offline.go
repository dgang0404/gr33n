package farmguardian

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"

	"gr33n-api/internal/ai"
)

// Safety tiers for field guides and guided procedures (Phase 37 WS4).
const (
	SafetyTierSafe                    = "safe"
	SafetyTierCaution                 = "caution"
	SafetyTierQualifiedPersonRequired = "qualified_person_required"
)

// FieldAssistantHealth summarizes offline-capable Guardian backends (Phase 37 WS1).
type FieldAssistantHealth struct {
	FieldMode               bool   `json:"field_mode"`
	LLMBaseURL              string `json:"llm_base_url,omitempty"`
	LLMReachable            bool   `json:"llm_reachable"`
	LLMReachableError       string `json:"llm_reachable_error,omitempty"`
	EmbeddingBaseURL        string `json:"embedding_base_url,omitempty"`
	EmbeddingConfigured     bool   `json:"embedding_configured"`
	EmbeddingLocalEndpoint  bool   `json:"embedding_local_endpoint"`
	EmbeddingReachable      bool   `json:"embedding_reachable"`
	EmbeddingReachableError string `json:"embedding_reachable_error,omitempty"`
	SplitInferenceHosts     bool   `json:"split_inference_hosts"`
	FieldGuideChunkCount    int64  `json:"field_guide_chunk_count"`
	PlatformDocChunkCount   int64  `json:"platform_doc_chunk_count"`
}

// IsLocalInferenceURL reports whether baseURL points at loopback or a private LAN host
// (typical on-prem Ollama / LM Studio — Phase 37 offline field mode).
func IsLocalInferenceURL(baseURL string) bool {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate()
}

// BuildFieldAssistantHealth probes env + optional DB chunk counts for GET /v1/chat/health.
func BuildFieldAssistantHealth(ctx context.Context, llmReachable func(context.Context, string, string) error, fieldGuideChunks, platformDocChunks int64) FieldAssistantHealth {
	llmBase := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	embBase := strings.TrimSpace(os.Getenv("EMBEDDING_BASE_URL"))
	if embBase == "" {
		embBase = llmBase
	}
	embKey := strings.TrimSpace(os.Getenv("EMBEDDING_API_KEY"))
	llmKey := strings.TrimSpace(os.Getenv("LLM_API_KEY"))

	out := FieldAssistantHealth{
		LLMBaseURL:             llmBase,
		EmbeddingBaseURL:       embBase,
		EmbeddingConfigured:    embKey != "" || embBase != "",
		EmbeddingLocalEndpoint: embBase != "" && IsLocalInferenceURL(embBase),
		SplitInferenceHosts:    InferenceHostsSplit(),
		FieldGuideChunkCount:   fieldGuideChunks,
		PlatformDocChunkCount:  platformDocChunks,
	}
	out.FieldMode = llmBase != "" && IsLocalInferenceURL(llmBase)

	check := llmReachable
	if check == nil {
		check = ai.VerifyChatBackend
	}

	if llmBase == "" {
		out.LLMReachableError = "LLM_BASE_URL not set"
	} else if err := check(ctx, llmBase, llmKey); err != nil {
		out.LLMReachableError = err.Error()
	} else {
		out.LLMReachable = true
	}

	if embBase == "" {
		out.EmbeddingReachableError = "EMBEDDING_BASE_URL not set"
	} else if err := check(ctx, embBase, embKey); err != nil {
		out.EmbeddingReachableError = err.Error()
	} else {
		out.EmbeddingReachable = true
	}
	return out
}
