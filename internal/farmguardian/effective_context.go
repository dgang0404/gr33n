package farmguardian

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// Built-in effective context caps for models whose Ollama /api/show advertises
// rope-extended windows far above CPU runtime (Phase 126).
var builtinEffectiveContextOverrides = map[string]int{
	"phi3:mini":  4096,
	"tinyllama":  2048,
}

var (
	effectiveOverrideOnce sync.Once
	effectiveOverrideEnv  map[string]int
)

func loadEffectiveContextOverrides() map[string]int {
	effectiveOverrideOnce.Do(func() {
		effectiveOverrideEnv = parseEffectiveContextOverrides(os.Getenv("GUARDIAN_EFFECTIVE_CONTEXT_OVERRIDES"))
	})
	out := make(map[string]int, len(builtinEffectiveContextOverrides)+len(effectiveOverrideEnv))
	for k, v := range builtinEffectiveContextOverrides {
		out[k] = v
	}
	for k, v := range effectiveOverrideEnv {
		out[k] = v
	}
	return out
}

// parseEffectiveContextOverrides parses "phi3:mini=4096,tinyllama=2048".
func parseEffectiveContextOverrides(raw string) map[string]int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	out := make(map[string]int)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil && n > 0 {
			out[k] = n
		}
	}
	return out
}

func effectiveContextOverride(name string) (int, bool) {
	overrides := loadEffectiveContextOverrides()
	for _, key := range modelLookupKeys(name) {
		bare := NormalizeModelName(key)
		if cap, ok := overrides[key]; ok {
			return cap, true
		}
		if cap, ok := overrides[bare]; ok {
			return cap, true
		}
	}
	return 0, false
}

// parseOriginalContextLength reads rope.scaling.original_context_length from Ollama model_info.
func parseOriginalContextLength(modelInfo map[string]any) int {
	if len(modelInfo) == 0 {
		return 0
	}
	min := 0
	for k, v := range modelInfo {
		if !strings.HasSuffix(k, ".rope.scaling.original_context_length") {
			continue
		}
		n := jsonNumberInt(v)
		if n <= 0 {
			continue
		}
		if min == 0 || n < min {
			min = n
		}
	}
	return min
}

// ResolveEffectiveContextWindow picks the runtime budget for prompt trimming.
// Advertised (max *.context_length) is kept separately for display and the 8192 gate.
func ResolveEffectiveContextWindow(name string, advertised int, modelInfo map[string]any) int {
	if cap, ok := effectiveContextOverride(name); ok {
		return cap
	}
	original := parseOriginalContextLength(modelInfo)
	if original > 0 {
		if advertised > 0 {
			if original < advertised {
				return original
			}
			return advertised
		}
		return original
	}
	return advertised
}

// PromptBudgetContextWindow returns the window used for ComputePromptBudget.
func PromptBudgetContextWindow(info ModelInfo) int {
	if info.EffectiveContextWindow > 0 {
		return info.EffectiveContextWindow
	}
	return info.ContextWindow
}
