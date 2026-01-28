package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	tokens    float64
	lastSeen  time.Time
}

// RateLimiter implements a per-IP token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rpm      int
}

// NewRateLimiter creates a rate limiter allowing rpm requests per minute per IP.
func NewRateLimiter(rpm int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rpm:      rpm,
	}
	// Clean up stale entries every 3 minutes
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(3 * time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:   float64(rl.rpm) - 1,
			lastSeen: now,
		}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.tokens += elapsed * (float64(rl.rpm) / 60.0)
	if v.tokens > float64(rl.rpm) {
		v.tokens = float64(rl.rpm)
	}
	v.lastSeen = now

	if v.tokens >= 1 {
		v.tokens--
		return true
	}

	return false
}

// Middleware returns a Gin middleware that enforces the rate limit.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)
		if !rl.allow(ip) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

func clientIP(c *gin.Context) string {
	// Trust X-Forwarded-For from Cloud Run's load balancer
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// First IP is the client
		if ip, _, err := net.SplitHostPort(xff); err == nil {
			return ip
		}
		// May not have port
		parts := splitFirst(xff, ",")
		return parts
	}
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

func splitFirst(s, sep string) string {
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			return s[:i]
		}
	}
	return s
}
