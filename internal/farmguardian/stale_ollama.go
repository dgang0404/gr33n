// Phase 130 WS6 — detect stray `ollama run` CLI sessions blocking the daemon.

package farmguardian

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// staleOllamaRunCount is overridden in tests.
var staleOllamaRunCount = defaultStaleOllamaRunCount

func defaultStaleOllamaRunCount(ctx context.Context) int {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pgrep", "-f", `ollama run`)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	n := 0
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

// DetectStaleOllamaCLI reports orphan terminal `ollama run` while Ollama ps is empty.
func DetectStaleOllamaCLI(ctx context.Context, llmBaseURL string) bool {
	base := OllamaNativeBase(llmBaseURL)
	if base == "" {
		return false
	}
	loaded, err := listOllamaPS(ctx, base, nil)
	if err == nil && len(loaded) > 0 {
		return false
	}
	return staleOllamaRunCount(ctx) > 0
}

const staleOllamaMessage = "Close stray terminal ollama run sessions — they can block the Ollama service for hours."
