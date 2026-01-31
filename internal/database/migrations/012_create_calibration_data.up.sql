-- Migration 012: Create calibration data table for feedback loop
CREATE TABLE IF NOT EXISTS calibration_data (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_name    VARCHAR(100) NOT NULL,
    decision_id      UUID NOT NULL REFERENCES moderation_decisions(id),
    category_scores  JSONB NOT NULL,
    automated_action VARCHAR(20) NOT NULL,
    review_outcome   VARCHAR(20) NOT NULL, -- 'agree', 'disagree', 'uncertain'
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_calibration_provider ON calibration_data (provider_name);
CREATE INDEX IF NOT EXISTS idx_calibration_created ON calibration_data (created_at);
CREATE INDEX IF NOT EXISTS idx_calibration_outcome ON calibration_data (review_outcome);

-- Materialized view for provider accuracy metrics
CREATE MATERIALIZED VIEW IF NOT EXISTS provider_accuracy AS
SELECT
    provider_name,
    COUNT(*) AS total_decisions,
    COUNT(*) FILTER (WHERE review_outcome = 'agree') AS agree_count,
    COUNT(*) FILTER (WHERE review_outcome = 'disagree') AS disagree_count,
    COUNT(*) FILTER (WHERE review_outcome = 'uncertain') AS uncertain_count,
    CASE
        WHEN COUNT(*) > 0
        THEN ROUND(COUNT(*) FILTER (WHERE review_outcome = 'agree')::numeric / COUNT(*)::numeric, 4)
        ELSE 0
    END AS accuracy_rate
FROM calibration_data
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY provider_name;

CREATE UNIQUE INDEX IF NOT EXISTS idx_provider_accuracy_name ON provider_accuracy (provider_name);
