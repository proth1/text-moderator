package middleware

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proth1/text-moderator/internal/cache"
	"go.uber.org/zap"
)

const (
	idempotencyHeader = "Idempotency-Key"
	idempotencyTTL    = 24 * time.Hour
	idempotencyPrefix = "idempotency:"
)

// idempotencyRecorder captures the response so it can be cached.
type idempotencyRecorder struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *idempotencyRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *idempotencyRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// IdempotencyMiddleware deduplicates requests using the Idempotency-Key header.
// When Redis is nil the middleware is a no-op (single-instance deployments
// already benefit from client-side dedup).
func IdempotencyMiddleware(redisCache *cache.RedisCache, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(idempotencyHeader)
		if key == "" || redisCache == nil {
			c.Next()
			return
		}

		cacheKey := idempotencyPrefix + key

		// Check if we've already processed this key
		cached, err := redisCache.Get(c.Request.Context(), cacheKey)
		if err == nil && cached != "" {
			// Return cached response
			c.Header("X-Idempotent-Replayed", "true")
			c.Data(http.StatusOK, "application/json", []byte(cached))
			c.Abort()
			return
		}

		// Record the response
		recorder := &idempotencyRecorder{
			ResponseWriter: c.Writer,
			body:           new(bytes.Buffer),
		}
		c.Writer = recorder
		c.Next()

		// Cache successful responses only (2xx)
		if recorder.statusCode >= 200 && recorder.statusCode < 300 {
			if err := redisCache.Set(c.Request.Context(), cacheKey, recorder.body.String(), idempotencyTTL); err != nil {
				logger.Warn("failed to cache idempotency response", zap.String("key", key), zap.Error(err))
			}
		}
	}
}
