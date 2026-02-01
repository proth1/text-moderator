-- Control: AUD-001 (Tamper-proof evidence hash chain)

ALTER TABLE evidence_records
    ADD COLUMN chain_hash VARCHAR(64),
    ADD COLUMN previous_hash VARCHAR(64);

CREATE INDEX idx_evidence_chain_hash ON evidence_records(chain_hash);

COMMENT ON COLUMN evidence_records.chain_hash IS 'SHA-256 hash of (previous_hash + record fields) forming a tamper-proof chain';
COMMENT ON COLUMN evidence_records.previous_hash IS 'chain_hash of the preceding evidence record';
