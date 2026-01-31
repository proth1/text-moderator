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
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/proth1/text-moderator/internal/behavior"
	"github.com/proth1/text-moderator/internal/cache"
	"github.com/proth1/text-moderator/internal/classifier"
	"github.com/proth1/text-moderator/internal/config"
	"github.com/proth1/text-moderator/internal/database"
	"github.com/proth1/text-moderator/internal/evidence"
	"github.com/proth1/text-moderator/internal/langdetect"
	"github.com/proth1/text-moderator/internal/middleware"
	"github.com/proth1/text-moderator/internal/models"
	"github.com/proth1/text-moderator/internal/normalizer"
	"github.com/proth1/text-moderator/internal/webhook"
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

	// Initialize Redis cache (optional - degrades gracefully if unavailable)
	var redisCache *cache.RedisCache
	if cfg.RedisURL != "" {
		redisCache, err = cache.NewRedisCache(ctx, cache.Config{
			URL:         cfg.RedisURL,
			MaxRetries:  3,
			DialTimeout: 5 * time.Second,
			ReadTimeout: 3 * time.Second,
		}, logger)
		if err != nil {
			logger.Warn("redis unavailable, classification caching disabled", zap.Error(err))
		} else {
			defer redisCache.Close()
			logger.Info("redis cache enabled for classification results")
		}
	}

	// Initialize HuggingFace client
	hfClient := client.NewHuggingFaceClient(client.Config{
		APIKey:   cfg.HuggingFaceAPIKey,
		ModelURL: cfg.HuggingFaceModelURL,
		Timeout:  cfg.HuggingFaceTimeout,
	}, logger)

	// Initialize multi-provider classification orchestrator
	providerConfigs := []classifier.ProviderConfig{
		{Name: "huggingface", Priority: 1, Enabled: true},
	}
	if cfg.PerspectiveAPIKey != "" {
		providerConfigs = append(providerConfigs, classifier.ProviderConfig{
			Name: "perspective", Priority: 2, Enabled: true,
		})
	}
	if cfg.OpenAIAPIKey != "" {
		providerConfigs = append(providerConfigs, classifier.ProviderConfig{
			Name: "openai", Priority: 3, Enabled: true,
		})
	}

	orchestrator := classifier.NewOrchestrator(classifier.OrchestratorConfig{
		Providers:       providerConfigs,
		FallbackEnabled: true,
	}, logger)

	// Register providers
	orchestrator.RegisterProvider(classifier.NewHuggingFaceProvider(hfClient))
	if cfg.PerspectiveAPIKey != "" {
		orchestrator.RegisterProvider(classifier.NewPerspectiveProvider(classifier.PerspectiveConfig{
			APIKey:  cfg.PerspectiveAPIKey,
			Timeout: 10 * time.Second,
		}, logger))
	}
	if cfg.OpenAIAPIKey != "" {
		orchestrator.RegisterProvider(classifier.NewOpenAIProvider(classifier.OpenAIConfig{
			APIKey:  cfg.OpenAIAPIKey,
			Timeout: 10 * time.Second,
		}, logger))
	}

	// Initialize optional score calibrator
	if cal := classifier.NewCalibratorFromJSON(cfg.CalibrationConfigJSON); cal != nil {
		orchestrator.SetCalibrator(cal)
		logger.Info("score calibration enabled")
	}

	// Initialize optional LLM provider for second-pass classification
	var llmProvider *classifier.LLMProvider
	if cfg.LLMProvider != "" && cfg.LLMAPIKey != "" {
		llmProvider = classifier.NewLLMProvider(classifier.LLMConfig{
			Provider: cfg.LLMProvider,
			APIKey:   cfg.LLMAPIKey,
			Model:    cfg.LLMModel,
		}, logger)
		logger.Info("LLM second-pass classification enabled",
			zap.String("provider", cfg.LLMProvider),
			zap.String("model", cfg.LLMModel),
		)
	}

	// Configure ensemble mode if enabled
	if cfg.EnsembleEnabled {
		orchestrator.SetEnsembleConfig(&classifier.EnsembleConfig{
			Enabled:            true,
			MinProviders:       2,
			AgreementThreshold: 0.3,
			Strategy:           cfg.EnsembleStrategy,
		})
		logger.Info("ensemble classification mode enabled", zap.String("strategy", cfg.EnsembleStrategy))
	}

	// Initialize text normalizer for Unicode evasion defense
	textNormalizer := normalizer.New()

	// Initialize language detector
	langDetector := langdetect.New()

	// Initialize policy evaluator
	evaluator := engine.NewEvaluator(db.Pool, logger)

	// Initialize behavioral scorer for user trust scores
	behaviorScorer := behavior.NewScorer(db.Pool, logger)

	// Initialize evidence writer
	evidenceWriter := evidence.NewWriter(db.Pool, logger)

	// Initialize webhook dispatcher
	webhookDispatcher := webhook.NewDispatcher(db.Pool, logger)

	// Create HTTP server
	router := setupRouter(cfg, logger, db, hfClient, evaluator, evidenceWriter, redisCache, orchestrator, webhookDispatcher, textNormalizer, langDetector, llmProvider, behaviorScorer)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.ModerationPort),
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
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

