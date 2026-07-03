package farmguardian

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
)

// GuardianMinContextWindow is the minimum reported context length for grounded chat.
const GuardianMinContextWindow = 8192

// ModelCache holds the last Ollama /api/tags discovery snapshot (server-wide).
type ModelCache struct {
	mu            sync.RWMutex
	models        []ModelInfo
	byName        map[string]ModelInfo
	serverDefault string
}

func NewModelCache() *ModelCache {
	return &ModelCache{byName: make(map[string]ModelInfo)}
}

// RefreshFromEnv re-queries Ollama using LLM_BASE_URL.
func (c *ModelCache) RefreshFromEnv(ctx context.Context) error {
	base := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	if base == "" {
		c.mu.Lock()
		c.models = nil
		c.byName = make(map[string]ModelInfo)
		c.serverDefault = EnvServerDefaultModel()
		c.mu.Unlock()
		return nil
	}
	models, err := DiscoverOllamaModels(ctx, base, http.DefaultClient)
	if err != nil {
		return err
	}
	c.Set(models, EnvServerDefaultModel())
	slog.Info("guardian: discovered ollama models", "count", len(models))
	return nil
}

// Set replaces the cache contents (tests and manual refresh).
func (c *ModelCache) Set(models []ModelInfo, serverDefault string) {
	byName := make(map[string]ModelInfo, len(models))
	for _, m := range models {
		byName[m.Name] = m
	}
	c.mu.Lock()
	c.models = append([]ModelInfo(nil), models...)
	c.byName = byName
	c.serverDefault = strings.TrimSpace(serverDefault)
	c.mu.Unlock()
}

func (c *ModelCache) Snapshot() (models []ModelInfo, serverDefault string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.models) == 0 {
		return nil, c.serverDefault
	}
	out := make([]ModelInfo, len(c.models))
	copy(out, c.models)
	return out, c.serverDefault
}

func (c *ModelCache) ServerDefault() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serverDefault
}

func (c *ModelCache) Get(name string) (ModelInfo, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ModelInfo{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.byName[name]
	return m, ok
}

func (c *ModelCache) Contains(name string) bool {
	_, ok := c.Get(name)
	return ok
}

// ResolveOutcome is the result of choosing a chat model for one turn.
type ResolveOutcome struct {
	ModelName    string
	Fallback     bool
	RejectReason string
}

// ResolveChatModel picks session → farm → env and validates against the cache.
func ResolveChatModel(cache *ModelCache, sessionModel string, farmPreferred *string, envDefault string, grounded bool) ResolveOutcome {
	envDefault = strings.TrimSpace(envDefault)
	requested := strings.TrimSpace(sessionModel)
	if requested == "" && farmPreferred != nil {
		requested = strings.TrimSpace(*farmPreferred)
	}
	if requested == "" {
		requested = envDefault
	}
	if requested == "" {
		return ResolveOutcome{RejectReason: "no chat model configured (set LLM_MODEL)"}
	}

	try := func(name string) (ResolveOutcome, bool) {
		name = strings.TrimSpace(name)
		if name == "" {
			return ResolveOutcome{}, false
		}
		if cache == nil {
			return ResolveOutcome{ModelName: name}, true
		}
		info, ok := cache.Get(name)
		if !ok {
			return ResolveOutcome{}, false
		}
		if grounded && info.ContextWindow > 0 && info.ContextWindow < GuardianMinContextWindow {
			return ResolveOutcome{
				RejectReason: formatContextReject(name, info.ContextWindow),
			}, true
		}
		return ResolveOutcome{ModelName: name}, true
	}

	if out, ok := try(requested); ok {
		return out
	}

	if envDefault != "" && envDefault != requested {
		if out, ok := try(envDefault); ok {
			out.Fallback = true
			return out
		}
	}

	// Unknown model and no cache match — still attempt env default path for bare installs.
	if cache == nil || len(cache.byNameSnapshot()) == 0 {
		return ResolveOutcome{ModelName: requested}
	}

	if envDefault != "" {
		return ResolveOutcome{ModelName: envDefault, Fallback: true}
	}
	return ResolveOutcome{ModelName: requested, Fallback: true}
}

func (c *ModelCache) byNameSnapshot() map[string]ModelInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.byName) == 0 {
		return nil
	}
	out := make(map[string]ModelInfo, len(c.byName))
	for k, v := range c.byName {
		out[k] = v
	}
	return out
}

func formatContextReject(name string, window int) string {
	return fmt.Sprintf(
		"Model %q context window (%d) is below the minimum required for grounded Guardian chat (%d). Switch to a larger model or use non-grounded chat.",
		name, window, GuardianMinContextWindow,
	)
}
