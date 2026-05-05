package rag

import (
	"os"
	"strconv"
	"sync"
	"time"
)

// synthesisLimiter gates POST /rag/answer (Phase 25 WS5 — optional per-farm fairness).
type synthesisLimiter interface {
	Allow(farmID int64) bool
}

type globalSynthLimiter struct {
	inner *minuteLimiter
}

func (g *globalSynthLimiter) Allow(farmID int64) bool {
	_ = farmID
	return g.inner.Allow()
}

type perFarmSynthLimiter struct {
	mu      sync.Mutex
	maxPer  int
	perFarm map[int64]*minuteLimiter
}

func newPerFarmSynthLimiter(maxPerMinute int) *perFarmSynthLimiter {
	return &perFarmSynthLimiter{
		maxPer:  maxPerMinute,
		perFarm: make(map[int64]*minuteLimiter),
	}
}

func (p *perFarmSynthLimiter) Allow(farmID int64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	lim, ok := p.perFarm[farmID]
	if !ok {
		lim = &minuteLimiter{maxPerMinute: p.maxPer, windowStart: time.Now()}
		p.perFarm[farmID] = lim
	}
	return lim.Allow()
}

// newSynthGateFromEnv chooses per-farm buckets when RAG_SYNTHESIS_MAX_PER_MINUTE_PER_FARM is set
// (integer > 0); otherwise uses process-wide RAG_SYNTHESIS_MAX_PER_MINUTE (default 30).
func newSynthGateFromEnv() synthesisLimiter {
	if s := os.Getenv("RAG_SYNTHESIS_MAX_PER_MINUTE_PER_FARM"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			return newPerFarmSynthLimiter(v)
		}
	}
	return &globalSynthLimiter{inner: newSynthLimiterFromEnv()}
}

// allowAllSynth is for tests only (no rate limit).
type allowAllSynth struct{}

func (allowAllSynth) Allow(int64) bool { return true }

// NewTestSynthGlobalLimiter returns a process-wide synthesis gate with an explicit per-minute cap (tests only).
func NewTestSynthGlobalLimiter(maxPerMinute int) synthesisLimiter {
	return &globalSynthLimiter{inner: &minuteLimiter{
		maxPerMinute: maxPerMinute,
		windowStart:  time.Now(),
	}}
}
