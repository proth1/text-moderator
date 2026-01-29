# Text Moderator - AI Content Moderation Platform

## Repository Structure

```
text-moderator/
├── .claude/              # AI integration (agents, config, scripts)
│   ├── agents/           # SubAgent definitions
│   ├── config/           # PM tool & Helix configuration
│   └── scripts/          # Automation scripts
├── apps/                 # Frontend applications
├── controls/             # Moderation control definitions
├── docs/                 # Documentation
├── internal/             # Go internal packages
├── schemas/              # Data schemas
├── scripts/              # Build/deploy scripts
├── services/             # Go microservices
└── tests/                # Test suites
```

## Tech Stack

- **Language**: Go
- **Architecture**: Microservices
- **Build**: Makefile, Docker Compose

---

## Project Management

**Tool**: GitHub Issues
**Repository**: `proth1/text-moderator`
**Work Item Format**: `#XXX`
**Branch Format**: `feature/{id}-{description}`

**Rules:**
- Use `github-issues-manager` subagent for ALL work item operations
- NEVER use `gh issue` CLI commands directly
- Use `Closes #XXX` in PR descriptions for auto-linking
- All stories MUST have Gherkin acceptance criteria (BDD)

---

## Helix Platform Integration

This project is connected to the Helix AI-Powered SDLC Platform.

**Tenant**: `text-moderator`
**API**: `https://helix-api.agentic-innovations.com`
**Config**: `.claude/config/helix-connection.yaml`

### Available Services

| Service | Purpose |
|---------|---------|
| Context Intelligence | Codebase analysis, pattern suggestions, risk identification |
| Critical Thinking Enforcer | Pre-task analysis, post-task validation, PR analysis |
| Tenant Management | Tenant configuration and status |

### Using Helix Services

```bash
# Check platform health
./.claude/scripts/helix-client.sh health

# Analyze context for a task
./.claude/scripts/helix-client.sh analyze "services/moderation/*.go" "content filtering"

# Pre-task critical thinking
./.claude/scripts/helix-client.sh think "Should we use regex or ML for spam detection?"
```

### SubAgents for Helix

```
> Use the helix-context-advisor subagent to analyze context for [task]
> Use the helix-critical-thinking subagent to evaluate [decision]
```

### Environment Setup

Source `.env.helix` or set these environment variables:
- `HELIX_PLATFORM_URL` - API base URL
- `HELIX_TENANT_SLUG` - Tenant identifier
- `HELIX_JWT_SECRET` - JWT signing secret
- `HELIX_ENABLED` - Enable/disable integration

---

## SubAgents

| Agent | Purpose |
|-------|---------|
| `github-issues-manager` | Work item management (GitHub Issues) |
| `helix-context-advisor` | Codebase analysis via Helix Context Intelligence |
| `helix-critical-thinking` | Decision support via Helix Critical Thinking Enforcer |

---

## Git Workflow

```bash
# BEFORE any branch:
git branch --show-current  # MUST be "main"
git status                 # MUST be clean
git pull origin main       # MUST be current

# Then create:
git checkout -b feature/123-description
```

**Rules:**
- Branch from main only
- All changes require Work Item -> Branch -> PR -> Approval
- Max 5 files per PR unless necessary
- Include `Closes #XXX` in PR description

---

## Development

```bash
# Build
make build

# Test
make test

# Run locally
docker-compose up
```

---

## File Creation Rules

- **DEFAULT**: Upgrade existing files in-place
- **NEVER**: Create `_v2`, `_new`, `_enhanced` variants
- **NO markdown** without explicit user approval
- **PR content** belongs IN the PR description, not separate files
