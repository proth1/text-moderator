-- Migration 009: Expand CategoryScores with self_harm, spam, pii fields
-- Backfill existing JSONB rows in moderation_decisions and evidence tables

-- Update moderation_decisions: add new fields with default 0 to category_scores JSONB
UPDATE moderation_decisions
SET category_scores = category_scores
    || '{"self_harm": 0, "spam": 0, "pii": 0}'::jsonb
WHERE NOT (category_scores ? 'self_harm');

-- Temporarily disable the evidence immutability trigger for backfill
ALTER TABLE evidence DISABLE TRIGGER evidence_no_update;

UPDATE evidence
SET category_scores = category_scores
    || '{"self_harm": 0, "spam": 0, "pii": 0}'::jsonb
WHERE category_scores IS NOT NULL
  AND NOT (category_scores ? 'self_harm');

-- Re-enable the evidence immutability trigger
ALTER TABLE evidence ENABLE TRIGGER evidence_no_update;
