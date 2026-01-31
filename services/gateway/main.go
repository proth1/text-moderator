package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/proth1/text-moderator/internal/config"
	"github.com/proth1/text-moderator/internal/middleware"
	"github.com/proth1/text-moderator/internal/models"
	"go.uber.org/zap"
)

// maxRequestBodySize limits request body to 1MB to prevent abuse
const maxRequestBodySize = 1 << 20 // 1MB

// allowedProxyHeaders defines which headers are forwarded to internal services
var allowedProxyHeaders = map[string]bool{
	"content-type":     true,
	"accept":           true,
	"x-correlation-id": true,
	"x-api-key":        true,
	"x-request-id":     true,
}

// internalServiceTokenHeader is the header for service-to-service auth
const internalServiceTokenHeader = "X-Internal-Service-Token"

// API Gateway routes requests to backend services

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	logger, err := cfg.NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("starting gateway service",
		zap.String("version", cfg.Version),
		zap.String("environment", cfg.Environment),
	)

	// Create HTTP server
	router := setupRouter(cfg, logger)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.GatewayPort),
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("gateway service listening", zap.String("port", cfg.GatewayPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gateway service")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("gateway service stopped")
}

func setupRouter(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.SecurityHeadersMiddleware(cfg.Environment))
	router.Use(middleware.CORSMiddleware(middleware.CORSConfigFromOrigins(cfg.AllowedOrigins)))

	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPM)
	router.Use(rateLimiter.Middleware())

	// Request body size limit
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodySize)
		c.Next()
	})

	// Content-Type validation for POST/PUT/PATCH requests
	router.Use(func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			// Allow empty body (Content-Length: 0) without Content-Type
			if c.Request.ContentLength > 0 {
				// SECURITY: Require explicit Content-Type for requests with body
				if contentType == "" {
					c.JSON(http.StatusUnsupportedMediaType, gin.H{
						"error": "Content-Type header is required",
					})
					c.Abort()
					return
				}
				// Only accept JSON for API requests
				if !strings.HasPrefix(contentType, "application/json") {
					c.JSON(http.StatusUnsupportedMediaType, gin.H{
						"error": "Content-Type must be application/json",
					})
					c.Abort()
					return
				}
			}
		}
		c.Next()
	})

	// Health check (no auth required)
	router.GET("/health", healthHandler())

	// API v1 routes with API key authentication
	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(nil, logger)) // nil db = header-only validation for gateway proxy
	{
		// Moderation service proxy
		v1.POST("/moderate", proxyHandler(cfg, logger, "moderation", "/moderate"))

		// Policy engine service proxy
		v1.GET("/policies", proxyHandler(cfg, logger, "policy-engine", "/policies"))
		v1.POST("/policies", proxyHandler(cfg, logger, "policy-engine", "/policies"))
		v1.GET("/policies/:id", proxyHandler(cfg, logger, "policy-engine", "/policies/:id"))
		v1.POST("/policies/:id/evaluate", proxyHandler(cfg, logger, "policy-engine", "/policies/:id/evaluate"))

		// Review service proxy
		v1.GET("/reviews", proxyHandler(cfg, logger, "review", "/reviews"))
		v1.GET("/reviews/:id", proxyHandler(cfg, logger, "review", "/reviews/:id"))
		v1.POST("/reviews/:id/action", proxyHandler(cfg, logger, "review", "/reviews/:id/action"))
		v1.GET("/evidence", proxyHandler(cfg, logger, "review", "/evidence"))
		v1.GET("/evidence/export", proxyHandler(cfg, logger, "review", "/evidence/export"))

		// Webhook management proxy
		v1.GET("/webhooks", proxyHandler(cfg, logger, "review", "/webhooks"))
		v1.POST("/webhooks", proxyHandler(cfg, logger, "review", "/webhooks"))
		v1.DELETE("/webhooks/:id", proxyHandler(cfg, logger, "review", "/webhooks/:id"))

		// Compliance report generation proxy
		v1.POST("/reports/generate", proxyHandler(cfg, logger, "review", "/reports/generate"))

		// Batch moderation proxy
		v1.POST("/moderate/batch", proxyHandler(cfg, logger, "moderation", "/moderate/batch"))
	}

	return router
}

func healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, models.HealthResponse{
			Status:  "healthy",
			Service: "gateway",
			Version: "0.1.0",
		})
	}
}

// serviceBaseURL returns the base URL for a backend service.
// If a full URL is configured (for Cloud Run), it takes precedence over host:port.
func serviceBaseURL(cfg *config.Config, service string) string {
	switch service {
	case "moderation":
		if cfg.ModerationURL != "" {
			return strings.TrimRight(cfg.ModerationURL, "/")
		}
		return fmt.Sprintf("http://%s:%s", cfg.ModerationHost, cfg.ModerationPort)
	case "policy-engine":
		if cfg.PolicyEngineURL != "" {
			return strings.TrimRight(cfg.PolicyEngineURL, "/")
		}
		return fmt.Sprintf("http://%s:%s", cfg.PolicyEngineHost, cfg.PolicyEnginePort)
	case "review":
		if cfg.ReviewURL != "" {
			return strings.TrimRight(cfg.ReviewURL, "/")
		}
		return fmt.Sprintf("http://%s:%s", cfg.ReviewHost, cfg.ReviewPort)
	default:
		return ""
	}
}

// sharedHTTPClient is reused across all proxy requests for connection pooling.
var sharedHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

func proxyHandler(cfg *config.Config, logger *zap.Logger, service string, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Determine target service URL
		baseURL := serviceBaseURL(cfg, service)
		if baseURL == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown service"})
			return
		}
		targetURL := baseURL + path

		// Replace path parameters
		for _, param := range c.Params {
			targetURL = replacePathParam(targetURL, param.Key, param.Value)
		}

		// Add query parameters
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		// Read request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		// Create proxy request
		proxyReq, err := http.NewRequestWithContext(
			c.Request.Context(),
			c.Request.Method,
			targetURL,
			bytes.NewReader(bodyBytes),
		)
		if err != nil {
			logger.Error("failed to create proxy request",
				zap.Error(err),
				zap.String("target_url", targetURL),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create proxy request"})
			return
		}

		// Copy only allowed headers to prevent header injection
		for key, values := range c.Request.Header {
			if allowedProxyHeaders[strings.ToLower(key)] {
				for _, value := range values {
					proxyReq.Header.Add(key, value)
				}
			}
		}

		// Add internal service token for service-to-service authentication
		if cfg.InternalServiceToken != "" {
			proxyReq.Header.Set(internalServiceTokenHeader, cfg.InternalServiceToken)
		}

		// Send request using shared client with connection pooling
		resp, err := sharedHTTPClient.Do(proxyReq)
		if err != nil {
			logger.Error("proxy request failed",
				zap.Error(err),
				zap.String("service", service),
				zap.String("target_url", targetURL),
			)
			c.JSON(http.StatusBadGateway, gin.H{"error": "service unavailable"})
			return
		}
		defer resp.Body.Close()

		// Read response body (bounded to 10MB to prevent memory exhaustion)
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
		if err != nil {
			logger.Error("failed to read proxy response",
				zap.Error(err),
				zap.String("service", service),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
			return
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Send response
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)

		logger.Debug("proxied request",
			zap.String("service", service),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", resp.StatusCode),
		)
	}
}

func replacePathParam(url, key, value string) string {
	return strings.Replace(url, ":"+key, value, 1)
}
