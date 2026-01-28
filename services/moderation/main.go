package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/proth1/text-moderator/internal/config"
	"github.com/proth1/text-moderator/internal/database"
	"github.com/proth1/text-moderator/internal/evidence"
	"github.com/proth1/text-moderator/internal/middleware"
	"github.com/proth1/text-moderator/internal/models"
	"github.com/proth1/text-moderator/services/moderation/client"
	"github.com/proth1/text-moderator/services/policy-engine/engine"
	"go.uber.org/zap"
)

// Control: MOD-001 (Content moderation service)

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

	logger.Info("starting moderation service",
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

	// Initialize HuggingFace client
	hfClient := client.NewHuggingFaceClient(client.Config{
		APIKey:   cfg.HuggingFaceAPIKey,
		ModelURL: cfg.HuggingFaceModelURL,
		Timeout:  cfg.HuggingFaceTimeout,
	}, logger)

	// Initialize policy evaluator
	evaluator := engine.NewEvaluator(db.Pool, logger)

	// Initialize evidence writer
	evidenceWriter := evidence.NewWriter(db.Pool, logger)

	// Create HTTP server
	router := setupRouter(cfg, logger, db, hfClient, evaluator, evidenceWriter)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ModerationPort),
		Handler: router,
	}

	// Start server
	go func() {
		logger.Info("moderation service listening", zap.String("port", cfg.ModerationPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down moderation service")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("moderation service stopped")
}

func setupRouter(cfg *config.Config, logger *zap.Logger, db *database.PostgresDB, hfClient *client.HuggingFaceClient, evaluator *engine.Evaluator, evidenceWriter *evidence.Writer) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware(middleware.DefaultCORSConfig()))

	// Health check
	router.GET("/health", healthHandler(db, hfClient, logger))

	// Moderation endpoint
	router.POST("/moderate", moderateHandler(db, hfClient, evaluator, evidenceWriter, cfg, logger))

	return router
}

func healthHandler(db *database.PostgresDB, hfClient *client.HuggingFaceClient, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]string)

		// Check database
		if err := db.Health(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
		} else {
			checks["database"] = "healthy"
		}

		// Check HuggingFace API (optional, can be slow)
		// Uncomment if needed
		// if err := hfClient.Health(ctx); err != nil {
		// 	checks["huggingface"] = "unhealthy: " + err.Error()
		// } else {
		// 	checks["huggingface"] = "healthy"
		// }

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
			Service: "moderation",
			Version: "0.1.0",
			Checks:  checks,
		})
	}
}

var sourcePattern = regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)

func moderateHandler(db *database.PostgresDB, hfClient *client.HuggingFaceClient, evaluator *engine.Evaluator, evidenceWriter *evidence.Writer, cfg *config.Config, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ModerationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate content length
		if len(req.Content) > cfg.MaxContentLength {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("content exceeds maximum length of %d characters", cfg.MaxContentLength),
			})
			return
		}

		// Validate source field
		if req.Source != "" {
			if len(req.Source) > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "source field exceeds maximum length of 100 characters"})
				return
			}
			if !sourcePattern.MatchString(req.Source) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "source field must contain only alphanumeric characters and hyphens"})
				return
			}
		}

		// Validate context_metadata
		if req.ContextMetadata != nil {
			if len(req.ContextMetadata) > 10 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "context_metadata exceeds maximum of 10 keys"})
				return
			}
			metaBytes, err := json.Marshal(req.ContextMetadata)
			if err == nil && len(metaBytes) > 1024 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "context_metadata exceeds maximum size of 1KB"})
				return
			}
		}

		ctx := c.Request.Context()

		// Hash content for deduplication
		hash := sha256.Sum256([]byte(req.Content))
		contentHash := hex.EncodeToString(hash[:])

		// Create submission record
		submission := &models.TextSubmission{
			ID:              uuid.New(),
			ContentHash:     contentHash,
			ContextMetadata: req.ContextMetadata,
			Source:          &req.Source,
		}

		query := `
			INSERT INTO text_submissions (id, content_hash, context_metadata, source)
			VALUES ($1, $2, $3, $4)
			RETURNING created_at
		`
		err := db.Pool.QueryRow(ctx, query, submission.ID, submission.ContentHash, submission.ContextMetadata, submission.Source).Scan(&submission.CreatedAt)
		if err != nil {
			logger.Error("failed to create submission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create submission"})
			return
		}

		// Call HuggingFace for classification
		scores, err := hfClient.ClassifyText(ctx, req.Content)
		if err != nil {
			logger.Error("failed to classify text", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to classify text"})
			return
		}

		// Get policy (use provided or default)
		var policy *models.Policy
		if req.PolicyID != nil {
			policy, err = evaluator.GetDefaultPolicy(ctx) // TODO: Get specific policy by ID
		} else {
			policy, err = evaluator.GetDefaultPolicy(ctx)
		}
		if err != nil {
			logger.Warn("no policy found, defaulting to allow", zap.Error(err))
			// If no policy, allow by default
			policy = &models.Policy{
				ID:      uuid.New(),
				Name:    "default",
				Version: 1,
			}
		}

		// Evaluate against policy
		var action models.PolicyAction
		if policy.Name != "default" {
			evalResult, err := evaluator.EvaluateScores(ctx, scores, policy.ID)
			if err != nil {
				logger.Error("failed to evaluate policy", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to evaluate policy"})
				return
			}
			action = evalResult.Action
		} else {
			action = models.ActionAllow
		}

		// Create decision record
		decision := &models.ModerationDecision{
			ID:              uuid.New(),
			SubmissionID:    submission.ID,
			ModelName:       "unitary/toxic-bert",
			ModelVersion:    "v1",
			CategoryScores:  *scores,
			PolicyID:        &policy.ID,
			PolicyVersion:   &policy.Version,
			AutomatedAction: action,
		}

		decisionQuery := `
			INSERT INTO moderation_decisions (
				id, submission_id, model_name, model_version, category_scores,
				policy_id, policy_version, automated_action
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING created_at
		`
		err = db.Pool.QueryRow(ctx, decisionQuery,
			decision.ID, decision.SubmissionID, decision.ModelName, decision.ModelVersion,
			decision.CategoryScores, decision.PolicyID, decision.PolicyVersion, decision.AutomatedAction,
		).Scan(&decision.CreatedAt)
		if err != nil {
			logger.Error("failed to create decision", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create decision"})
			return
		}

		// Write evidence record
		if err := evidenceWriter.RecordModerationDecision(ctx, decision); err != nil {
			logger.Error("failed to write evidence", zap.Error(err))
			// Don't fail the request, but log the error
		}

		// Prepare response
		requiresReview := action == models.ActionEscalate
		response := models.ModerationResponse{
			DecisionID:     decision.ID,
			SubmissionID:   submission.ID,
			Action:         action,
			CategoryScores: *scores,
			RequiresReview: requiresReview,
		}

		if policy.Name != "default" {
			response.PolicyApplied = &policy.Name
			response.PolicyVersion = &policy.Version
		}

		c.JSON(http.StatusOK, response)
	}
}
