.PHONY: up down build test migrate test-all test-unit test-integration test-cdd test-bdd test-e2e test-coverage

# Development
up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

# Testing - Development (no Docker)
test:
	go test ./... -v -cover

test-unit:
	go test ./tests/unit/... -v -cover

test-integration:
	go test ./tests/integration/... -v -cover

test-cdd:
	TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/civitas_test?sslmode=disable" \
	TEST_REDIS_ADDR="localhost:6380" \
	go test ./tests/cdd/... -v -cover

# Testing - Docker-based
test-up:
	docker compose -f docker-compose.yml -f docker-compose.test.yml up -d

test-down:
	docker compose -f docker-compose.yml -f docker-compose.test.yml down -v

test-logs:
	docker compose -f docker-compose.yml -f docker-compose.test.yml logs -f

test-bdd-docker:
	docker compose -f docker-compose.yml -f docker-compose.test.yml run --rm bdd-tests

test-cdd-docker:
	docker compose -f docker-compose.yml -f docker-compose.test.yml run --rm cdd-tests

test-e2e-docker:
	docker compose -f docker-compose.yml -f docker-compose.test.yml run --rm e2e-tests

# E2E Tests (local)
test-e2e:
	cd tests/e2e && npm test

test-e2e-ui:
	cd tests/e2e && npm run test:ui

test-e2e-headed:
	cd tests/e2e && npm run test:headed

# BDD Feature Tests
test-bdd:
	@echo "BDD step definitions not yet implemented"
	@echo "Run 'make test-bdd-docker' when implementation is ready"

# Test All
test-all: test-unit test-integration test-cdd test-e2e

test-all-docker: test-up test-cdd-docker test-bdd-docker test-e2e-docker test-down

# Coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-func:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

# Database migrations
migrate-up:
	migrate -path internal/database/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path internal/database/migrations -database "$(DATABASE_URL)" down

# Linting
lint:
	golangci-lint run ./...

# Development servers
dev-gateway:
	go run ./services/gateway

dev-moderation:
	go run ./services/moderation

dev-policy:
	go run ./services/policy-engine

dev-review:
	go run ./services/review

dev-web:
	cd apps/web && npm run dev

# Clean
clean:
	rm -rf coverage.out coverage.html
	rm -rf tests/e2e/test-results tests/e2e/playwright-report
	docker compose -f docker-compose.yml -f docker-compose.test.yml down -v

# Help
help:
	@echo "Civitas AI Text Moderator - Makefile Commands"
	@echo ""
	@echo "Development:"
	@echo "  make up              - Start development environment"
	@echo "  make down            - Stop development environment"
	@echo "  make build           - Build all services"
	@echo "  make logs            - View logs"
	@echo ""
	@echo "Testing (Local):"
	@echo "  make test            - Run all Go tests"
	@echo "  make test-unit       - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-cdd        - Run control-driven tests"
	@echo "  make test-e2e        - Run E2E Playwright tests"
	@echo "  make test-all        - Run all test suites"
	@echo ""
	@echo "Testing (Docker):"
	@echo "  make test-up         - Start test environment"
	@echo "  make test-down       - Stop test environment"
	@echo "  make test-cdd-docker - Run CDD tests in Docker"
	@echo "  make test-bdd-docker - Run BDD tests in Docker"
	@echo "  make test-e2e-docker - Run E2E tests in Docker"
	@echo "  make test-all-docker - Run all tests in Docker"
	@echo ""
	@echo "Coverage:"
	@echo "  make test-coverage   - Generate HTML coverage report"
	@echo "  make test-coverage-func - Show coverage by function"
	@echo ""
	@echo "Utilities:"
	@echo "  make migrate-up      - Run database migrations"
	@echo "  make migrate-down    - Rollback database migrations"
	@echo "  make lint            - Run linters"
	@echo "  make clean           - Clean test artifacts"
	@echo "  make help            - Show this help message"
