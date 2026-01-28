#!/bin/bash
# Verification script to check test infrastructure setup

echo "=========================================="
echo "  Test Infrastructure Verification"
echo "=========================================="
echo ""

ERRORS=0
WARNINGS=0

check_file() {
    if [ -f "$1" ]; then
        echo "✓ $1"
    else
        echo "✗ MISSING: $1"
        ERRORS=$((ERRORS + 1))
    fi
}

check_dir() {
    if [ -d "$1" ]; then
        echo "✓ $1/"
    else
        echo "✗ MISSING: $1/"
        ERRORS=$((ERRORS + 1))
    fi
}

echo "Checking directories..."
check_dir "tests/fixtures"
check_dir "tests/helpers"
check_dir "tests/bdd/features"
check_dir "tests/cdd"
check_dir "tests/e2e/specs"
check_dir "tests/e2e/fixtures"
echo ""

echo "Checking fixture files..."
check_file "tests/fixtures/seed-data.sql"
check_file "tests/fixtures/moderation_responses.json"
check_file "tests/fixtures/policies.json"
echo ""

echo "Checking helper files..."
check_file "tests/helpers/testdb.go"
check_file "tests/helpers/mockserver.go"
check_file "tests/helpers/testredis.go"
check_file "tests/helpers/mock-api-server.js"
echo ""

echo "Checking BDD feature files..."
check_file "tests/bdd/features/moderation.feature"
check_file "tests/bdd/features/policy_engine.feature"
check_file "tests/bdd/features/review_workflow.feature"
check_file "tests/bdd/features/evidence.feature"
check_file "tests/bdd/features/admin.feature"
echo ""

echo "Checking CDD files..."
check_file "tests/cdd/controls_test.go"
check_file "tests/cdd/control-registry.yaml"
echo ""

echo "Checking E2E files..."
check_file "tests/e2e/package.json"
check_file "tests/e2e/playwright.config.ts"
check_file "tests/e2e/tsconfig.json"
check_file "tests/e2e/Dockerfile.e2e"
check_file "tests/e2e/fixtures/auth.ts"
check_file "tests/e2e/specs/moderation-demo.spec.ts"
check_file "tests/e2e/specs/moderator-queue.spec.ts"
check_file "tests/e2e/specs/policy-management.spec.ts"
check_file "tests/e2e/specs/audit-log.spec.ts"
echo ""

echo "Checking documentation files..."
check_file "tests/README.md"
check_file "tests/TESTING_GUIDE.md"
check_file "tests/FILES_CREATED.md"
check_file "tests/.env.test"
check_file "tests/setup-tests.sh"
echo ""

echo "Checking configuration files..."
check_file "docker-compose.test.yml"
check_file "Makefile"
check_file ".github/workflows/tests.yml"
echo ""

echo "=========================================="
if [ $ERRORS -eq 0 ]; then
    echo "✓ All files present! Setup is complete."
else
    echo "✗ $ERRORS file(s) missing. Please check the output above."
fi

if [ $WARNINGS -gt 0 ]; then
    echo "⚠ $WARNINGS warning(s) found."
fi
echo "=========================================="

exit $ERRORS
