-- Control: GOV-002 (Review queue assignment and SLA tracking)

ALTER TABLE moderation_decisions
    ADD COLUMN assigned_reviewer UUID REFERENCES api_users(id),
    ADD COLUMN assigned_at TIMESTAMPTZ,
    ADD COLUMN sla_deadline TIMESTAMPTZ;

CREATE INDEX idx_decisions_assigned ON moderation_decisions(assigned_reviewer) WHERE automated_action = 'escalate';
CREATE INDEX idx_decisions_sla ON moderation_decisions(sla_deadline) WHERE automated_action = 'escalate' AND sla_deadline IS NOT NULL;

COMMENT ON COLUMN moderation_decisions.assigned_reviewer IS 'Reviewer assigned to this escalated decision';
COMMENT ON COLUMN moderation_decisions.assigned_at IS 'When the reviewer was assigned';
COMMENT ON COLUMN moderation_decisions.sla_deadline IS 'SLA deadline for review completion';
