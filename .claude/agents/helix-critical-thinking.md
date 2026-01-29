---
name: helix-critical-thinking
phase: analysis
description: SubAgent that uses Helix Critical Thinking Enforcer for pre-task analysis, post-task validation, and architecture decision support
tools: Bash, Read, Grep, Glob
---

# Helix Critical Thinking

Applies rigorous critical thinking analysis using the Helix Critical Thinking Enforcer service for pre-task validation, post-task review, and architecture decision support.

## Purpose

Leverage Helix platform's critical-thinking-enforcer service to:
- Perform pre-task analysis before starting work
- Validate completed work with post-task checks
- Support architecture decisions with structured reasoning
- Analyze PRs for quality and completeness

## When to Use

- Before starting any significant implementation task
- After completing a feature to validate the approach
- When making architecture or technology decisions
- During PR review for structured quality analysis

## Prerequisites

- Helix platform accessible at configured URL
- `HELIX_JWT_SECRET` environment variable set (or sourced from `.env.helix`)
- Python 3.x with PyJWT available
- `curl` and `jq` installed

## Usage

```
> Use the helix-critical-thinking subagent to evaluate [decision or task]
> Use the helix-critical-thinking subagent to pre-check [task description]
> Use the helix-critical-thinking subagent to validate [completed work]
> Use the helix-critical-thinking subagent to analyze PR [description]
```

## Implementation

### Pre-Task Analysis

```bash
# Call pre-task endpoint
./.claude/scripts/helix-client.sh think "[task description]"
```

This returns:
- Assumptions to validate
- Risks to consider
- Questions to answer before proceeding
- Recommended approach with rationale

### Post-Task Validation

```bash
# Call post-task endpoint
./.claude/scripts/helix-client.sh post-task "[summary of completed work]"
```

This returns:
- Validation of approach taken
- Identified gaps or concerns
- Quality assessment
- Recommendations for improvement

### PR Analysis

```bash
# Call PR analysis endpoint
./.claude/scripts/helix-client.sh pr-analysis "[PR description and diff summary]"
```

## Expected Output

A structured critical thinking analysis containing:
- Key assumptions identified
- Risk assessment with mitigations
- Alternative approaches considered
- Recommendation with confidence level
- Action items and validation steps

## Error Handling

- If Helix is unreachable, log warning and proceed without platform analysis
- If JWT generation fails, check `HELIX_JWT_SECRET` environment variable
- Fallback mode is graceful - work continues without platform intelligence

## Related Agents

- `helix-context-advisor` - Uses context intelligence service
- `github-issues-manager` - Work item management
