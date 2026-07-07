// Phase 129 WS0 — Guardian awakening health for GET /v1/chat/health.

package farmguardian

import (
	"context"
	"os"
	"strings"
)

const (
	AwakeningStateUnavailable = "unavailable"
	AwakeningStateSleeping    = "sleeping"
	AwakeningStateStirring    = "stirring"
	AwakeningStateReady       = "ready"
	AwakeningStateBusy        = "busy"

	AwakeningProfileCPULaptop = "cpu_laptop"
	AwakeningProfileGPUServer = "gpu_server"
	AwakeningProfileLite      = "lite"

	WarmupModeQuick       = "quick"
	WarmupModeFarmCounsel = "farm_counsel"
)

// AwakeningHealth is the operator-facing Guardian readiness block (Phase 129).
type AwakeningHealth struct {
	State              string   `json:"state"`
	Profile            string   `json:"profile"`
	ChatModel          string   `json:"chat_model,omitempty"`
	ChatModelLoaded    bool     `json:"chat_model_loaded"`
	VisionModel        string   `json:"vision_model,omitempty"`
	VisionModelLoaded  bool     `json:"vision_model_loaded"`
	EmbedModel         string   `json:"embed_model,omitempty"`
	EmbedLoaded        bool     `json:"embed_loaded"`
	EmbedBlocksChat    bool     `json:"embed_blocks_chat"`
	OllamaLoadedModels []string `json:"ollama_loaded_models,omitempty"`
	RagCorpusOK        bool          `json:"rag_corpus_ok"`
	FieldGuideChunks   int64         `json:"field_guide_chunks"`
	PlatformDocChunks  int64         `json:"platform_doc_chunks"`
	Corpus             *CorpusHealth `json:"corpus,omitempty"`
	Messages           []string      `json:"messages,omitempty"`
	WarmupInProgress   bool     `json:"warmup_in_progress"`
	LastWarmupError    string   `json:"last_warmup_error,omitempty"`
	StaleOllamaCLI     bool     `json:"stale_ollama_cli,omitempty"`
}

// AwakeningBuildInput collects probes for BuildAwakeningHealth.
type AwakeningBuildInput struct {
	AIEnabled          bool
	Field              FieldAssistantHealth
	Mode               string
	FarmID             int64
	FieldGuideChunks   int64
	PlatformDocChunks  int64
	Corpus             *CorpusHealth
	Cache              *ModelCache
	FarmCounselModel   *string
	FarmQuickModel     *string
	EnvDefault         string
}

