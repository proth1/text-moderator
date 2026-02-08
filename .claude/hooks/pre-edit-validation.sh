#!/bin/bash
RED='\033[0;31m'; GREEN='\033[0;32m'; NC='\033[0m'

validate_sdlc_compliance() {
    local file_path=$1 operation=$2
    BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
    if [[ "$BRANCH" == "main" || "$BRANCH" == "master" ]]; then
        echo -e "${RED}BLOCKED: Cannot $operation files on main/master branch${NC}"
        return 1
    fi
    echo -e "${GREEN}Pre-edit validation passed${NC}"
    return 0
}

export -f validate_sdlc_compliance
if [[ $# -ge 2 ]]; then validate_sdlc_compliance "$1" "$2"; exit $?; fi
