#!/bin/bash

# Verification script for Civitas AI Text Moderator setup

set -e

echo "=========================================="
echo "Civitas AI Text Moderator - Setup Verification"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check functions
check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} $1"
        return 0
    else
        echo -e "${RED}✗${NC} $1 (missing)"
        return 1
    fi
}

check_dir() {
    if [ -d "$1" ]; then
        echo -e "${GREEN}✓${NC} $1/"
        return 0
    else
        echo -e "${RED}✗${NC} $1/ (missing)"
        return 1
    fi
}

# Track errors
ERRORS=0

echo "Checking Project Structure..."
echo "------------------------------"

# Core files
check_file "go.mod" || ((ERRORS++))
check_file ".env.example" || ((ERRORS++))
check_file "README.md" || ((ERRORS++))
check_file "QUICKSTART.md" || ((ERRORS++))
check_file "IMPLEMENTATION_STATUS.md" || ((ERRORS++))

echo ""
echo "Checking Internal Packages..."
echo "------------------------------"

# Internal packages
check_dir "internal/models" || ((ERRORS++))
check_file "internal/models/models.go" || ((ERRORS++))

check_dir "internal/database" || ((ERRORS++))
check_file "internal/database/postgres.go" || ((ERRORS++))

check_dir "internal/cache" || ((ERRORS++))
check_file "internal/cache/redis.go" || ((ERRORS++))

check_dir "internal/config" || ((ERRORS++))
check_file "internal/config/config.go" || ((ERRORS++))

check_dir "internal/middleware" || ((ERRORS++))
check_file "internal/middleware/auth.go" || ((ERRORS++))
check_file "internal/middleware/cors.go" || ((ERRORS++))
check_file "internal/middleware/logging.go" || ((ERRORS++))

check_dir "internal/evidence" || ((ERRORS++))
check_file "internal/evidence/writer.go" || ((ERRORS++))

echo ""
echo "Checking Database Migrations..."
echo "--------------------------------"

check_dir "internal/database/migrations" || ((ERRORS++))
check_file "internal/database/migrations/001_create_users.up.sql" || ((ERRORS++))
check_file "internal/database/migrations/001_create_users.down.sql" || ((ERRORS++))
check_file "internal/database/migrations/002_create_policies.up.sql" || ((ERRORS++))
check_file "internal/database/migrations/002_create_policies.down.sql" || ((ERRORS++))
check_file "internal/database/migrations/003_create_submissions.up.sql" || ((ERRORS++))
check_file "internal/database/migrations/003_create_submissions.down.sql" || ((ERRORS++))
check_file "internal/database/migrations/004_create_decisions.up.sql" || ((ERRORS++))
check_file "internal/database/migrations/004_create_decisions.down.sql" || ((ERRORS++))
check_file "internal/database/migrations/005_create_review_actions.up.sql" || ((ERRORS++))
check_file "internal/database/migrations/005_create_review_actions.down.sql" || ((ERRORS++))
check_file "internal/database/migrations/006_create_evidence.up.sql" || ((ERRORS++))
check_file "internal/database/migrations/006_create_evidence.down.sql" || ((ERRORS++))

echo ""
echo "Checking Services..."
echo "--------------------"

# Gateway
check_dir "services/gateway" || ((ERRORS++))
check_file "services/gateway/main.go" || ((ERRORS++))

# Moderation
check_dir "services/moderation" || ((ERRORS++))
check_file "services/moderation/main.go" || ((ERRORS++))
check_dir "services/moderation/client" || ((ERRORS++))
check_file "services/moderation/client/huggingface.go" || ((ERRORS++))

# Policy Engine
check_dir "services/policy-engine" || ((ERRORS++))
check_file "services/policy-engine/main.go" || ((ERRORS++))
check_dir "services/policy-engine/engine" || ((ERRORS++))
check_file "services/policy-engine/engine/evaluator.go" || ((ERRORS++))

# Review
check_dir "services/review" || ((ERRORS++))
check_file "services/review/main.go" || ((ERRORS++))

echo ""
echo "=========================================="
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}All checks passed!${NC} ✓"
    echo ""
    echo "Your project structure is complete and ready."
    echo ""
    echo "Next steps:"
    echo "1. Copy .env.example to .env and configure"
    echo "2. Start PostgreSQL and Redis: docker-compose up -d"
    echo "3. Run migrations: See QUICKSTART.md"
    echo "4. Start services: go run services/<service>/main.go"
    echo ""
    echo "See QUICKSTART.md for detailed instructions."
else
    echo -e "${RED}${ERRORS} checks failed!${NC} ✗"
    echo ""
    echo "Please review the missing files/directories above."
fi
echo "=========================================="