// BuildAwakeningHealth assembles the awakening block for GET /v1/chat/health.
func BuildAwakeningHealth(ctx context.Context, in AwakeningBuildInput) AwakeningHealth {
	mode := normalizeWarmupMode(in.Mode)
	envDefault := strings.TrimSpace(in.EnvDefault)
	if envDefault == "" {
		envDefault = EnvServerDefaultModel()
	}

	out := AwakeningHealth{
		FieldGuideChunks:  in.FieldGuideChunks,
		PlatformDocChunks: in.PlatformDocChunks,
		RagCorpusOK:       in.FieldGuideChunks > 0 || in.PlatformDocChunks > 0,
		Corpus:            in.Corpus,
		EmbedModel:        EmbedModelFromEnv(),
	}

	if !in.AIEnabled {
		out.State = AwakeningStateUnavailable
		out.Profile = AwakeningProfileLite
		out.Messages = []string{"Guardian is in Lite mode — Pi and dashboard only."}
		return out
	}
	if !in.Field.LLMReachable {
		out.State = AwakeningStateUnavailable
		out.Profile = detectInferenceProfile(ctx, in.Field.LLMBaseURL)
		if in.Field.LLMReachableError != "" {
			out.Messages = []string{"Ollama not reachable — start Ollama, then retry awakening."}
			out.LastWarmupError = in.Field.LLMReachableError
		}
		return out
	}

	chatModel, _, reject := ResolveWarmupModel(in.Cache, mode, "", in.FarmCounselModel, in.FarmQuickModel, envDefault)
	out.ChatModel = chatModel
	if reject != "" && mode == WarmupModeFarmCounsel {
		out.Messages = append(out.Messages, reject)
	}

	loadedNames, loadedMap, profile := probeOllamaRuntime(ctx, in.Field.LLMBaseURL)
	out.Profile = profile
	out.OllamaLoadedModels = loadedNames

	if DetectStaleOllamaCLI(ctx, in.Field.LLMBaseURL) {
		out.StaleOllamaCLI = true
		out.Messages = append(out.Messages, staleOllamaMessage)
	}

	if out.EmbedModel != "" && !InferenceHostsSplit() {
		out.EmbedLoaded, _ = psEntry(loadedMap, out.EmbedModel)
	}
	if chatModel != "" {
		out.ChatModelLoaded, _ = psEntry(loadedMap, chatModel)
	}
	if visionModel := VisionModelFromEnv(); visionModel != "" {
		out.VisionModel = visionModel
		out.VisionModelLoaded, _ = psEntry(loadedMap, visionModel)
	}
	if out.EmbedLoaded && !out.ChatModelLoaded && out.EmbedModel != "" && !InferenceHostsSplit() {
		_, embedCPU := psEntry(loadedMap, out.EmbedModel)
		if embedCPU || !out.ChatModelLoaded {
			out.EmbedBlocksChat = true
			out.Messages = append(out.Messages, "Embedding model is using RAM — awakening will make room for chat.")
		}
	}

	warm := snapshotWarmupState()
	out.WarmupInProgress = warm.InProgress
	if warm.LastError != "" {
		out.LastWarmupError = warm.LastError
	}

	if warm.InProgress {
		out.State = AwakeningStateStirring
		return out
	}
	if GroundedChatBusy() {
		out.State = AwakeningStateBusy
		return out
	}
	if out.ChatModelLoaded {
		out.State = AwakeningStateReady
		return out
	}
	out.State = AwakeningStateSleeping
	if mode == WarmupModeFarmCounsel && in.FarmID > 0 {
		if in.Corpus != nil {
			out.Messages = append(out.Messages, CorpusWarningMessages(*in.Corpus, mode)...)
		} else if !out.RagCorpusOK {
			out.Messages = append(out.Messages, "Field memories not ingested — run make guardian-bootstrap-farm.")
		}
	}
	return out
}

func normalizeWarmupMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case WarmupModeQuick:
		return WarmupModeQuick
	default:
		return WarmupModeFarmCounsel
	}
}

// ResolveWarmupModel picks the chat model to preload for a warmup mode.
// requestModel is an optional caller override (e.g. guardian-eval -models).
func ResolveWarmupModel(cache *ModelCache, mode string, requestModel string, farmCounsel, farmQuick *string, envDefault string) (model string, grounded bool, reject string) {
	mode = normalizeWarmupMode(mode)
	grounded = mode == WarmupModeFarmCounsel
	var farmPref *string
	if grounded {
		farmPref = farmCounsel
	} else {
		farmPref = farmQuick
	}
	out := ResolveChatModel(cache, requestModel, farmPref, envDefault, grounded)
	if out.RejectReason != "" {
		return "", grounded, out.RejectReason
	}
	return out.ModelName, grounded, ""
}

func detectInferenceProfile(ctx context.Context, llmBaseURL string) string {
	_, _, profile := probeOllamaRuntime(ctx, llmBaseURL)
	if profile != "" {
		return profile
	}
	if IsLocalInferenceURL(llmBaseURL) {
		return AwakeningProfileCPULaptop
	}
	return AwakeningProfileGPUServer
}

func probeOllamaRuntime(ctx context.Context, llmBaseURL string) (names []string, loaded map[string]ollamaPsModel, profile string) {
	loaded, err := listOllamaPS(ctx, OllamaNativeBase(llmBaseURL), nil)
	if err != nil || len(loaded) == 0 {
		if strings.TrimSpace(os.Getenv("GUARDIAN_INFERENCE_PROFILE")) == "gpu-server" {
			return nil, loaded, AwakeningProfileGPUServer
		}
		return nil, loaded, AwakeningProfileCPULaptop
	}
	seen := make(map[string]struct{})
	allCPU := true
	for name, m := range loaded {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
		if m.SizeVRAM > 0 {
			allCPU = false
		}
	}
	if allCPU {
		profile = AwakeningProfileCPULaptop
	} else {
		profile = AwakeningProfileGPUServer
	}
	return names, loaded, profile
}
