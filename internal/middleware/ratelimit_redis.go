package middleware

import (
	"context"
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
