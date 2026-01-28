# Civitas AI Text Moderator - Test Infrastructure

This directory contains the complete test infrastructure for the Civitas AI Text Moderator platform.

## Test Architecture

The test suite follows multiple testing methodologies:

- **TDD (Test-Driven Development)**: Go unit tests for core business logic
- **BDD (Behavior-Driven Development)**: Gherkin feature files for user-facing behavior
- **CDD (Control-Driven Development)**: Tests mapped to governance controls
- **E2E (End-to-End)**: Playwright tests for full user workflows

## Directory Structure

```
tests/
├── fixtures/              # Test data and fixtures
│   ├── seed-data.sql      # Database seed data for tests
│   ├── moderation_responses.json  # Mock API responses
│   └── policies.json      # Test policy definitions
├── helpers/               # Test helper utilities
│   ├── testdb.go          # Database setup and teardown
│   ├── mockserver.go      # Mock HuggingFace API server
│   ├── testredis.go       # Redis test utilities
│   └── mock-api-server.js # Node.js mock API server
├── bdd/                   # Behavior-Driven Development tests
│   └── features/          # Gherkin feature files
│       ├── moderation.feature
│       ├── policy_engine.feature
│       ├── review_workflow.feature
│       ├── evidence.feature
│       └── admin.feature
├── cdd/                   # Control-Driven Development tests
│   ├── controls_test.go   # Control verification tests
│   └── control-registry.yaml  # Control definitions
├── e2e/                   # End-to-End Playwright tests
│   ├── specs/             # Test specifications
│   │   ├── moderation-demo.spec.ts
│   │   ├── moderator-queue.spec.ts
│   │   ├── policy-management.spec.ts
│   │   └── audit-log.spec.ts
│   ├── fixtures/          # E2E fixtures
│   │   └── auth.ts        # Authentication fixtures
│   ├── playwright.config.ts
│   ├── package.json
│   └── Dockerfile.e2e
├── integration/           # Integration tests
└── unit/                  # Unit tests

```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.22+
- Node.js 20+
- Make

### Running All Tests

```bash
# Run all test suites
make test-all

# Run specific test suites
make test-unit          # Go unit tests
make test-integration   # Integration tests
make test-cdd           # Control-driven tests
make test-bdd           # Behavior-driven tests
make test-e2e           # End-to-end Playwright tests
```

### Running Tests in Docker

```bash
# Start test environment
docker compose -f docker-compose.yml -f docker-compose.test.yml up -d

# Run CDD tests
docker compose -f docker-compose.yml -f docker-compose.test.yml run cdd-tests

# Run BDD tests
docker compose -f docker-compose.yml -f docker-compose.test.yml run bdd-tests

# Run E2E tests
docker compose -f docker-compose.yml -f docker-compose.test.yml run e2e-tests

# Stop test environment
docker compose -f docker-compose.yml -f docker-compose.test.yml down
```

## Test Fixtures

### Database Seed Data

The `fixtures/seed-data.sql` file contains realistic test data:

- **Users**: admin, moderator, viewer (3 users with different roles)
- **Policies**: Standard, Youth Safe Mode, Relaxed (3 policies)
- **Submissions**: Safe and toxic content examples (5 submissions)
- **Decisions**: Automated moderation decisions (4 decisions)
- **Reviews**: Human review actions (2 reviews)
- **Evidence**: Immutable audit records (3 evidence records)

### Mock API Responses

The `fixtures/moderation_responses.json` contains mock HuggingFace API responses:

- **safe_text**: Low toxicity scores across all categories
- **toxic_text**: High toxicity and insult scores
- **hate_speech**: Very high hate and severe toxic scores
- **mild_profanity**: Moderate obscene scores
- **borderline**: Moderate toxicity, just below thresholds
- **api_timeout**: Simulates API timeout (504)
- **api_rate_limit**: Simulates rate limiting (429)

## Test Types

### 1. Control-Driven Development (CDD)

CDD tests verify that governance controls are properly implemented:

