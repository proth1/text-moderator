-- Migration 011: Create user behavior stats table for trust scoring
CREATE TABLE IF NOT EXISTS user_behavior_stats (
    user_id         VARCHAR(255) NOT NULL,
    window_start    DATE NOT NULL,
    total_decisions INTEGER NOT NULL DEFAULT 0,
    allowed_count   INTEGER NOT NULL DEFAULT 0,
    blocked_count   INTEGER NOT NULL DEFAULT 0,
    escalated_count INTEGER NOT NULL DEFAULT 0,
    warned_count    INTEGER NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, window_start)
);

CREATE INDEX IF NOT EXISTS idx_user_behavior_user_id ON user_behavior_stats (user_id);
CREATE INDEX IF NOT EXISTS idx_user_behavior_window ON user_behavior_stats (window_start);
