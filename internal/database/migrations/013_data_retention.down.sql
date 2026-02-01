-- Rollback migration 013: Remove data retention columns

DROP INDEX IF EXISTS idx_decisions_retention;
DROP INDEX IF EXISTS idx_submissions_retention;

ALTER TABLE moderation_decisions DROP COLUMN IF EXISTS retention_expires_at;
ALTER TABLE text_submissions DROP COLUMN IF EXISTS retention_expires_at;
