#!/bin/bash
# Test Infrastructure Setup Script
# This script sets up the test environment for Civitas AI Text Moderator

set -e  # Exit on error

echo "================================================"
echo "  Civitas AI Text Moderator - Test Setup"
echo "================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}➜ $1${NC}"
}

# Check prerequisites
print_info "Checking prerequisites..."

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi
print_success "Docker found"

# Check Docker Compose
if ! command -v docker compose &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi
print_success "Docker Compose found"

# Check Node.js
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js 20+ first."
    exit 1
fi
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 20 ]; then
    print_error "Node.js version 20+ is required. Current version: $(node -v)"
    exit 1
fi
print_success "Node.js $(node -v) found"

# Check Go (optional, for local tests)
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | cut -d'o' -f2)
    print_success "Go $GO_VERSION found"
else
    print_info "Go not found (optional for local tests)"
fi

echo ""
print_info "Setting up test environment..."

# Navigate to project root
cd "$(dirname "$0")/.."

# Install E2E test dependencies
print_info "Installing E2E test dependencies..."
cd tests/e2e
if [ ! -d "node_modules" ]; then
    npm install
    print_success "E2E dependencies installed"
else
    print_success "E2E dependencies already installed"
fi

# Install Playwright browsers
print_info "Installing Playwright browsers..."
npx playwright install --with-deps chromium firefox
print_success "Playwright browsers installed"

cd ../..

# Create .env.test if it doesn't exist
if [ ! -f "tests/.env.test" ]; then
    print_info "Creating tests/.env.test..."
    cp tests/.env.test.example tests/.env.test 2>/dev/null || true
    print_success ".env.test created"
else
    print_success ".env.test already exists"
fi

# Start test environment
print_info "Starting test environment..."
docker compose -f docker-compose.yml -f docker-compose.test.yml up -d

echo ""
print_info "Waiting for services to be healthy..."
sleep 5

# Check service health
RETRY_COUNT=0
MAX_RETRIES=30

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker compose -f docker-compose.test.yml ps | grep -q "healthy"; then
        print_success "Services are healthy"
        break
    fi

    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo -n "."
    sleep 2

    if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
        echo ""
        print_error "Services failed to become healthy. Check logs with: make test-logs"
        exit 1
    fi
done

echo ""
echo "================================================"
print_success "Test environment setup complete!"
echo "================================================"
echo ""
echo "Available test commands:"
echo "  make test-cdd          - Run control-driven tests"
echo "  make test-e2e          - Run E2E Playwright tests"
echo "  make test-all          - Run all test suites"
echo "  make test-coverage     - Generate coverage report"
echo "  make test-logs         - View test service logs"
echo "  make test-down         - Stop test environment"
echo ""
echo "Quick start:"
echo "  1. Run control tests:    make test-cdd"
echo "  2. Run E2E tests:        make test-e2e"
echo "  3. View test logs:       make test-logs"
echo "  4. Stop environment:     make test-down"
echo ""
echo "For more information, see tests/README.md"
echo ""
