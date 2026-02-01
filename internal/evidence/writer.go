package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/proth1/text-moderator/internal/models"
	"go.uber.org/zap"
)

// Control: AUD-001 (Immutable evidence generation for compliance)

// Writer handles writing evidence records to the database
type Writer struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewWriter creates a new evidence writer
func NewWriter(db *pgxpool.Pool, logger *zap.Logger) *Writer {
	return &Writer{
		db:     db,
		logger: logger,
	}
}

// WriteEvidenceInTx writes an evidence record within an existing transaction.
// This ensures decision + evidence are atomically committed together.
func (w *Writer) WriteEvidenceInTx(ctx context.Context, tx pgx.Tx, evidence *models.EvidenceRecord) error {
	// Compute hash chain (reads previous hash from DB outside tx for simplicity)
	w.computeChainHash(ctx, evidence)

	query := `
		INSERT INTO evidence_records (
			id, control_id, policy_id, policy_version, decision_id, review_id,
			model_name, model_version, category_scores, automated_action,
			human_override, submission_hash, immutable, chain_hash, previous_hash
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := tx.Exec(ctx, query,
		evidence.ID,
		evidence.ControlID,
		evidence.PolicyID,
		evidence.PolicyVersion,
		evidence.DecisionID,
		evidence.ReviewID,
		evidence.ModelName,
		evidence.ModelVersion,
		evidence.CategoryScores,
		evidence.AutomatedAction,
		evidence.HumanOverride,
		evidence.SubmissionHash,
		evidence.Immutable,
		evidence.ChainHash,
		evidence.PreviousHash,
	)

	if err != nil {
		w.logger.Error("failed to write evidence record in transaction",
			zap.String("control_id", evidence.ControlID),
			zap.String("evidence_id", evidence.ID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to write evidence record: %w", err)
	}

	w.logger.Info("evidence record created in transaction",
		zap.String("control_id", evidence.ControlID),
		zap.String("evidence_id", evidence.ID.String()),
	)

	return nil
}

// BeginTx starts a new database transaction
func (w *Writer) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return w.db.Begin(ctx)
}

// RecordModerationDecision creates an evidence record for a moderation decision
func (w *Writer) RecordModerationDecision(ctx context.Context, decision *models.ModerationDecision) error {
	evidence := &models.EvidenceRecord{
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

	return w.writeEvidence(ctx, evidence)
}

// RecordReviewAction creates an evidence record for a human review action
func (w *Writer) RecordReviewAction(ctx context.Context, review *models.ReviewAction, decision *models.ModerationDecision) error {
	evidence := &models.EvidenceRecord{
		ID:             uuid.New(),
		ControlID:      "GOV-002",
		PolicyID:       decision.PolicyID,
		PolicyVersion:  decision.PolicyVersion,
		DecisionID:     &decision.ID,
		ReviewID:       &review.ID,
		ModelName:      &decision.ModelName,
		ModelVersion:   &decision.ModelVersion,
		CategoryScores: &decision.CategoryScores,
		HumanOverride:  &review.Action,
		Immutable:      true,
	}

	return w.writeEvidence(ctx, evidence)
}

// RecordPolicyApplication creates an evidence record for policy application
func (w *Writer) RecordPolicyApplication(ctx context.Context, policy *models.Policy, decisionID uuid.UUID) error {
	evidence := &models.EvidenceRecord{
		ID:            uuid.New(),
		ControlID:     "POL-001",
		PolicyID:      &policy.ID,
		PolicyVersion: &policy.Version,
		DecisionID:    &decisionID,
		Immutable:     true,
	}

	return w.writeEvidence(ctx, evidence)
}

// computeChainHash builds the hash chain by fetching the latest chain_hash
// and computing SHA-256(previous_hash + record fields).
func (w *Writer) computeChainHash(ctx context.Context, evidence *models.EvidenceRecord) {
	// Get the most recent chain_hash
	var prevHash *string
	err := w.db.QueryRow(ctx,
		`SELECT chain_hash FROM evidence_records WHERE chain_hash IS NOT NULL ORDER BY created_at DESC LIMIT 1`,
	).Scan(&prevHash)
	if err != nil {
		// No previous record or query error — start a new chain
		prevHash = nil
	}
	evidence.PreviousHash = prevHash

	// Build canonical data string for hashing
	prev := ""
	if prevHash != nil {
		prev = *prevHash
	}
	data := fmt.Sprintf("%s|%s|%s", prev, evidence.ID.String(), evidence.ControlID)
	if evidence.DecisionID != nil {
		data += "|" + evidence.DecisionID.String()
	}
	if evidence.CategoryScores != nil {
		if scoreBytes, err := json.Marshal(evidence.CategoryScores); err == nil {
			data += "|" + string(scoreBytes)
		}
	}
	if evidence.SubmissionHash != nil {
		data += "|" + *evidence.SubmissionHash
	}

	h := sha256.Sum256([]byte(data))
	chainHash := hex.EncodeToString(h[:])
	evidence.ChainHash = &chainHash
}

// writeEvidence writes an evidence record to the database with hash chain.
func (w *Writer) writeEvidence(ctx context.Context, evidence *models.EvidenceRecord) error {
	// Compute tamper-proof hash chain
	w.computeChainHash(ctx, evidence)

	query := `
		INSERT INTO evidence_records (
			id, control_id, policy_id, policy_version, decision_id, review_id,
			model_name, model_version, category_scores, automated_action,
			human_override, submission_hash, immutable, chain_hash, previous_hash
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := w.db.Exec(ctx, query,
		evidence.ID,
		evidence.ControlID,
		evidence.PolicyID,
		evidence.PolicyVersion,
		evidence.DecisionID,
		evidence.ReviewID,
		evidence.ModelName,
		evidence.ModelVersion,
		evidence.CategoryScores,
		evidence.AutomatedAction,
		evidence.HumanOverride,
		evidence.SubmissionHash,
		evidence.Immutable,
		evidence.ChainHash,
		evidence.PreviousHash,
	)

	if err != nil {
		w.logger.Error("failed to write evidence record",
			zap.String("control_id", evidence.ControlID),
			zap.String("evidence_id", evidence.ID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to write evidence record: %w", err)
	}

	w.logger.Info("evidence record created",
		zap.String("control_id", evidence.ControlID),
		zap.String("evidence_id", evidence.ID.String()),
	)

	return nil
}

// ListEvidence retrieves evidence records with optional filtering
func (w *Writer) ListEvidence(ctx context.Context, controlID *string, limit int, offset int) ([]models.EvidenceRecord, error) {
	query := `
		SELECT id, control_id, policy_id, policy_version, decision_id, review_id,
		       model_name, model_version, category_scores, automated_action,
		       human_override, submission_hash, immutable, chain_hash, previous_hash, created_at
		FROM evidence_records
		WHERE ($1::text IS NULL OR control_id = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := w.db.Query(ctx, query, controlID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query evidence records: %w", err)
	}
	defer rows.Close()

	var records []models.EvidenceRecord
	for rows.Next() {
		var record models.EvidenceRecord
		err := rows.Scan(
			&record.ID,
			&record.ControlID,
			&record.PolicyID,
			&record.PolicyVersion,
			&record.DecisionID,
			&record.ReviewID,
			&record.ModelName,
			&record.ModelVersion,
			&record.CategoryScores,
			&record.AutomatedAction,
			&record.HumanOverride,
			&record.SubmissionHash,
			&record.Immutable,
			&record.ChainHash,
			&record.PreviousHash,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan evidence record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating evidence records: %w", err)
	}

	return records, nil
}
