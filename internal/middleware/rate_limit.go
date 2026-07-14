package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type rateLimitVisitor struct {
	requests int
	resetAt  time.Time
}

type rateLimiter struct {
	mu          sync.Mutex
	visitors    map[string]rateLimitVisitor
	limit       int
	window      time.Duration
	lastCleanup time.Time
}

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := &rateLimiter{
		visitors:    make(map[string]rateLimitVisitor),
		limit:       limit,
		window:      window,
		lastCleanup: time.Now(),
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, retryAfter := limiter.allow(rateLimitKey(r))
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (l *rateLimiter) allow(key string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if now.Sub(l.lastCleanup) >= l.window {
		for key, visitor := range l.visitors {
			if !now.Before(visitor.resetAt) {
				delete(l.visitors, key)
			}
		}
		l.lastCleanup = now
	}

	visitor, exists := l.visitors[key]
	if !exists || !now.Before(visitor.resetAt) {
		l.visitors[key] = rateLimitVisitor{
			requests: 1,
			resetAt:  now.Add(l.window),
		}
		return true, 0
	}

	if visitor.requests >= l.limit {
		retryAfter := int(time.Until(visitor.resetAt).Seconds()) + 1
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, retryAfter
	}

	visitor.requests++
	l.visitors[key] = visitor
	return true, 0
}

func rateLimitKey(r *http.Request) string {
	if userID, ok := GetUserID(r); ok {
		return fmt.Sprintf("user:%d", userID)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return "ip:" + host
	}

	return "ip:" + r.RemoteAddr
}
