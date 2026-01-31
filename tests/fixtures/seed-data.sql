-- Test Users
-- Control: GOV-002 (User management and access control)
-- API keys stored as SHA-256 hashes (SECURITY: never store plaintext keys)
-- tk_admin_test_key_001 -> sha256 hash
-- tk_mod_test_key_002 -> sha256 hash
-- tk_viewer_test_key_003 -> sha256 hash
INSERT INTO users (id, email, api_key_hash, role, created_at, updated_at) VALUES
('a0000000-0000-0000-0000-000000000001', 'admin@civitas.test', encode(sha256('tk_admin_test_key_001'::bytea), 'hex'), 'admin', NOW(), NOW()),
('a0000000-0000-0000-0000-000000000002', 'moderator@civitas.test', encode(sha256('tk_mod_test_key_002'::bytea), 'hex'), 'moderator', NOW(), NOW()),
('a0000000-0000-0000-0000-000000000003', 'viewer@civitas.test', encode(sha256('tk_viewer_test_key_003'::bytea), 'hex'), 'viewer', NOW(), NOW());

-- Test Policies
-- Control: POL-001 (Policy definition and versioning)
INSERT INTO policies (id, name, version, thresholds, actions, scope, status, effective_date, created_at, created_by) VALUES
('b0000000-0000-0000-0000-000000000001', 'Standard Community Guidelines', 1,
 '{"toxicity": 0.8, "hate": 0.7, "harassment": 0.75, "sexual_content": 0.8, "violence": 0.85, "profanity": 0.9}'::jsonb,
 '{"toxicity": "block", "hate": "block", "harassment": "warn", "sexual_content": "block", "violence": "block", "profanity": "warn"}'::jsonb,
 '{"region": "global", "content_type": "user_generated"}'::jsonb,
 'published', NOW(), NOW(), 'a0000000-0000-0000-0000-000000000001'),
('b0000000-0000-0000-0000-000000000002', 'Youth Safe Mode', 1,
 '{"toxicity": 0.5, "hate": 0.4, "harassment": 0.5, "sexual_content": 0.3, "violence": 0.4, "profanity": 0.6}'::jsonb,
 '{"toxicity": "block", "hate": "block", "harassment": "block", "sexual_content": "block", "violence": "block", "profanity": "warn"}'::jsonb,
 '{"region": "global", "content_type": "youth", "age_group": "under_13"}'::jsonb,
 'published', NOW(), NOW(), 'a0000000-0000-0000-0000-000000000001'),
('b0000000-0000-0000-0000-000000000003', 'Relaxed Forum Policy', 1,
 '{"toxicity": 0.95, "hate": 0.85, "harassment": 0.9, "sexual_content": 0.9, "violence": 0.95, "profanity": 0.99}'::jsonb,
 '{"toxicity": "warn", "hate": "block", "harassment": "warn", "sexual_content": "warn", "violence": "warn", "profanity": "allow"}'::jsonb,
 '{"region": "US", "content_type": "forum"}'::jsonb,
 'draft', NULL, NOW(), 'a0000000-0000-0000-0000-000000000001');

-- Test Submissions (mix of safe and toxic content)
-- Control: MOD-001 (Input tracking and hashing)
INSERT INTO text_submissions (id, content_hash, content_encrypted, context_metadata, source, created_at) VALUES
('c0000000-0000-0000-0000-000000000001', 'hash_safe_001', 'Hello, this is a friendly message!', '{"channel": "chat"}'::jsonb, 'web', NOW()),
('c0000000-0000-0000-0000-000000000002', 'hash_toxic_001', 'This content is mildly problematic', '{"channel": "comments"}'::jsonb, 'web', NOW()),
('c0000000-0000-0000-0000-000000000003', 'hash_hate_001', 'Extreme hateful content example', '{"channel": "chat"}'::jsonb, 'mobile', NOW()),
('c0000000-0000-0000-0000-000000000004', 'hash_safe_002', 'Great product, highly recommend!', '{"channel": "reviews"}'::jsonb, 'api', NOW()),
('c0000000-0000-0000-0000-000000000005', 'hash_harass_001', 'Targeted harassment example', '{"channel": "messaging"}'::jsonb, 'web', NOW());

