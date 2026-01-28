-- Control: GOV-002 (Human review workflow and oversight)

CREATE TYPE review_action_type AS ENUM ('approve', 'reject', 'edit', 'escalate');

CREATE TABLE review_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id UUID NOT NULL REFERENCES moderation_decisions(id),
    reviewer_id UUID NOT NULL REFERENCES users(id),
    action review_action_type NOT NULL,
    rationale TEXT,
    edited_content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_decision ON review_actions(decision_id);
CREATE INDEX idx_reviews_reviewer ON review_actions(reviewer_id);
CREATE INDEX idx_reviews_created ON review_actions(created_at);
CREATE INDEX idx_reviews_action ON review_actions(action);

COMMENT ON TABLE review_actions IS 'Human review actions for moderation decisions';
COMMENT ON COLUMN review_actions.action IS 'Type of review action taken';
COMMENT ON COLUMN review_actions.rationale IS 'Explanation for the review decision';
COMMENT ON COLUMN review_actions.edited_content IS 'Modified content (if action is edit)';
