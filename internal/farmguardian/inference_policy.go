package farmguardian

import (
	"os"
	"strings"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/llm"
)

// FarmCounselModel returns the grounded counsel model for a farm (Phase 138).
// Falls back to guardian_preferred_model for farms not yet migrated in UI.
func FarmCounselModel(farm *db.Gr33ncoreFarm) *string {
	if farm == nil {
		return nil
	}
	if s := trimModelPtr(farm.GuardianCounselModel); s != "" {
		return &s
	}
	return farm.GuardianPreferredModel
}

// FarmQuickModel returns the quick-chat model for a farm (Phase 138).
func FarmQuickModel(farm *db.Gr33ncoreFarm) *string {
	if farm == nil {
		return nil
	}
	if s := trimModelPtr(farm.GuardianQuickModel); s != "" {
		return &s
	}
	return nil
}

// FarmGroundedChatTimeout returns per-farm grounded timeout or env default.
func FarmGroundedChatTimeout(farm *db.Gr33ncoreFarm) time.Duration {
	if farm != nil && farm.GuardianGroundedTimeoutSeconds != nil && *farm.GuardianGroundedTimeoutSeconds > 0 {
		return time.Duration(*farm.GuardianGroundedTimeoutSeconds) * time.Second
	}
	return llm.GroundedChatTimeoutFromEnv()
}

func trimModelPtr(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

// InferenceHostsSplit reports whether chat and embed use different base URLs.
func InferenceHostsSplit() bool {
	llm := normalizeInferenceBase(os.Getenv("LLM_BASE_URL"))
	emb := normalizeInferenceBase(os.Getenv("EMBEDDING_BASE_URL"))
	if emb == "" {
		return false
	}
	return llm != emb
}

func normalizeInferenceBase(base string) string {
	return strings.TrimSuffix(strings.TrimSpace(base), "/")
}
