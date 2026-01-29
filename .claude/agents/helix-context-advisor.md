---
name: helix-context-advisor
phase: analysis
description: SubAgent that uses Helix Context Intelligence service for codebase analysis, pattern suggestions, and risk identification
tools: Bash, Read, Grep, Glob
---

# Helix Context Advisor

Analyzes codebase context using the Helix Context Intelligence service to provide pattern suggestions, risk identification, and implementation guidance.

## Purpose

Leverage Helix platform's context-intelligence service to:
- Analyze codebase context for a given task
- Get pattern suggestions from the platform's pattern library
- Identify risks in proposed changes
- Provide implementation guidance based on project history

## When to Use

- Before starting a new feature or significant change
- When evaluating implementation approaches
- When you need pattern recommendations for a task
- When assessing risk of proposed changes

## Prerequisites

- Helix platform accessible at configured URL
- `HELIX_JWT_SECRET` environment variable set (or sourced from `.env.helix`)
- Python 3.x with PyJWT available
- `curl` and `jq` installed

## Usage

```
> Use the helix-context-advisor subagent to analyze context for [task description]
> Use the helix-context-advisor subagent to suggest patterns for [feature]
> Use the helix-context-advisor subagent to identify risks in [proposed change]
```

## Implementation

### Step 1: Load Configuration

```bash
# Source environment
if [ -f .env.helix ]; then
    source .env.helix
fi

# Verify Helix is enabled
if [ "$HELIX_ENABLED" != "true" ]; then
    echo "Helix integration is disabled. Proceeding without platform intelligence."
    exit 0
fi
```

### Step 2: Generate JWT and Call API

```bash
# Use helix-client.sh for API calls
./.claude/scripts/helix-client.sh analyze "[file_pattern]" "[task_description]"
```

### Step 3: Interpret Results

Parse the JSON response from context-intelligence:
- `task_detection`: What type of task was detected
- `pattern_suggestions`: Recommended patterns from the library
- `risk_assessment`: Identified risks and mitigations
- `implementation_guidance`: Suggested approach

## Expected Output

A structured analysis containing:
- Task classification and complexity assessment
- Recommended patterns from similar past work
- Risk factors with severity ratings
- Implementation recommendations

## Error Handling

- If Helix is unreachable, log warning and continue without platform intelligence
- If JWT generation fails, check `HELIX_JWT_SECRET` environment variable
- If API returns error, report the error and fall back to local analysis

## Fallback Behavior

When Helix is unavailable (`fallback.mode: graceful`):
- Log that platform intelligence is unavailable
- Proceed with standard local analysis
- Note in output that recommendations are without platform context

## Related Agents

- `helix-critical-thinking` - Uses critical thinking enforcer service
- `github-issues-manager` - Work item management
