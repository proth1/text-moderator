package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a restrictive default CORS configuration.
// In production, you MUST set ALLOWED_ORIGINS environment variable.
// SECURITY: Wildcard "*" is only allowed in development mode.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		// SECURITY: Empty by default - requires explicit configuration
		// Use CORSConfigFromOrigins() with ALLOWED_ORIGINS env var in production
		AllowedOrigins: []string{},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-API-Key",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"X-Request-ID",
		},
		AllowCredentials: false,
		MaxAge:           3600,
	}
}

// DevelopmentCORSConfig returns a permissive CORS config for local development only.
// WARNING: Never use this in production!
func DevelopmentCORSConfig() CORSConfig {
	cfg := DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"http://localhost:3000", "http://localhost:5173", "http://127.0.0.1:3000", "http://127.0.0.1:5173"}
	cfg.AllowCredentials = true
	return cfg
}

// CORSConfigFromOrigins creates a CORS config with specific allowed origins.
// Pass a comma-separated string of origins (e.g. "https://example.com,https://other.com").
// SECURITY: Empty string results in no CORS headers being sent (restrictive default).
// For development, explicitly pass "http://localhost:3000" or similar.
func CORSConfigFromOrigins(origins string) CORSConfig {
	cfg := DefaultCORSConfig()
	if origins == "" {
		// SECURITY: No origins configured - CORS headers won't be sent
		// This is the secure default for production
		return cfg
	}

	parsed := strings.Split(origins, ",")
	trimmed := make([]string, 0, len(parsed))
	for _, o := range parsed {
		o = strings.TrimSpace(o)
		if o != "" {
			// SECURITY: Reject wildcard in production - must be explicit origins
			if o == "*" {
				// Log warning but don't add - require explicit origins
				continue
			}
			trimmed = append(trimmed, o)
		}
	}
	if len(trimmed) > 0 {
		cfg.AllowedOrigins = trimmed
		cfg.AllowCredentials = true
	}
	return cfg
}

// CORSMiddleware creates a CORS middleware with the given configuration.
// SECURITY: If no origins are configured, no CORS headers are sent (restrictive default).
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// SECURITY: If no allowed origins configured, don't set any CORS headers
		// This prevents cross-origin requests entirely
		if len(config.AllowedOrigins) == 0 {
			// Handle preflight but don't allow the request
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(403)
				return
			}
			c.Next()
			return
		}

		// Set CORS headers only for configured origins
		if origin != "" && contains(config.AllowedOrigins, origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		} else if origin != "" {
			// Origin not in allowed list - reject preflight, continue without CORS headers for regular requests
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(403)
				return
			}
		}

		if len(config.AllowedMethods) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods))
		}

		if len(config.AllowedHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders))
		}

		if len(config.ExposedHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposedHeaders))
		}

		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// joinStrings joins a string slice with commas
func joinStrings(slice []string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
