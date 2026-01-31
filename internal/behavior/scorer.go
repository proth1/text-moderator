package behavior

import (
	"context"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Scorer tracks per-user moderation outcomes and computes trust scores.
type Scorer struct {
	db     *pgxpool.Pool
	logger *zap.Logger
	window time.Duration // rolling window for trust calculation
}

// NewScorer creates a new behavioral scorer with a 30-day rolling window.
func NewScorer(db *pgxpool.Pool, logger *zap.Logger) *Scorer {
	return &Scorer{
		db:     db,
		logger: logger,
		window: 30 * 24 * time.Hour,
	}
}

// GetTrustScore computes a trust score for the given user ID.
// Returns 0.5 (neutral) if no history exists.
// Formula: trust = clamp(allowed/total - blocked*0.1 - escalated*0.05, 0, 1)
func (s *Scorer) GetTrustScore(ctx context.Context, userID string) float64 {
	if userID == "" {
		return 0.5
	}

	query := `
		SELECT
			COALESCE(SUM(total_decisions), 0) AS total,
			COALESCE(SUM(allowed_count), 0) AS allowed,
			COALESCE(SUM(blocked_count), 0) AS blocked,
			COALESCE(SUM(escalated_count), 0) AS escalated
		FROM user_behavior_stats
		WHERE user_id = $1 AND window_start >= $2
	`

	windowStart := time.Now().Add(-s.window)

	var total, allowed, blocked, escalated int
	err := s.db.QueryRow(ctx, query, userID, windowStart).Scan(&total, &allowed, &blocked, &escalated)
	if err != nil {
		// No history or error - return neutral trust
		return 0.5
	}

	if total == 0 {
		return 0.5
	}

	trust := float64(allowed)/float64(total) - float64(blocked)*0.1 - float64(escalated)*0.05
	return math.Max(0.0, math.Min(1.0, trust))
}

// RecordOutcome records a moderation outcome for a user.
func (s *Scorer) RecordOutcome(ctx context.Context, userID string, action string) {
	if userID == "" {
		return
	}

	windowStart := time.Now().Truncate(24 * time.Hour) // daily buckets

	query := `
		INSERT INTO user_behavior_stats (user_id, window_start, total_decisions, allowed_count, blocked_count, escalated_count, warned_count, updated_at)
		VALUES ($1, $2, 1,
			CASE WHEN $3 = 'allow' THEN 1 ELSE 0 END,
			CASE WHEN $3 = 'block' THEN 1 ELSE 0 END,
			CASE WHEN $3 = 'escalate' THEN 1 ELSE 0 END,
			CASE WHEN $3 = 'warn' THEN 1 ELSE 0 END,
			NOW()
		)
		ON CONFLICT (user_id, window_start) DO UPDATE SET
			total_decisions = user_behavior_stats.total_decisions + 1,
			allowed_count = user_behavior_stats.allowed_count + CASE WHEN $3 = 'allow' THEN 1 ELSE 0 END,
			blocked_count = user_behavior_stats.blocked_count + CASE WHEN $3 = 'block' THEN 1 ELSE 0 END,
			escalated_count = user_behavior_stats.escalated_count + CASE WHEN $3 = 'escalate' THEN 1 ELSE 0 END,
			warned_count = user_behavior_stats.warned_count + CASE WHEN $3 = 'warn' THEN 1 ELSE 0 END,
			updated_at = NOW()
	`

	if _, err := s.db.Exec(ctx, query, userID, windowStart, action); err != nil {
		s.logger.Warn("failed to record behavior outcome",
			zap.String("user_id", userID),
			zap.String("action", action),
			zap.Error(err),
		)
	}
}
