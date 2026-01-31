package engine

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/proth1/text-moderator/internal/models"
	"go.uber.org/zap"
)

// Control: POL-001 (Deterministic policy evaluation)

// EvaluationOptions provides optional context for policy evaluation.
type EvaluationOptions struct {
	// ContextMetadata from the moderation request (e.g., audience, platform).
	ContextMetadata map[string]interface{}
	// TrustScore from user behavioral scoring (0.0-1.0, nil if not available).
	TrustScore *float64
}

// ContextOverride defines a threshold adjustment based on context metadata matching.
type ContextOverride struct {
	Match                map[string]interface{} `json:"match"`
	ThresholdAdjustments map[string]float64     `json:"threshold_adjustments"`
}

// Evaluator handles policy evaluation logic
type Evaluator struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewEvaluator creates a new policy evaluator
func NewEvaluator(db *pgxpool.Pool, logger *zap.Logger) *Evaluator {
	return &Evaluator{
		db:     db,
		logger: logger,
	}
}

// Pool returns the database connection pool
func (e *Evaluator) Pool() *pgxpool.Pool {
	return e.db
}

// EvaluateScores evaluates category scores against a policy.
// Pass nil for opts if no context or trust score adjustments are needed.
func (e *Evaluator) EvaluateScores(ctx context.Context, scores *models.CategoryScores, policyID uuid.UUID, opts *EvaluationOptions) (*models.PolicyEvaluationResponse, error) {
	// Fetch policy from database
	policy, err := e.getPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if policy.Status != models.PolicyStatusPublished {
		return nil, fmt.Errorf("policy %s is not published (status: %s)", policyID, policy.Status)
	}

	// Build effective thresholds (start from policy defaults)
	effectiveThresholds := make(map[string]float64, len(policy.Thresholds))
	for k, v := range policy.Thresholds {
		effectiveThresholds[k] = v
	}

	// Apply context-aware overrides from policy scope
	if opts != nil && opts.ContextMetadata != nil {
		e.applyContextOverrides(effectiveThresholds, policy.Scope, opts.ContextMetadata)
	}

	// Apply trust score adjustments (low trust = stricter thresholds)
	if opts != nil && opts.TrustScore != nil {
		trust := *opts.TrustScore
		if trust < 0.5 {
			// Lower thresholds for untrusted users (stricter)
			adjustment := (0.5 - trust) * 0.2 // max 0.1 reduction at trust=0
			for k, v := range effectiveThresholds {
				effectiveThresholds[k] = v - adjustment
				if effectiveThresholds[k] < 0.1 {
					effectiveThresholds[k] = 0.1
				}
			}
		}
	}

	// Evaluate each category against thresholds
	triggeredRules := []string{}
	highestAction := models.ActionAllow

	categories := map[string]float64{
		"toxicity":        scores.Toxicity,
		"hate":            scores.Hate,
		"harassment":      scores.Harassment,
		"sexual_content":  scores.SexualContent,
		"violence":        scores.Violence,
		"profanity":       scores.Profanity,
		"self_harm":       scores.SelfHarm,
		"spam":            scores.Spam,
		"pii":             scores.PII,
	}

	for category, score := range categories {
		threshold, hasThreshold := effectiveThresholds[category]
		if !hasThreshold {
			continue
		}

		// Check if score exceeds threshold
		if score >= threshold {
			triggeredRules = append(triggeredRules, fmt.Sprintf("%s >= %.2f", category, threshold))

			// Get action for this category
			actionStr, hasAction := policy.Actions[category]
			if hasAction {
				action := models.PolicyAction(actionStr)
				if actionPriority(action) > actionPriority(highestAction) {
					highestAction = action
				}
			}
		}
	}

	e.logger.Info("policy evaluation completed",
		zap.String("policy_id", policyID.String()),
		zap.String("policy_name", policy.Name),
		zap.Int("policy_version", policy.Version),
		zap.String("action", string(highestAction)),
		zap.Strings("triggered_rules", triggeredRules),
	)

	return &models.PolicyEvaluationResponse{
		Action:         highestAction,
		PolicyID:       policy.ID,
		PolicyVersion:  policy.Version,
		TriggeredRules: triggeredRules,
	}, nil
}

