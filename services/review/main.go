package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/proth1/text-moderator/internal/compliance"
	"github.com/proth1/text-moderator/internal/config"
	"github.com/proth1/text-moderator/internal/feedback"
	"github.com/proth1/text-moderator/internal/database"
	"github.com/proth1/text-moderator/internal/evidence"
	"github.com/proth1/text-moderator/internal/middleware"
	"github.com/proth1/text-moderator/internal/models"
	"github.com/proth1/text-moderator/internal/webhook"
	"go.uber.org/zap"
)

// Control: GOV-002 (Human review workflow service)

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

	logger.Info("starting review service",
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

	// Initialize evidence writer
	evidenceWriter := evidence.NewWriter(db.Pool, logger)

	// Initialize webhook dispatcher
	webhookDispatcher := webhook.NewDispatcher(db.Pool, logger)

	// Initialize feedback tracker for calibration loop
	feedbackTracker := feedback.NewTracker(db.Pool, logger)

	// Initialize compliance reporter
	complianceReporter := compliance.NewReporter(db.Pool, logger)

	// Create HTTP server
	router := setupRouter(cfg, logger, db, evidenceWriter, webhookDispatcher, complianceReporter, feedbackTracker)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.ReviewPort),
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("review service listening", zap.String("port", cfg.ReviewPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down review service")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("review service stopped")
}

func setupRouter(cfg *config.Config, logger *zap.Logger, db *database.PostgresDB, evidenceWriter *evidence.Writer, webhookDispatcher *webhook.Dispatcher, complianceReporter *compliance.Reporter, feedbackTracker *feedback.Tracker) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware(middleware.DefaultCORSConfig()))

	// Health check
	router.GET("/health", healthHandler(db))

	// Review endpoints (require authentication)
	api := router.Group("/")
	api.Use(middleware.AuthMiddleware(db.Pool, logger))
	{
		api.GET("/reviews", middleware.RequireRole("admin", "moderator"), listReviewsHandler(db, logger))
		api.GET("/reviews/:id", middleware.RequireRole("admin", "moderator"), getReviewHandler(db, logger))
		api.POST("/reviews/:id/action", middleware.RequireRole("admin", "moderator"), submitReviewActionHandler(db, evidenceWriter, feedbackTracker, logger))

		// Evidence endpoints
		api.GET("/evidence", middleware.RequireRole("admin"), listEvidenceHandler(evidenceWriter))
		api.GET("/evidence/export", middleware.RequireRole("admin"), exportEvidenceHandler(evidenceWriter))

		// Webhook management endpoints
		api.GET("/webhooks", middleware.RequireRole("admin"), listWebhooksHandler(webhookDispatcher))
		api.POST("/webhooks", middleware.RequireRole("admin"), createWebhookHandler(webhookDispatcher, logger))
		api.DELETE("/webhooks/:id", middleware.RequireRole("admin"), deleteWebhookHandler(webhookDispatcher, logger))

		// Compliance report generation
		api.POST("/reports/generate", middleware.RequireRole("admin"), generateReportHandler(complianceReporter, logger))
	}

	return router
}

func healthHandler(db *database.PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]string)

		// Check database
		if err := db.Health(ctx); err != nil {
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
			Service: "review",
			Version: "0.1.0",
			Checks:  checks,
		})
	}
}

