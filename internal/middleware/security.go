package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security headers to all responses.
// In non-development environments, HSTS is also enabled.
func SecurityHeadersMiddleware(environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Writer.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")

		if environment != "development" {
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}
