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
	"github.com/jackc/pgx/v5"
	"github.com/proth1/text-moderator/internal/compliance"
	"github.com/proth1/text-moderator/internal/config"
	"github.com/proth1/text-moderator/internal/apikey"
	"github.com/proth1/text-moderator/internal/database"
	"github.com/proth1/text-moderator/internal/evidence"
	"github.com/proth1/text-moderator/internal/fairness"
	"github.com/proth1/text-moderator/internal/feedback"
	"github.com/proth1/text-moderator/internal/retention"
	"github.com/proth1/text-moderator/internal/middleware"
	"github.com/proth1/text-moderator/internal/models"
	"github.com/proth1/text-moderator/internal/observability"
	"github.com/proth1/text-moderator/internal/webhook"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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

	// Initialize data retention purger
	purger := retention.NewPurger(db.Pool, logger)

	// Initialize API key manager
	keyManager := apikey.NewManager(db.Pool, logger)

	// Initialize distributed tracing
	tracingShutdown, err := observability.InitTracing(context.Background(), "review", cfg.Version, cfg.OTLPEndpoint, logger)
	if err != nil {
		logger.Warn("failed to initialize tracing", zap.Error(err))
	} else {
		defer tracingShutdown(context.Background())
	}

	// Initialize Prometheus metrics
	metrics := observability.NewMetrics("review")

	// Start background goroutine to update review queue size gauge
	gaugeCtx, gaugeCancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-gaugeCtx.Done():
				return
			case <-ticker.C:
				var count float64
				err := db.Pool.QueryRow(gaugeCtx,
					`SELECT COUNT(*) FROM moderation_decisions d
					 LEFT JOIN review_actions r ON r.decision_id = d.id
					 WHERE d.automated_action = 'escalate' AND r.id IS NULL`,
				).Scan(&count)
				if err != nil {
					logger.Warn("failed to query review queue size", zap.Error(err))
					continue
				}
				metrics.ReviewQueueSize.Set(count)
			}
		}
	}()

	// Create HTTP server
	router := setupRouter(cfg, logger, db, evidenceWriter, webhookDispatcher, complianceReporter, feedbackTracker, purger, keyManager, metrics)
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
	gaugeCancel() // stop background queue size polling

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("review service stopped")
}

