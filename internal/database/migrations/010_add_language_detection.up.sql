-- Migration 010: Add language detection column to moderation decisions
ALTER TABLE moderation_decisions
    ADD COLUMN IF NOT EXISTS detected_language VARCHAR(10);
