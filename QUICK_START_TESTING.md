# Quick Start: Testing

## 🚀 5-Minute Setup

```bash
# 1. Install E2E dependencies
cd tests/e2e && npm install && cd ../..

# 2. Start test environment
make test-up

# 3. Run your first tests
make test-cdd
```

## 📋 Essential Commands

| Command | What It Does | Time |
|---------|--------------|------|
| `make test-up` | Start test environment | 30s |
| `make test-cdd` | Run control tests | 5s |
| `make test-e2e` | Run browser tests | 30s |
| `make test-all` | Run all tests | 1m |
| `make test-down` | Stop test environment | 10s |
| `make test-logs` | View service logs | - |
| `make test-coverage` | Generate coverage report | 10s |

## 📂 Key Files

| File | Purpose |
|------|---------|
| `tests/README.md` | Full documentation |
| `tests/TESTING_GUIDE.md` | How to write tests |
| `tests/fixtures/seed-data.sql` | Test data |
| `docker-compose.test.yml` | Test environment |
| `.github/workflows/tests.yml` | CI/CD pipeline |

## 🧪 Test Types

### 1. Control Tests (CDD)
```bash
make test-cdd              # Run locally
make test-cdd-docker       # Run in Docker
```
**What**: Verifies governance controls (MOD-001, GOV-002, etc.)
**Where**: `tests/cdd/controls_test.go`

### 2. E2E Tests
```bash
make test-e2e              # Run all
make test-e2e-ui           # Interactive mode
make test-e2e-headed       # See browser
```
**What**: Full user workflows in real browsers
**Where**: `tests/e2e/specs/*.spec.ts`

### 3. BDD Scenarios
```bash
make test-bdd              # (Step defs not yet implemented)
```
**What**: User stories in Gherkin format
**Where**: `tests/bdd/features/*.feature`

## 🔍 Debugging

### View Logs
```bash
make test-logs                              # All services
docker compose -f docker-compose.test.yml logs postgres
```

### Connect to Database
```bash
psql postgres://postgres:postgres@localhost:5433/civitas_test
```

### Test Redis
```bash
redis-cli -p 6380
```

### Run Single Test
```bash
go test ./tests/cdd -run TestMOD001
```

## 📊 Test Data

### Users (in database)
- `admin@civitas.test` - Admin role
- `moderator@civitas.test` - Moderator role
- `viewer@civitas.test` - Viewer role

### API Keys
- Admin: `tk_admin_test_key_001`
- Moderator: `tk_mod_test_key_002`
- Viewer: `tk_viewer_test_key_003`

### Policies
- Standard Community Guidelines (Published)
- Youth Safe Mode (Published)
- Relaxed Forum Policy (Draft)

## 🆘 Common Issues

### Services Won't Start
```bash
# Clean and restart
make test-down
docker system prune -f
make test-up
```

### Tests Fail
```bash
# Check services are healthy
docker compose -f docker-compose.test.yml ps

# View specific service logs
make test-logs
```

### E2E Tests Timeout
```bash
# Verify frontend is accessible
curl http://localhost:3001/health

# Check API gateway
curl http://localhost:8081/health
```

## 📚 Learn More

- **Full Guide**: `tests/README.md`
- **Writing Tests**: `tests/TESTING_GUIDE.md`
- **All Files**: `tests/FILES_CREATED.md`
- **Summary**: `TEST_INFRASTRUCTURE_SUMMARY.md`

## ✅ Verification

```bash
# Verify setup
./tests/verify-setup.sh

# Expected output:
# ✓ All files present! Setup is complete.
```

## 🎯 Quick Test Run

```bash
# Complete test run (5 minutes)
make test-up && \
make test-cdd && \
make test-e2e && \
make test-coverage && \
make test-down
```

## 📈 Coverage Goals

- **Overall**: 80% minimum
- **Critical Controls**: 100%
- **Automated Ratio**: 90%

## 🏗️ CI/CD

Tests run automatically on:
- ✅ Push to main/develop
- ✅ Pull requests
- ✅ Manual trigger

View results: `.github/workflows/tests.yml`

---

**Need Help?** See `tests/TESTING_GUIDE.md` for detailed troubleshooting.
