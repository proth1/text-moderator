#!/bin/bash
# Jira API Wrapper for Rival Project
# Provides CLI interface to Jira REST API
# Created to fix broken jira-manager SubAgent

set -euo pipefail

# Jira Configuration - reads from environment variables
JIRA_URL="${JIRA_URL:-https://agentic-sdlc.atlassian.net}"
JIRA_USER="${JIRA_USER:-}"
JIRA_TOKEN="${JIRA_API_TOKEN:-}"

# Validate required environment variables
if [ -z "$JIRA_USER" ] || [ -z "$JIRA_TOKEN" ]; then
    echo "ERROR: Missing required environment variables." >&2
    echo "Please set:" >&2
    echo "  export JIRA_USER='your-email@example.com'" >&2
    echo "  export JIRA_API_TOKEN='your-api-token'" >&2
    echo "" >&2
    echo "Get your API token from: https://id.atlassian.com/manage-profile/security/api-tokens" >&2
    exit 1
fi
API_BASE="${JIRA_URL}/rest/api/3"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

error() { echo -e "${RED}ERROR: $1${NC}" >&2; exit 1; }
success() { echo -e "${GREEN}$1${NC}"; }

api_call() {
    local method="$1" endpoint="$2" data="${3:-}"
    if [ -n "$data" ]; then
        curl -s -X "$method" -H "Content-Type: application/json" -H "Accept: application/json" \
            -u "${JIRA_USER}:${JIRA_TOKEN}" -d "$data" "${API_BASE}${endpoint}"
    else
        curl -s -X "$method" -H "Accept: application/json" -u "${JIRA_USER}:${JIRA_TOKEN}" "${API_BASE}${endpoint}"
    fi
}

cmd_view() {
    local issue="${1:-}"; [ -z "$issue" ] && error "Usage: jira-api.sh view <issue>"
    local response=$(api_call GET "/issue/${issue}?fields=key,summary,status,priority,parent,assignee,description,issuetype")
    echo "$response" | jq -e '.errorMessages' >/dev/null 2>&1 && error "$(echo "$response" | jq -r '.errorMessages[0]')"
    echo "$response" | jq '{key:.key,summary:.fields.summary,status:.fields.status.name,priority:.fields.priority.name,type:.fields.issuetype.name,epic:(.fields.parent.key//null),assignee:(.fields.assignee.displayName//"Unassigned")}'
}

cmd_list() {
    local project="${1:-}" status="${2:-}"; [ -z "$project" ] && error "Usage: jira-api.sh list <project> [status]"
    local jql="project=${project}"; [ -n "$status" ] && jql="${jql} AND status='${status}'"
    jql="${jql} ORDER BY created DESC"
    local data=$(jq -n --arg jql "$jql" '{jql:$jql,fields:["key","summary","status","priority","issuetype"],maxResults:100}')
    local response=$(api_call POST "/search/jql" "$data")
    echo "$response" | jq -e '.errorMessages' >/dev/null 2>&1 && error "$(echo "$response" | jq -r '.errorMessages[0]')"
    echo "$response" | jq -r '.issues[]|[.key,.fields.status.name,.fields.priority.name,.fields.issuetype.name,.fields.summary]|@tsv' | \
        awk 'BEGIN{printf "%-12s %-15s %-10s %-10s %s\n","KEY","STATUS","PRIORITY","TYPE","SUMMARY";print "------------------------------------------------------------------------------------"}{printf "%-12s %-15s %-10s %-10s %s\n",$1,$2,$3,$4,substr($0,index($0,$5))}'
}

cmd_create() {
    local project="${1:-}" type="${2:-}" summary="${3:-}" description="${4:-No description}"
    [ -z "$project" ] || [ -z "$type" ] || [ -z "$summary" ] && error "Usage: jira-api.sh create <project> <type> <summary> [description]"

    # Convert markdown to ADF format using co-located script
    local script_dir="$(dirname "$0")"
    local adf_description
    if [ -f "$script_dir/markdown-to-adf.js" ]; then
        adf_description=$(node "$script_dir/markdown-to-adf.js" "$description")
    else
        adf_description=$(jq -n --arg d "$description" '{type:"doc",version:1,content:[{type:"paragraph",content:[{type:"text",text:$d}]}]}')
    fi

    local data=$(jq -n --arg p "$project" --arg t "$type" --arg s "$summary" --argjson d "$adf_description" \
        '{fields:{project:{key:$p},issuetype:{name:$t},summary:$s,description:$d}}')
    local response=$(api_call POST "/issue" "$data")
    echo "$response" | jq -e '.errorMessages' >/dev/null 2>&1 && error "$(echo "$response" | jq -r '.errorMessages[]')"
    success "Created issue: $(echo "$response" | jq -r '.key')"
    echo "$response" | jq '{key:.key,id:.id,self:.self}'
}

