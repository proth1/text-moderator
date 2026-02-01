package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proth1/text-moderator/internal/cache"
)

// RedisRateLimiter implements a sliding window rate limiter backed by Redis.
// Unlike the in-memory RateLimiter, this works across multiple service instances.
// Control: SEC-002 (Distributed rate limiting)
type RedisRateLimiter struct {
	redis  *cache.RedisCache
	rpm    int
	prefix string
}

// NewRedisRateLimiter creates a distributed rate limiter using Redis.
func NewRedisRateLimiter(redis *cache.RedisCache, rpm int) *RedisRateLimiter {
	return &RedisRateLimiter{
		redis:  redis,
		rpm:    rpm,
		prefix: "ratelimit:",
	}
}

// Middleware returns a Gin middleware that enforces the distributed rate limit.
func (rl *RedisRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)
		key := fmt.Sprintf("%s%s", rl.prefix, ip)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Increment the counter for this window
		count, err := rl.redis.Incr(ctx, key)
		if err != nil {
			// On Redis failure, fall through (fail-open to avoid blocking all traffic)
			c.Next()
			return
		}

		// Set expiry on first request in window
		if count == 1 {
			rl.redis.Expire(ctx, key, time.Minute)
		}

		if count > int64(rl.rpm) {
			remaining := int64(rl.rpm) - count
			if remaining < 0 {
				remaining = 0
			}

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.rpm))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.rpm))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(rl.rpm)-count))

		c.Next()
	}
}

// RateLimitLookup is a callback to look up per-key RPM limits.
type RateLimitLookup func(ctx context.Context, apiKeyHash string) (int, error)

// KeyAwareRateLimiter checks the API key first for per-key rate limits,
// falling back to IP-based limiting if no key is present.
// Control: SEC-004 (Distributed rate limiting with per-key support)
type KeyAwareRateLimiter struct {
	redis      *cache.RedisCache
	defaultRPM int
	lookup     RateLimitLookup
	prefix     string
}

// NewKeyAwareRateLimiter creates a rate limiter that supports per-key limits.
func NewKeyAwareRateLimiter(redis *cache.RedisCache, defaultRPM int, lookup RateLimitLookup) *KeyAwareRateLimiter {
	return &KeyAwareRateLimiter{
		redis:      redis,
		defaultRPM: defaultRPM,
		lookup:     lookup,
		prefix:     "ratelimit:",
	}
}

// Middleware returns a Gin middleware that enforces per-key or per-IP rate limits.
func (rl *KeyAwareRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Determine rate limit key and RPM
		var rateLimitKey string
		rpm := rl.defaultRPM

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try Bearer token
			if auth := c.GetHeader("Authorization"); len(auth) > 7 {
				apiKey = auth[7:] // strip "Bearer "
			}
		}

		if apiKey != "" {
			// Per-key rate limiting
			h := sha256.Sum256([]byte(apiKey))
			apiKeyHash := hex.EncodeToString(h[:])
			rateLimitKey = fmt.Sprintf("%skey:%s", rl.prefix, apiKeyHash[:16])

			// Look up per-key RPM (cached in Redis for 5 min)
			cacheKey := fmt.Sprintf("keyrpm:%s", apiKeyHash[:16])
			if cached, err := rl.redis.Get(ctx, cacheKey); err == nil {
				var cachedRPM int
				if _, err := fmt.Sscanf(cached, "%d", &cachedRPM); err == nil {
					rpm = cachedRPM
				}
			} else if rl.lookup != nil {
				if keyRPM, err := rl.lookup(ctx, apiKeyHash); err == nil {
					rpm = keyRPM
					rl.redis.Set(ctx, cacheKey, fmt.Sprintf("%d", keyRPM), 5*time.Minute)
				}
			}
		} else {
			// IP-based fallback
			ip := clientIP(c)
			rateLimitKey = fmt.Sprintf("%sip:%s", rl.prefix, ip)
		}

		// Increment counter
		count, err := rl.redis.Incr(ctx, rateLimitKey)
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rl.redis.Expire(ctx, rateLimitKey, time.Minute)
		}

		if count > int64(rpm) {
			remaining := int64(rpm) - count
			if remaining < 0 {
				remaining = 0
			}
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rpm))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rpm))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(rpm)-count))
		c.Next()
	}
}
