#!/bin/bash
set -e
RED='\033[0;31m'; GREEN='\033[0;32m'; NC='\033[0m'
echo "MANDATORY SDLC Check"
CURRENT_BRANCH=$(git branch --show-current)
if [[ ! "$CURRENT_BRANCH" =~ ^feature/[0-9]+-.*$ ]]; then
    echo -e "${RED}BLOCKING: Not on a valid feature branch${NC}"
    echo "1. Create GitHub Issue (#XXX)"
    echo "2. git checkout -b feature/XXX-description"
    exit 1
fi
ISSUE_NUMBER=$(echo "$CURRENT_BRANCH" | grep -oE '^feature/[0-9]+' | cut -d'/' -f2)
echo -e "${GREEN}Valid feature branch: $CURRENT_BRANCH (Issue #$ISSUE_NUMBER)${NC}"
