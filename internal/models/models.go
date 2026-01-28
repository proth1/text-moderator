package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user in the system
// Control: GOV-002 (Role-based access control)
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleModerator UserRole = "moderator"
	RoleViewer    UserRole = "viewer"
)

// User represents a system user
// Control: GOV-002 (User management and access control)
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	APIKey    *string   `json:"api_key,omitempty" db:"api_key"`
	Role      UserRole  `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PolicyStatus represents the status of a policy
type PolicyStatus string

const (
	PolicyStatusDraft     PolicyStatus = "draft"
	PolicyStatusPublished PolicyStatus = "published"
	PolicyStatusArchived  PolicyStatus = "archived"
)

// PolicyAction represents the action to be taken for moderation
// Control: POL-001 (Policy-driven decision making)
type PolicyAction string

const (
	ActionAllow    PolicyAction = "allow"
	ActionWarn     PolicyAction = "warn"
	ActionBlock    PolicyAction = "block"
	ActionEscalate PolicyAction = "escalate"
)

// Policy represents a moderation policy
// Control: POL-001 (Policy definition and versioning)
type Policy struct {
	ID            uuid.UUID               `json:"id" db:"id"`
	Name          string                  `json:"name" db:"name"`
	Version       int                     `json:"version" db:"version"`
	Thresholds    map[string]float64      `json:"thresholds" db:"thresholds"`
	Actions       map[string]PolicyAction `json:"actions" db:"actions"`
	Scope         map[string]interface{}  `json:"scope" db:"scope"`
	Status        PolicyStatus            `json:"status" db:"status"`
	EffectiveDate *time.Time              `json:"effective_date,omitempty" db:"effective_date"`
	CreatedAt     time.Time               `json:"created_at" db:"created_at"`
	CreatedBy     *uuid.UUID              `json:"created_by,omitempty" db:"created_by"`
}

// CategoryScores represents the confidence scores for each moderation category
// Control: MOD-001 (AI model output structure)
type CategoryScores struct {
	Toxicity       float64 `json:"toxicity"`
	Hate           float64 `json:"hate"`
	Harassment     float64 `json:"harassment"`
	SexualContent  float64 `json:"sexual_content"`
	Violence       float64 `json:"violence"`
	Profanity      float64 `json:"profanity"`
}

// TextSubmission represents a text submission for moderation
// Control: MOD-001 (Input tracking and hashing)
type TextSubmission struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	ContentHash      string                 `json:"content_hash" db:"content_hash"`
	ContentEncrypted *string                `json:"-" db:"content_encrypted"` // Encrypted, not exposed in JSON
	ContextMetadata  map[string]interface{} `json:"context_metadata" db:"context_metadata"`
	Source           *string                `json:"source,omitempty" db:"source"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// ModerationDecision represents the result of content moderation
// Control: MOD-001 (Decision tracking and traceability)
type ModerationDecision struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	SubmissionID    uuid.UUID       `json:"submission_id" db:"submission_id"`
	ModelName       string          `json:"model_name" db:"model_name"`
	ModelVersion    string          `json:"model_version" db:"model_version"`
	CategoryScores  CategoryScores  `json:"category_scores" db:"category_scores"`
	PolicyID        *uuid.UUID      `json:"policy_id,omitempty" db:"policy_id"`
	PolicyVersion   *int            `json:"policy_version,omitempty" db:"policy_version"`
	AutomatedAction PolicyAction    `json:"automated_action" db:"automated_action"`
	Confidence      *float64        `json:"confidence,omitempty" db:"confidence"`
	Explanation     *string         `json:"explanation,omitempty" db:"explanation"`
	CorrelationID   *uuid.UUID      `json:"correlation_id,omitempty" db:"correlation_id"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// ReviewActionType represents the type of action taken during human review
// Control: GOV-002 (Human oversight actions)
type ReviewActionType string

const (
	ReviewActionApprove  ReviewActionType = "approve"
	ReviewActionReject   ReviewActionType = "reject"
	ReviewActionEdit     ReviewActionType = "edit"
	ReviewActionEscalate ReviewActionType = "escalate"
)

// ReviewAction represents a human review action
// Control: GOV-002 (Human review workflow)
type ReviewAction struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	DecisionID    uuid.UUID        `json:"decision_id" db:"decision_id"`
	ReviewerID    uuid.UUID        `json:"reviewer_id" db:"reviewer_id"`
	Action        ReviewActionType `json:"action" db:"action"`
	Rationale     *string          `json:"rationale,omitempty" db:"rationale"`
	EditedContent *string          `json:"edited_content,omitempty" db:"edited_content"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
}

