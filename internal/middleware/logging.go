package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"
)

// LoggingMiddleware creates a structured logging middleware
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or extract request ID
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get user info if available
		userID, _ := GetUserID(c)
		userRole, _ := GetUserRole(c)

		// Build log fields
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Add user context if available
		if userID != uuid.Nil {
			fields = append(fields,
				zap.String("user_id", userID.String()),
				zap.String("user_role", userRole),
			)
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		// Log based on status code
		statusCode := c.Writer.Status()
		switch {
		case statusCode >= 500:
			logger.Error("request completed", fields...)
		case statusCode >= 400:
			logger.Warn("request completed", fields...)
		default:
			logger.Info("request completed", fields...)
		}
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *gin.Context) string {
	requestID, exists := c.Get(RequestIDKey)
	if !exists {
		return ""
	}
	return requestID.(string)
}
