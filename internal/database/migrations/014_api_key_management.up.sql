-- Migration 014: API key management enhancements
-- Control: SEC-002 (API Key and OAuth Authentication)

ALTER TABLE users ADD COLUMN IF NOT EXISTS api_key_name VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS api_key_prefix VARCHAR(8);
ALTER TABLE users ADD COLUMN IF NOT EXISTS api_key_last_used_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS rate_limit_rpm INTEGER DEFAULT 60;

CREATE INDEX IF NOT EXISTS idx_users_api_key_prefix ON users(api_key_prefix) WHERE api_key_prefix IS NOT NULL;
