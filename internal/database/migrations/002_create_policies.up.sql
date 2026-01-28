-- Control: POL-001 (Policy definition, versioning, and lifecycle management)

CREATE TYPE policy_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE moderation_action AS ENUM ('allow', 'warn', 'block', 'escalate');

CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    thresholds JSONB NOT NULL,
    actions JSONB NOT NULL,
    scope JSONB DEFAULT '{}',
    status policy_status NOT NULL DEFAULT 'draft',
    effective_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    UNIQUE(name, version)
);

CREATE INDEX idx_policies_status ON policies(status);
CREATE INDEX idx_policies_name_version ON policies(name, version);
CREATE INDEX idx_policies_effective_date ON policies(effective_date);
CREATE INDEX idx_policies_created_by ON policies(created_by);

COMMENT ON TABLE policies IS 'Moderation policies with versioning and lifecycle management';
COMMENT ON COLUMN policies.thresholds IS 'JSON map of category thresholds (e.g., {"toxicity": 0.7, "hate": 0.8})';
COMMENT ON COLUMN policies.actions IS 'JSON map of category actions (e.g., {"toxicity": "warn", "hate": "block"})';
COMMENT ON COLUMN policies.scope IS 'JSON object defining policy scope (e.g., tenant, region, content type)';
