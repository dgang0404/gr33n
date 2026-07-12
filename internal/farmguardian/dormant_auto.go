// Phase 163 WS3 — optional auto-dormant after idle minutes (solar/battery sites).

package farmguardian

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const autoDormantCheckInterval = time.Minute

var (
	activityMu              sync.Mutex
	lastActivityAt          time.Time
	lastActivityChatModel   string
	lastActivityVisionModel string
)

// AutoDormantMinutesFromEnv returns how long Guardian may stay warm with no
// chat turns before auto-rest. Zero disables (default).
func AutoDormantMinutesFromEnv() time.Duration {
	s := strings.TrimSpace(os.Getenv("GUARDIAN_AUTO_DORMANT_MINUTES"))
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0
	}
	return time.Duration(n) * time.Minute
}

// NoteGuardianActivity records a successful warmup or chat turn — starts or
// resets the idle clock and clears any deliberate rest state.
func NoteGuardianActivity(chatModel string) {
	chatModel = strings.TrimSpace(chatModel)
	if chatModel == "" {
		return
	}
	activityMu.Lock()
	lastActivityAt = time.Now()
	lastActivityChatModel = chatModel
	lastActivityVisionModel = ""
	if vm := VisionModelFromEnv(); vm != "" && vm != chatModel {
		lastActivityVisionModel = vm
	}
	activityMu.Unlock()
	ClearDormantFlag()
}

func snapshotActivity() (at time.Time, chatModel, visionModel string) {
	activityMu.Lock()
	defer activityMu.Unlock()
	return lastActivityAt, lastActivityChatModel, lastActivityVisionModel
}

// AutoDormantIdleRemaining reports whether auto-rest is enabled and how long
// until the idle timer would unload the warm model (0 when due now).
func AutoDormantIdleRemaining() (enabled bool, remaining time.Duration) {
	limit := AutoDormantMinutesFromEnv()
	if limit <= 0 {
		return false, 0
	}
	at, chatModel, _ := snapshotActivity()
	if chatModel == "" || at.IsZero() {
		return true, limit
	}
	elapsed := time.Since(at)
	if elapsed >= limit {
		return true, 0
	}
	return true, limit - elapsed
}

// MaybeAutoDormant unloads the warm chat model when idle longer than
// GUARDIAN_AUTO_DORMANT_MINUTES. Safe to call from health polls or the
// background loop — no-ops when disabled, busy, stirring, or already resting.
func MaybeAutoDormant(ctx context.Context) (bool, error) {
	idleLimit := AutoDormantMinutesFromEnv()
	if idleLimit <= 0 {
		return false, nil
	}
	if GroundedChatBusy() {
		return false, nil
	}
	if warm := snapshotWarmupState(); warm.InProgress {
		return false, nil
	}
	if requested, _, _ := snapshotDormantState(); requested {
		return false, nil
	}

	at, chatModel, visionModel := snapshotActivity()
	if chatModel == "" || at.IsZero() {
		return false, nil
	}
	if time.Since(at) < idleLimit {
		return false, nil
	}

	llmBase := LLMBaseURLFromEnv()
	if llmBase == "" {
		return false, nil
	}
	_, loadedMap, _ := probeOllamaRuntime(ctx, llmBase)
	if ok, _ := psEntry(loadedMap, chatModel); !ok {
		return false, nil
	}

	if err := RequestDormant(ctx, llmBase, chatModel, visionModel, true); err != nil {
		return false, err
	}
	slog.Info("guardian: auto-dormant after idle",
		"idle_minutes", int(idleLimit/time.Minute),
		"chat_model", chatModel,
	)
	return true, nil
}

// StartAutoDormantLoop runs a background idle check when
// GUARDIAN_AUTO_DORMANT_MINUTES is set. Works even when no browser tab is open.
func StartAutoDormantLoop(ctx context.Context) {
	if AutoDormantMinutesFromEnv() <= 0 {
		return
	}
	slog.Info("guardian: auto-dormant loop enabled",
		"idle_minutes", int(AutoDormantMinutesFromEnv()/time.Minute),
	)
	go func() {
		ticker := time.NewTicker(autoDormantCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				_, _ = MaybeAutoDormant(runCtx)
				cancel()
			}
		}
	}()
}
