-- Migration 013: Add data retention columns for GDPR compliance
-- Control: SEC-003 (Data Retention Controls)

ALTER TABLE text_submissions ADD COLUMN IF NOT EXISTS retention_expires_at TIMESTAMPTZ;
ALTER TABLE moderation_decisions ADD COLUMN IF NOT EXISTS retention_expires_at TIMESTAMPTZ;

-- Set default retention periods for existing data
UPDATE text_submissions SET retention_expires_at = created_at + INTERVAL '90 days' WHERE retention_expires_at IS NULL;
UPDATE moderation_decisions SET retention_expires_at = created_at + INTERVAL '365 days' WHERE retention_expires_at IS NULL;

-- Indexes for efficient purge queries
CREATE INDEX IF NOT EXISTS idx_submissions_retention ON text_submissions(retention_expires_at) WHERE retention_expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_decisions_retention ON moderation_decisions(retention_expires_at) WHERE retention_expires_at IS NOT NULL;
