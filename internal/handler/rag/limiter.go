package rag

import (
	"os"
	"strconv"
	"sync"
	"time"
)

type minuteLimiter struct {
	mu           sync.Mutex
	windowStart  time.Time
	count        int
	maxPerMinute int
}

func newSynthLimiterFromEnv() *minuteLimiter {
	n := 30
	if s := os.Getenv("RAG_SYNTHESIS_MAX_PER_MINUTE"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			n = v
		}
	}
	return &minuteLimiter{maxPerMinute: n}
}

// Allow reports whether one request may proceed (fixed window per minute).
func (w *minuteLimiter) Allow() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	if now.Sub(w.windowStart) >= time.Minute {
		w.windowStart = now
		w.count = 0
	}
	if w.count >= w.maxPerMinute {
		return false
	}
	w.count++
	return true
}