-- Test Decisions
-- Control: MOD-001 (Decision tracking and traceability)
INSERT INTO moderation_decisions (id, submission_id, model_name, model_version, category_scores, policy_id, policy_version, automated_action, confidence, correlation_id, created_at) VALUES
('d0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'hf-friendly-text-moderator', '2025-11',
 '{"toxicity": 0.05, "hate": 0.02, "harassment": 0.03, "sexual_content": 0.01, "violence": 0.01, "profanity": 0.02}'::jsonb,
 'b0000000-0000-0000-0000-000000000001', 1, 'allow', 0.95, 'e0000000-0000-0000-0000-000000000001', NOW()),
('d0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000002', 'hf-friendly-text-moderator', '2025-11',
 '{"toxicity": 0.65, "hate": 0.3, "harassment": 0.45, "sexual_content": 0.05, "violence": 0.1, "profanity": 0.55}'::jsonb,
 'b0000000-0000-0000-0000-000000000001', 1, 'warn', 0.78, 'e0000000-0000-0000-0000-000000000002', NOW()),
('d0000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000003', 'hf-friendly-text-moderator', '2025-11',
 '{"toxicity": 0.92, "hate": 0.95, "harassment": 0.88, "sexual_content": 0.1, "violence": 0.45, "profanity": 0.7}'::jsonb,
 'b0000000-0000-0000-0000-000000000001', 1, 'block', 0.95, 'e0000000-0000-0000-0000-000000000003', NOW()),
('d0000000-0000-0000-0000-000000000004', 'c0000000-0000-0000-0000-000000000005', 'hf-friendly-text-moderator', '2025-11',
 '{"toxicity": 0.78, "hate": 0.35, "harassment": 0.82, "sexual_content": 0.02, "violence": 0.15, "profanity": 0.4}'::jsonb,
 'b0000000-0000-0000-0000-000000000001', 1, 'escalate', 0.82, 'e0000000-0000-0000-0000-000000000004', NOW());

-- Test Review Actions
-- Control: GOV-002 (Human review workflow)
INSERT INTO review_actions (id, decision_id, reviewer_id, action, rationale, created_at) VALUES
('f0000000-0000-0000-0000-000000000001', 'd0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000002', 'approve', 'Confirmed hate speech, block is correct', NOW()),
('f0000000-0000-0000-0000-000000000002', 'd0000000-0000-0000-0000-000000000004', 'a0000000-0000-0000-0000-000000000002', 'reject', 'Context shows this is quoted text being discussed, not directed harassment', NOW());

-- Test Evidence Records
-- Control: AUD-001 (Immutable evidence generation)
INSERT INTO evidence_records (id, control_id, policy_id, policy_version, decision_id, model_name, model_version, category_scores, automated_action, submission_hash, immutable, created_at) VALUES
('e1000000-0000-0000-0000-000000000001', 'MOD-001', 'b0000000-0000-0000-0000-000000000001', 1, 'd0000000-0000-0000-0000-000000000001', 'hf-friendly-text-moderator', '2025-11',
 '{"toxicity": 0.05, "hate": 0.02}'::jsonb, 'allow', 'hash_safe_001', true, NOW()),
('e1000000-0000-0000-0000-000000000002', 'MOD-001', 'b0000000-0000-0000-0000-000000000001', 1, 'd0000000-0000-0000-0000-000000000003', 'hf-friendly-text-moderator', '2025-11',
 '{"toxicity": 0.92, "hate": 0.95}'::jsonb, 'block', 'hash_hate_001', true, NOW()),
('e1000000-0000-0000-0000-000000000003', 'GOV-002', 'b0000000-0000-0000-0000-000000000001', 1, 'd0000000-0000-0000-0000-000000000003', 'hf-friendly-text-moderator', '2025-11',
 '{"toxicity": 0.92, "hate": 0.95}'::jsonb, 'block', 'hash_hate_001', true, NOW());
