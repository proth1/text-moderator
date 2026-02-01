ALTER TABLE moderation_decisions
    DROP COLUMN IF EXISTS assigned_reviewer,
    DROP COLUMN IF EXISTS assigned_at,
    DROP COLUMN IF EXISTS sla_deadline;
