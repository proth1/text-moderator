-- Rollback Migration 010: Remove language detection column
ALTER TABLE moderation_decisions
    DROP COLUMN IF EXISTS detected_language;