cmd_move() {
    local issue="${1:-}" status="${2:-}"; [ -z "$issue" ] || [ -z "$status" ] && error "Usage: jira-api.sh move <issue> <status>"
    local tid; case "$status" in "To Do"|"TODO"|"todo"|11) tid="11";; "In Progress"|"IN_PROGRESS"|"in_progress"|21) tid="21";; "Done"|"DONE"|"done"|31) tid="31";; *) error "Invalid status. Use: 'To Do', 'In Progress', or 'Done'";; esac
    local response=$(api_call POST "/issue/${issue}/transitions" "$(jq -n --arg id "$tid" '{transition:{id:$id}}')")
    [ -z "$response" ] || [ "$response" = "{}" ] && { success "Moved ${issue} to ${status}"; return 0; }
    echo "$response" | jq -e '.errorMessages' >/dev/null 2>&1 && error "$(echo "$response" | jq -r '.errorMessages[]')"
    success "Moved ${issue} to ${status}"
}

cmd_delete() {
    local issue="${1:-}"; [ -z "$issue" ] && error "Usage: jira-api.sh delete <issue>"
    local response=$(api_call DELETE "/issue/${issue}")
    [ -z "$response" ] || [ "$response" = "{}" ] && { success "Deleted issue: ${issue}"; return 0; }
    echo "$response" | jq -e '.errorMessages' >/dev/null 2>&1 && error "$(echo "$response" | jq -r '.errorMessages[]')"
    success "Deleted issue: ${issue}"
}

cmd_comment() {
    local issue="${1:-}" text="${2:-}"; [ -z "$issue" ] || [ -z "$text" ] && error "Usage: jira-api.sh comment <issue> <text>"

    # Convert markdown to ADF format
    local script_dir="$(dirname "$0")"
    local adf_body
    if [ -f "$script_dir/markdown-to-adf.js" ]; then
        adf_body=$(node "$script_dir/markdown-to-adf.js" "$text")
    else
        adf_body=$(jq -n --arg t "$text" '{type:"doc",version:1,content:[{type:"paragraph",content:[{type:"text",text:$t}]}]}')
    fi

    local data=$(jq -n --argjson b "$adf_body" '{body:$b}')
    local response=$(api_call POST "/issue/${issue}/comment" "$data")
    echo "$response" | jq -e '.errorMessages' >/dev/null 2>&1 && error "$(echo "$response" | jq -r '.errorMessages[]')"
    success "Added comment to ${issue}"
    echo "$response" | jq '{id:.id,created:.created}'
}

cmd_link() {
    local issue="${1:-}" epic="${2:-}"; [ -z "$issue" ] || [ -z "$epic" ] && error "Usage: jira-api.sh link <issue> <epic>"
    local response=$(api_call PUT "/issue/${issue}" "$(jq -n --arg e "$epic" '{fields:{parent:{key:$e}}}')")
    [ -z "$response" ] || [ "$response" = "{}" ] && { success "Linked ${issue} to epic ${epic}"; return 0; }
    echo "$response" | jq -e '.errorMessages' >/dev/null 2>&1 && error "$(echo "$response" | jq -r '.errorMessages[]')"
    success "Linked ${issue} to epic ${epic}"
}

cmd_assign() {
    local issue="${1:-}" account="${2:-}"; [ -z "$issue" ] || [ -z "$account" ] && error "Usage: jira-api.sh assign <issue> <accountId>"
    local response=$(api_call PUT "/issue/${issue}/assignee" "$(jq -n --arg a "$account" '{accountId:$a}')")
    [ -z "$response" ] || [ "$response" = "{}" ] && { success "Assigned ${issue}"; return 0; }
    echo "$response" | jq -e '.errorMessages' >/dev/null 2>&1 && error "$(echo "$response" | jq -r '.errorMessages[]')"
    success "Assigned ${issue}"
}

main() {
    local cmd="${1:-}"; [ -z "$cmd" ] && { echo "Usage: jira-api.sh <command> [args...]"; echo "Commands: view, list, create, move, delete, comment, link, assign"; exit 1; }
    shift; case "$cmd" in view) cmd_view "$@";; list) cmd_list "$@";; create) cmd_create "$@";; move) cmd_move "$@";; delete) cmd_delete "$@";; comment) cmd_comment "$@";; link) cmd_link "$@";; assign) cmd_assign "$@";; *) error "Unknown command: $cmd";; esac
}

main "$@"