func setupRouter(cfg *config.Config, logger *zap.Logger, db *database.PostgresDB, evidenceWriter *evidence.Writer, webhookDispatcher *webhook.Dispatcher, complianceReporter *compliance.Reporter, feedbackTracker *feedback.Tracker, purger *retention.Purger, keyManager *apikey.Manager, metrics *observability.Metrics) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware("review"))
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware(middleware.DefaultCORSConfig()))
	router.Use(observability.MetricsMiddleware(metrics))

	// Health check and metrics
	router.GET("/health", healthHandler(db))
	router.GET("/metrics", observability.PrometheusHandler())

	// Review endpoints (require authentication)
	api := router.Group("/")
	api.Use(middleware.AuthMiddleware(db.Pool, logger))
	{
		api.GET("/reviews", middleware.RequireRole("admin", "moderator"), listReviewsHandler(db, logger))
		api.GET("/reviews/:id", middleware.RequireRole("admin", "moderator"), getReviewHandler(db, logger))
		api.POST("/reviews/:id/action", middleware.RequireRole("admin", "moderator"), submitReviewActionHandler(db, evidenceWriter, feedbackTracker, logger, metrics))
		api.POST("/reviews/:id/claim", middleware.RequireRole("admin", "moderator"), claimReviewHandler(db, cfg, logger))
		api.POST("/reviews/:id/unclaim", middleware.RequireRole("admin", "moderator"), unclaimReviewHandler(db, logger))

		// Evidence endpoints
		api.GET("/evidence", middleware.RequireRole("admin"), listEvidenceHandler(evidenceWriter))
		api.GET("/evidence/export", middleware.RequireRole("admin"), exportEvidenceHandler(evidenceWriter))

		// Webhook management endpoints
		api.GET("/webhooks", middleware.RequireRole("admin"), listWebhooksHandler(webhookDispatcher))
		api.POST("/webhooks", middleware.RequireRole("admin"), createWebhookHandler(webhookDispatcher, logger))
		api.DELETE("/webhooks/:id", middleware.RequireRole("admin"), deleteWebhookHandler(webhookDispatcher, logger))

		// Compliance report generation
		api.POST("/reports/generate", middleware.RequireRole("admin"), generateReportHandler(complianceReporter, logger))

		// Fairness and bias detection
		api.GET("/reports/fairness", middleware.RequireRole("admin"), fairnessReportHandler(db, logger))

		// GDPR erasure endpoint
		api.DELETE("/submissions/:hash", middleware.RequireRole("admin"), erasureHandler(purger, logger))

		// API key management endpoints
		api.GET("/api-keys", middleware.RequireRole("admin"), listAPIKeysHandler(keyManager))
		api.POST("/api-keys", middleware.RequireRole("admin"), generateAPIKeyHandler(keyManager, logger))
		api.DELETE("/api-keys/:user_id", middleware.RequireRole("admin"), revokeAPIKeyHandler(keyManager, logger))
		api.POST("/api-keys/:user_id/rotate", middleware.RequireRole("admin"), rotateAPIKeyHandler(keyManager, logger))
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

		// Parse pagination parameters
		limitStr := c.DefaultQuery("limit", "50")
		cursor := c.Query("cursor") // cursor = created_at of last item (ISO 8601)

		limit := 50
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}

		// Build query with cursor-based pagination
		var rows pgx.Rows
		var err error

		selectCols := `
					d.id AS decision_id,
					d.submission_id,
					s.content_hash,
					d.category_scores,
					d.automated_action,
					p.name AS policy_name,
					d.assigned_reviewer,
					d.assigned_at,
					d.sla_deadline,
					d.created_at`
		fromClause := `
				FROM moderation_decisions d
				JOIN text_submissions s ON s.id = d.submission_id
				LEFT JOIN policies p ON p.id = d.policy_id
				LEFT JOIN review_actions r ON r.decision_id = d.id
				WHERE d.automated_action = 'escalate' AND r.id IS NULL`

		if cursor != "" {
			cursorTime, parseErr := time.Parse(time.RFC3339Nano, cursor)
			if parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor format, expected RFC3339"})
				return
			}
			query := `SELECT` + selectCols + fromClause + `
				  AND d.created_at > $1
				ORDER BY d.created_at ASC
				LIMIT $2
			`
			rows, err = db.Pool.Query(ctx, query, cursorTime, limit+1)
		} else {
			query := `SELECT` + selectCols + fromClause + `
				ORDER BY d.created_at ASC
				LIMIT $1
			`
			rows, err = db.Pool.Query(ctx, query, limit+1)
		}

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
				&item.AssignedReviewer,
				&item.AssignedAt,
				&item.SLADeadline,
				&item.CreatedAt,
			)
			if err != nil {
				logger.Error("failed to scan review item", zap.Error(err))
				continue
			}
			items = append(items, item)
		}

		// Determine if there are more results
		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
		}

		var nextCursor string
		if hasMore && len(items) > 0 {
			nextCursor = items[len(items)-1].CreatedAt.Format(time.RFC3339Nano)
		}

		c.JSON(http.StatusOK, gin.H{
			"items":       items,
			"next_cursor": nextCursor,
			"has_more":    hasMore,
		})
	}
}