func setupRouter(cfg *config.Config, logger *zap.Logger, db *database.PostgresDB, hfClient *client.HuggingFaceClient, evaluator *engine.Evaluator, evidenceWriter *evidence.Writer, redisCache *cache.RedisCache, orchestrator *classifier.Orchestrator, webhookDispatcher *webhook.Dispatcher, textNormalizer *normalizer.Normalizer, langDetector *langdetect.Detector, llmProvider *classifier.LLMProvider, behaviorScorer *behavior.Scorer) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.CORSMiddleware(middleware.CORSConfigFromOrigins(cfg.AllowedOrigins)))

	// Health check - public for load balancer probes
	router.GET("/health", healthHandler(db, hfClient, logger))

	// Moderation endpoint - requires internal service authentication
	// In production, INTERNAL_SERVICE_TOKEN must be set
	// In development, we allow requests if token is not configured (with a warning)
	api := router.Group("/")
	if cfg.InternalServiceToken != "" {
		api.Use(middleware.InternalServiceAuthMiddleware(cfg.InternalServiceToken, logger))
	} else if cfg.Environment == "production" {
		logger.Error("CRITICAL: INTERNAL_SERVICE_TOKEN not configured in production - all moderation requests will fail")
		api.Use(middleware.InternalServiceAuthMiddleware("", logger)) // This will reject all requests
	} else {
		logger.Warn("INTERNAL_SERVICE_TOKEN not configured - internal endpoints are unprotected (development mode only)")
	}
	api.POST("/moderate", moderateHandler(db, orchestrator, evaluator, evidenceWriter, redisCache, webhookDispatcher, cfg, logger, textNormalizer, langDetector, llmProvider, behaviorScorer))
	api.POST("/moderate/batch", batchModerateHandler(db, orchestrator, evaluator, evidenceWriter, redisCache, webhookDispatcher, cfg, logger, textNormalizer, langDetector, llmProvider, behaviorScorer))

	return router
}

