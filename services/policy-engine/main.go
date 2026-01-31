package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/proth1/text-moderator/internal/config"
	"github.com/proth1/text-moderator/internal/database"
	"github.com/proth1/text-moderator/internal/middleware"
	"github.com/proth1/text-moderator/internal/models"
	"github.com/proth1/text-moderator/services/policy-engine/engine"
	"go.uber.org/zap"
)

// Control: POL-001 (Policy management and evaluation service)

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

	logger.Info("starting policy-engine service",
		zap.String("version", cfg.Version),
		zap.String("environment", cfg.Environment),
	)

	// Initialize database connection
	ctx := context.Background()
	db, err := database.NewPostgresDB(ctx, database.Config{
		URL:             cfg.DatabaseURL,
		MaxConns:        cfg.DatabaseMaxConns,
		MinConns:        cfg.DatabaseMinConns,
		MaxConnLifetime: cfg.DatabaseMaxLifetime,
	}, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize policy evaluator
	evaluator := engine.NewEvaluator(db.Pool, logger)

	// Create HTTP server
	router := setupRouter(cfg, logger, db, evaluator)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.PolicyEnginePort),
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("policy-engine service listening", zap.String("port", cfg.PolicyEnginePort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down policy-engine service")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("policy-engine service stopped")
}

func setupRouter(cfg *config.Config, logger *zap.Logger, db *database.PostgresDB, evaluator *engine.Evaluator) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware(middleware.DefaultCORSConfig()))

	// Health check
	router.GET("/health", healthHandler(db))

	// Policy endpoints (require authentication)
	api := router.Group("/")
	api.Use(middleware.AuthMiddleware(db.Pool, logger))
	{
		api.GET("/policies", listPoliciesHandler(evaluator))
		api.POST("/policies", middleware.RequireRole("admin"), createPolicyHandler(evaluator))
		api.GET("/policies/:id", getPolicyHandler(evaluator))
		api.POST("/policies/:id/evaluate", evaluatePolicyHandler(evaluator))
	}

	return router
}

func healthHandler(db *database.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]string)

		// Check database
		// SECURITY: Don't expose error details in health check responses
		if err := db.Health(ctx); err != nil {
			// Log error internally but don't expose to clients
			checks["database"] = "unhealthy"
		} else {
			checks["database"] = "healthy"
		}

		status := "healthy"
		for _, check := range checks {
			if check != "healthy" {
				status = "unhealthy"
				break
			}
		}

		statusCode := http.StatusOK
		if status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, models.HealthResponse{
			Status:  status,
			Service: "policy-engine",
			Version: "0.1.0",
			Checks:  checks,
		})
	}
}

func listPoliciesHandler(evaluator *engine.Evaluator) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Optional status filter
		var status *models.PolicyStatus
		if statusParam := c.Query("status"); statusParam != "" {
			s := models.PolicyStatus(statusParam)
			status = &s
		}

		policies, err := evaluator.ListPolicies(ctx, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list policies"})
			return
		}

		c.JSON(http.StatusOK, policies)
	}
}

func createPolicyHandler(evaluator *engine.Evaluator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreatePolicyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// SECURITY: Don't expose detailed parsing errors to clients
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		ctx := c.Request.Context()
		userID := middleware.MustGetUserID(c)

		policy, err := evaluator.CreatePolicy(ctx, &req, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create policy"})
			return
		}

		c.JSON(http.StatusCreated, policy)
	}
}

func getPolicyHandler(evaluator *engine.Evaluator) gin.HandlerFunc {
	return func(c *gin.Context) {
		policyID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy ID"})
			return
		}

		ctx := c.Request.Context()

		query := `
			SELECT id, name, version, thresholds, actions, scope, status, effective_date, created_at, created_by
			FROM policies
			WHERE id = $1
		`

		var policy models.Policy
		err = evaluator.Pool().QueryRow(ctx, query, policyID).Scan(
			&policy.ID,
			&policy.Name,
			&policy.Version,
			&policy.Thresholds,
			&policy.Actions,
			&policy.Scope,
			&policy.Status,
			&policy.EffectiveDate,
			&policy.CreatedAt,
			&policy.CreatedBy,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}

		c.JSON(http.StatusOK, policy)
	}
}

func evaluatePolicyHandler(evaluator *engine.Evaluator) gin.HandlerFunc {
	return func(c *gin.Context) {
		policyID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy ID"})
			return
		}

		var req models.PolicyEvaluationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// SECURITY: Don't expose detailed parsing errors to clients
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		ctx := c.Request.Context()

		result, err := evaluator.EvaluateScores(ctx, &req.CategoryScores, policyID, nil)
		if err != nil {
			// SECURITY: Don't expose internal error details
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to evaluate policy"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
