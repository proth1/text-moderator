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

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
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

// CORSConfigFromOrigins creates a CORS config with specific allowed origins.
// Pass a comma-separated string of origins (e.g. "https://example.com,https://other.com").
// Empty string defaults to "*" (development mode).
func CORSConfigFromOrigins(origins string) CORSConfig {
	cfg := DefaultCORSConfig()
	if origins != "" {
		parsed := strings.Split(origins, ",")
		trimmed := make([]string, 0, len(parsed))
		for _, o := range parsed {
			o = strings.TrimSpace(o)
			if o != "" {
				trimmed = append(trimmed, o)
			}
		}
		if len(trimmed) > 0 {
			cfg.AllowedOrigins = trimmed
			cfg.AllowCredentials = true
		}
	}
	return cfg
}

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Set CORS headers
		if len(config.AllowedOrigins) > 0 {
			if config.AllowedOrigins[0] == "*" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" && contains(config.AllowedOrigins, origin) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Vary", "Origin")
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