func healthHandler(db *database.PostgresDB, hfClient *client.HuggingFaceClient, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]string)

		// Check database
		// SECURITY: Don't expose error details in health check responses
		if err := db.Health(ctx); err != nil {
			logger.Warn("database health check failed", zap.Error(err))
			checks["database"] = "unhealthy"
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

// classificationCacheTTL is how long cached classification results remain valid.
const classificationCacheTTL = 15 * time.Minute

func moderateHandler(db *database.PostgresDB, orchestrator *classifier.Orchestrator, evaluator *engine.Evaluator, evidenceWriter *evidence.Writer, redisCache *cache.RedisCache, webhookDispatcher *webhook.Dispatcher, cfg *config.Config, logger *zap.Logger, textNormalizer *normalizer.Normalizer, langDetector *langdetect.Detector, llmProvider *classifier.LLMProvider, behaviorScorer *behavior.Scorer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ModerationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// SECURITY: Don't expose detailed parsing errors to clients
			logger.Debug("invalid request body", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
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

		// Normalize text to defeat Unicode evasion before hashing and classification
		normalizedContent := textNormalizer.Normalize(req.Content)

		// Hash normalized content for deduplication
		hash := sha256.Sum256([]byte(normalizedContent))
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

		// Check cache for classification results by content hash
		// Control: MOD-004 (Latency Optimization and Caching)
		var scores *models.CategoryScores
		cacheKey := "classify:" + contentHash
		cacheHit := false

		if redisCache != nil {
			cached, err := redisCache.Get(ctx, cacheKey)
			if err == nil {
				var cachedScores models.CategoryScores
				if json.Unmarshal([]byte(cached), &cachedScores) == nil {
					scores = &cachedScores
					cacheHit = true
					logger.Debug("classification cache hit", zap.String("content_hash", contentHash))
				}
			}
		}

		// Detect language before classification
		langResult := langDetector.Detect(normalizedContent)

		// Classify normalized text using orchestrator if not cached
		var classResult *classifier.ClassificationResult
		var ensembleResult *classifier.EnsembleResult
		if !cacheHit {
			if orchestrator.IsEnsembleEnabled() {
				// Ensemble mode: run multiple providers in parallel
				ensembleResult, err = orchestrator.ClassifyEnsemble(ctx, normalizedContent)
				if err != nil {
					logger.Error("failed to classify text (ensemble)", zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to classify text"})
					return
				}
				scores = ensembleResult.CombinedScores
				if len(ensembleResult.ProviderResults) > 0 {
					classResult = &ensembleResult.ProviderResults[0]
				}
			} else {
				classResult, err = orchestrator.ClassifyWithLanguage(ctx, normalizedContent, langResult.Language)
				if err != nil {
					logger.Error("failed to classify text", zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to classify text"})
					return
				}
				scores = classResult.Scores
			}

			// Store in cache
			if redisCache != nil {
				if scoresJSON, err := json.Marshal(scores); err == nil {
					if err := redisCache.Set(ctx, cacheKey, string(scoresJSON), classificationCacheTTL); err != nil {
						logger.Warn("failed to cache classification result", zap.Error(err))
					}
				}
			}
		}

		// LLM second-pass for ambiguous scores (0.3-0.7 range)
		if llmProvider != nil && !cacheHit && classifier.IsAmbiguous(scores, 0.3, 0.7) {
			llmScores, llmErr := llmProvider.Classify(ctx, normalizedContent)
			if llmErr != nil {
				logger.Warn("LLM second-pass failed, using primary scores", zap.Error(llmErr))
			} else {
				scores = classifier.MergeAmbiguousScores(scores, llmScores, 0.3, 0.7)
				logger.Debug("LLM second-pass merged ambiguous scores")
			}
		}

		// Determine model info (from orchestrator result or default for cache hits)
		modelName := "s-nlp/roberta_toxicity_classifier"
		modelVersion := "v1"
		if classResult != nil {
			modelName = classResult.ModelName
			modelVersion = classResult.ModelVersion
		}

		// Get policy (use provided or default)
		var policy *models.Policy
		if req.PolicyID != nil {
			policy, err = evaluator.GetPolicyByID(ctx, *req.PolicyID)
			if err != nil {
				logger.Warn("requested policy not found, falling back to default",
					zap.String("requested_policy_id", req.PolicyID.String()),
					zap.Error(err),
				)
				policy, err = evaluator.GetDefaultPolicy(ctx)
			}
		} else {
			policy, err = evaluator.GetDefaultPolicy(ctx)
		}
		if err != nil {
			logger.Warn("no policy found, defaulting to allow", zap.Error(err))
			policy = &models.Policy{
				ID:      uuid.New(),
				Name:    "default",
				Version: 1,
			}
		}

		// Evaluate against policy with context metadata and trust score
		var action models.PolicyAction
		evalOpts := &engine.EvaluationOptions{
			ContextMetadata: req.ContextMetadata,
		}

		// Lookup user trust score if user_id is in context metadata
		var userID string
		if req.ContextMetadata != nil {
			if uid, ok := req.ContextMetadata["user_id"]; ok {
				userID = fmt.Sprintf("%v", uid)
				trustScore := behaviorScorer.GetTrustScore(ctx, userID)
				evalOpts.TrustScore = &trustScore
			}
		}

		if policy.Name != "default" {
			evalResult, err := evaluator.EvaluateScores(ctx, scores, policy.ID, evalOpts)
			if err != nil {
				logger.Error("failed to evaluate policy", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to evaluate policy"})
				return
			}
			action = evalResult.Action
		} else {
			action = models.ActionAllow
		}

		// Create decision record and evidence atomically in a transaction
		decision := &models.ModerationDecision{
			ID:              uuid.New(),
			SubmissionID:    submission.ID,
			ModelName:       modelName,
			ModelVersion:    modelVersion,
			CategoryScores:  *scores,
			PolicyID:        &policy.ID,
			PolicyVersion:   &policy.Version,
			AutomatedAction: action,
		}

		tx, err := evidenceWriter.BeginTx(ctx)
		if err != nil {
			logger.Error("failed to begin transaction", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer tx.Rollback(ctx)

		decisionQuery := `
			INSERT INTO moderation_decisions (
				id, submission_id, model_name, model_version, category_scores,
				policy_id, policy_version, automated_action
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING created_at
		`
		err = tx.QueryRow(ctx, decisionQuery,
			decision.ID, decision.SubmissionID, decision.ModelName, decision.ModelVersion,
			decision.CategoryScores, decision.PolicyID, decision.PolicyVersion, decision.AutomatedAction,
		).Scan(&decision.CreatedAt)
		if err != nil {
			logger.Error("failed to create decision", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create decision"})
			return
		}

		// Write evidence record within the same transaction
		evidenceRecord := &models.EvidenceRecord{
			ID:              uuid.New(),
			ControlID:       "MOD-001",
			PolicyID:        decision.PolicyID,
			PolicyVersion:   decision.PolicyVersion,
			DecisionID:      &decision.ID,
			ModelName:       &decision.ModelName,
			ModelVersion:    &decision.ModelVersion,
			CategoryScores:  &decision.CategoryScores,
			AutomatedAction: &decision.AutomatedAction,
			Immutable:       true,
		}
		if err := evidenceWriter.WriteEvidenceInTx(ctx, tx, evidenceRecord); err != nil {
			logger.Error("failed to write evidence in transaction", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record evidence"})
			return
		}

		if err := tx.Commit(ctx); err != nil {
			logger.Error("failed to commit transaction", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		// Auto-escalate on ensemble disagreement
		if ensembleResult != nil && ensembleResult.HasDisagreement && action != models.ActionBlock {
			action = models.ActionEscalate
			logger.Info("auto-escalated due to ensemble disagreement",
				zap.Strings("disagreed_categories", ensembleResult.DisagreedCategories),
			)
		}

		// Prepare response
		requiresReview := action == models.ActionEscalate
		response := models.ModerationResponse{
			DecisionID:       decision.ID,
			SubmissionID:     submission.ID,
			Action:           action,
			CategoryScores:   *scores,
			RequiresReview:   requiresReview,
			DetectedLanguage: langResult.Language,
		}

		if policy.Name != "default" {
			response.PolicyApplied = &policy.Name
			response.PolicyVersion = &policy.Version
		}

		c.JSON(http.StatusOK, response)

		// Dispatch webhook events and record behavior asynchronously (non-blocking)
		go func() {
			bgCtx := context.Background()
			webhookDispatcher.Dispatch(bgCtx, models.EventModerationCompleted, response)
			if requiresReview {
				webhookDispatcher.Dispatch(bgCtx, models.EventReviewRequired, response)
			}
			// Record user behavior outcome
			if userID != "" {
				behaviorScorer.RecordOutcome(bgCtx, userID, string(action))
			}
		}()
	}
}

// maxBatchSize limits the number of items in a single batch request.
const maxBatchSize = 100

// batchWorkerPool controls concurrent classification requests.
const batchWorkerPool = 10

func batchModerateHandler(db *database.PostgresDB, orchestrator *classifier.Orchestrator, evaluator *engine.Evaluator, evidenceWriter *evidence.Writer, redisCache *cache.RedisCache, webhookDispatcher *webhook.Dispatcher, cfg *config.Config, logger *zap.Logger, textNormalizer *normalizer.Normalizer, langDetector *langdetect.Detector, llmProvider *classifier.LLMProvider, behaviorScorer *behavior.Scorer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.BatchModerationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "items array is empty"})
			return
		}
		if len(req.Items) > maxBatchSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("batch size exceeds maximum of %d items", maxBatchSize),
			})
			return
		}

		ctx := c.Request.Context()
		results := make([]models.BatchModerationResult, len(req.Items))

		// Process items concurrently with a worker pool
		sem := make(chan struct{}, batchWorkerPool)
		var wg sync.WaitGroup

		for i, item := range req.Items {
			wg.Add(1)
			go func(idx int, item models.BatchModerationItem) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				result := processBatchItem(ctx, db, orchestrator, evaluator, evidenceWriter, redisCache, cfg, logger, textNormalizer, langDetector, llmProvider, behaviorScorer, item)
				results[idx] = result
			}(i, item)
		}

		wg.Wait()

		// Build summary
		summary := models.BatchSummary{Total: len(results)}
		for _, r := range results {
			if r.Error != "" {
				summary.Failed++
				continue
			}
			switch r.Action {
			case models.ActionAllow:
				summary.Allowed++
			case models.ActionWarn:
				summary.Warned++
			case models.ActionBlock:
				summary.Blocked++
			case models.ActionEscalate:
				summary.Escalated++
			}
		}

		c.JSON(http.StatusOK, models.BatchModerationResponse{
			Results: results,
			Summary: summary,
		})
	}
}

func processBatchItem(ctx context.Context, db *database.PostgresDB, orchestrator *classifier.Orchestrator, evaluator *engine.Evaluator, evidenceWriter *evidence.Writer, redisCache *cache.RedisCache, cfg *config.Config, logger *zap.Logger, textNormalizer *normalizer.Normalizer, langDetector *langdetect.Detector, llmProvider *classifier.LLMProvider, behaviorScorer *behavior.Scorer, item models.BatchModerationItem) models.BatchModerationResult {
	result := models.BatchModerationResult{ItemID: item.ID}

	if len(item.Content) > cfg.MaxContentLength {
		result.Error = "content exceeds maximum length"
		return result
	}

	// Normalize text to defeat Unicode evasion
	normalizedContent := textNormalizer.Normalize(item.Content)

	// Hash normalized content
	hash := sha256.Sum256([]byte(normalizedContent))
	contentHash := hex.EncodeToString(hash[:])

	// Create submission
	submission := &models.TextSubmission{
		ID:              uuid.New(),
		ContentHash:     contentHash,
		ContextMetadata: item.ContextMetadata,
		Source:          &item.Source,
	}

	query := `INSERT INTO text_submissions (id, content_hash, context_metadata, source) VALUES ($1, $2, $3, $4) RETURNING created_at`
	if err := db.Pool.QueryRow(ctx, query, submission.ID, submission.ContentHash, submission.ContextMetadata, submission.Source).Scan(&submission.CreatedAt); err != nil {
		result.Error = "failed to create submission"
		return result
	}

	// Check cache
	var scores *models.CategoryScores
	cacheKey := "classify:" + contentHash
	var classResult *classifier.ClassificationResult

	if redisCache != nil {
		cached, err := redisCache.Get(ctx, cacheKey)
		if err == nil {
			var cachedScores models.CategoryScores
			if json.Unmarshal([]byte(cached), &cachedScores) == nil {
				scores = &cachedScores
			}
		}
	}

	if scores == nil {
		var err error
		langResult := langDetector.Detect(normalizedContent)
		classResult, err = orchestrator.ClassifyWithLanguage(ctx, normalizedContent, langResult.Language)
		if err != nil {
			result.Error = "classification failed"
			return result
		}
		scores = classResult.Scores

		if redisCache != nil {
			if scoresJSON, err := json.Marshal(scores); err == nil {
				redisCache.Set(ctx, cacheKey, string(scoresJSON), classificationCacheTTL)
			}
		}
	}

	// Get policy
	var policy *models.Policy
	var err error
	if item.PolicyID != nil {
		policy, err = evaluator.GetPolicyByID(ctx, *item.PolicyID)
		if err != nil {
			policy, err = evaluator.GetDefaultPolicy(ctx)
		}
	} else {
		policy, err = evaluator.GetDefaultPolicy(ctx)
	}
	if err != nil {
		policy = &models.Policy{ID: uuid.New(), Name: "default", Version: 1}
	}

	// Evaluate with context metadata
	var action models.PolicyAction
	evalOpts := &engine.EvaluationOptions{
		ContextMetadata: item.ContextMetadata,
	}
	if policy.Name != "default" {
		evalResult, err := evaluator.EvaluateScores(ctx, scores, policy.ID, evalOpts)
		if err != nil {
			result.Error = "policy evaluation failed"
			return result
		}
		action = evalResult.Action
	} else {
		action = models.ActionAllow
	}

	// Create decision
	modelName := "s-nlp/roberta_toxicity_classifier"
	modelVersion := "v1"
	if classResult != nil {
		modelName = classResult.ModelName
		modelVersion = classResult.ModelVersion
	}

	decision := &models.ModerationDecision{
		ID: uuid.New(), SubmissionID: submission.ID,
		ModelName: modelName, ModelVersion: modelVersion,
		CategoryScores: *scores, PolicyID: &policy.ID,
		PolicyVersion: &policy.Version, AutomatedAction: action,
	}

	tx, err := evidenceWriter.BeginTx(ctx)
	if err != nil {
		result.Error = "internal error"
		return result
	}
	defer tx.Rollback(ctx)

	decisionQuery := `INSERT INTO moderation_decisions (id, submission_id, model_name, model_version, category_scores, policy_id, policy_version, automated_action) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at`
	if err := tx.QueryRow(ctx, decisionQuery, decision.ID, decision.SubmissionID, decision.ModelName, decision.ModelVersion, decision.CategoryScores, decision.PolicyID, decision.PolicyVersion, decision.AutomatedAction).Scan(&decision.CreatedAt); err != nil {
		result.Error = "failed to create decision"
		return result
	}

	evidenceRecord := &models.EvidenceRecord{
		ID: uuid.New(), ControlID: "MOD-001",
		PolicyID: decision.PolicyID, PolicyVersion: decision.PolicyVersion,
		DecisionID: &decision.ID, ModelName: &decision.ModelName,
		ModelVersion: &decision.ModelVersion, CategoryScores: &decision.CategoryScores,
		AutomatedAction: &decision.AutomatedAction, Immutable: true,
	}
	if err := evidenceWriter.WriteEvidenceInTx(ctx, tx, evidenceRecord); err != nil {
		result.Error = "failed to record evidence"
		return result
	}

	if err := tx.Commit(ctx); err != nil {
		result.Error = "internal error"
		return result
	}

	result.DecisionID = decision.ID
	result.Action = action
	result.CategoryScores = scores
	result.RequiresReview = action == models.ActionEscalate
	return result
}
