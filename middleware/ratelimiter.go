package middleware

import (
	"encoding/json"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// 5 requests per second, burst of 10 per IP
const (
	requestsPerSecond = 5
	burst             = 10
)

type ipLimiter struct {
	limiter *rate.Limiter
}

type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*ipLimiter),
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if l, exists := rl.limiters[ip]; exists {
		return l.limiter
	}

	l := &ipLimiter{limiter: rate.NewLimiter(requestsPerSecond, burst)}
	rl.limiters[ip] = l
	return l.limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		if !rl.getLimiter(ip).Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rate limit exceeded, slow down",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
