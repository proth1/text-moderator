-- Migration 015: Table partitioning for evidence_records and moderation_decisions
-- Partitions by month for efficient range queries and retention management.
--
-- NOTE: PostgreSQL does not support ALTER TABLE ... PARTITION BY on existing tables.
-- This migration creates partitioned replacements and migrates data.

-- Step 1: Rename existing tables
ALTER TABLE evidence_records RENAME TO evidence_records_old;
ALTER TABLE moderation_decisions RENAME TO moderation_decisions_old;

-- Step 2: Create partitioned evidence_records table
CREATE TABLE evidence_records (
    id UUID NOT NULL,
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Step 3: Create partitioned moderation_decisions table
CREATE TABLE moderation_decisions (
    id UUID NOT NULL,
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
    retention_expires_at TIMESTAMPTZ,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Step 4: Create monthly partitions (2025-01 through 2026-06)
CREATE TABLE evidence_records_2025_01 PARTITION OF evidence_records FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE evidence_records_2025_02 PARTITION OF evidence_records FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE evidence_records_2025_03 PARTITION OF evidence_records FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE evidence_records_2025_04 PARTITION OF evidence_records FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE evidence_records_2025_05 PARTITION OF evidence_records FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE evidence_records_2025_06 PARTITION OF evidence_records FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE evidence_records_2025_07 PARTITION OF evidence_records FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE evidence_records_2025_08 PARTITION OF evidence_records FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE evidence_records_2025_09 PARTITION OF evidence_records FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE evidence_records_2025_10 PARTITION OF evidence_records FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE evidence_records_2025_11 PARTITION OF evidence_records FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE evidence_records_2025_12 PARTITION OF evidence_records FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE evidence_records_2026_01 PARTITION OF evidence_records FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE evidence_records_2026_02 PARTITION OF evidence_records FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE evidence_records_2026_03 PARTITION OF evidence_records FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE evidence_records_2026_04 PARTITION OF evidence_records FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE evidence_records_2026_05 PARTITION OF evidence_records FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE evidence_records_2026_06 PARTITION OF evidence_records FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE TABLE moderation_decisions_2025_01 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE moderation_decisions_2025_02 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE moderation_decisions_2025_03 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE moderation_decisions_2025_04 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE moderation_decisions_2025_05 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE moderation_decisions_2025_06 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE moderation_decisions_2025_07 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE moderation_decisions_2025_08 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE moderation_decisions_2025_09 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE moderation_decisions_2025_10 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE moderation_decisions_2025_11 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE moderation_decisions_2025_12 PARTITION OF moderation_decisions FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE moderation_decisions_2026_01 PARTITION OF moderation_decisions FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE moderation_decisions_2026_02 PARTITION OF moderation_decisions FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE moderation_decisions_2026_03 PARTITION OF moderation_decisions FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE moderation_decisions_2026_04 PARTITION OF moderation_decisions FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE moderation_decisions_2026_05 PARTITION OF moderation_decisions FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE moderation_decisions_2026_06 PARTITION OF moderation_decisions FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

-- Step 5: Migrate existing data
INSERT INTO evidence_records SELECT * FROM evidence_records_old;
INSERT INTO moderation_decisions SELECT * FROM moderation_decisions_old;

-- Step 6: Recreate indexes on partitioned tables
CREATE INDEX idx_evidence_control_id ON evidence_records(control_id);
CREATE INDEX idx_evidence_decision_id ON evidence_records(decision_id);
CREATE INDEX idx_decisions_submission_id ON moderation_decisions(submission_id);
CREATE INDEX idx_decisions_retention ON moderation_decisions(retention_expires_at) WHERE retention_expires_at IS NOT NULL;

-- Step 7: Drop old tables
DROP TABLE evidence_records_old;
DROP TABLE moderation_decisions_old;
