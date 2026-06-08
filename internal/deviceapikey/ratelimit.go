package deviceapikey

import (
	"sync"
	"time"
)

// RateLimiter caps requests per device key row (Phase 57 WS4).
type RateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	max     int
	buckets map[int64][]time.Time
}

func NewRateLimiter(maxPerMinute int) *RateLimiter {
	if maxPerMinute <= 0 {
		maxPerMinute = 120
	}
	return &RateLimiter{
		window:  time.Minute,
		max:     maxPerMinute,
		buckets: make(map[int64][]time.Time),
	}
}

func (l *RateLimiter) Allow(keyRowID int64) bool {
	now := time.Now()
	cutoff := now.Add(-l.window)
	l.mu.Lock()
	defer l.mu.Unlock()
	history := l.buckets[keyRowID]
	filtered := history[:0]
	for _, t := range history {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= l.max {
		l.buckets[keyRowID] = filtered
		return false
	}
	filtered = append(filtered, now)
	l.buckets[keyRowID] = filtered
	return true
}
