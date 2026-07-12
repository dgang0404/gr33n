// Phase 163 WS1 — deliberate Guardian rest state (power saving on solar/battery sites).
//
// RequestDormant unloads the chat (and optional vision) model from Ollama's
// RAM/VRAM — the same keep_alive:0 mechanism Phase 130 uses for embed
// contention — and records that the operator asked for this, so
// BuildAwakeningHealth can report "dormant" (resting on purpose) instead of
// "sleeping" (never warmed) until the next warmup wakes it back up.
//
// This is only ever set by an explicit operator action (POST
// /guardian/dormant). Nothing in the chat pipeline calls this per-turn —
// that would reintroduce the cold-start-every-message problem Phase 129
// exists to avoid.
package farmguardian

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	dormantMu        sync.Mutex
	dormantRequested bool
	dormantAt        time.Time
)

// RequestDormant unloads the given chat model (and vision model, if set and
// different) from Ollama. Returns an error only if the chat model unload
// fails — vision unload is best-effort.
func RequestDormant(ctx context.Context, llmBaseURL, chatModel, visionModel string) error {
	chatModel = strings.TrimSpace(chatModel)
	if chatModel == "" {
		return fmt.Errorf("empty chat model")
	}
	base := OllamaNativeBase(llmBaseURL)
	if base == "" {
		return fmt.Errorf("not an Ollama base URL")
	}
	if err := unloadOllamaModel(ctx, base, chatModel, nil); err != nil {
		return err
	}
	visionModel = strings.TrimSpace(visionModel)
	if visionModel != "" && visionModel != chatModel {
		_ = unloadOllamaModel(ctx, base, visionModel, nil)
	}
	dormantMu.Lock()
	dormantRequested = true
	dormantAt = time.Now()
	dormantMu.Unlock()
	return nil
}

// ClearDormantFlag marks Guardian as no longer deliberately resting — called
// at the start of any wake path (StartWarmup) so awakening health stops
// reporting "dormant" once the operator asks Guardian to wake up.
func ClearDormantFlag() {
	dormantMu.Lock()
	dormantRequested = false
	dormantMu.Unlock()
}

// snapshotDormantState reports whether Guardian is currently in a
// deliberate rest state and when that was requested.
func snapshotDormantState() (bool, time.Time) {
	dormantMu.Lock()
	defer dormantMu.Unlock()
	return dormantRequested, dormantAt
}
