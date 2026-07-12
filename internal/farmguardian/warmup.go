// Phase 129 WS1–2 — Guardian warmup orchestration.

package farmguardian

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const warmupKeepAlive = "30m"

var (
	warmupMu    sync.Mutex
	warmupState warmupSnapshot
)

type warmupSnapshot struct {
	InProgress bool
	Mode       string
	ChatModel  string
	LastError  string
	StartedAt  time.Time
}

func snapshotWarmupState() warmupSnapshot {
	warmupMu.Lock()
	defer warmupMu.Unlock()
	return warmupState
}

// StartWarmup kicks off async model preload when not already warm or stirring.
// Returns current state: stirring (202), ready (200), or unavailable.
// requestModel is an optional override (POST /guardian/warmup chat_model, eval runner).
// When includeVision is true and LLM_VISION_MODEL is set, the vision model is preloaded after chat.
func StartWarmup(ctx context.Context, llmBaseURL, mode string, requestModel string, farmCounsel, farmQuick *string, envDefault string, cache *ModelCache, includeVision bool) (state string, chatModel string) {
	ClearDormantFlag()
	mode = normalizeWarmupMode(mode)
	chatModel, _, reject := ResolveWarmupModel(cache, mode, requestModel, farmCounsel, farmQuick, envDefault)
	if reject != "" {
		return AwakeningStateUnavailable, ""
	}
	if chatModel == "" {
		return AwakeningStateUnavailable, ""
	}

	field := BuildFieldAssistantHealth(ctx, nil, 0, 0)
	if !field.LLMReachable {
		return AwakeningStateUnavailable, chatModel
	}

	_, loadedMap, _ := probeOllamaRuntime(ctx, llmBaseURL)
	if chatLoaded, _ := psEntry(loadedMap, chatModel); chatLoaded {
		return AwakeningStateReady, chatModel
	}

	warmupMu.Lock()
	if warmupState.InProgress {
		state = AwakeningStateStirring
		chatModel = warmupState.ChatModel
		warmupMu.Unlock()
		return state, chatModel
	}
	warmupState = warmupSnapshot{
		InProgress: true,
		Mode:       mode,
		ChatModel:  chatModel,
		LastError:  "",
		StartedAt:  time.Now(),
	}
	warmupMu.Unlock()

	go runWarmup(context.WithoutCancel(ctx), llmBaseURL, mode, chatModel, includeVision)
	return AwakeningStateStirring, chatModel
}

func runWarmup(ctx context.Context, llmBaseURL, mode, chatModel string, includeVision bool) {
	var lastErr string
	defer func() {
		warmupMu.Lock()
		warmupState.InProgress = false
		warmupState.LastError = lastErr
		warmupMu.Unlock()
	}()

	embedModel := EmbedModelFromEnv()
	if mode == WarmupModeFarmCounsel {
		MaybeUnloadEmbedForChat(ctx, llmBaseURL, embedModel, chatModel)
	}

	if err := preloadOllamaChatModel(ctx, llmBaseURL, chatModel, warmupKeepAlive); err != nil {
		lastErr = err.Error()
		slog.Warn("guardian: warmup preload failed", "chat_model", chatModel, "mode", mode, "err", err)
		return
	}

	_, loadedMap, _ := probeOllamaRuntime(ctx, llmBaseURL)
	if ok, _ := psEntry(loadedMap, chatModel); !ok {
		lastErr = "chat model did not stay loaded after warmup"
		slog.Warn("guardian: warmup verify failed", "chat_model", chatModel)
		return
	}
	slog.Info("guardian: warmup ready", "chat_model", chatModel, "mode", mode)
	NoteGuardianActivity(chatModel)

	if includeVision {
		visionModel := VisionModelFromEnv()
		if visionModel != "" && visionModel != chatModel {
			if err := preloadOllamaChatModel(ctx, llmBaseURL, visionModel, warmupKeepAlive); err != nil {
				slog.Warn("guardian: vision warmup preload failed", "vision_model", visionModel, "err", err)
			}
		}
	}
}
