package middleware

import (
	"net"
	"net/http"
	"strings"
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

// trustedProxyCIDRs defines networks from which X-Forwarded-For headers are trusted.
// These should be configured based on your infrastructure.
// For Cloud Run: Google's frontend proxies add the client IP.
// For local dev: Trust localhost.
var trustedProxyCIDRs = []string{
	"10.0.0.0/8",      // Private networks (Cloud Run internal)
	"172.16.0.0/12",   // Private networks (Docker)
	"192.168.0.0/16",  // Private networks (local)
	"127.0.0.1/8",     // Localhost
	"169.254.0.0/16",  // Link-local (GCP metadata)
	"35.191.0.0/16",   // Google Cloud Load Balancer
	"130.211.0.0/22",  // Google Cloud Load Balancer
}

func clientIP(c *gin.Context) string {
	remoteAddr := c.Request.RemoteAddr
	remoteIP, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		remoteIP = remoteAddr
	}

	// SECURITY: Only trust X-Forwarded-For if the request comes from a trusted proxy
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		if isTrustedProxy(remoteIP) {
			// Use the rightmost non-trusted IP in the chain (most reliable)
			// This prevents spoofing by malicious clients
			ips := parseXForwardedFor(xff)
			for i := len(ips) - 1; i >= 0; i-- {
				if !isTrustedProxy(ips[i]) {
					return ips[i]
				}
			}
		}
		// Not from trusted proxy - ignore X-Forwarded-For, use direct connection IP
	}

	return remoteIP
}

// parseXForwardedFor extracts individual IPs from X-Forwarded-For header
func parseXForwardedFor(xff string) []string {
	var ips []string
	for _, part := range strings.Split(xff, ",") {
		ip := strings.TrimSpace(part)
		// Remove port if present
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

// isTrustedProxy checks if an IP is from a trusted proxy network
func isTrustedProxy(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, cidr := range trustedProxyCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func splitFirst(s, sep string) string {
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			return s[:i]
		}
	}
	return s
}
