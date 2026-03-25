package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter provides per-IP and per-user rate limiting using Redis sliding
// window counters.
type RateLimiter struct {
	redis         *redis.Client
	ipLimit       int           // requests per window for unauthenticated users
	userLimit     int           // requests per window for authenticated users
	windowSeconds int           // window duration in seconds
	window        time.Duration // window duration
}

// NewRateLimiter creates a rate limiter.
// ipLimit is the per-IP rate (unauthenticated), userLimit is the per-user rate (authenticated).
// window is the sliding window duration.
func NewRateLimiter(redisClient *redis.Client, ipLimit, userLimit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:         redisClient,
		ipLimit:       ipLimit,
		userLimit:     userLimit,
		windowSeconds: int(window.Seconds()),
		window:        window,
	}
}

// Middleware returns an http middleware that enforces rate limits.
// Authenticated users (identified by ctxUserID) get the higher user limit.
// Unauthenticated requests use the client IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Determine identifier and limit
		var key string
		var limit int

		userID := UserIDFromContext(ctx)
		if userID.String() != "00000000-0000-0000-0000-000000000000" {
			key = fmt.Sprintf("rl:user:%s", userID.String())
			limit = rl.userLimit
		} else {
			ip := extractIP(r)
			key = fmt.Sprintf("rl:ip:%s", ip)
			limit = rl.ipLimit
		}

		count, err := rl.increment(ctx, key)
		if err != nil {
			// On Redis failure, allow the request (fail open)
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		remaining := limit - count
		if remaining < 0 {
			remaining = 0
		}
		resetAt := time.Now().Add(rl.window).Unix()
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if count > limit {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.Itoa(rl.windowSeconds))
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"code":"RATE_LIMITED","message":"Too many requests. Please try again later."}}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// increment atomically increments the counter for the key using a sliding window.
// Returns the current count within the window.
func (rl *RateLimiter) increment(ctx context.Context, key string) (int, error) {
	pipe := rl.redis.Pipeline()

	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return int(incr.Val()), nil
}

// extractIP extracts the client IP from the request, respecting X-Forwarded-For
// and X-Real-IP headers (trusted behind a reverse proxy).
func extractIP(r *http.Request) string {
	// X-Real-IP is typically set by nginx
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// X-Forwarded-For contains a comma-separated list; first is the client
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
