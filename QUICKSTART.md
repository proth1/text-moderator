# Quick Start Guide

Get the Civitas AI Text Moderator running in 5 minutes.

## Prerequisites

- Go 1.22 or later installed
- Docker and Docker Compose installed
- HuggingFace API key ([Get one here](https://huggingface.co/settings/tokens))

## Step 1: Configure Environment

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env and add your HuggingFace API key
# HUGGINGFACE_API_KEY=hf_xxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Step 2: Start Database Services

```bash
# Start PostgreSQL and Redis with Docker
docker-compose up -d postgres redis

# Wait for services to be ready (about 10 seconds)
sleep 10
```

## Step 3: Run Database Migrations

```bash
# Install migrate tool if you don't have it
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Load environment variables
source .env

# Run migrations
migrate -path internal/database/migrations -database "${DATABASE_URL}" up
```

You should see:
```
6/u create_evidence (xxx.xxxms)
```

## Step 4: Create Admin User

```bash
# Connect to PostgreSQL
docker exec -it text-moderator-postgres psql -U postgres -d text_moderator

# In the psql prompt, run:
INSERT INTO users (email, api_key, role)
VALUES ('admin@example.com', 'test-api-key-12345', 'admin');

# Exit psql
\q
```

## Step 5: Start All Services

Open 4 terminal windows and run one command in each:

### Terminal 1: Gateway
```bash
go run services/gateway/main.go
```

### Terminal 2: Moderation Service
```bash
go run services/moderation/main.go
```

### Terminal 3: Policy Engine
```bash
go run services/policy-engine/main.go
```

### Terminal 4: Review Service
```bash
go run services/review/main.go
```

All services should start without errors. You should see:
- Gateway: `gateway service listening port=8080`
- Moderation: `moderation service listening port=8081`
- Policy Engine: `policy-engine service listening port=8082`
- Review: `review service listening port=8083`

## Step 6: Test the System

### Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"healthy","service":"gateway","version":"0.1.0"}
```

### Create a Policy

```bash
curl -X POST http://localhost:8080/api/v1/policies \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-12345" \
  -d '{
    "name": "Default Policy",
    "thresholds": {
      "toxicity": 0.7,
      "hate": 0.6,
      "harassment": 0.7,
      "violence": 0.8,
      "profanity": 0.6,
      "sexual_content": 0.8
    },
    "actions": {
      "toxicity": "warn",
      "hate": "block",
      "harassment": "escalate",
      "violence": "block",
      "profanity": "warn",
      "sexual_content": "block"
    }
  }'
```

Save the policy ID from the response.

### Publish the Policy

```bash
# Update policy status to published (direct SQL for now)
docker exec -it text-moderator-postgres psql -U postgres -d text_moderator -c \
  "UPDATE policies SET status = 'published' WHERE name = 'Default Policy';"
```

### Moderate Some Text

```bash
curl -X POST http://localhost:8080/api/v1/moderate \
  -H "Content-Type: application/json" \
  -d '{
    "content": "This is a nice and friendly message!",
    "source": "test"
  }'
```

Expected response (with category scores and action):
```json
{
  "decision_id": "xxx-xxx-xxx",
  "submission_id": "xxx-xxx-xxx",
  "action": "allow",
  "category_scores": {
    "toxicity": 0.05,
    "hate": 0.01,
    "harassment": 0.02,
    ...
  },
  "requires_review": false,
  "policy_applied": "Default Policy",
  "policy_version": 1
}
```

### Test with Toxic Content

```bash
curl -X POST http://localhost:8080/api/v1/moderate \
  -H "Content-Type: application/json" \
  -d '{
    "content": "I hate you, you are terrible!",
    "source": "test"
  }'
```

This should return a higher toxicity score and potentially trigger a warning or block action.

### List Policies

```bash
curl http://localhost:8080/api/v1/policies \
  -H "X-API-Key: test-api-key-12345"
```

### View Evidence Records

```bash
curl http://localhost:8080/api/v1/evidence \
  -H "X-API-Key: test-api-key-12345"
```

## Troubleshooting

### "Connection refused" errors
- Make sure PostgreSQL and Redis are running: `docker-compose ps`
- Check that services are listening on the correct ports: `netstat -an | grep LISTEN | grep -E '8080|8081|8082|8083'`

### "Invalid API key"
- Make sure you created the admin user in Step 4
- Use the exact API key you inserted: `test-api-key-12345`

### "No default policy found"
- Make sure you created and published a policy (Steps 6.2 and 6.3)
- Check policy status: `docker exec -it text-moderator-postgres psql -U postgres -d text_moderator -c "SELECT name, status FROM policies;"`

### HuggingFace API errors
- Verify your API key is correct in `.env`
- Check HuggingFace API status: https://status.huggingface.co/
- The model might be loading (first request can take 30-60 seconds)

### Migration errors
- Drop and recreate database: `docker-compose down -v && docker-compose up -d`
- Re-run migrations

## Next Steps

1. **Explore the API**: See the README.md for all available endpoints
2. **Create more policies**: Experiment with different thresholds and actions
3. **Test review workflow**: Create decisions that require human review (action: "escalate")
4. **Check evidence**: View the immutable audit trail in the evidence table
5. **Build the frontend**: Connect a React/Vue app to the API gateway

## Clean Up

To stop all services:

```bash
# Stop Go services: Press Ctrl+C in each terminal

# Stop and remove database containers
docker-compose down

# Remove all data (WARNING: This deletes your database)
docker-compose down -v
```

## Production Considerations

Before deploying to production:

1. Change all default passwords and API keys
2. Enable TLS/HTTPS
3. Set up proper database backups
4. Configure log aggregation
5. Set up monitoring and alerting
6. Review and harden security settings
7. Implement rate limiting
8. Set up CI/CD pipelines

For production deployment guidance, see the full README.md.