func listReviewsHandler(db *database.PostgresDB, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get decisions that require review (escalated decisions without review actions)
		query := `
			SELECT
				d.id AS decision_id,
				d.submission_id,
				s.content_hash,
				d.category_scores,
				d.automated_action,
				p.name AS policy_name,
				d.created_at
			FROM moderation_decisions d
			JOIN text_submissions s ON s.id = d.submission_id
			LEFT JOIN policies p ON p.id = d.policy_id
			LEFT JOIN review_actions r ON r.decision_id = d.id
			WHERE d.automated_action = 'escalate' AND r.id IS NULL
			ORDER BY d.created_at ASC
			LIMIT 100
		`

		rows, err := db.Pool.Query(ctx, query)
		if err != nil {
			logger.Error("failed to query review queue", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query review queue"})
			return
		}
		defer rows.Close()

		var items []models.ReviewQueueItem
		for rows.Next() {
			var item models.ReviewQueueItem
			err := rows.Scan(
				&item.DecisionID,
				&item.SubmissionID,
				&item.ContentHash,
				&item.CategoryScores,
				&item.AutomatedAction,
				&item.PolicyName,
				&item.CreatedAt,
			)
			if err != nil {
				logger.Error("failed to scan review item", zap.Error(err))
				continue
			}
			items = append(items, item)
		}

		c.JSON(http.StatusOK, items)
	}
}

func getReviewHandler(db *database.PostgresDB, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		decisionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid decision ID"})
			return
		}

		ctx := c.Request.Context()

		query := `
			SELECT
				d.id, d.submission_id, d.model_name, d.model_version, d.category_scores,
				d.policy_id, d.policy_version, d.automated_action, d.confidence,
				d.explanation, d.created_at,
				s.content_hash, s.context_metadata, s.source
			FROM moderation_decisions d
			JOIN text_submissions s ON s.id = d.submission_id
			WHERE d.id = $1
		`

		var decision models.ModerationDecision
		var submission models.TextSubmission
		err = db.Pool.QueryRow(ctx, query, decisionID).Scan(
			&decision.ID,
			&decision.SubmissionID,
			&decision.ModelName,
			&decision.ModelVersion,
			&decision.CategoryScores,
			&decision.PolicyID,
			&decision.PolicyVersion,
			&decision.AutomatedAction,
			&decision.Confidence,
			&decision.Explanation,
			&decision.CreatedAt,
			&submission.ContentHash,
			&submission.ContextMetadata,
			&submission.Source,
		)

		if err != nil {
			logger.Error("failed to get decision", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"error": "decision not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"decision":   decision,
			"submission": submission,
		})
	}
}

func submitReviewActionHandler(db *database.PostgresDB, evidenceWriter *evidence.Writer, feedbackTracker *feedback.Tracker, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		decisionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid decision ID"})
			return
		}

		var req models.SubmitReviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()
		reviewerID := middleware.MustGetUserID(c)

		// Get the decision first
		var decision models.ModerationDecision
		decisionQuery := `
			SELECT id, submission_id, model_name, model_version, category_scores,
			       policy_id, policy_version, automated_action
			FROM moderation_decisions
			WHERE id = $1
		`
		err = db.Pool.QueryRow(ctx, decisionQuery, decisionID).Scan(
			&decision.ID,
			&decision.SubmissionID,
			&decision.ModelName,
			&decision.ModelVersion,
			&decision.CategoryScores,
			&decision.PolicyID,
			&decision.PolicyVersion,
			&decision.AutomatedAction,
		)
		if err != nil {
			logger.Error("failed to get decision", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"error": "decision not found"})
			return
		}

		// Create review action
		review := &models.ReviewAction{
			ID:         uuid.New(),
			DecisionID: decisionID,
			ReviewerID: reviewerID,
			Action:     req.Action,
		}

		if req.Rationale != "" {
			review.Rationale = &req.Rationale
		}
		if req.EditedContent != "" {
			review.EditedContent = &req.EditedContent
		}

		query := `
			INSERT INTO review_actions (id, decision_id, reviewer_id, action, rationale, edited_content)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING created_at
		`
		err = db.Pool.QueryRow(ctx, query,
			review.ID, review.DecisionID, review.ReviewerID, review.Action,
			review.Rationale, review.EditedContent,
		).Scan(&review.CreatedAt)

		if err != nil {
			logger.Error("failed to create review action", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review action"})
			return
		}

		// Write evidence record
		if err := evidenceWriter.RecordReviewAction(ctx, review, &decision); err != nil {
			logger.Error("failed to write evidence", zap.Error(err))
			// Don't fail the request, but log the error
		}

		// Record feedback for calibration loop (async)
		go feedbackTracker.RecordFeedback(context.Background(), &decision, req.Action)

		c.JSON(http.StatusCreated, review)
	}
}

