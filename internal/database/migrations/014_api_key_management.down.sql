-- Rollback migration 014: Remove API key management columns

DROP INDEX IF EXISTS idx_users_api_key_prefix;

ALTER TABLE users DROP COLUMN IF EXISTS rate_limit_rpm;
ALTER TABLE users DROP COLUMN IF EXISTS api_key_last_used_at;
ALTER TABLE users DROP COLUMN IF EXISTS api_key_prefix;
ALTER TABLE users DROP COLUMN IF EXISTS api_key_name;
