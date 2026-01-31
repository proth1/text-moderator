-- Rollback: restore plaintext api_key column
ALTER TABLE users ADD COLUMN api_key VARCHAR(255);
DROP INDEX IF EXISTS idx_users_api_key_hash;
ALTER TABLE users DROP COLUMN api_key_hash;
