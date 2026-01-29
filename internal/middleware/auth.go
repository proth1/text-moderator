package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Control: GOV-002 (API key authentication and authorization)

const (
	AuthorizationHeader = "Authorization"
	APIKeyHeader        = "X-API-Key"
	UserContextKey      = "user_id"
	UserRoleContextKey  = "user_role"
)

// AuthMiddleware validates API keys and populates user context.
// If db is nil, it only validates that an API key is present (gateway proxy mode).
func AuthMiddleware(db *pgxpool.Pool, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			logger.Warn("missing API key",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing API key",
			})
			c.Abort()
			return
		}

		// Gateway proxy mode: only validate header presence, downstream services handle full auth
		if db == nil {
			c.Next()
			return
		}

		// Query user by API key
		ctx := c.Request.Context()
		var userID uuid.UUID
		var userRole string

		query := `SELECT id, role FROM users WHERE api_key = $1`
		err := db.QueryRow(ctx, query, apiKey).Scan(&userID, &userRole)
		if err != nil {
			logger.Warn("invalid API key",
				zap.String("error", err.Error()),
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set(UserContextKey, userID)
		c.Set(UserRoleContextKey, userRole)

		logger.Debug("authenticated request",
			zap.String("user_id", userID.String()),
			zap.String("role", userRole),
			zap.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}

// RequireRole creates middleware that enforces role requirements
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(UserRoleContextKey)
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "user role not found in context",
			})
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions",
		})
		c.Abort()
	}
}

// extractAPIKey extracts the API key from request headers
func extractAPIKey(c *gin.Context) string {
	// Try X-API-Key header first
	if apiKey := c.GetHeader(APIKeyHeader); apiKey != "" {
		return apiKey
	}

	// Try Authorization header with Bearer scheme
	if auth := c.GetHeader(AuthorizationHeader); auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	return ""
}

// GetUserID retrieves the authenticated user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(UserContextKey)
	if !exists {
		return uuid.Nil, false
	}
	return userID.(uuid.UUID), true
}

// GetUserRole retrieves the authenticated user role from context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(UserRoleContextKey)
	if !exists {
		return "", false
	}
	return role.(string), true
}

// MustGetUserID retrieves the user ID from context or panics
func MustGetUserID(c *gin.Context) uuid.UUID {
	userID, exists := GetUserID(c)
	if !exists {
		panic("user ID not found in context")
	}
	return userID
}

// WithUserContext adds user information to a standard context
func WithUserContext(ctx context.Context, userID uuid.UUID, role string) context.Context {
	ctx = context.WithValue(ctx, UserContextKey, userID)
	ctx = context.WithValue(ctx, UserRoleContextKey, role)
	return ctx
}

// InternalServiceHeader is the header used for service-to-service authentication
const InternalServiceHeader = "X-Internal-Service-Token"

// InternalServiceAuthMiddleware validates requests from other internal services.
// This provides defense-in-depth for internal service endpoints.
// The token should be set via INTERNAL_SERVICE_TOKEN environment variable.
func InternalServiceAuthMiddleware(expectedToken string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If no token configured, reject all requests (fail secure)
		if expectedToken == "" {
			logger.Error("internal service token not configured - rejecting request")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "service not properly configured",
			})
			c.Abort()
			return
		}

		token := c.GetHeader(InternalServiceHeader)
		if token == "" {
			// Also check X-API-Key for backwards compatibility with gateway proxy
			token = extractAPIKey(c)
		}

		if token == "" {
			logger.Warn("missing internal service token",
				zap.String("path", c.Request.URL.Path),
				zap.String("remote_addr", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		// Constant-time comparison to prevent timing attacks
		if !secureCompare(token, expectedToken) {
			logger.Warn("invalid internal service token",
				zap.String("path", c.Request.URL.Path),
				zap.String("remote_addr", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// secureCompare performs constant-time string comparison to prevent timing attacks
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
