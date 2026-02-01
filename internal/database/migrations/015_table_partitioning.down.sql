-- Rollback migration 015: Revert table partitioning
-- WARNING: This converts partitioned tables back to regular tables.

-- Step 1: Create non-partitioned replacements
CREATE TABLE evidence_records_flat (
    id UUID PRIMARY KEY,
    control_id VARCHAR(50) NOT NULL,
    policy_id UUID,
    policy_version INTEGER,
    decision_id UUID,
    review_id UUID,
    model_name VARCHAR(255),
    model_version VARCHAR(50),
    category_scores JSONB,
    automated_action VARCHAR(50),
    human_override VARCHAR(50),
    submission_hash VARCHAR(255),
    immutable BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE moderation_decisions_flat (
    id UUID PRIMARY KEY,
    submission_id UUID NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    model_version VARCHAR(50) NOT NULL,
    category_scores JSONB NOT NULL,
    policy_id UUID,
    policy_version INTEGER,
    automated_action VARCHAR(50) NOT NULL,
    confidence DOUBLE PRECISION,
    explanation TEXT,
    correlation_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    retention_expires_at TIMESTAMPTZ
);

-- Step 2: Migrate data
INSERT INTO evidence_records_flat SELECT * FROM evidence_records;
INSERT INTO moderation_decisions_flat SELECT * FROM moderation_decisions;

-- Step 3: Drop partitioned tables and rename
DROP TABLE evidence_records CASCADE;
DROP TABLE moderation_decisions CASCADE;

ALTER TABLE evidence_records_flat RENAME TO evidence_records;
ALTER TABLE moderation_decisions_flat RENAME TO moderation_decisions;
