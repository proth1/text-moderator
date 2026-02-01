package retention

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Purger handles data retention policy enforcement and GDPR erasure.
// Control: SEC-003 (Data Retention Controls)
type Purger struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewPurger creates a new data retention purger.
func NewPurger(pool *pgxpool.Pool, logger *zap.Logger) *Purger {
	return &Purger{pool: pool, logger: logger}
}

// PurgeExpired deletes submissions and decisions past their retention date.
// Returns the count of deleted submissions and decisions.
func (p *Purger) PurgeExpired(ctx context.Context) (int64, int64, error) {
	now := time.Now()

	// Delete expired decisions first (foreign key dependency)
	decisionResult, err := p.pool.Exec(ctx,
		`DELETE FROM moderation_decisions WHERE retention_expires_at IS NOT NULL AND retention_expires_at < $1`,
		now,
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to purge expired decisions: %w", err)
	}
	decisionsDeleted := decisionResult.RowsAffected()

	// Delete expired submissions
	submissionResult, err := p.pool.Exec(ctx,
		`DELETE FROM text_submissions WHERE retention_expires_at IS NOT NULL AND retention_expires_at < $1`,
		now,
	)
	if err != nil {
		return 0, decisionsDeleted, fmt.Errorf("failed to purge expired submissions: %w", err)
	}
	submissionsDeleted := submissionResult.RowsAffected()

	p.logger.Info("purged expired data",
		zap.Int64("submissions_deleted", submissionsDeleted),
		zap.Int64("decisions_deleted", decisionsDeleted),
	)

	return submissionsDeleted, decisionsDeleted, nil
}

// PurgeByContentHash performs GDPR erasure for a specific content hash.
// Instead of deleting records (which would break the audit trail), this anonymizes
// the data while preserving the evidence that processing occurred and was erased.
func (p *Purger) PurgeByContentHash(ctx context.Context, contentHash string) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Anonymize the submission (replace hash with erasure marker)
	erasureMarker := fmt.Sprintf("ERASED:%s", uuid.New().String())
	result, err := tx.Exec(ctx,
		`UPDATE text_submissions SET content_hash = $1, context_metadata = NULL, source = NULL WHERE content_hash = $2`,
		erasureMarker, contentHash,
	)
	if err != nil {
		return fmt.Errorf("failed to anonymize submissions: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no submissions found with hash %s", contentHash)
	}

	// Write evidence record documenting the erasure (control SEC-003)
	_, err = tx.Exec(ctx,
		`INSERT INTO evidence_records (id, control_id, immutable, created_at)
		 VALUES ($1, 'SEC-003', true, NOW())`,
		uuid.New(),
	)
	if err != nil {
		return fmt.Errorf("failed to write erasure evidence: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit erasure: %w", err)
	}

	p.logger.Info("GDPR erasure completed",
		zap.String("content_hash", contentHash),
		zap.Int64("records_anonymized", result.RowsAffected()),
	)

	return nil
}
