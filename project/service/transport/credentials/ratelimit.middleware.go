package credentials

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// rateLimiter is a simple in-memory, per-process sliding-window limiter
// keyed by an arbitrary string (here, client IP). It is deliberately not a
// distributed limiter -- this app runs as a single instance today, and the
// bar to clear is "some cap" vs the previous "none at all" on the
// authentication endpoints (see RateLimitMiddleware).
type rateLimiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	visitors map[string][]time.Time
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{max: max, window: window, visitors: make(map[string][]time.Time)}
	go rl.cleanupLoop()
	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-rl.window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	kept := rl.visitors[key][:0]
	for _, t := range rl.visitors[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= rl.max {
		rl.visitors[key] = kept
		return false
	}
	rl.visitors[key] = append(kept, now)
	return true
}

// cleanupLoop periodically drops visitors with no requests left inside the
// window, so long-running processes don't accumulate one map entry per IP
// ever seen.
func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		cutoff := time.Now().Add(-rl.window)
		rl.mu.Lock()
		for key, times := range rl.visitors {
			kept := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					kept = append(kept, t)
				}
			}
			if len(kept) == 0 {
				delete(rl.visitors, key)
			} else {
				rl.visitors[key] = kept
			}
		}
		rl.mu.Unlock()
	}
}

// clientIP prefers a leftmost X-Forwarded-For entry (set by a reverse
// proxy/load balancer in front of this app) and falls back to the raw
// connection's RemoteAddr. Not spoof-proof against an attacker who can set
// arbitrary headers directly against this process, but that's true of any
// XFF-based scheme without an explicit trusted-proxy allowlist -- the goal
// here is a real cap on the common case, not a bulletproof one.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first := strings.TrimSpace(strings.Split(xff, ",")[0]); first != "" {
			return first
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// RateLimitMiddleware returns 429 once the calling IP has made more than
// max requests to the wrapped route within window. Meant for
// authentication-sensitive public endpoints (login, password reset/
// recovery, register, refresh-token) which previously had no request cap
// at all, making them brute-forceable/enumerable/spammable without limit.
// Each call creates its own limiter+state, so wrap each route (or route
// group) once at startup, not per-request.
func RateLimitMiddleware(max int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(max, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(clientIP(r)) {
				http.Error(w, "too many requests, try again later", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
