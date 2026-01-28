# Civitas AI Text Moderator -- Implementation Plan

## Document Metadata

| Field | Value |
|-------|-------|
| Version | 1.0 |
| Date | 2026-01-27 |
| Status | Draft |
| PRD Reference | `docs/requirements/civitas_ai_product_requirements_document.md` |

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Technology Stack](#2-technology-stack)
3. [Project Structure](#3-project-structure)
4. [Control ID Registry](#4-control-id-registry)
5. [Phase 1: Foundation](#5-phase-1-foundation)
6. [Phase 2: Core Moderation](#6-phase-2-core-moderation)
7. [Phase 3: Policy Engine](#7-phase-3-policy-engine)
8. [Phase 4: User Experience](#8-phase-4-user-experience)
9. [Phase 5: Moderator Experience](#9-phase-5-moderator-experience)
10. [Phase 6: Admin and Governance](#10-phase-6-admin-and-governance)
11. [Phase 7: Evidence and Compliance](#11-phase-7-evidence-and-compliance)
12. [Phase 8: Analytics and Monitoring](#12-phase-8-analytics-and-monitoring)
13. [Phase 9: Security Hardening](#13-phase-9-security-hardening)
14. [Phase 10: E2E Testing and Fixtures](#14-phase-10-e2e-testing-and-fixtures)
15. [PR Strategy and Merge Order](#15-pr-strategy-and-merge-order)
16. [Test Strategy](#16-test-strategy)
17. [Docker Compose Deployment Plan](#17-docker-compose-deployment-plan)

---

## 1. Architecture Overview

```
                    +------------------+
                    |   React Web App  |
                    | (Vite + TS + TW) |
                    +--------+---------+
                             |
                             v
                    +------------------+
                    |   API Gateway    |
                    |   (Go service)   |
                    +--------+---------+
                             |
              +--------------+--------------+
              |              |              |
              v              v              v
     +--------+---+  +------+------+  +----+-------+
     | Moderation |  | Policy      |  | Review     |
     | Service    |  | Engine      |  | Service    |
     +--------+---+  +------+------+  +----+-------+
              |              |              |
              v              v              v
     +--------+---+  +------+------+  +----+-------+
     | HuggingFace|  | PostgreSQL  |  | PostgreSQL |
     | API        |  |             |  |            |
     +------------+  +------+------+  +------------+
                             |
                      +------+------+
                      |   Redis     |
                      | (cache)     |
                      +-------------+
```

Services communicate via REST over the internal Docker network. The gateway is the single external entry point. PostgreSQL is the system of record. Redis provides caching for moderation results and session data.

---

## 2. Technology Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Backend | Go | 1.22+ |
| Frontend | React + TypeScript | React 18, TS 5.x |
| Build Tool | Vite | 5.x |
| CSS | TailwindCSS + shadcn/ui | TW 3.x |
| Database | PostgreSQL | 16 |
| Cache | Redis | 7 |
| Containerization | Docker Compose | 3.8 spec |
| API Schema | OpenAPI 3.0 | -- |
| JSON Schema | JSON Schema Draft 2020-12 | -- |
| BDD | Cucumber / Gherkin | godog for Go |
| E2E | Playwright | latest |
| Observability | OpenTelemetry + Prometheus | -- |

---

## 3. Project Structure

```
/text-moderator
  /apps
    /web                         # React frontend
      /src
        /components              # shadcn/ui based components
        /pages                   # Route pages
        /hooks                   # Custom React hooks
        /lib                     # API client, utilities
        /types                   # TypeScript type definitions
      package.json
      vite.config.ts
      tailwind.config.ts
      tsconfig.json
      Dockerfile
  /services
    /gateway                     # Go API gateway
      /cmd/gateway/main.go
      /internal
        /handler                 # HTTP handlers
        /middleware               # Auth, CORS, rate limiting
        /router                  # Route definitions
      go.mod
      Dockerfile
    /moderation                  # Go moderation service
      /cmd/moderation/main.go
      /internal
        /handler                 # HTTP handlers
        /service                 # Business logic
        /client                  # HuggingFace API client
        /model                   # Domain models
        /repository              # DB access
        /cache                   # Redis cache layer
      go.mod
      Dockerfile
    /policy-engine               # Go policy engine
      /cmd/policy-engine/main.go
      /internal
        /handler
        /service
        /model
        /repository
        /evaluator               # Policy evaluation logic
      go.mod
      Dockerfile
    /review                      # Go review service
      /cmd/review/main.go
      /internal
        /handler
        /service
        /model
        /repository
      go.mod
      Dockerfile
  /schemas
    policy.json                  # JSON Schema for policy
    decision.json                # JSON Schema for moderation decision
    evidence.json                # JSON Schema for evidence records
    submission.json              # JSON Schema for text submission
  /controls
    control-registry.yaml        # CDD control definitions
  /migrations
    /000001_initial_schema.up.sql
    /000001_initial_schema.down.sql
    ...
  /tests
    /unit                        # Go unit tests (co-located + here)
    /integration                 # Cross-service integration tests
    /bdd
      /features                  # Gherkin .feature files
      /steps                     # Step definitions (Go)
    /e2e                         # Playwright tests
      /specs
      /fixtures
      playwright.config.ts
    /fixtures                    # Shared seed data, mock responses
  /docs
    /requirements
    /architecture
    implementation-plan.md
  /scripts
    /seed                        # DB seed scripts
  docker-compose.yml
  docker-compose.dev.yml
  .env.example
  Makefile
```

---

## 4. Control ID Registry

These control IDs are defined in the PRD and will be referenced throughout implementation. Every PR must declare which control IDs it implements or supports.

| Control ID | Name | Type | PRD Section |
|-----------|------|------|-------------|
| MOD-001 | Automated Text Classification | Automated | 6.1 |
| MOD-002 | Real-Time User Feedback Enforcement | Automated | 6.5 |
| MOD-003 | Moderation Request Pipeline | Automated | Epic A2 |
| MOD-004 | Latency Optimization and Caching | Automated | Epic A3 |
| POL-001 | Threshold-Based Moderation Policy | Policy | 6.3 |
| POL-002 | Policy Versioning and Rollback | Policy | Epic B3 |
| POL-003 | Regional Policy Resolution | Policy | Epic B4 |
| GOV-001 | Role-Based Access Control | Governance | 10 |
| GOV-002 | Human-in-the-Loop Review | Compensating | 6.4 |
| GOV-003 | Separation of Duties | Governance | 8.2 |
| GOV-004 | Policy Management UI | Governance | Epic E1 |
| AUD-001 | Immutable Evidence Storage | Audit | 8.3 |
| AUD-002 | Evidence Export | Audit | Epic F3 |
| AUD-003 | Audit Log Viewer | Audit | Epic E3 |
| SEC-001 | TLS for Data in Transit | Security | 10 |
| SEC-002 | API Key and OAuth Authentication | Security | 10 |
| SEC-003 | Data Retention Controls | Security | Epic H2 |
| OBS-001 | Observability (Logs, Metrics, Traces) | Operational | 9 |
| ANL-001 | Moderation Metrics | Analytics | 12 |
| ANL-002 | Override Analytics | Analytics | Epic G2 |
| ANL-003 | Policy Effectiveness Dashboard | Analytics | Epic G3 |

---

## 5. Phase 1: Foundation

**Goal:** Establish project scaffold, database schema, Docker Compose, and base service skeletons with health checks. Produces a running but empty system.

**Depends on:** Nothing (greenfield start)

**Epics touched:** H (Platform)

### PR 1.1: Project Scaffold and Go Module Init

**Description:** Initialize all Go modules, React app, and top-level config files.

**Files to create:**

```
# Root
Makefile
.env.example
.gitignore (update)

# Gateway service
services/gateway/go.mod
services/gateway/go.sum
services/gateway/cmd/gateway/main.go
services/gateway/internal/router/router.go
services/gateway/internal/handler/health.go
services/gateway/internal/middleware/cors.go
services/gateway/internal/middleware/logging.go
services/gateway/Dockerfile

# Moderation service
services/moderation/go.mod
services/moderation/go.sum
services/moderation/cmd/moderation/main.go
services/moderation/internal/handler/health.go
services/moderation/internal/model/submission.go
services/moderation/internal/model/decision.go
services/moderation/Dockerfile

# Policy engine
services/policy-engine/go.mod
services/policy-engine/go.sum
services/policy-engine/cmd/policy-engine/main.go
services/policy-engine/internal/handler/health.go
services/policy-engine/internal/model/policy.go
services/policy-engine/Dockerfile

# Review service
services/review/go.mod
services/review/go.sum
services/review/cmd/review/main.go
services/review/internal/handler/health.go
services/review/internal/model/review.go
services/review/Dockerfile

# React app
apps/web/package.json
apps/web/vite.config.ts
apps/web/tsconfig.json
apps/web/tsconfig.node.json
apps/web/tailwind.config.ts
apps/web/postcss.config.js
apps/web/index.html
apps/web/src/main.tsx
apps/web/src/App.tsx
apps/web/src/index.css
apps/web/src/vite-env.d.ts
apps/web/Dockerfile
apps/web/.env.example
```

**Tests:**
- TDD: Health endpoint unit tests for each Go service
- Go test files co-located: `services/*/internal/handler/health_test.go`

**Control IDs:** OBS-001 (health endpoints are the first observability surface)

**Docker Compose:** Not yet (next PR)

---

### PR 1.2: Docker Compose and Infrastructure Services

**Description:** Full Docker Compose with PostgreSQL, Redis, all four Go services, and the React dev server.

**Files to create:**

```
docker-compose.yml
docker-compose.dev.yml           # Dev overrides (hot reload, volumes)
scripts/wait-for-it.sh           # Container readiness script
```

**Docker Compose services defined:**
- `postgres` (PostgreSQL 16, port 5432)
- `redis` (Redis 7, port 6379)
- `gateway` (port 8080, depends on moderation, policy-engine, review)
- `moderation` (port 8081, depends on postgres, redis)
- `policy-engine` (port 8082, depends on postgres)
- `review` (port 8083, depends on postgres)
- `web` (port 3000, depends on gateway)

**Tests:**
- Integration: `tests/integration/docker_compose_test.sh` -- verifies all services start and respond to health checks

**Control IDs:** OBS-001 (infrastructure is running and observable)

---

### PR 1.3: Database Schema and Migrations

**Description:** Initial PostgreSQL schema covering all core domain entities. Uses `golang-migrate` for migration management.

**Files to create:**

```
migrations/000001_create_submissions.up.sql
migrations/000001_create_submissions.down.sql
migrations/000002_create_decisions.up.sql
migrations/000002_create_decisions.down.sql
migrations/000003_create_policies.up.sql
migrations/000003_create_policies.down.sql
migrations/000004_create_reviews.up.sql
migrations/000004_create_reviews.down.sql
migrations/000005_create_evidence.up.sql
migrations/000005_create_evidence.down.sql
migrations/000006_create_users_roles.up.sql
migrations/000006_create_users_roles.down.sql
```

**Schema overview:**

```sql
-- submissions
CREATE TABLE text_submissions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_hash    TEXT NOT NULL,           -- SHA-256 hash for dedup/privacy
    content_encrypted BYTEA,                -- Optional encrypted content
    context_type    TEXT,                    -- chat, comment, review, ai-prompt
    context_region  TEXT,                    -- ISO 3166-1 region code
    source_id       TEXT,                    -- External correlation ID
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- moderation_decisions
CREATE TABLE moderation_decisions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id     UUID NOT NULL REFERENCES text_submissions(id),
    model_name        TEXT NOT NULL,
    model_version     TEXT NOT NULL,
    scores            JSONB NOT NULL,        -- {hate: 0.91, toxicity: 0.88, ...}
    policy_id         UUID,
    policy_version    INT,
    automated_action  TEXT NOT NULL,          -- allow, warn, block, escalate
    correlation_id    TEXT NOT NULL,
    latency_ms        INT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- policies
CREATE TABLE policies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    version         INT NOT NULL DEFAULT 1,
    status          TEXT NOT NULL DEFAULT 'draft', -- draft, published, archived
    thresholds      JSONB NOT NULL,
    actions         JSONB NOT NULL,
    scope_context   TEXT,
    scope_region    TEXT,
    effective_from  TIMESTAMPTZ,
    effective_to    TIMESTAMPTZ,
    created_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, version)
);

-- review_actions
CREATE TABLE review_actions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    decision_id     UUID NOT NULL REFERENCES moderation_decisions(id),
    reviewer_id     UUID NOT NULL,
    action          TEXT NOT NULL,           -- approve, reject, edit, escalate
    rationale       TEXT,
    override_action TEXT,                    -- New action if overridden
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- evidence_records
CREATE TABLE evidence_records (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    evidence_type     TEXT NOT NULL,
    submission_id     UUID REFERENCES text_submissions(id),
    decision_id       UUID REFERENCES moderation_decisions(id),
    review_id         UUID REFERENCES review_actions(id),
    control_id        TEXT NOT NULL,
    policy_id         TEXT,
    policy_version    INT,
    payload           JSONB NOT NULL,
    immutable         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Prevent UPDATE/DELETE on evidence_records via trigger

-- users
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL UNIQUE,
    display_name    TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'viewer', -- viewer, moderator, admin, auditor
    api_key_hash    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Tests:**
- TDD: Migration up/down roundtrip test
- Integration: Schema validation test (all tables exist, constraints enforced)

**Control IDs:** AUD-001 (evidence table with immutability trigger), GOV-001 (users/roles table)

---

### PR 1.4: JSON Schemas and Control Registry

**Description:** Define JSON Schemas for API contracts and the YAML control registry.

**Files to create:**

```
schemas/submission.json
schemas/decision.json
schemas/policy.json
schemas/evidence.json
schemas/review-action.json

controls/control-registry.yaml
```

**Tests:**
- TDD: Schema validation tests using Go `jsonschema` library
- CDD: Validate control registry references match schema fields

**Control IDs:** All control IDs registered in `control-registry.yaml`

---

### PR 1.5: Shared Go Libraries

**Description:** Create shared packages used across services -- database connection, config, logging, error handling, and HTTP helpers.

**Files to create:**

```
pkg/database/postgres.go         # Connection pool setup
pkg/database/postgres_test.go
pkg/cache/redis.go               # Redis client wrapper
pkg/cache/redis_test.go
pkg/config/config.go             # Env-based config loading
pkg/config/config_test.go
pkg/logging/logger.go            # Structured logging (slog)
pkg/logging/logger_test.go
pkg/httputil/response.go         # Standard JSON response helpers
pkg/httputil/response_test.go
pkg/httputil/errors.go           # Standard error responses
pkg/middleware/correlation.go    # Correlation ID middleware
pkg/middleware/correlation_test.go
go.mod                           # Root module for shared packages
```

**Tests:**
- TDD: Unit tests for every shared package

**Control IDs:** MOD-003 (correlation ID), OBS-001 (structured logging)

---

### Phase 1 Deliverable

After merging PRs 1.1 through 1.5:
- `docker-compose up` starts all services with health checks passing
- PostgreSQL schema is applied via migrations
- All Go services compile and serve `/health`
- React app renders a blank shell at `localhost:3000`
- JSON schemas and control registry are committed
- Shared libraries are available for import

---

## 6. Phase 2: Core Moderation

**Goal:** Implement the moderation pipeline -- HuggingFace API integration, request handling, caching, and the moderation decision flow.

**Depends on:** Phase 1

**Epics touched:** A (Core Moderation Platform)

### PR 2.1: HuggingFace API Client

**Description:** Go client for the HuggingFace Friendly Text Moderator API with retry, timeout, and circuit breaker.

**Files to create/modify:**

```
services/moderation/internal/client/huggingface.go
services/moderation/internal/client/huggingface_test.go
services/moderation/internal/client/types.go
services/moderation/internal/client/mock.go          # Mock for testing
tests/fixtures/huggingface_responses.json            # Fixture responses
```

**Tests:**
- TDD: Client unit tests with mock HTTP server
- TDD: Timeout, retry, and circuit breaker tests
- TDD: Response parsing for all six categories

**Control IDs:** MOD-001 (automated text classification)

---

### PR 2.2: Moderation Service Business Logic

**Description:** Core moderation service -- accept text, call HuggingFace, persist submission and decision, return result.

**Files to create/modify:**

```
services/moderation/internal/service/moderate.go
services/moderation/internal/service/moderate_test.go
services/moderation/internal/repository/submission.go
services/moderation/internal/repository/submission_test.go
services/moderation/internal/repository/decision.go
services/moderation/internal/repository/decision_test.go
services/moderation/internal/handler/moderate.go
services/moderation/internal/handler/moderate_test.go
```

**Tests:**
- TDD: Service layer unit tests (mock client, mock repo)
- TDD: Repository tests against test database
- TDD: Handler tests (HTTP request/response)
- BDD: `tests/bdd/features/moderation.feature`
  ```gherkin
  Feature: Text Moderation
    Scenario: Moderate clean text
      Given a text submission "Hello, how are you?"
      When the moderation service processes it
      Then the automated action should be "allow"
      And all category scores should be below 0.5

    Scenario: Moderate toxic text
      Given a text submission containing hate speech
      When the moderation service processes it
      Then the automated action should be "block"
      And the hate score should be above 0.8
  ```
- CDD: Verify MOD-001 evidence is generated for every decision

**Control IDs:** MOD-001, MOD-003 (pipeline with correlation ID)

---

### PR 2.3: Redis Caching Layer

**Description:** Cache moderation results by content hash. Respect policy version for cache invalidation.

**Files to create/modify:**

```
services/moderation/internal/cache/moderation_cache.go
services/moderation/internal/cache/moderation_cache_test.go
services/moderation/internal/service/moderate.go      # Add cache lookup
```

**Tests:**
- TDD: Cache hit/miss tests
- TDD: Cache invalidation on policy change
- TDD: TTL expiration test
- BDD: `tests/bdd/features/caching.feature`
  ```gherkin
  Feature: Moderation Caching
    Scenario: Repeated text returns cached result
      Given text "Hello world" was previously moderated
      When the same text is submitted again
      Then the result should be returned from cache
      And the HuggingFace API should not be called
  ```

**Control IDs:** MOD-004 (latency optimization and caching)

---

### PR 2.4: Gateway Routes for Moderation

**Description:** Wire up the gateway to proxy moderation requests to the moderation service.

**Files to create/modify:**

```
services/gateway/internal/handler/moderate.go
services/gateway/internal/handler/moderate_test.go
services/gateway/internal/router/router.go          # Add moderation routes
services/gateway/internal/client/moderation.go      # Internal HTTP client
```

**API endpoint:** `POST /api/v1/moderate`

**Request:**
```json
{
  "text": "string",
  "context": {
    "type": "chat|comment|review|ai-prompt",
    "region": "US|EU|...",
    "source_id": "string"
  }
}
```

**Response:**
```json
{
  "decision_id": "uuid",
  "action": "allow|warn|block|escalate",
  "scores": {
    "toxicity": 0.12,
    "hate": 0.03,
    "harassment": 0.08,
    "sexual": 0.01,
    "violence": 0.02,
    "profanity": 0.15
  },
  "policy": {
    "id": "uuid",
    "version": 1,
    "name": "string"
  },
  "explanation": "string"
}
```

**Tests:**
- TDD: Gateway handler tests
- Integration: End-to-end moderation flow (gateway -> moderation -> mock HF)

**Control IDs:** MOD-001, MOD-003

---

### Phase 2 Deliverable

After merging PRs 2.1 through 2.4:
- `POST /api/v1/moderate` accepts text and returns moderation decisions
- HuggingFace API is called with retry and circuit breaker
- Results are cached in Redis
- Submissions and decisions are persisted in PostgreSQL
- Correlation IDs trace requests end-to-end

---

## 7. Phase 3: Policy Engine

**Goal:** Implement configurable, versioned policies with deterministic evaluation, region awareness, and rollback.

**Depends on:** Phase 2 (moderation decisions reference policies)

**Epics touched:** B (Policy Engine)

### PR 3.1: Policy Data Model and Repository

**Description:** CRUD operations for policies with versioning semantics. Published policies are immutable.

**Files to create/modify:**

```
services/policy-engine/internal/model/policy.go
services/policy-engine/internal/repository/policy.go
services/policy-engine/internal/repository/policy_test.go
services/policy-engine/internal/service/policy_crud.go
services/policy-engine/internal/service/policy_crud_test.go
services/policy-engine/internal/handler/policy.go
services/policy-engine/internal/handler/policy_test.go
```

**Tests:**
- TDD: Repository CRUD tests
- TDD: Immutability enforcement (cannot update published policy)
- TDD: Version increment on publish
- BDD: `tests/bdd/features/policy_management.feature`
  ```gherkin
  Feature: Policy Management
    Scenario: Create and publish a policy
      Given an admin creates a policy "Youth Safe Mode"
      And sets hate threshold to 0.6 with action "block"
      When the admin publishes the policy
      Then the policy status should be "published"
      And the policy version should be 1

    Scenario: Published policy is immutable
      Given a published policy "Youth Safe Mode" v1
      When an admin attempts to modify thresholds
      Then the modification should be rejected
  ```

**Control IDs:** POL-001 (threshold-based policy), POL-002 (versioning)

---

### PR 3.2: Policy Evaluation Engine

**Description:** Deterministic evaluation: given model scores and a resolved policy, return the correct action.

**Files to create/modify:**

```
services/policy-engine/internal/evaluator/evaluate.go
services/policy-engine/internal/evaluator/evaluate_test.go
services/policy-engine/internal/evaluator/resolver.go    # Policy resolution
services/policy-engine/internal/evaluator/resolver_test.go
```

**Evaluation logic:**
1. Resolve applicable policy by context type and region
2. For each category score, check against policy threshold
3. Determine highest-priority action (block > escalate > warn > allow)
4. Return action with explanation referencing policy version

**Tests:**
- TDD: Exhaustive evaluation tests (every action type, edge cases)
- TDD: Policy resolution by region and context
- TDD: Tie-breaking rules
- BDD: `tests/bdd/features/policy_evaluation.feature`
  ```gherkin
  Feature: Policy Evaluation
    Scenario: Block when hate score exceeds threshold
      Given a published policy with hate threshold 0.6 action "block"
      And moderation scores with hate at 0.91
      When the policy engine evaluates
      Then the action should be "block"
      And the explanation should reference the policy name and version

    Scenario: Allow when all scores are below thresholds
      Given a published policy with all thresholds at 0.5
      And moderation scores all below 0.3
      When the policy engine evaluates
      Then the action should be "allow"
  ```

**Control IDs:** POL-001

---

### PR 3.3: Policy Versioning, Rollback, and Region Awareness

**Description:** Support publishing new versions, rolling back to previous versions, and region-based policy resolution.

**Files to create/modify:**

```
services/policy-engine/internal/service/versioning.go
services/policy-engine/internal/service/versioning_test.go
services/policy-engine/internal/service/rollback.go
services/policy-engine/internal/service/rollback_test.go
services/policy-engine/internal/evaluator/resolver.go    # Add region logic
services/policy-engine/internal/evaluator/resolver_test.go
```

**Tests:**
- TDD: New version creation from existing policy
- TDD: Rollback restores previous effective version
- TDD: Region-specific policy resolution
- TDD: Default policy fallback when no region match
- BDD: `tests/bdd/features/policy_versioning.feature`
  ```gherkin
  Feature: Policy Versioning
    Scenario: Rollback to previous version
      Given policy "Social Standard" v2 is published
      When an admin rolls back to v1
      Then the effective policy should be v1
      And v2 should be archived

    Scenario: Regional policy resolution
      Given a policy "EU Strict" scoped to region "EU"
      And a policy "US Standard" scoped to region "US"
      When a submission from region "EU" is evaluated
      Then the "EU Strict" policy should be applied
  ```

**Control IDs:** POL-002 (versioning/rollback), POL-003 (regional resolution)

---

### PR 3.4: Integrate Policy Engine with Moderation Service

**Description:** Wire the moderation service to call the policy engine for evaluation before returning decisions.

**Files to modify:**

```
services/moderation/internal/service/moderate.go
services/moderation/internal/client/policy_engine.go
services/moderation/internal/client/policy_engine_test.go
services/gateway/internal/router/router.go           # Policy API routes
services/gateway/internal/handler/policy.go
services/gateway/internal/client/policy_engine.go
```

**Gateway API endpoints:**
- `GET /api/v1/policies` -- List policies
- `POST /api/v1/policies` -- Create policy
- `GET /api/v1/policies/:id` -- Get policy
- `PUT /api/v1/policies/:id/publish` -- Publish policy
- `PUT /api/v1/policies/:id/rollback` -- Rollback to previous version

**Tests:**
- Integration: Full flow -- submit text -> moderation -> policy evaluation -> decision
- TDD: Policy engine client tests

**Control IDs:** MOD-001, POL-001

---

### Phase 3 Deliverable

After merging PRs 3.1 through 3.4:
- Policies can be created, versioned, published, and rolled back
- Moderation decisions are evaluated against the resolved policy
- Region and context type drive policy resolution
- Policy CRUD is available via gateway API
- Historical decisions are unaffected by policy changes

---

## 8. Phase 4: User Experience

**Goal:** Build the end-user facing React components for inline moderation feedback, warnings, and rewrite suggestions.

**Depends on:** Phase 2 (moderation API), Phase 3 (policy-driven actions)

**Epics touched:** C (End-User Experience)

### PR 4.1: API Client and Type Definitions

**Description:** TypeScript API client for the gateway, shared types, and React query hooks.

**Files to create:**

```
apps/web/src/lib/api-client.ts
apps/web/src/lib/api-client.test.ts
apps/web/src/types/moderation.ts
apps/web/src/types/policy.ts
apps/web/src/hooks/use-moderate.ts
apps/web/src/hooks/use-moderate.test.ts
```

**Tests:**
- TDD: API client unit tests (mocked fetch)
- TDD: Hook tests with React Testing Library

**Control IDs:** MOD-002 (real-time feedback enforcement)

---

### PR 4.2: Text Input Component with Inline Feedback

**Description:** Chat/comment input component with real-time moderation feedback. States: neutral, warning, blocked.

**Files to create:**

```
apps/web/src/components/text-input/TextInput.tsx
apps/web/src/components/text-input/TextInput.test.tsx
apps/web/src/components/text-input/ModerationFeedback.tsx
apps/web/src/components/text-input/ModerationFeedback.test.tsx
apps/web/src/components/ui/                          # shadcn/ui components
```

**Tests:**
- TDD: Component renders in neutral state
- TDD: Warning state displays category and confidence
- TDD: Blocked state disables submit button
- TDD: Feedback appears within UI (200ms SLA tested in E2E)
- BDD: `tests/bdd/features/user_feedback.feature`
  ```gherkin
  Feature: Inline User Feedback
    Scenario: Warning shown for borderline content
      Given the user types text that triggers a "warn" action
      When the moderation result is received
      Then an inline warning should be displayed
      And the warning should show the category and confidence
      And the submit button should remain enabled

    Scenario: Submit blocked for harmful content
      Given the user types text that triggers a "block" action
      When the moderation result is received
      Then a block message should be displayed
      And the submit button should be disabled
  ```

**Control IDs:** MOD-002

---

### PR 4.3: User Messaging and Rewrite Suggestions

**Description:** Human-readable messaging (no internal model details exposed). Optional rewrite suggestion feature.

**Files to create:**

```
apps/web/src/components/text-input/MessageDisplay.tsx
apps/web/src/components/text-input/MessageDisplay.test.tsx
apps/web/src/components/text-input/RewriteSuggestion.tsx
apps/web/src/components/text-input/RewriteSuggestion.test.tsx
apps/web/src/lib/message-templates.ts
apps/web/src/lib/message-templates.test.ts
```

**Tests:**
- TDD: Messages are human-readable (no score numbers in user-facing text)
- TDD: Rewrite suggestion renders when available
- TDD: Feature flag disables rewrite suggestions

**Control IDs:** MOD-002

---

### Phase 4 Deliverable

After merging PRs 4.1 through 4.3:
- Users see real-time inline feedback as they type
- Warnings display category and human-readable explanation
- Blocked content disables submission
- Rewrite suggestions are optionally displayed
- No internal model details are exposed to end users

---

## 9. Phase 5: Moderator Experience

**Goal:** Build the moderator queue, detail view, review actions, and feedback capture workflow.

**Depends on:** Phase 2 (decisions exist), Phase 4 (UI foundation)

**Epics touched:** D (Moderator Experience)

### PR 5.1: Review Service Business Logic

**Description:** Queue management, filtering, review action persistence, and feedback capture in the Go review service.

**Files to create/modify:**

```
services/review/internal/service/queue.go
services/review/internal/service/queue_test.go
services/review/internal/service/review_action.go
services/review/internal/service/review_action_test.go
services/review/internal/repository/review.go
services/review/internal/repository/review_test.go
services/review/internal/handler/queue.go
services/review/internal/handler/queue_test.go
services/review/internal/handler/review_action.go
services/review/internal/handler/review_action_test.go
```

**Tests:**
- TDD: Queue returns only actionable items (pending decisions with escalate/block)
- TDD: Sort and filter by category, confidence, date
- TDD: Review actions are persisted with reviewer ID
- TDD: Override records capture structured feedback
- BDD: `tests/bdd/features/moderation_queue.feature`
  ```gherkin
  Feature: Moderation Queue
    Scenario: Queue shows pending reviews
      Given 3 decisions with action "escalate" exist
      And 2 decisions with action "allow" exist
      When a moderator views the queue
      Then only the 3 escalated decisions should appear

    Scenario: Moderator approves a block
      Given a decision with action "block" is in the queue
      When a moderator approves the block
      Then the review action should be "approve"
      And an evidence record should be generated
  ```

**Control IDs:** GOV-002 (human-in-the-loop review)

---

### PR 5.2: Gateway Routes for Review

**Description:** Wire review API endpoints through the gateway.

**Files to create/modify:**

```
services/gateway/internal/handler/review.go
services/gateway/internal/handler/review_test.go
services/gateway/internal/client/review.go
services/gateway/internal/router/router.go    # Add review routes
```

**API endpoints:**
- `GET /api/v1/review/queue` -- List queue items (with filtering)
- `GET /api/v1/review/queue/:id` -- Get queue item detail
- `POST /api/v1/review/queue/:id/action` -- Submit review action
- `GET /api/v1/review/actions` -- List review actions (audit trail)

**Tests:**
- TDD: Handler tests for each endpoint
- Integration: Full review flow

**Control IDs:** GOV-002

---

### PR 5.3: Moderator UI Components

**Description:** React components for the moderator queue, detail view, and review action forms.

**Files to create:**

```
apps/web/src/pages/ModeratorQueue.tsx
apps/web/src/pages/ModeratorQueue.test.tsx
apps/web/src/pages/ModerationDetail.tsx
apps/web/src/pages/ModerationDetail.test.tsx
apps/web/src/components/moderator/QueueTable.tsx
apps/web/src/components/moderator/QueueTable.test.tsx
apps/web/src/components/moderator/QueueFilters.tsx
apps/web/src/components/moderator/QueueFilters.test.tsx
apps/web/src/components/moderator/ScoreDisplay.tsx
apps/web/src/components/moderator/ScoreDisplay.test.tsx
apps/web/src/components/moderator/ReviewActionBar.tsx
apps/web/src/components/moderator/ReviewActionBar.test.tsx
apps/web/src/hooks/use-queue.ts
apps/web/src/hooks/use-review-action.ts
```

**Tests:**
- TDD: Queue table renders items with correct columns
- TDD: Filters update query params
- TDD: Detail view shows text, scores, policy, recommendation
- TDD: Action bar allows approve, reject, edit, escalate
- TDD: Feedback capture form persists rationale

**Control IDs:** GOV-002, AUD-002

---

### Phase 5 Deliverable

After merging PRs 5.1 through 5.3:
- Moderators see a queue of actionable items
- Queue supports sorting and filtering by category, confidence, status
- Detail view shows full context: text, scores, policy, recommendation
- Review actions (approve, reject, edit, escalate) are persisted
- Override feedback is captured as structured data
- Evidence records generated for every review action

---

## 10. Phase 6: Admin and Governance

**Goal:** Build the admin dashboard with policy management UI, RBAC enforcement, and audit log viewer.

**Depends on:** Phase 3 (policy CRUD), Phase 5 (review service)

**Epics touched:** E (Admin and Governance)

### PR 6.1: RBAC Middleware and User Management

**Description:** Role-based access control enforced at the gateway. Roles: viewer, moderator, admin, auditor.

**Files to create/modify:**

```
services/gateway/internal/middleware/auth.go
services/gateway/internal/middleware/auth_test.go
services/gateway/internal/middleware/rbac.go
services/gateway/internal/middleware/rbac_test.go
services/gateway/internal/handler/users.go
services/gateway/internal/handler/users_test.go
services/gateway/internal/repository/user.go
services/gateway/internal/repository/user_test.go
```

**RBAC matrix:**

| Endpoint | viewer | moderator | admin | auditor |
|----------|--------|-----------|-------|---------|
| POST /api/v1/moderate | yes | yes | yes | no |
| GET /api/v1/review/queue | no | yes | yes | no |
| POST /api/v1/review/queue/:id/action | no | yes | yes | no |
| CRUD /api/v1/policies | no | no | yes | no |
| GET /api/v1/audit/* | no | no | yes | yes |
| CRUD /api/v1/users | no | no | yes | no |

**Tests:**
- TDD: Unauthorized access returns 403
- TDD: Each role can only access permitted endpoints
- TDD: Separation of duties -- moderators cannot create policies
- BDD: `tests/bdd/features/rbac.feature`
  ```gherkin
  Feature: Role-Based Access Control
    Scenario: Moderator cannot create policies
      Given a user with role "moderator"
      When they attempt to create a policy
      Then the request should be rejected with 403

    Scenario: Admin can create policies
      Given a user with role "admin"
      When they create a policy "Test Policy"
      Then the policy should be created successfully
  ```

**Control IDs:** GOV-001 (RBAC), GOV-003 (separation of duties)

---

### PR 6.2: Policy Management UI

**Description:** Admin UI for creating, editing, publishing, and rolling back policies.

**Files to create:**

```
apps/web/src/pages/PolicyList.tsx
apps/web/src/pages/PolicyList.test.tsx
apps/web/src/pages/PolicyEditor.tsx
apps/web/src/pages/PolicyEditor.test.tsx
apps/web/src/components/admin/PolicyForm.tsx
apps/web/src/components/admin/PolicyForm.test.tsx
apps/web/src/components/admin/ThresholdEditor.tsx
apps/web/src/components/admin/ThresholdEditor.test.tsx
apps/web/src/components/admin/PolicyVersionHistory.tsx
apps/web/src/components/admin/PolicyVersionHistory.test.tsx
apps/web/src/hooks/use-policies.ts
```

**Tests:**
- TDD: Form validation for thresholds (0.0-1.0 range)
- TDD: Draft/published states reflected in UI
- TDD: Version history displays all versions
- TDD: Publish triggers confirmation dialog

**Control IDs:** GOV-004 (policy management UI), POL-001, POL-002

---

### PR 6.3: Audit Log Viewer

**Description:** Searchable, exportable audit log UI for admins and auditors.

**Files to create/modify:**

```
services/gateway/internal/handler/audit.go
services/gateway/internal/handler/audit_test.go
apps/web/src/pages/AuditLog.tsx
apps/web/src/pages/AuditLog.test.tsx
apps/web/src/components/admin/AuditTable.tsx
apps/web/src/components/admin/AuditTable.test.tsx
apps/web/src/components/admin/AuditFilters.tsx
apps/web/src/components/admin/AuditFilters.test.tsx
apps/web/src/hooks/use-audit-log.ts
```

**API endpoints:**
- `GET /api/v1/audit/evidence` -- Search evidence records
- `GET /api/v1/audit/evidence/:id` -- Get evidence detail
- `GET /api/v1/audit/export` -- Export evidence as JSON/CSV

**Tests:**
- TDD: Search by date range, control ID, policy ID
- TDD: Export matches stored evidence exactly
- TDD: Pagination works correctly

**Control IDs:** AUD-003 (audit log viewer)

---

### Phase 6 Deliverable

After merging PRs 6.1 through 6.3:
- RBAC enforced on all API endpoints
- Admins can create, edit, publish, and rollback policies via UI
- Moderators cannot modify policies (separation of duties)
- Audit log is searchable and exportable
- Auditors can view evidence without modification capability

---

## 11. Phase 7: Evidence and Compliance

**Goal:** Implement the full CDD evidence pipeline -- schema validation, immutable storage, export, and compliance attestation.

**Depends on:** Phase 5 (review actions generate evidence), Phase 6 (audit viewer)

**Epics touched:** F (Agentic SDLC and Evidence)

### PR 7.1: Evidence Schema Validation and Generation

**Description:** Every moderation decision and review action generates a validated evidence record.

**Files to create/modify:**

```
services/moderation/internal/service/evidence.go
services/moderation/internal/service/evidence_test.go
services/review/internal/service/evidence.go
services/review/internal/service/evidence_test.go
pkg/evidence/generator.go
pkg/evidence/generator_test.go
pkg/evidence/validator.go
pkg/evidence/validator_test.go
```

**Tests:**
- TDD: Evidence records reference policy, control, and outcome
- TDD: Evidence validates against JSON schema
- TDD: Missing fields rejected
- CDD: Every code path that produces a decision also produces evidence
  ```
  // CDD Control Annotation
  // Control: AUD-001 - Immutable Evidence Storage
  // Verify: evidence record generated for every moderation decision
  ```

**Control IDs:** AUD-001

---

### PR 7.2: Immutable Evidence Storage

**Description:** Database trigger preventing UPDATE/DELETE on evidence_records. Append-only enforcement.

**Files to create/modify:**

```
migrations/000007_evidence_immutability.up.sql
migrations/000007_evidence_immutability.down.sql
pkg/evidence/immutability_test.go
```

**Migration content:**
```sql
-- Prevent any UPDATE or DELETE on evidence_records
CREATE OR REPLACE FUNCTION prevent_evidence_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Evidence records are immutable. UPDATE and DELETE are prohibited.';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER evidence_immutable_trigger
    BEFORE UPDATE OR DELETE ON evidence_records
    FOR EACH ROW
    EXECUTE FUNCTION prevent_evidence_mutation();
```

**Tests:**
- TDD: INSERT succeeds
- TDD: UPDATE throws exception
- TDD: DELETE throws exception
- CDD: Immutability verified against AUD-001

**Control IDs:** AUD-001

---

### PR 7.3: Evidence Export and Compliance Attestation

**Description:** Export evidence as JSON and CSV. Generate compliance attestation documents.

**Files to create/modify:**

```
services/gateway/internal/handler/audit.go            # Add export endpoints
services/gateway/internal/service/export.go
services/gateway/internal/service/export_test.go
services/gateway/internal/service/attestation.go
services/gateway/internal/service/attestation_test.go
```

**API endpoints:**
- `GET /api/v1/audit/export?format=json&from=DATE&to=DATE`
- `GET /api/v1/audit/export?format=csv&from=DATE&to=DATE`
- `GET /api/v1/audit/attestation?from=DATE&to=DATE`

**Tests:**
- TDD: JSON export matches stored evidence exactly
- TDD: CSV export includes all required fields
- TDD: Attestation includes workflow ID, phases completed, evidence count
- BDD: `tests/bdd/features/evidence_export.feature`
  ```gherkin
  Feature: Evidence Export
    Scenario: Export evidence as JSON
      Given 10 evidence records exist
      When an auditor exports evidence as JSON
      Then the export should contain exactly 10 records
      And each record should match the evidence schema

    Scenario: Export matches stored evidence
      Given a specific evidence record with ID "ev-001"
      When exported and compared to the database record
      Then all fields should match exactly
  ```

**Control IDs:** AUD-002 (evidence export), AUD-001

---

### Phase 7 Deliverable

After merging PRs 7.1 through 7.3:
- Every moderation decision generates validated evidence
- Every review action generates validated evidence
- Evidence records are append-only (database-enforced)
- Evidence can be exported as JSON or CSV
- Compliance attestation documents can be generated
- Full traceability from control ID to evidence record

---

## 12. Phase 8: Analytics and Monitoring

**Goal:** Implement moderation metrics, override analytics, policy effectiveness dashboards, and observability infrastructure.

**Depends on:** Phase 5 (review data), Phase 7 (evidence data)

**Epics touched:** G (Analytics and Monitoring)

### PR 8.1: Metrics Collection and Prometheus Integration

**Description:** Instrument all services with OpenTelemetry metrics. Expose Prometheus endpoints.

**Files to create/modify:**

```
pkg/metrics/metrics.go
pkg/metrics/metrics_test.go
services/moderation/internal/middleware/metrics.go
services/policy-engine/internal/middleware/metrics.go
services/review/internal/middleware/metrics.go
services/gateway/internal/middleware/metrics.go
docker-compose.yml                                   # Add Prometheus
configs/prometheus/prometheus.yml
```

**Metrics collected:**
- `moderation_requests_total` (counter, labels: action, category)
- `moderation_latency_ms` (histogram)
- `policy_evaluations_total` (counter, labels: policy_id, action)
- `review_actions_total` (counter, labels: action_type)
- `cache_hit_ratio` (gauge)

**Tests:**
- TDD: Metrics are incremented correctly
- TDD: Prometheus endpoint returns expected format
- Integration: Prometheus scrapes all services

**Control IDs:** OBS-001 (observability), ANL-001 (moderation metrics)

**Docker Compose changes:** Add `prometheus` service (port 9090)

---

### PR 8.2: Analytics API Endpoints

**Description:** Backend endpoints for moderation volume, override trends, and policy effectiveness.

**Files to create/modify:**

```
services/gateway/internal/handler/analytics.go
services/gateway/internal/handler/analytics_test.go
services/moderation/internal/repository/analytics.go
services/moderation/internal/repository/analytics_test.go
services/review/internal/repository/analytics.go
services/review/internal/repository/analytics_test.go
```

**API endpoints:**
- `GET /api/v1/analytics/moderation/volume` -- Volume by category over time
- `GET /api/v1/analytics/moderation/actions` -- Action distribution
- `GET /api/v1/analytics/overrides` -- Override rate and trends
- `GET /api/v1/analytics/policies/effectiveness` -- Policy effectiveness

**Tests:**
- TDD: Aggregation queries return correct counts
- TDD: Date range filtering works
- TDD: Override attribution is correct

**Control IDs:** ANL-001, ANL-002

---

### PR 8.3: Analytics Dashboard UI

**Description:** React dashboard with charts for moderation volume, trends, overrides, and policy effectiveness.

**Files to create:**

```
apps/web/src/pages/Dashboard.tsx
apps/web/src/pages/Dashboard.test.tsx
apps/web/src/components/analytics/VolumeChart.tsx
apps/web/src/components/analytics/VolumeChart.test.tsx
apps/web/src/components/analytics/ActionDistribution.tsx
apps/web/src/components/analytics/ActionDistribution.test.tsx
apps/web/src/components/analytics/OverrideTrends.tsx
apps/web/src/components/analytics/OverrideTrends.test.tsx
apps/web/src/components/analytics/PolicyEffectiveness.tsx
apps/web/src/components/analytics/PolicyEffectiveness.test.tsx
apps/web/src/hooks/use-analytics.ts
```

**Tests:**
- TDD: Charts render with mock data
- TDD: Date range picker updates chart data
- TDD: Loading and error states handled

**Control IDs:** ANL-001, ANL-002, ANL-003

---

### PR 8.4: Structured Logging and Tracing

**Description:** OpenTelemetry distributed tracing across all services. Structured JSON logging with correlation IDs.

**Files to create/modify:**

```
pkg/tracing/tracer.go
pkg/tracing/tracer_test.go
services/gateway/cmd/gateway/main.go          # Init tracer
services/moderation/cmd/moderation/main.go    # Init tracer
services/policy-engine/cmd/policy-engine/main.go
services/review/cmd/review/main.go
docker-compose.yml                             # Add Jaeger
configs/jaeger/                                # Jaeger config (if needed)
```

**Tests:**
- TDD: Trace context propagated across service calls
- TDD: Correlation ID appears in all log entries
- Integration: Jaeger receives traces from all services

**Control IDs:** OBS-001

**Docker Compose changes:** Add `jaeger` service (port 16686 UI, 4318 OTLP)

---

### Phase 8 Deliverable

After merging PRs 8.1 through 8.4:
- All services emit Prometheus metrics
- Distributed tracing across gateway -> moderation -> policy-engine
- Analytics dashboard shows volume, trends, overrides, policy effectiveness
- Structured JSON logging with correlation IDs
- Prometheus and Jaeger accessible via Docker Compose

---

## 13. Phase 9: Security Hardening

**Goal:** Implement authentication, API key management, TLS, data retention controls, and a final security audit.

**Depends on:** Phase 6 (RBAC foundation), Phase 8 (observability)

**Epics touched:** H (Platform and Security)

### PR 9.1: Authentication -- API Keys and OAuth

**Description:** API key authentication for programmatic access. OAuth 2.0 for web UI. JWT session tokens.

**Files to create/modify:**

```
services/gateway/internal/middleware/auth.go           # Enhance existing
services/gateway/internal/middleware/apikey.go
services/gateway/internal/middleware/apikey_test.go
services/gateway/internal/middleware/oauth.go
services/gateway/internal/middleware/oauth_test.go
services/gateway/internal/handler/auth.go
services/gateway/internal/handler/auth_test.go
services/gateway/internal/service/apikey.go
services/gateway/internal/service/apikey_test.go
migrations/000008_api_keys.up.sql
migrations/000008_api_keys.down.sql
```

**Tests:**
- TDD: Requests without API key or JWT are rejected (401)
- TDD: Valid API key grants access
- TDD: Expired JWT is rejected
- TDD: API keys are stored as hashes (not plaintext)
- BDD: `tests/bdd/features/authentication.feature`
  ```gherkin
  Feature: Authentication
    Scenario: Valid API key grants access
      Given a user with a valid API key
      When they call POST /api/v1/moderate
      Then the request should succeed

    Scenario: Missing credentials rejected
      Given no authentication credentials
      When they call POST /api/v1/moderate
      Then the request should return 401
  ```

**Control IDs:** SEC-002 (API key and OAuth authentication)

---

### PR 9.2: TLS Configuration

**Description:** TLS termination at the gateway. Self-signed certs for development, configuration for production certs.

**Files to create/modify:**

```
scripts/generate-certs.sh
configs/tls/README.md
services/gateway/cmd/gateway/main.go    # TLS listener option
docker-compose.yml                       # TLS port mapping
```

**Tests:**
- Integration: Gateway serves HTTPS
- Integration: HTTP redirects to HTTPS (when TLS enabled)

**Control IDs:** SEC-001 (TLS for data in transit)

---

### PR 9.3: Data Retention Controls

**Description:** Configurable retention periods. Automated purge of expired submissions (not evidence -- evidence is retained separately).

**Files to create/modify:**

```
services/moderation/internal/service/retention.go
services/moderation/internal/service/retention_test.go
services/moderation/internal/repository/retention.go
services/moderation/internal/repository/retention_test.go
migrations/000009_retention_policy.up.sql
migrations/000009_retention_policy.down.sql
configs/retention.yaml
```

**Retention rules:**
- Text submissions: configurable (default 90 days)
- Moderation decisions: configurable (default 1 year)
- Evidence records: never deleted (immutable)
- Review actions: follow decision retention

**Tests:**
- TDD: Expired submissions are purged
- TDD: Evidence records are never purged
- TDD: Retention settings are configurable
- BDD: `tests/bdd/features/data_retention.feature`
  ```gherkin
  Feature: Data Retention
    Scenario: Expired submissions are purged
      Given a submission older than the retention period
      When the retention job runs
      Then the submission should be deleted
      But the associated evidence record should remain
  ```

**Control IDs:** SEC-003 (data retention controls)

---

### PR 9.4: Security Audit and Hardening

**Description:** Rate limiting, input sanitization, SQL injection prevention audit, CORS tightening, security headers.

**Files to create/modify:**

```
services/gateway/internal/middleware/ratelimit.go
services/gateway/internal/middleware/ratelimit_test.go
services/gateway/internal/middleware/security_headers.go
services/gateway/internal/middleware/security_headers_test.go
services/gateway/internal/middleware/cors.go           # Tighten CORS
pkg/httputil/sanitize.go
pkg/httputil/sanitize_test.go
```

**Tests:**
- TDD: Rate limiting enforced per API key
- TDD: Oversized payloads rejected
- TDD: SQL injection attempts blocked
- TDD: Security headers present (X-Content-Type-Options, X-Frame-Options, CSP)
- TDD: CORS only allows configured origins

**Control IDs:** SEC-001, SEC-002, GOV-001

---

### Phase 9 Deliverable

After merging PRs 9.1 through 9.4:
- API key and OAuth authentication on all endpoints
- TLS configured for gateway
- Data retention enforced with configurable periods
- Rate limiting, security headers, input sanitization
- Evidence records are exempt from retention purge
- CORS tightened to specific origins

---

## 14. Phase 10: E2E Testing and Fixtures

**Goal:** Comprehensive Playwright E2E test suite, test fixtures, seed data, and complete test harness.

**Depends on:** All prior phases

**Epics touched:** Cross-cutting

### PR 10.1: Test Fixtures and Seed Data

**Description:** Deterministic seed data for all entities. Fixture files for mock HuggingFace responses.

**Files to create:**

```
tests/fixtures/seed/users.json
tests/fixtures/seed/policies.json
tests/fixtures/seed/submissions.json
tests/fixtures/seed/decisions.json
tests/fixtures/seed/reviews.json
tests/fixtures/seed/evidence.json
tests/fixtures/huggingface/toxic_response.json
tests/fixtures/huggingface/clean_response.json
tests/fixtures/huggingface/borderline_response.json
tests/fixtures/huggingface/error_response.json
scripts/seed/seed.go
scripts/seed/seed_test.go
```

**Tests:**
- TDD: Seed script populates all tables
- TDD: Seed data is idempotent (re-runnable)

**Control IDs:** None (testing infrastructure)

---

### PR 10.2: Playwright E2E Test Suite -- User Flows

**Description:** E2E tests for end-user text submission, inline feedback, and warnings.

**Files to create:**

```
tests/e2e/playwright.config.ts
tests/e2e/specs/text-submission.spec.ts
tests/e2e/specs/inline-feedback.spec.ts
tests/e2e/specs/blocked-content.spec.ts
tests/e2e/helpers/auth.ts
tests/e2e/helpers/seed.ts
tests/e2e/package.json
```

**Test scenarios:**
```typescript
// text-submission.spec.ts
test('clean text is submitted successfully', async ({ page }) => {
  // Type clean text -> no warning -> submit succeeds
});

test('warning displayed for borderline content', async ({ page }) => {
  // Type borderline text -> warning appears within 200ms -> can still submit
});

test('blocked content prevents submission', async ({ page }) => {
  // Type harmful text -> block message -> submit button disabled
});
```

**Control IDs:** MOD-002 (E2E verification of real-time feedback)

---

### PR 10.3: Playwright E2E Test Suite -- Moderator Flows

**Description:** E2E tests for moderator queue, detail view, and review actions.

**Files to create:**

```
tests/e2e/specs/moderator-queue.spec.ts
tests/e2e/specs/moderation-detail.spec.ts
tests/e2e/specs/review-action.spec.ts
```

**Test scenarios:**
```typescript
// moderator-queue.spec.ts
test('moderator sees pending items in queue', async ({ page }) => {
  // Login as moderator -> navigate to queue -> verify items listed
});

test('moderator approves a block decision', async ({ page }) => {
  // Click item -> view detail -> click approve -> verify status updated
});

test('moderator overrides with rationale', async ({ page }) => {
  // Click item -> click override -> enter rationale -> verify feedback captured
});
```

**Control IDs:** GOV-002 (E2E verification of human-in-the-loop)

---

### PR 10.4: Playwright E2E Test Suite -- Admin Flows

**Description:** E2E tests for policy management, RBAC enforcement, and audit log.

**Files to create:**

```
tests/e2e/specs/policy-management.spec.ts
tests/e2e/specs/rbac-enforcement.spec.ts
tests/e2e/specs/audit-log.spec.ts
```

**Test scenarios:**
```typescript
// policy-management.spec.ts
test('admin creates and publishes a policy', async ({ page }) => {
  // Login as admin -> create policy -> set thresholds -> publish -> verify
});

// rbac-enforcement.spec.ts
test('moderator cannot access policy editor', async ({ page }) => {
  // Login as moderator -> navigate to /policies -> verify 403 or redirect
});

// audit-log.spec.ts
test('auditor can search and export evidence', async ({ page }) => {
  // Login as auditor -> search by date range -> export -> verify file downloaded
});
```

**Control IDs:** GOV-001, GOV-003, GOV-004, AUD-003

---

### PR 10.5: CI Test Harness and Makefile Targets

**Description:** Makefile targets for running all test levels. Docker Compose test configuration.

**Files to create/modify:**

```
Makefile                                    # Add test targets
docker-compose.test.yml                    # Test environment overrides
scripts/run-tests.sh                       # Master test runner
tests/bdd/godog_test.go                    # BDD test runner entry point
.github/workflows/ci.yml                   # CI pipeline (optional)
```

**Makefile targets:**
```makefile
test-unit:        # Run all Go unit tests
test-integration: # Run integration tests against Docker services
test-bdd:         # Run Gherkin/godog BDD tests
test-e2e:         # Run Playwright E2E tests
test-cdd:         # Run compliance control verification
test-all:         # Run all test levels in sequence
lint:             # Run golangci-lint + eslint
```

**Control IDs:** All (test harness validates all controls)

---

### Phase 10 Deliverable

After merging PRs 10.1 through 10.5:
- Deterministic seed data for all entities
- Playwright E2E tests cover user, moderator, and admin flows
- BDD feature files cover all critical business scenarios
- CDD control verification tests validate evidence generation
- Single Makefile target runs all test levels
- CI pipeline ready

---

## 15. PR Strategy and Merge Order

### Merge Dependency Graph

```
Phase 1: 1.1 → 1.2 → 1.3 → 1.4 (parallel with 1.3) → 1.5
Phase 2: 2.1 → 2.2 → 2.3 → 2.4
Phase 3: 3.1 → 3.2 → 3.3 → 3.4
Phase 4: 4.1 → 4.2 → 4.3
Phase 5: 5.1 → 5.2 → 5.3
Phase 6: 6.1 → 6.2 (parallel with 6.3) → 6.3
Phase 7: 7.1 → 7.2 → 7.3
Phase 8: 8.1 → 8.2 → 8.3 → 8.4 (parallel with 8.3)
Phase 9: 9.1 → 9.2 → 9.3 → 9.4
Phase 10: 10.1 → 10.2 → 10.3 → 10.4 → 10.5
```

### Cross-Phase Dependencies

| Phase | Depends On |
|-------|-----------|
| Phase 1 | None |
| Phase 2 | Phase 1 |
| Phase 3 | Phase 2 |
| Phase 4 | Phase 2, Phase 3 |
| Phase 5 | Phase 2, Phase 4 |
| Phase 6 | Phase 3, Phase 5 |
| Phase 7 | Phase 5, Phase 6 |
| Phase 8 | Phase 5, Phase 7 |
| Phase 9 | Phase 6, Phase 8 |
| Phase 10 | All prior phases |

### PR Naming Convention

```
feat(phase-N/service): brief description

Examples:
feat(phase-1/scaffold): initialize Go modules and React app
feat(phase-2/moderation): HuggingFace API client with circuit breaker
feat(phase-3/policy): deterministic policy evaluation engine
feat(phase-6/rbac): role-based access control middleware
feat(phase-10/e2e): Playwright moderator flow tests
```

### PR Checklist (Every PR)

- [ ] All new files listed in implementation plan
- [ ] Unit tests written and passing
- [ ] BDD feature files added (if applicable)
- [ ] CDD control IDs declared in PR description
- [ ] Docker Compose changes documented (if applicable)
- [ ] JSON schema compliance verified (if applicable)
- [ ] No secrets or credentials committed

### Total PR Count

| Phase | PR Count |
|-------|---------|
| Phase 1 | 5 |
| Phase 2 | 4 |
| Phase 3 | 4 |
| Phase 4 | 3 |
| Phase 5 | 3 |
| Phase 6 | 3 |
| Phase 7 | 3 |
| Phase 8 | 4 |
| Phase 9 | 4 |
| Phase 10 | 5 |
| **Total** | **38** |

---

## 16. Test Strategy

### Test Pyramid

```
         /  E2E  \          ← Playwright (Phase 10)
        /   BDD   \         ← Gherkin + godog (every phase)
       / Integration\       ← Cross-service (every phase)
      /   CDD Control\     ← Control verification (Phase 7+)
     /     Unit Tests  \   ← Go test + React Testing Library (every phase)
    /____________________\
```

### Test Level Details

#### TDD -- Unit Tests (Every PR)

- **Go services:** Table-driven tests, mock interfaces, `testify` assertions
- **React components:** React Testing Library, `vitest`
- **Coverage target:** 80% line coverage per service
- **Naming:** `*_test.go` co-located with source, `*.test.ts(x)` co-located

#### BDD -- Gherkin Feature Files (Business-Critical PRs)

- **Runner:** `godog` (Go BDD framework)
- **Location:** `tests/bdd/features/*.feature`
- **Step definitions:** `tests/bdd/steps/*_steps.go`
- **Scenarios written for:**
  - Text moderation flow (Phase 2)
  - Policy evaluation (Phase 3)
  - Inline user feedback (Phase 4)
  - Moderator queue and review (Phase 5)
  - RBAC enforcement (Phase 6)
  - Evidence export (Phase 7)
  - Authentication (Phase 9)
  - Data retention (Phase 9)

#### CDD -- Compliance Control Tests (Phase 7+)

- **Purpose:** Verify that every control ID generates evidence
- **Approach:** Test that exercises a control path and asserts evidence record exists
- **Location:** `tests/cdd/controls_test.go`
- **Pattern:**
  ```go
  func TestControl_MOD001_GeneratesEvidence(t *testing.T) {
      // Submit text -> moderate -> assert evidence record with control_id="MOD-001"
  }
  func TestControl_AUD001_EvidenceIsImmutable(t *testing.T) {
      // Insert evidence -> attempt UPDATE -> assert failure
  }
  ```

#### E2E -- Playwright (Phase 10)

- **Browser:** Chromium
- **Auth:** Helper to login as different roles
- **Seed:** Fresh database seed before each test suite
- **Scenarios:** User flows, moderator flows, admin flows
- **SLA validation:** Assert inline feedback appears within 200ms

### BDD Feature File Inventory

| Feature File | Phase | Epic |
|-------------|-------|------|
| `moderation.feature` | 2 | A |
| `caching.feature` | 2 | A |
| `policy_management.feature` | 3 | B |
| `policy_evaluation.feature` | 3 | B |
| `policy_versioning.feature` | 3 | B |
| `user_feedback.feature` | 4 | C |
| `moderation_queue.feature` | 5 | D |
| `rbac.feature` | 6 | E |
| `evidence_export.feature` | 7 | F |
| `authentication.feature` | 9 | H |
| `data_retention.feature` | 9 | H |

---

## 17. Docker Compose Deployment Plan

### Service Evolution by Phase

| Phase | Services Added | Ports |
|-------|---------------|-------|
| 1 | postgres, redis, gateway, moderation, policy-engine, review, web | 5432, 6379, 8080-8083, 3000 |
| 8 | prometheus, jaeger | 9090, 16686 |
| 9 | (TLS configuration on gateway) | 443 |

### Final Docker Compose Topology

```yaml
services:
  # Infrastructure
  postgres:
    image: postgres:16
    ports: ["5432:5432"]
    volumes: [pgdata:/var/lib/postgresql/data]
    environment:
      POSTGRES_DB: civitas
      POSTGRES_USER: civitas
      POSTGRES_PASSWORD: ${DB_PASSWORD}

  redis:
    image: redis:7
    ports: ["6379:6379"]

  # Application
  gateway:
    build: services/gateway
    ports: ["8080:8080"]
    depends_on: [moderation, policy-engine, review]
    environment:
      MODERATION_URL: http://moderation:8081
      POLICY_ENGINE_URL: http://policy-engine:8082
      REVIEW_URL: http://review:8083

  moderation:
    build: services/moderation
    ports: ["8081:8081"]
    depends_on: [postgres, redis]
    environment:
      DATABASE_URL: postgres://civitas:${DB_PASSWORD}@postgres:5432/civitas
      REDIS_URL: redis://redis:6379
      HUGGINGFACE_API_KEY: ${HF_API_KEY}

  policy-engine:
    build: services/policy-engine
    ports: ["8082:8082"]
    depends_on: [postgres]
    environment:
      DATABASE_URL: postgres://civitas:${DB_PASSWORD}@postgres:5432/civitas

  review:
    build: services/review
    ports: ["8083:8083"]
    depends_on: [postgres]
    environment:
      DATABASE_URL: postgres://civitas:${DB_PASSWORD}@postgres:5432/civitas

  web:
    build: apps/web
    ports: ["3000:3000"]
    depends_on: [gateway]
    environment:
      VITE_API_URL: http://localhost:8080

  # Observability (Phase 8)
  prometheus:
    image: prom/prometheus
    ports: ["9090:9090"]
    volumes: [./configs/prometheus:/etc/prometheus]

  jaeger:
    image: jaegertracing/all-in-one
    ports: ["16686:16686", "4318:4318"]

volumes:
  pgdata:
```

### Environment Files

```
# .env.example
DB_PASSWORD=changeme
HF_API_KEY=hf_your_key_here
JWT_SECRET=changeme
CORS_ORIGINS=http://localhost:3000
TLS_ENABLED=false
RETENTION_DAYS_SUBMISSIONS=90
RETENTION_DAYS_DECISIONS=365
```

---

## Appendix A: Control ID to Phase Mapping

| Control ID | Phase Introduced | Phase Completed | PRs |
|-----------|-----------------|----------------|-----|
| MOD-001 | 2 | 2 | 2.1, 2.2, 2.4 |
| MOD-002 | 4 | 4 | 4.1, 4.2, 4.3 |
| MOD-003 | 1 | 2 | 1.5, 2.2 |
| MOD-004 | 2 | 2 | 2.3 |
| POL-001 | 3 | 3 | 3.1, 3.2, 3.4 |
| POL-002 | 3 | 3 | 3.1, 3.3 |
| POL-003 | 3 | 3 | 3.3 |
| GOV-001 | 1 | 6 | 1.3, 6.1 |
| GOV-002 | 5 | 5 | 5.1, 5.2, 5.3 |
| GOV-003 | 6 | 6 | 6.1 |
| GOV-004 | 6 | 6 | 6.2 |
| AUD-001 | 1 | 7 | 1.3, 7.1, 7.2 |
| AUD-002 | 7 | 7 | 7.3 |
| AUD-003 | 6 | 6 | 6.3 |
| SEC-001 | 9 | 9 | 9.2, 9.4 |
| SEC-002 | 9 | 9 | 9.1 |
| SEC-003 | 9 | 9 | 9.3 |
| OBS-001 | 1 | 8 | 1.1, 1.5, 8.1, 8.4 |
| ANL-001 | 8 | 8 | 8.1, 8.2, 8.3 |
| ANL-002 | 8 | 8 | 8.2, 8.3 |
| ANL-003 | 8 | 8 | 8.2, 8.3 |

---

## Appendix B: Epic to Phase Mapping

| Epic | Name | Primary Phase | Supporting Phases |
|------|------|--------------|-------------------|
| A | Core Moderation Platform | 2 | 1 |
| B | Policy Engine | 3 | -- |
| C | End-User Experience | 4 | -- |
| D | Moderator Experience | 5 | -- |
| E | Admin and Governance | 6 | -- |
| F | Agentic SDLC and Evidence | 7 | 1, 2, 5 |
| G | Analytics and Monitoring | 8 | -- |
| H | Platform and Security | 9 | 1 |
| -- | E2E Testing | 10 | All |

---

## Appendix C: Estimated Effort

| Phase | Estimated Duration | PRs | Key Risk |
|-------|--------------------|-----|----------|
| Phase 1: Foundation | 1 week | 5 | Toolchain setup issues |
| Phase 2: Core Moderation | 1 week | 4 | HuggingFace API reliability |
| Phase 3: Policy Engine | 1 week | 4 | Evaluation logic complexity |
| Phase 4: User Experience | 1 week | 3 | Real-time latency SLA |
| Phase 5: Moderator Experience | 1 week | 3 | Queue performance at scale |
| Phase 6: Admin and Governance | 1 week | 3 | RBAC edge cases |
| Phase 7: Evidence and Compliance | 0.5 week | 3 | Immutability enforcement |
| Phase 8: Analytics and Monitoring | 1 week | 4 | Aggregation query performance |
| Phase 9: Security Hardening | 1 week | 4 | OAuth integration complexity |
| Phase 10: E2E Testing | 1 week | 5 | Test environment stability |
| **Total** | **~9.5 weeks** | **38** | -- |
