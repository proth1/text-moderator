# Testing Guide

## Overview

This guide provides detailed instructions for writing, running, and maintaining tests for the Civitas AI Text Moderator platform.

## Test Philosophy

Our testing approach follows these principles:

1. **Test Pyramid**: Most tests are fast unit tests, fewer integration tests, even fewer E2E tests
2. **Control-Driven**: All tests map to governance controls for compliance
3. **Behavior-Driven**: User-facing features have Gherkin specifications
4. **Isolation**: Each test is independent and can run in any order
5. **Fast Feedback**: Tests run quickly to enable rapid iteration

## Test Types

### Unit Tests

**Purpose**: Test individual functions and methods in isolation

**Location**: `tests/unit/`

**Example**:
```go
func TestPolicyThresholdEvaluation(t *testing.T) {
    policy := &models.Policy{
        Thresholds: map[string]float64{"toxicity": 0.8},
    }

    score := 0.9
    if score <= policy.Thresholds["toxicity"] {
        t.Errorf("Expected score %f to exceed threshold %f",
            score, policy.Thresholds["toxicity"])
    }
}
```

**When to write**:
- Testing business logic
- Testing utility functions
- Testing data transformations
- Testing validators

### Integration Tests

**Purpose**: Test interactions between components

**Location**: `tests/integration/`

**Example**:
```go
func TestModerationEndToEnd(t *testing.T) {
    db, cleanup := helpers.SetupTestDB(t)
    defer cleanup()

    redis, redisCleanup := helpers.SetupTestRedis(t)
    defer redisCleanup()

    // Test full moderation flow
    submission := createSubmission(t, db, "Test content")
    decision := moderate(t, submission)
    evidence := verifyEvidence(t, db, decision.ID)

    assert.NotNil(t, evidence)
}
```

**When to write**:
- Testing database interactions
- Testing cache behavior
- Testing API integrations
- Testing service-to-service communication

### Control-Driven Tests (CDD)

**Purpose**: Verify governance controls are implemented

**Location**: `tests/cdd/`

**Example**:
```go
func TestMOD001_AutomatedClassification(t *testing.T) {
    // Control: MOD-001 ensures all text is classified
    db, cleanup := helpers.SetupTestDB(t)
    defer cleanup()

    // Create submission
    submissionID := db.CreateTestSubmission(t, "content", "hash")

    // Verify classification happened
    var count int
    db.Pool.QueryRow(ctx,
        "SELECT COUNT(*) FROM moderation_decisions WHERE submission_id = $1",
        submissionID).Scan(&count)

    if count != 1 {
        t.Error("MOD-001 violated: text not classified")
    }
}
```

**When to write**:
- New control is added to `control-registry.yaml`
- Control implementation changes
- Compliance audit requires verification

### Behavior-Driven Tests (BDD)

**Purpose**: Specify user-facing behavior in plain language

**Location**: `tests/bdd/features/`

**Example**:
```gherkin
Feature: Text Moderation
  Scenario: Toxic text is blocked
    Given a text submission with high toxicity scores
    When the text is submitted for moderation
    Then the moderation decision should be "block"
    And an evidence record should be created
```

**When to write**:
- New user-facing feature
- User workflow changes
- Acceptance criteria from product requirements

### End-to-End Tests (E2E)

**Purpose**: Test complete user workflows through the UI

**Location**: `tests/e2e/specs/`

**Example**:
```typescript
test('moderator can review and approve decision', async ({ moderatorPage }) => {
    await moderatorPage.goto('/moderation/queue');

    const firstItem = moderatorPage.locator('[data-testid="queue-item"]').first();
    await firstItem.locator('button:has-text("Approve")').click();

    await moderatorPage.locator('textarea[name="rationale"]').fill('Confirmed');
    await moderatorPage.locator('button:has-text("Submit")').click();

    await expect(moderatorPage.locator('[data-testid="success-message"]'))
        .toBeVisible();
});
```

**When to write**:
- Critical user workflows
- UI interactions
- Multi-step processes
- Cross-browser compatibility needed

## Writing Good Tests

### Test Structure

Follow the **Arrange-Act-Assert** pattern:

```go
func TestExample(t *testing.T) {
    // Arrange: Set up test data and dependencies
    db, cleanup := helpers.SetupTestDB(t)
    defer cleanup()

    submission := createTestSubmission(t, db)

    // Act: Execute the code under test
    result := moderateContent(submission)

    // Assert: Verify the outcome
    if result.Action != models.ActionBlock {
        t.Errorf("Expected block, got %s", result.Action)
    }
}
```

### Test Naming

Use descriptive names that explain what is being tested:

```go
// Good
func TestModerationReturnsBlockForHighToxicityScore(t *testing.T) {}
func TestPolicyThresholdExceedanceTriggersAction(t *testing.T) {}

// Bad
func TestModeration(t *testing.T) {}
func TestPolicy(t *testing.T) {}
```

### Assertions

Make assertions specific and meaningful:

