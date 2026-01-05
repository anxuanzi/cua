package safety

import (
	"sync"
	"time"
)

// RateLimiter controls the rate of actions to prevent overwhelming the system.
// It uses a sliding window approach to track actions per minute.
type RateLimiter struct {
	mu sync.Mutex

	// maxPerMinute is the maximum actions allowed per minute.
	maxPerMinute int

	// actions is a slice of timestamps for recent actions.
	actions []time.Time

	// windowDuration is how long to track actions (1 minute).
	windowDuration time.Duration
}

// NewRateLimiter creates a new rate limiter with the specified limit.
func NewRateLimiter(maxPerMinute int) *RateLimiter {
	if maxPerMinute <= 0 {
		maxPerMinute = 60 // default
	}
	return &RateLimiter{
		maxPerMinute:   maxPerMinute,
		actions:        make([]time.Time, 0, maxPerMinute),
		windowDuration: time.Minute,
	}
}

// Allow checks if an action is allowed and records it if so.
// Returns true if the action is allowed, false if rate limited.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.pruneExpired(now)

	if len(r.actions) >= r.maxPerMinute {
		return false
	}

	r.actions = append(r.actions, now)
	return true
}

// Wait blocks until an action is allowed.
// Returns the time waited.
func (r *RateLimiter) Wait() time.Duration {
	start := time.Now()

	for {
		r.mu.Lock()
		now := time.Now()
		r.pruneExpired(now)

		if len(r.actions) < r.maxPerMinute {
			r.actions = append(r.actions, now)
			r.mu.Unlock()
			return time.Since(start)
		}

		// Calculate how long until the oldest action expires
		oldest := r.actions[0]
		waitTime := oldest.Add(r.windowDuration).Sub(now)
		r.mu.Unlock()

		if waitTime > 0 {
			time.Sleep(waitTime)
		}
	}
}

// Available returns the number of actions available in the current window.
func (r *RateLimiter) Available() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pruneExpired(time.Now())
	return r.maxPerMinute - len(r.actions)
}

// Reset clears all tracked actions.
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actions = r.actions[:0]
}

// pruneExpired removes actions outside the time window.
// Must be called with lock held.
func (r *RateLimiter) pruneExpired(now time.Time) {
	cutoff := now.Add(-r.windowDuration)
	i := 0
	for i < len(r.actions) && r.actions[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		r.actions = r.actions[i:]
	}
}
