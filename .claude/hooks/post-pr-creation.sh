#!/bin/bash
set -e
echo "Running Post-PR Creation Validation..."
get_issue_number() {
    local branch=$(git branch --show-current)
    if [[ "$branch" =~ ^feature/([0-9]+)- ]]; then echo "${BASH_REMATCH[1]}"; fi
}
main() {
    local issue_number=$(get_issue_number)
    if [ -n "$issue_number" ]; then echo "GitHub Issue: #$issue_number"; fi
    echo "MANDATORY: Wait for human approval before merge"
}
main "$@"
