-- Rollback Migration 012
DROP MATERIALIZED VIEW IF EXISTS provider_accuracy;
DROP TABLE IF EXISTS calibration_data;
