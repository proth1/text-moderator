-- Control: MOD-001 (Content submission tracking and hashing for deduplication)

CREATE TABLE text_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_hash VARCHAR(64) NOT NULL,
    content_encrypted TEXT,
    context_metadata JSONB DEFAULT '{}',
    source VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submissions_hash ON text_submissions(content_hash);
CREATE INDEX idx_submissions_created ON text_submissions(created_at);
CREATE INDEX idx_submissions_source ON text_submissions(source);

COMMENT ON TABLE text_submissions IS 'Text submissions for moderation with content hashing';
COMMENT ON COLUMN text_submissions.content_hash IS 'SHA-256 hash of content for deduplication';
COMMENT ON COLUMN text_submissions.content_encrypted IS 'Encrypted content (optional, for audit trail)';
COMMENT ON COLUMN text_submissions.context_metadata IS 'Additional context (e.g., user_id, channel, timestamp)';
COMMENT ON COLUMN text_submissions.source IS 'Source system or application identifier';
