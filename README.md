# Civitas AI Text Moderator

Enterprise-grade AI-powered text moderation system with policy-driven decision making, human review workflows, and immutable audit trails for compliance (SOC 2, ISO 27001, GDPR).

## Architecture

This is a Go monorepo with multiple microservices:

- **Gateway**: API gateway routing requests to backend services (port 8080)
- **Moderation**: HuggingFace integration for AI-powered text classification (port 8081)
- **Policy Engine**: Policy management and evaluation (port 8082)
- **Review**: Human review workflows and evidence management (port 8083)

### Technology Stack

- **Language**: Go 1.22+
- **Web Framework**: Gin
- **Database**: PostgreSQL (with pgx driver)
- **Cache**: Redis
- **AI Model**: HuggingFace Inference API (toxic-bert)
- **Logging**: Zap (structured JSON logging)

## Project Structure

```
/services
  /gateway/          # API gateway
  /moderation/       # Moderation service with HuggingFace client
  /policy-engine/    # Policy management and evaluation
  /review/           # Human review and evidence management
/internal
  /models/           # Shared domain models
  /database/         # PostgreSQL connection and migrations
  /cache/            # Redis client wrapper
  /config/           # Configuration management
  /middleware/       # Shared middleware (auth, CORS, logging)
  /evidence/         # Evidence generation for compliance
```

## Compliance Controls

The system implements controls for multiple compliance frameworks:

- **MOD-001**: AI model integration, decision tracking, and traceability
- **POL-001**: Policy-driven decision making with versioning
- **GOV-002**: Role-based access control and human oversight
- **AUD-001**: Immutable evidence generation for audit trails

## Getting Started

### Prerequisites

- Go 1.22 or later
- PostgreSQL 14+
- Redis 7+
- HuggingFace API key

### Configuration

Copy the example environment file and configure:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/text_moderator?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0

# HuggingFace
HUGGINGFACE_API_KEY=your_api_key_here
HUGGINGFACE_MODEL_URL=https://api-inference.huggingface.co/models/unitary/toxic-bert

# Service Ports
GATEWAY_PORT=8080
MODERATION_PORT=8081
POLICY_ENGINE_PORT=8082
REVIEW_PORT=8083

# Application
ENVIRONMENT=development
LOG_LEVEL=info
```

### Database Setup

Run migrations to set up the database schema:

```bash
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path internal/database/migrations -database "${DATABASE_URL}" up
```

### Running Services

Install dependencies:

```bash
go mod download
```

Run each service in a separate terminal:

```bash
# Terminal 1: Gateway
go run services/gateway/main.go

# Terminal 2: Moderation Service
go run services/moderation/main.go

# Terminal 3: Policy Engine Service
go run services/policy-engine/main.go

# Terminal 4: Review Service
go run services/review/main.go
```

## API Endpoints

### Gateway (Port 8080)

All requests are proxied through the gateway at `http://localhost:8080/api/v1`

#### Moderation

- `POST /api/v1/moderate` - Submit text for moderation

#### Policies

- `GET /api/v1/policies` - List policies
- `POST /api/v1/policies` - Create policy (requires admin role)
- `GET /api/v1/policies/:id` - Get policy by ID
- `POST /api/v1/policies/:id/evaluate` - Evaluate scores against policy

#### Reviews

- `GET /api/v1/reviews` - List review queue (requires moderator role)
- `GET /api/v1/reviews/:id` - Get review details
- `POST /api/v1/reviews/:id/action` - Submit review action

#### Evidence

- `GET /api/v1/evidence` - List evidence records (requires admin role)
- `GET /api/v1/evidence/export` - Export evidence (requires admin role)

## Authentication

All API requests (except `/health` endpoints) require authentication via API key.

Include your API key in one of these headers:

- `X-API-Key: your_api_key`
- `Authorization: Bearer your_api_key`

## Example Usage

### Submit text for moderation

```bash
curl -X POST http://localhost:8080/api/v1/moderate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_api_key" \
  -d '{
    "content": "This is some text to moderate",
    "context_metadata": {
      "user_id": "user123",
      "channel": "comments"
    },
    "source": "web_app"
  }'
```

### Create a policy

```bash
curl -X POST http://localhost:8080/api/v1/policies \
  -H "Content-Type: application/json" \
  -H "X-API-Key: admin_api_key" \
  -d '{
    "name": "Strict Policy",
    "thresholds": {
      "toxicity": 0.7,
      "hate": 0.6,
      "harassment": 0.7
    },
    "actions": {
      "toxicity": "warn",
      "hate": "block",
      "harassment": "escalate"
    }
  }'
```

## Development

### Code Organization

- Each service is independently runnable
- Shared code lives in `/internal`
- Database migrations are versioned in `/internal/database/migrations`
- All services use structured logging with correlation IDs

### Testing

```bash
go test ./...
```

## License

Proprietary
