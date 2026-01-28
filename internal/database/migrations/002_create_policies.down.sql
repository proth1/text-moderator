DROP INDEX IF EXISTS idx_policies_created_by;
DROP INDEX IF EXISTS idx_policies_effective_date;
DROP INDEX IF EXISTS idx_policies_name_version;
DROP INDEX IF EXISTS idx_policies_status;
DROP TABLE IF EXISTS policies;
DROP TYPE IF EXISTS moderation_action;
DROP TYPE IF EXISTS policy_status;
