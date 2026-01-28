-- Control: AUD-001 (Immutable evidence generation and audit trail)

CREATE TABLE evidence_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    control_id VARCHAR(20) NOT NULL,
    policy_id UUID REFERENCES policies(id),
    policy_version INTEGER,
    decision_id UUID REFERENCES moderation_decisions(id),
    review_id UUID REFERENCES review_actions(id),
    model_name VARCHAR(255),
    model_version VARCHAR(50),
    category_scores JSONB,
    automated_action moderation_action,
    human_override review_action_type,
    submission_hash VARCHAR(64),
    immutable BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evidence_control ON evidence_records(control_id);
CREATE INDEX idx_evidence_policy ON evidence_records(policy_id);
CREATE INDEX idx_evidence_decision ON evidence_records(decision_id);
CREATE INDEX idx_evidence_review ON evidence_records(review_id);
CREATE INDEX idx_evidence_created ON evidence_records(created_at);
CREATE INDEX idx_evidence_hash ON evidence_records(submission_hash);

COMMENT ON TABLE evidence_records IS 'Immutable audit evidence for compliance (SOC 2, ISO 27001, GDPR)';
COMMENT ON COLUMN evidence_records.control_id IS 'Control framework reference (e.g., MOD-001, POL-001, GOV-002, AUD-001)';
COMMENT ON COLUMN evidence_records.immutable IS 'Enforcement flag for immutability (always true)';

-- Prevent updates/deletes on evidence (append-only)
CREATE OR REPLACE FUNCTION prevent_evidence_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Evidence records are immutable and cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER evidence_no_update
    BEFORE UPDATE ON evidence_records
    FOR EACH ROW EXECUTE FUNCTION prevent_evidence_modification();

CREATE TRIGGER evidence_no_delete
    BEFORE DELETE ON evidence_records
    FOR EACH ROW EXECUTE FUNCTION prevent_evidence_modification();
