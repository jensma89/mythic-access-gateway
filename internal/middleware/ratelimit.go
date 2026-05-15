// ratelimit.go
// Per-IP rate limiting

package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ipLimiter holds a rate limiter and the last time it was seen.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter manages per-IP rate limiters with automatic cleanup.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	r        rate.Limit // requests per second
	burst    int        // max burst size
}

// NewRateLimiter creates a RateLimiter and starts a background cleanup goroutine.
// rps - sustained requests per second per IP
// burst - maximum burst above the sustained rate
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*ipLimiter),
		r:        rate.Limit(rps),
		burst:    burst,
	}
	go rl.cleanup()
	return rl
}

// getLimiter returns the rate limiter for the given IP, creating one if needed.
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, ok := rl.limiters[ip]
	if !ok {
		entry = &ipLimiter{limiter: rate.NewLimiter(rl.r, rl.burst)}
		rl.limiters[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

// cleanup removes limiters for IPs not seen in the last 10 minutes
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, entry := range rl.limiters {
			if time.Since(entry.lastSeen) > 10*time.Minute {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit returns HTTP middleware that enforces the per-IP rate limit.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		if !rl.getLimiter(ip).Allow() {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// realIP extracts the client IP, respecting common proxy headers.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