```go
// Example: MOD-001 verification
func TestMOD001_AutomatedClassification(t *testing.T) {
    // Verify text classification generates evidence
}
```

Controls are defined in `cdd/control-registry.yaml` and include:

- **MOD-001**: Automated Text Classification
- **MOD-002**: Real-Time User Feedback
- **POL-001**: Threshold-Based Moderation Policy
- **POL-003**: Regional Policy Resolution
- **GOV-002**: Human-in-the-Loop Review
- **AUD-001**: Immutable Evidence Storage

### 2. Behavior-Driven Development (BDD)

BDD tests use Gherkin syntax to describe user-facing behavior:

```gherkin
Feature: Text Moderation
  Scenario: Safe text is allowed
    Given a text submission "Hello, this is a friendly message!"
    When the text is submitted for moderation
    Then the moderation decision should be "allow"
```

### 3. End-to-End Tests (E2E)

E2E tests use Playwright to test complete user workflows:

```typescript
test('user can type text and see real-time feedback', async ({ page }) => {
  await page.goto('/demo');
  await page.fill('textarea[name="content"]', 'Hello world');
  await expect(page.locator('[data-testid="moderation-feedback"]')).toContainText('safe');
});
```

## Test Helpers

### Database Helpers

```go
import "github.com/proth1/text-moderator/tests/helpers"

func TestExample(t *testing.T) {
    db, cleanup := helpers.SetupTestDB(t)
    defer cleanup()

    // Use db.Pool for queries
}
```

### Redis Helpers

```go
redis, cleanup := helpers.SetupTestRedis(t)
defer cleanup()

redis.SetTestCache(t, "key", "value", 5*time.Minute)
```

### Mock API Server

```go
mock, cleanup := helpers.SetupMockServer(t)
defer cleanup()

// Use mock.URL() in your tests
```

## Environment Variables

### Test Database

```bash
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5433/civitas_test?sslmode=disable
TEST_REDIS_ADDR=localhost:6380
```

### E2E Tests

```bash
BASE_URL=http://localhost:3001
API_URL=http://localhost:8081/api/v1
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: |
          docker compose -f docker-compose.yml -f docker-compose.test.yml up --abort-on-container-exit
```

## Coverage Reports

```bash
# Generate coverage report
make test-coverage

# View HTML coverage report
open coverage.html
```

## Writing New Tests

### Adding a New Control Test

1. Define the control in `cdd/control-registry.yaml`
2. Write the test in `cdd/controls_test.go`
3. Map the test to the control

### Adding a New BDD Feature

1. Create a `.feature` file in `bdd/features/`
2. Write scenarios in Gherkin syntax
3. Implement step definitions (when framework is set up)

### Adding a New E2E Test

1. Create a `.spec.ts` file in `e2e/specs/`
2. Use the auth fixtures for authenticated tests
3. Follow the page object pattern for reusability

## Troubleshooting

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker compose ps postgres

# View logs
docker compose logs postgres

# Connect manually
psql postgres://postgres:postgres@localhost:5433/civitas_test
```

### Redis Connection Issues

```bash
# Check Redis is running
docker compose ps redis

# Test connection
redis-cli -p 6380 ping
```

### E2E Test Failures

```bash
# Run with UI mode
cd tests/e2e
npm run test:ui

# Run in headed mode
npm run test:headed

# Debug specific test
npm run test:debug -- specs/moderation-demo.spec.ts
```

## Best Practices

1. **Isolation**: Each test should be independent and not rely on others
2. **Cleanup**: Always use cleanup functions to reset state
3. **Fixtures**: Use fixtures for consistent test data
4. **Assertions**: Make assertions specific and meaningful
5. **Documentation**: Document complex test scenarios
6. **Performance**: Keep tests fast (< 5s for unit, < 30s for E2E)

## Resources

- [Playwright Documentation](https://playwright.dev)
- [Gherkin Reference](https://cucumber.io/docs/gherkin/reference/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Control-Driven Development](https://github.com/proth1/text-moderator/blob/main/controls/README.md)