```go
// Good
if score < 0.0 || score > 1.0 {
    t.Errorf("Score %f out of valid range [0.0, 1.0]", score)
}

// Bad
if score != 0.5 {
    t.Error("Wrong score")
}
```

### Test Data

Use fixtures for consistent test data:

```go
// Load from fixture
policies := loadFixture(t, "policies.json")

// Create minimal test data
user := &models.User{
    Email: "test@example.com",
    Role:  models.RoleAdmin,
}
```

### Cleanup

Always clean up resources:

```go
func TestExample(t *testing.T) {
    db, cleanup := helpers.SetupTestDB(t)
    defer cleanup() // Runs even if test fails

    // Test code
}
```

## Test Helpers

### Database Helpers

```go
// Setup
db, cleanup := helpers.SetupTestDB(t)
defer cleanup()

// Create test data
submissionID := db.CreateTestSubmission(t, "content", "hash")

// Get test user
userID, apiKey := db.GetTestUser(t, "admin@civitas.test")

// Cleanup specific tables
db.Cleanup(t)
```

### Redis Helpers

```go
// Setup
redis, cleanup := helpers.SetupTestRedis(t)
defer cleanup()

// Set cache
redis.SetTestCache(t, "key", "value", 5*time.Minute)

// Get cache
value := redis.GetTestCache(t, "key")

// Check existence
exists := redis.KeyExists(t, "key")
```

### Mock Server

```go
// Setup mock HuggingFace API
mock, cleanup := helpers.SetupMockServer(t)
defer cleanup()

// Use mock URL
apiURL := mock.URL()

// Add custom response
mock.AddCustomResponse("custom", ModerationResponse{
    Input: "custom input",
    Response: []byte(`[{"label": "toxic", "score": 0.9}]`),
})
```

## Running Tests

### Local Development

```bash
# Run all tests
make test

# Run specific test suite
make test-unit
make test-integration
make test-cdd
make test-e2e

# Run with coverage
make test-coverage

# Run specific test
go test ./tests/cdd -run TestMOD001
```

### Docker Environment

```bash
# Start test environment
make test-up

# Run tests in Docker
make test-cdd-docker
make test-e2e-docker

# View logs
make test-logs

# Stop test environment
make test-down
```

### Continuous Integration

Tests run automatically on:
- Push to `main` or `develop`
- Pull request creation
- Manual workflow dispatch

## Debugging Tests

### Go Tests

```bash
# Run with verbose output
go test -v ./tests/cdd/...

# Run specific test
go test -run TestMOD001 ./tests/cdd/...

# Enable race detection
go test -race ./tests/cdd/...

# Print test logs
go test -v -args -test.v
```

### E2E Tests

```bash
# Run with UI mode
cd tests/e2e
npm run test:ui

# Run in headed mode (see browser)
npm run test:headed

# Debug specific test
npm run test:debug -- specs/moderation-demo.spec.ts

# Update snapshots
npm test -- --update-snapshots
```

### Database Inspection

```bash
# Connect to test database
psql postgres://postgres:postgres@localhost:5433/civitas_test

# View evidence records
SELECT * FROM evidence_records ORDER BY created_at DESC LIMIT 10;

# Check test data
SELECT * FROM users WHERE email LIKE '%@civitas.test';
```

## Common Issues

### Database Connection Errors

**Problem**: `pq: database "civitas_test" does not exist`

**Solution**:
```bash
# Ensure test database is running
make test-up

# Check database status
docker compose -f docker-compose.test.yml ps postgres
```

### Redis Connection Errors

**Problem**: `dial tcp 127.0.0.1:6380: connect: connection refused`

**Solution**:
```bash
# Start Redis
docker compose -f docker-compose.test.yml up redis -d

# Verify Redis is running
redis-cli -p 6380 ping
```

### E2E Test Timeouts

**Problem**: Tests timeout waiting for services

**Solution**:
```bash
# Increase timeout in playwright.config.ts
timeout: 60 * 1000 // 60 seconds

# Check services are healthy
curl http://localhost:3001/health
curl http://localhost:8081/health
```

### Flaky Tests

**Problem**: Tests pass sometimes, fail other times

**Solution**:
- Add explicit waits instead of fixed timeouts
- Ensure proper cleanup between tests
- Check for race conditions
- Verify test isolation

## Best Practices

1. **Keep tests fast**: Unit tests < 100ms, Integration < 1s, E2E < 30s
2. **Test behavior, not implementation**: Focus on what, not how
3. **Use meaningful assertions**: Make failures easy to understand
4. **Maintain test fixtures**: Keep fixtures realistic and up-to-date
5. **Document complex tests**: Add comments explaining why, not what
6. **Run tests before committing**: Ensure all tests pass locally
7. **Keep tests independent**: No test should depend on another
8. **Use test helpers**: Reduce duplication with shared helpers

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Playwright Documentation](https://playwright.dev)
- [Gherkin Reference](https://cucumber.io/docs/gherkin/reference/)
- [Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
