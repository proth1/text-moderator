-- Control: MOD-001 (Moderation decision tracking and traceability)

CREATE TABLE moderation_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES text_submissions(id),
    model_name VARCHAR(255) NOT NULL,
    model_version VARCHAR(50) NOT NULL,
    category_scores JSONB NOT NULL,
    policy_id UUID REFERENCES policies(id),
    policy_version INTEGER,
    automated_action moderation_action NOT NULL,
    confidence FLOAT,
    explanation TEXT,
    correlation_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_decisions_submission ON moderation_decisions(submission_id);
CREATE INDEX idx_decisions_policy ON moderation_decisions(policy_id);
CREATE INDEX idx_decisions_action ON moderation_decisions(automated_action);
CREATE INDEX idx_decisions_created ON moderation_decisions(created_at);
CREATE INDEX idx_decisions_correlation ON moderation_decisions(correlation_id);
CREATE INDEX idx_decisions_model ON moderation_decisions(model_name, model_version);

COMMENT ON TABLE moderation_decisions IS 'AI model moderation decisions with full traceability';
COMMENT ON COLUMN moderation_decisions.category_scores IS 'JSON object with category confidence scores';
COMMENT ON COLUMN moderation_decisions.automated_action IS 'Action determined by policy evaluation';
COMMENT ON COLUMN moderation_decisions.confidence IS 'Overall confidence score (0.0-1.0)';
COMMENT ON COLUMN moderation_decisions.explanation IS 'Human-readable explanation of decision';
COMMENT ON COLUMN moderation_decisions.correlation_id IS 'Optional ID for request tracing';
