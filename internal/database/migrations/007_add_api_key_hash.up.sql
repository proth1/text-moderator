-- Migration: Add api_key_hash column for secure API key storage
-- API keys are now stored as SHA-256 hashes, never in plaintext

ALTER TABLE users ADD COLUMN api_key_hash VARCHAR(64);

-- Populate hash column from existing plaintext keys
UPDATE users SET api_key_hash = encode(sha256(api_key::bytea), 'hex') WHERE api_key IS NOT NULL;

-- Create index on the hash column for lookups
CREATE UNIQUE INDEX idx_users_api_key_hash ON users(api_key_hash) WHERE api_key_hash IS NOT NULL;

-- Drop old plaintext column and index
DROP INDEX IF EXISTS users_api_key_key;
ALTER TABLE users DROP COLUMN api_key;
