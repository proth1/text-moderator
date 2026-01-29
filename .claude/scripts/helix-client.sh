#!/usr/bin/env bash
# Helix Platform API Client
# Provides authenticated access to Helix platform services for text-moderator tenant.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Save caller's overrides before sourcing .env.helix
_CALLER_HELIX_ENABLED="${HELIX_ENABLED:-}"
_CALLER_HELIX_PLATFORM_URL="${HELIX_PLATFORM_URL:-}"
_CALLER_HELIX_TENANT_SLUG="${HELIX_TENANT_SLUG:-}"
_CALLER_HELIX_JWT_SECRET="${HELIX_JWT_SECRET:-}"

# Load environment defaults from file
if [ -f "$PROJECT_ROOT/.env.helix" ]; then
    source "$PROJECT_ROOT/.env.helix"
fi

# Restore caller overrides (take precedence over .env.helix)
[ -n "$_CALLER_HELIX_ENABLED" ] && HELIX_ENABLED="$_CALLER_HELIX_ENABLED"
[ -n "$_CALLER_HELIX_PLATFORM_URL" ] && HELIX_PLATFORM_URL="$_CALLER_HELIX_PLATFORM_URL"
[ -n "$_CALLER_HELIX_TENANT_SLUG" ] && HELIX_TENANT_SLUG="$_CALLER_HELIX_TENANT_SLUG"
[ -n "$_CALLER_HELIX_JWT_SECRET" ] && HELIX_JWT_SECRET="$_CALLER_HELIX_JWT_SECRET"

# Configuration
PLATFORM_URL="${HELIX_PLATFORM_URL:-https://helix-api.agentic-innovations.com}"
TENANT_SLUG="${HELIX_TENANT_SLUG:-text-moderator}"
JWT_SECRET="${HELIX_JWT_SECRET:-}"
ENABLED="${HELIX_ENABLED:-true}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# --- Helper Functions ---

check_enabled() {
    if [ "$ENABLED" != "true" ]; then
        echo -e "${YELLOW}Helix integration is disabled. Set HELIX_ENABLED=true to enable.${NC}"
        exit 0
    fi
}

check_dependencies() {
    local missing=()
    command -v curl &>/dev/null || missing+=("curl")
    command -v jq &>/dev/null || missing+=("jq")
    command -v python3 &>/dev/null || missing+=("python3")

    if [ ${#missing[@]} -gt 0 ]; then
        echo -e "${RED}Missing dependencies: ${missing[*]}${NC}"
        exit 1
    fi
}

generate_jwt() {
    if [ -z "$JWT_SECRET" ]; then
        echo -e "${RED}ERROR: HELIX_JWT_SECRET not set. Source .env.helix or set the variable.${NC}" >&2
        exit 1
    fi

    # Pass secret and tenant via environment variables (not command line args)
    # to avoid exposure in process list (ps aux)
    HELIX_JWT_SECRET_INTERNAL="$JWT_SECRET" \
    HELIX_TENANT_INTERNAL="$TENANT_SLUG" \
    python3 -c "
import json, hmac, hashlib, base64, time, os

secret = os.environ['HELIX_JWT_SECRET_INTERNAL']
tenant = os.environ['HELIX_TENANT_INTERNAL']

def b64url(data):
    return base64.urlsafe_b64encode(data).rstrip(b'=').decode()

header = b64url(json.dumps({'alg': 'HS256', 'typ': 'JWT'}).encode())
payload = b64url(json.dumps({
    'sub': 'claude-agent',
    'tenant': tenant,
    'iss': 'agentic-sdlc-platform',
    'aud': 'agentic-sdlc-platform',
    'iat': int(time.time()),
    'exp': int(time.time()) + 3600
}).encode())

signing_input = f'{header}.{payload}'.encode()
signature = b64url(hmac.new(
    secret.encode(),
    signing_input,
    hashlib.sha256
).digest())

print(f'{header}.{payload}.{signature}')
"
}

api_call() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"

    local token
    token=$(generate_jwt)

    local args=(
        -s -w "\n%{http_code}"
        -X "$method"
        -H "Authorization: Bearer $token"
        -H "Content-Type: application/json"
        -H "X-Tenant: $TENANT_SLUG"
    )

    if [ -n "$data" ]; then
        args+=(-d "$data")
    fi

    local response
    response=$(curl "${args[@]}" "${PLATFORM_URL}${endpoint}")

    local http_code
    http_code=$(echo "$response" | tail -1)
    local body
    body=$(echo "$response" | sed '$d')

    if [[ "$http_code" -ge 200 && "$http_code" -lt 300 ]]; then
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        echo -e "${RED}API Error (HTTP $http_code): $body${NC}" >&2
        return 1
    fi
}

# --- Commands ---

cmd_health() {
    echo "Checking Helix platform health..."
    echo ""

    # API gateway health (unauthenticated)
    local gw_status
    if gw_status=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 "${PLATFORM_URL}/health" 2>/dev/null); then
        if [ "$gw_status" = "200" ]; then
            echo -e "  ${GREEN}[OK]${NC} api-gateway"
        else
            echo -e "  ${RED}[FAIL]${NC} api-gateway (HTTP $gw_status)"
        fi
    else
        echo -e "  ${RED}[UNREACHABLE]${NC} api-gateway"
        return 1
    fi

    # Authenticated service checks
    local token
    token=$(generate_jwt 2>/dev/null) || {
        echo -e "  ${YELLOW}[SKIP]${NC} Authenticated checks (JWT secret not configured)"
        return 0
    }

    local services=("context-intelligence" "critical-thinking-enforcer" "tenant-management")
    local endpoints=("/api/v1/analysis/full-analysis" "/api/v1/pre-task" "/tenants")

    for i in "${!services[@]}"; do
        local service="${services[$i]}"
        local endpoint="${endpoints[$i]}"

        local status
        if status=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 \
            -H "Authorization: Bearer $token" \
            -H "X-Tenant: $TENANT_SLUG" \
            "${PLATFORM_URL}${endpoint}" 2>/dev/null); then
            # 200, 400, 405, 422 all indicate the service is alive and authenticated
            if [[ "$status" -ge 200 && "$status" -lt 500 ]]; then
                echo -e "  ${GREEN}[OK]${NC} $service"
            else
                echo -e "  ${RED}[FAIL]${NC} $service (HTTP $status)"
            fi
        else
            echo -e "  ${RED}[UNREACHABLE]${NC} $service"
        fi
    done
}