func claimReviewHandler(db *database.PostgresDB, cfg *config.Config, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		decisionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid decision ID"})
			return
		}

		reviewerID := middleware.MustGetUserID(c)
		ctx := c.Request.Context()

		// Set SLA deadline (default 4 hours from now)
		slaHours := 4 * time.Hour
		slaDeadline := time.Now().Add(slaHours)

		// Atomically claim only if unclaimed
		tag, err := db.Pool.Exec(ctx,
			`UPDATE moderation_decisions
			 SET assigned_reviewer = $1, assigned_at = NOW(), sla_deadline = $2
			 WHERE id = $3 AND automated_action = 'escalate' AND assigned_reviewer IS NULL`,
			reviewerID, slaDeadline, decisionID,
		)
		if err != nil {
			logger.Error("failed to claim review", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to claim review"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "review is already claimed or does not exist"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"decision_id":  decisionID,
			"assigned_to":  reviewerID,
			"sla_deadline": slaDeadline,
		})
	}
}

func unclaimReviewHandler(db *database.PostgresDB, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		decisionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid decision ID"})
			return
		}

		reviewerID := middleware.MustGetUserID(c)
		ctx := c.Request.Context()

		// Only allow the assigned reviewer (or admin) to unclaim
		tag, err := db.Pool.Exec(ctx,
			`UPDATE moderation_decisions
			 SET assigned_reviewer = NULL, assigned_at = NULL, sla_deadline = NULL
			 WHERE id = $1 AND assigned_reviewer = $2`,
			decisionID, reviewerID,
		)
		if err != nil {
			logger.Error("failed to unclaim review", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unclaim review"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "review is not assigned to you"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"decision_id": decisionID, "status": "unclaimed"})
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

func submitReviewActionHandler(db *database.PostgresDB, evidenceWriter *evidence.Writer, feedbackTracker *feedback.Tracker, logger *zap.Logger, metrics *observability.Metrics) gin.HandlerFunc {
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

		metrics.ReviewTotal.WithLabelValues(string(req.Action)).Inc()

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

// --- API Key Management Handlers ---
// Control: SEC-002 (API Key and OAuth Authentication)

func listAPIKeysHandler(keyManager *apikey.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		keys, err := keyManager.ListKeys(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list keys"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"keys": keys})
	}
}

func generateAPIKeyHandler(keyManager *apikey.Manager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID uuid.UUID `json:"user_id" binding:"required"`
			Name   string    `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		plaintext, err := keyManager.GenerateKey(c.Request.Context(), req.UserID, req.Name)
		if err != nil {
			logger.Error("failed to generate API key", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"api_key": plaintext,
			"message": "Store this key securely. It will not be shown again.",
		})
	}
}

func revokeAPIKeyHandler(keyManager *apikey.Manager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		if err := keyManager.RevokeKey(c.Request.Context(), userID); err != nil {
			logger.Error("failed to revoke API key", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke key"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "revoked"})
	}
}

func rotateAPIKeyHandler(keyManager *apikey.Manager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		plaintext, err := keyManager.RotateKey(c.Request.Context(), userID, req.Name)
		if err != nil {
			logger.Error("failed to rotate API key", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate key"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"api_key": plaintext,
			"message": "Store this key securely. It will not be shown again.",
		})
	}
}

// --- GDPR Erasure Handler ---
// Control: SEC-003 (Data Retention Controls)

func erasureHandler(purger *retention.Purger, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentHash := c.Param("hash")
		if contentHash == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content hash is required"})
			return
		}

		if err := purger.PurgeByContentHash(c.Request.Context(), contentHash); err != nil {
			logger.Error("GDPR erasure failed", zap.Error(err), zap.String("hash", contentHash))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erasure failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "erased",
			"message": "Content data anonymized per GDPR erasure request",
		})
	}
}

func fairnessReportHandler(db *database.PostgresDB, logger *zap.Logger) gin.HandlerFunc {
	detector := fairness.NewBiasDetector(db.Pool, logger)
	return func(c *gin.Context) {
		windowStr := c.DefaultQuery("window", "24h")
		window, err := time.ParseDuration(windowStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid window duration"})
			return
		}
		if window > 720*time.Hour { // max 30 days
			c.JSON(http.StatusBadRequest, gin.H{"error": "window cannot exceed 720h"})
			return
		}

		report, err := detector.GenerateReport(c.Request.Context(), window)
		if err != nil {
			logger.Error("failed to generate fairness report", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}
