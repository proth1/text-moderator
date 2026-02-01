ALTER TABLE evidence_records
    DROP COLUMN IF EXISTS chain_hash,
    DROP COLUMN IF EXISTS previous_hash;