func listEvidenceHandler(evidenceWriter *evidence.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		controlID := c.Query("control_id")
		var controlIDPtr *string
		if controlID != "" {
			controlIDPtr = &controlID
		}

		// Parse pagination with safe defaults and max limits
		limit := 100
		offset := 0
		if l := c.Query("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
				limit = parsed
			}
		}
		if o := c.Query("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		records, err := evidenceWriter.ListEvidence(ctx, controlIDPtr, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list evidence"})
			return
		}

		c.JSON(http.StatusOK, records)
	}
}

func exportEvidenceHandler(evidenceWriter *evidence.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		controlID := c.Query("control_id")
		var controlIDPtr *string
		if controlID != "" {
			controlIDPtr = &controlID
		}

		records, err := evidenceWriter.ListEvidence(ctx, controlIDPtr, 5000, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export evidence"})
			return
		}

		// Set headers for CSV download
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=evidence_export.json")

		c.JSON(http.StatusOK, records)
	}
}

// --- Webhook Management Handlers ---

func listWebhooksHandler(dispatcher *webhook.Dispatcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		subs, err := dispatcher.ListSubscriptions(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list webhooks"})
			return
		}
		c.JSON(http.StatusOK, subs)
	}
}

func createWebhookHandler(dispatcher *webhook.Dispatcher, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateWebhookRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		// Validate event types
		validTypes := map[string]bool{
			string(models.EventModerationCompleted): true,
			string(models.EventReviewRequired):      true,
			string(models.EventReviewCompleted):      true,
			string(models.EventPolicyUpdated):        true,
		}
		for _, et := range req.EventTypes {
			if !validTypes[et] {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid event type: %s", et)})
				return
			}
		}

		userID := middleware.MustGetUserID(c)
		sub, err := dispatcher.CreateSubscription(c.Request.Context(), &req, &userID)
		if err != nil {
			logger.Error("failed to create webhook", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create webhook"})
			return
		}

		// Return the subscription including the secret (only shown once at creation)
		c.JSON(http.StatusCreated, gin.H{
			"id":          sub.ID,
			"url":         sub.URL,
			"secret":      sub.Secret,
			"event_types": sub.EventTypes,
			"active":      sub.Active,
			"created_at":  sub.CreatedAt,
			"message":     "Store the secret securely - it will not be shown again",
		})
	}
}

func deleteWebhookHandler(dispatcher *webhook.Dispatcher, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook ID"})
			return
		}

		if err := dispatcher.DeleteSubscription(c.Request.Context(), id); err != nil {
			logger.Error("failed to delete webhook", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete webhook"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "deactivated"})
	}
}

// --- Compliance Report Handler ---

func generateReportHandler(reporter *compliance.Reporter, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req compliance.ReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		// Validate report type
		validTypes := map[compliance.ReportType]bool{
			compliance.ReportSOC2:     true,
			compliance.ReportISO42001: true,
			compliance.ReportEUAIAct:  true,
			compliance.ReportGDPR:     true,
		}
		if !validTypes[req.Type] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":       "invalid report type",
				"valid_types": []string{"soc2", "iso42001", "eu-ai-act", "gdpr"},
			})
			return
		}

		report, err := reporter.Generate(c.Request.Context(), req)
		if err != nil {
			logger.Error("failed to generate compliance report", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
			return
		}

		// Return HTML if requested, otherwise JSON
		if c.GetHeader("Accept") == "text/html" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(report.HTMLContent))
			return
		}

		c.JSON(http.StatusOK, report)
	}
}
