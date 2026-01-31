package feedback

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/proth1/text-moderator/internal/models"
	"go.uber.org/zap"
)

// Tracker records human review outcomes for calibration feedback.
type Tracker struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewTracker creates a new feedback tracker.
func NewTracker(db *pgxpool.Pool, logger *zap.Logger) *Tracker {
	return &Tracker{
		db:     db,
		logger: logger,
	}
}

// RecordFeedback records a human review outcome and maps it to provider calibration data.
// approve/edit → agree (model was right or close enough)
// reject → disagree (model was wrong)
// escalate → uncertain (needs more review)
func (t *Tracker) RecordFeedback(ctx context.Context, decision *models.ModerationDecision, reviewAction models.ReviewActionType) {
	outcome := mapReviewToOutcome(reviewAction)

	query := `
		INSERT INTO calibration_data (id, provider_name, decision_id, category_scores, automated_action, review_outcome, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := t.db.Exec(ctx, query,
		uuid.New(),
		decision.ModelName,
		decision.ID,
		decision.CategoryScores,
		decision.AutomatedAction,
		outcome,
		time.Now(),
	)
	if err != nil {
		t.logger.Warn("failed to record calibration feedback",
			zap.String("decision_id", decision.ID.String()),
			zap.String("outcome", outcome),
			zap.Error(err),
		)
	}
}

// ProviderAccuracy holds accuracy metrics for a classification provider.
type ProviderAccuracy struct {
	ProviderName   string  `json:"provider_name"`
	TotalDecisions int     `json:"total_decisions"`
	AgreeCount     int     `json:"agree_count"`
	DisagreeCount  int     `json:"disagree_count"`
	UncertainCount int     `json:"uncertain_count"`
	AccuracyRate   float64 `json:"accuracy_rate"`
}

// GetProviderAccuracy retrieves accuracy metrics per provider from calibration data.
func (t *Tracker) GetProviderAccuracy(ctx context.Context) ([]ProviderAccuracy, error) {
	query := `
		SELECT
			provider_name,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE review_outcome = 'agree') AS agree,
			COUNT(*) FILTER (WHERE review_outcome = 'disagree') AS disagree,
			COUNT(*) FILTER (WHERE review_outcome = 'uncertain') AS uncertain
		FROM calibration_data
		WHERE created_at >= NOW() - INTERVAL '30 days'
		GROUP BY provider_name
		ORDER BY provider_name
	`

	rows, err := t.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ProviderAccuracy
	for rows.Next() {
		var pa ProviderAccuracy
		if err := rows.Scan(&pa.ProviderName, &pa.TotalDecisions, &pa.AgreeCount, &pa.DisagreeCount, &pa.UncertainCount); err != nil {
			return nil, err
		}
		if pa.TotalDecisions > 0 {
			pa.AccuracyRate = float64(pa.AgreeCount) / float64(pa.TotalDecisions)
		}
		results = append(results, pa)
	}

	return results, nil
}

func mapReviewToOutcome(action models.ReviewActionType) string {
	switch action {
	case models.ReviewActionApprove, models.ReviewActionEdit:
		return "agree"
	case models.ReviewActionReject:
		return "disagree"
	case models.ReviewActionEscalate:
		return "uncertain"
	default:
		return "uncertain"
	}
}