// EvidenceRecord represents an immutable audit record
// Control: AUD-001 (Immutable evidence generation)
type EvidenceRecord struct {
	ID              uuid.UUID         `json:"id" db:"id"`
	ControlID       string            `json:"control_id" db:"control_id"`
	PolicyID        *uuid.UUID        `json:"policy_id,omitempty" db:"policy_id"`
	PolicyVersion   *int              `json:"policy_version,omitempty" db:"policy_version"`
	DecisionID      *uuid.UUID        `json:"decision_id,omitempty" db:"decision_id"`
	ReviewID        *uuid.UUID        `json:"review_id,omitempty" db:"review_id"`
	ModelName       *string           `json:"model_name,omitempty" db:"model_name"`
	ModelVersion    *string           `json:"model_version,omitempty" db:"model_version"`
	CategoryScores  *CategoryScores   `json:"category_scores,omitempty" db:"category_scores"`
	AutomatedAction *PolicyAction     `json:"automated_action,omitempty" db:"automated_action"`
	HumanOverride   *ReviewActionType `json:"human_override,omitempty" db:"human_override"`
	SubmissionHash  *string           `json:"submission_hash,omitempty" db:"submission_hash"`
	Immutable       bool              `json:"immutable" db:"immutable"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}

// ModerationRequest represents an incoming moderation request
type ModerationRequest struct {
	Content         string                 `json:"content" binding:"required"`
	ContextMetadata map[string]interface{} `json:"context_metadata,omitempty"`
	Source          string                 `json:"source,omitempty"`
	PolicyID        *uuid.UUID             `json:"policy_id,omitempty"`
}

// ModerationResponse represents the response from moderation
type ModerationResponse struct {
	DecisionID      uuid.UUID       `json:"decision_id"`
	SubmissionID    uuid.UUID       `json:"submission_id"`
	Action          PolicyAction    `json:"action"`
	CategoryScores  CategoryScores  `json:"category_scores"`
	Confidence      *float64        `json:"confidence,omitempty"`
	Explanation     *string         `json:"explanation,omitempty"`
	PolicyApplied   *string         `json:"policy_applied,omitempty"`
	PolicyVersion   *int            `json:"policy_version,omitempty"`
	RequiresReview  bool            `json:"requires_review"`
}

// PolicyEvaluationRequest represents a request to evaluate scores against a policy
type PolicyEvaluationRequest struct {
	CategoryScores CategoryScores `json:"category_scores" binding:"required"`
	PolicyID       uuid.UUID      `json:"policy_id" binding:"required"`
}

// PolicyEvaluationResponse represents the result of policy evaluation
type PolicyEvaluationResponse struct {
	Action         PolicyAction `json:"action"`
	PolicyID       uuid.UUID    `json:"policy_id"`
	PolicyVersion  int          `json:"policy_version"`
	TriggeredRules []string     `json:"triggered_rules,omitempty"`
}

// CreatePolicyRequest represents a request to create a new policy
type CreatePolicyRequest struct {
	Name       string                  `json:"name" binding:"required"`
	Thresholds map[string]float64      `json:"thresholds" binding:"required"`
	Actions    map[string]PolicyAction `json:"actions" binding:"required"`
	Scope      map[string]interface{}  `json:"scope,omitempty"`
}

// ReviewQueueItem represents an item in the review queue
type ReviewQueueItem struct {
	DecisionID      uuid.UUID      `json:"decision_id"`
	SubmissionID    uuid.UUID      `json:"submission_id"`
	ContentHash     string         `json:"content_hash"`
	CategoryScores  CategoryScores `json:"category_scores"`
	AutomatedAction PolicyAction   `json:"automated_action"`
	PolicyName      *string        `json:"policy_name,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

// SubmitReviewRequest represents a request to submit a review action
type SubmitReviewRequest struct {
	Action        ReviewActionType `json:"action" binding:"required"`
	Rationale     string           `json:"rationale,omitempty"`
	EditedContent string           `json:"edited_content,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks,omitempty"`
}