cmd_analyze() {
    local file_pattern="${1:-}"
    local task_description="${2:-}"

    if [ -z "$file_pattern" ]; then
        echo "Usage: helix-client.sh analyze <file_pattern> [task_description]"
        exit 1
    fi

    local payload
    payload=$(jq -n \
        --arg tenant "$TENANT_SLUG" \
        --arg files "$file_pattern" \
        --arg task "$task_description" \
        '{tenant: $tenant, file_pattern: $files, task_description: $task}')

    echo "Analyzing context for: $file_pattern"
    api_call "POST" "/context_intelligence/api/v1/analysis/full-analysis" "$payload"
}

cmd_think() {
    local task_description="${1:-}"

    if [ -z "$task_description" ]; then
        echo "Usage: helix-client.sh think <task_description>"
        exit 1
    fi

    local payload
    payload=$(jq -n \
        --arg tenant "$TENANT_SLUG" \
        --arg task "$task_description" \
        '{tenant: $tenant, task_description: $task}')

    echo "Running pre-task critical thinking analysis..."
    api_call "POST" "/critical_thinking/api/v1/pre-task" "$payload"
}

cmd_post_task() {
    local summary="${1:-}"

    if [ -z "$summary" ]; then
        echo "Usage: helix-client.sh post-task <work_summary>"
        exit 1
    fi

    local payload
    payload=$(jq -n \
        --arg tenant "$TENANT_SLUG" \
        --arg summary "$summary" \
        '{tenant: $tenant, work_summary: $summary}')

    echo "Running post-task validation..."
    api_call "POST" "/critical_thinking/api/v1/post-task" "$payload"
}

cmd_pr_analysis() {
    local description="${1:-}"

    if [ -z "$description" ]; then
        echo "Usage: helix-client.sh pr-analysis <pr_description>"
        exit 1
    fi

    local payload
    payload=$(jq -n \
        --arg tenant "$TENANT_SLUG" \
        --arg desc "$description" \
        '{tenant: $tenant, pr_description: $desc}')

    echo "Running PR analysis..."
    api_call "POST" "/critical_thinking/api/v1/pr-analysis" "$payload"
}

# --- Main ---

check_enabled
check_dependencies

command="${1:-help}"
shift || true

case "$command" in
    health)     cmd_health ;;
    analyze)    cmd_analyze "$@" ;;
    think)      cmd_think "$@" ;;
    post-task)  cmd_post_task "$@" ;;
    pr-analysis) cmd_pr_analysis "$@" ;;
    help|*)
        echo "Helix Platform Client - text-moderator tenant"
        echo ""
        echo "Usage: helix-client.sh <command> [args]"
        echo ""
        echo "Commands:"
        echo "  health                         Check all service health"
        echo "  analyze <files> [task]         Analyze codebase context"
        echo "  think <task>                   Pre-task critical thinking"
        echo "  post-task <summary>            Post-task validation"
        echo "  pr-analysis <description>      Analyze a PR"
        echo "  help                           Show this help"
        echo ""
        echo "Environment:"
        echo "  HELIX_PLATFORM_URL   API base URL (default: https://helix-api.agentic-innovations.com)"
        echo "  HELIX_TENANT_SLUG    Tenant identifier (default: text-moderator)"
        echo "  HELIX_JWT_SECRET     JWT signing secret (required)"
        echo "  HELIX_ENABLED        Enable/disable (default: true)"
        ;;
esac
