package authsecurity

import (
	"sync"
	"time"
)

// LoginLimiter caps login attempts per (IP, username) sliding window.
type LoginLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	max     int
	buckets map[string][]time.Time
}

func NewLoginLimiter(maxPerMinute int) *LoginLimiter {
	if maxPerMinute <= 0 {
		maxPerMinute = 10
	}
	return &LoginLimiter{
		window:  time.Minute,
		max:     maxPerMinute,
		buckets: make(map[string][]time.Time),
	}
}

func loginBucketKey(ip, username string) string {
	return ip + "\x00" + username
}

// Allow returns false when the bucket is over the cap.
func (l *LoginLimiter) Allow(ip, username string) bool {
	key := loginBucketKey(ip, username)
	now := time.Now()
	cutoff := now.Add(-l.window)
	l.mu.Lock()
	defer l.mu.Unlock()
	history := l.buckets[key]
	filtered := history[:0]
	for _, t := range history {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= l.max {
		l.buckets[key] = filtered
		return false
	}
	filtered = append(filtered, now)
	l.buckets[key] = filtered
	return true
}

// RetryAfter returns time until the oldest attempt in the window expires.
func (l *LoginLimiter) RetryAfter(ip, username string) time.Duration {
	key := loginBucketKey(ip, username)
	cutoff := time.Now().Add(-l.window)
	l.mu.Lock()
	defer l.mu.Unlock()
	history := l.buckets[key]
	var oldest time.Time
	for _, t := range history {
		if t.After(cutoff) && (oldest.IsZero() || t.Before(oldest)) {
			oldest = t
		}
	}
	if oldest.IsZero() {
		return l.window
	}
	until := oldest.Add(l.window).Sub(time.Now())
	if until < time.Second {
		return time.Second
	}
	return until
}