// applyContextOverrides adjusts thresholds based on policy scope context_overrides and request metadata.
func (e *Evaluator) applyContextOverrides(thresholds map[string]float64, scope map[string]interface{}, metadata map[string]interface{}) {
	if scope == nil {
		return
	}

	overridesRaw, ok := scope["context_overrides"]
	if !ok {
		return
	}

	overrides, ok := overridesRaw.([]interface{})
	if !ok {
		return
	}

	for _, overrideRaw := range overrides {
		override, ok := overrideRaw.(map[string]interface{})
		if !ok {
			continue
		}

		matchRaw, ok := override["match"]
		if !ok {
			continue
		}
		match, ok := matchRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if all match conditions are met
		allMatch := true
		for key, expected := range match {
			actual, exists := metadata[key]
			if !exists || fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
				allMatch = false
				break
			}
		}

		if !allMatch {
			continue
		}

		// Apply threshold adjustments
		adjustmentsRaw, ok := override["threshold_adjustments"]
		if !ok {
			continue
		}
		adjustments, ok := adjustmentsRaw.(map[string]interface{})
		if !ok {
			continue
		}

		for category, adjRaw := range adjustments {
			adj, ok := adjRaw.(float64)
			if !ok {
				continue
			}
			if currentThreshold, exists := thresholds[category]; exists {
				newThreshold := currentThreshold + adj
				if newThreshold < 0.05 {
					newThreshold = 0.05
				}
				if newThreshold > 1.0 {
					newThreshold = 1.0
				}
				thresholds[category] = newThreshold
			}
		}
	}
}

// GetPolicyByID retrieves a specific policy from the database by ID
func (e *Evaluator) GetPolicyByID(ctx context.Context, policyID uuid.UUID) (*models.Policy, error) {
	return e.getPolicy(ctx, policyID)
}

// getPolicy retrieves a policy from the database
func (e *Evaluator) getPolicy(ctx context.Context, policyID uuid.UUID) (*models.Policy, error) {
	query := `
		SELECT id, name, version, thresholds, actions, scope, status, effective_date, created_at, created_by
		FROM policies
		WHERE id = $1
	`

	var policy models.Policy
	err := e.db.QueryRow(ctx, query, policyID).Scan(
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
		return nil, fmt.Errorf("failed to query policy: %w", err)
	}

	return &policy, nil
}

// GetDefaultPolicy retrieves the default published policy
func (e *Evaluator) GetDefaultPolicy(ctx context.Context) (*models.Policy, error) {
	query := `
		SELECT id, name, version, thresholds, actions, scope, status, effective_date, created_at, created_by
		FROM policies
		WHERE status = 'published'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var policy models.Policy
	err := e.db.QueryRow(ctx, query).Scan(
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
		return nil, fmt.Errorf("no default policy found: %w", err)
	}

	return &policy, nil
}

// actionPriority returns the priority of an action (higher = more restrictive)
func actionPriority(action models.PolicyAction) int {
	switch action {
	case models.ActionAllow:
		return 0
	case models.ActionWarn:
		return 1
	case models.ActionEscalate:
		return 2
	case models.ActionBlock:
		return 3
	default:
		return 0
	}
}

// CreatePolicy creates a new policy in the database
func (e *Evaluator) CreatePolicy(ctx context.Context, req *models.CreatePolicyRequest, createdBy uuid.UUID) (*models.Policy, error) {
	// Check if policy with this name already exists
	var maxVersion int
	versionQuery := `SELECT COALESCE(MAX(version), 0) FROM policies WHERE name = $1`
	err := e.db.QueryRow(ctx, versionQuery, req.Name).Scan(&maxVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing policy versions: %w", err)
	}

	newVersion := maxVersion + 1

	// Create policy
	policy := &models.Policy{
		ID:         uuid.New(),
		Name:       req.Name,
		Version:    newVersion,
		Thresholds: req.Thresholds,
		Actions:    req.Actions,
		Scope:      req.Scope,
		Status:     models.PolicyStatusDraft,
		CreatedBy:  &createdBy,
	}

	query := `
		INSERT INTO policies (id, name, version, thresholds, actions, scope, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`

	err = e.db.QueryRow(ctx, query,
		policy.ID,
		policy.Name,
		policy.Version,
		policy.Thresholds,
		policy.Actions,
		policy.Scope,
		policy.Status,
		policy.CreatedBy,
	).Scan(&policy.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	e.logger.Info("policy created",
		zap.String("policy_id", policy.ID.String()),
		zap.String("name", policy.Name),
		zap.Int("version", policy.Version),
	)

	return policy, nil
}

// ListPolicies retrieves policies with optional filtering
func (e *Evaluator) ListPolicies(ctx context.Context, status *models.PolicyStatus) ([]models.Policy, error) {
	query := `
		SELECT id, name, version, thresholds, actions, scope, status, effective_date, created_at, created_by
		FROM policies
		WHERE ($1::policy_status IS NULL OR status = $1)
		ORDER BY created_at DESC
	`

	rows, err := e.db.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	var policies []models.Policy
	for rows.Next() {
		var policy models.Policy
		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		policies = append(policies, policy)
	}

	return policies, nil
}
