package middleware

import (
	"math/rand/v2"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientLimiter struct {
	tokens     float64
	lastUpdate time.Time
}

type IPRateLimiter struct {
	mu         sync.Mutex
	clients    map[string]*clientLimiter
	rateLimit  float64
	burstLimit float64
}

func NewIPRateLimiter(rateLimit float64, burstLimit float64) *IPRateLimiter {
	return &IPRateLimiter{
		clients:    make(map[string]*clientLimiter),
		rateLimit:  rateLimit,
		burstLimit: burstLimit,
	}
}

func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	if len(l.clients) > 1000 && rand.IntN(100) == 0 {
		for k, v := range l.clients {
			if now.Sub(v.lastUpdate) > 10*time.Minute {
				delete(l.clients, k)
			}
		}
	}

	lim, exists := l.clients[ip]
	if !exists {
		l.clients[ip] = &clientLimiter{
			tokens:     l.burstLimit - 1.0,
			lastUpdate: now,
		}
		return true
	}

	elapsed := now.Sub(lim.lastUpdate).Seconds()
	lim.lastUpdate = now

	lim.tokens += elapsed * l.rateLimit
	if lim.tokens > l.burstLimit {
		lim.tokens = l.burstLimit
	}

	if lim.tokens >= 1.0 {
		lim.tokens -= 1.0
		return true
	}
	return false
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	return ip
}

func RateLimitMiddleware(limiter *IPRateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !limiter.Allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
