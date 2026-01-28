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

// EvaluateScores evaluates category scores against a policy
func (e *Evaluator) EvaluateScores(ctx context.Context, scores *models.CategoryScores, policyID uuid.UUID) (*models.PolicyEvaluationResponse, error) {
	// Fetch policy from database
	policy, err := e.getPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if policy.Status != models.PolicyStatusPublished {
		return nil, fmt.Errorf("policy %s is not published (status: %s)", policyID, policy.Status)
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
	}

	for category, score := range categories {
		threshold, hasThreshold := policy.Thresholds[category]
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
