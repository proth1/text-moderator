#!/bin/bash
# text-moderator - Hooks Setup Script
set -e
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; CYAN='\033[0;36m'; NC='\033[0m'

echo -e "${CYAN}------------------------------------------------------------${NC}"
echo -e "${BLUE}  text-moderator - Hooks Setup Script${NC}"
echo -e "${CYAN}------------------------------------------------------------${NC}"

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"
if [ -z "$REPO_ROOT" ]; then echo -e "${RED}ERROR: Not inside a git repository.${NC}"; exit 1; fi
cd "$REPO_ROOT"
echo -e "${GREEN}Repo root:${NC} $REPO_ROOT"
ERRORS=0

echo -e "${BLUE}[1/4] Configuring git hooks path...${NC}"
if [ ! -d ".githooks" ]; then echo -e "${RED}ERROR: .githooks/ not found.${NC}"; exit 1; fi
git config core.hooksPath .githooks
echo -e "  ${GREEN}[PASS]${NC} core.hooksPath = .githooks"

echo -e "${BLUE}[2/4] Setting executable permissions...${NC}"
for f in .githooks/pre-commit .githooks/commit-msg; do
    if [ -f "$f" ]; then chmod +x "$f"; echo -e "  ${GREEN}[OK]${NC} $f"; fi
done
for f in .claude/hooks/*.sh; do
    if [ -f "$f" ]; then chmod +x "$f"; echo -e "  ${GREEN}[OK]${NC} $f"; fi
done

echo -e "${BLUE}[3/4] Checking settings...${NC}"
if [ -f ".claude/settings.json" ]; then echo -e "  ${GREEN}[PASS]${NC} .claude/settings.json"; else echo -e "  ${RED}[FAIL]${NC} .claude/settings.json not found"; ERRORS=$((ERRORS+1)); fi
if [ -f ".claude/config/project-management.yaml" ]; then
    PM_TOOL=$(grep -E "^project_management_tool:" .claude/config/project-management.yaml | sed 's/project_management_tool: *//' | tr -d '"')
    echo -e "  ${GREEN}[PASS]${NC} PM Tool: $PM_TOOL"
fi

echo -e "${BLUE}[4/4] Verification...${NC}"
HOOKSPATH=$(git config core.hooksPath 2>/dev/null)
if [ "$HOOKSPATH" = ".githooks" ]; then echo -e "  ${GREEN}[PASS]${NC} core.hooksPath = .githooks"; else echo -e "  ${RED}[FAIL]${NC} core.hooksPath"; ERRORS=$((ERRORS+1)); fi
for f in .githooks/pre-commit .githooks/commit-msg; do
    if [ -x "$f" ]; then echo -e "  ${GREEN}[PASS]${NC} $f is executable"; else echo -e "  ${RED}[FAIL]${NC} $f not executable"; ERRORS=$((ERRORS+1)); fi
done
if command -v go &>/dev/null; then echo -e "  ${GREEN}[PASS]${NC} go $(go version | awk '{print $3}')"; fi
if command -v gh &>/dev/null; then echo -e "  ${GREEN}[PASS]${NC} GitHub CLI available"; fi

echo ""
echo -e "${CYAN}------------------------------------------------------------${NC}"
if [ $ERRORS -eq 0 ]; then echo -e "${GREEN}  Setup complete! All checks passed.${NC}"; else echo -e "${YELLOW}  Setup complete with ${ERRORS} issue(s).${NC}"; fi
echo -e "${CYAN}------------------------------------------------------------${NC}"
echo ""
echo -e "${BLUE}Git Hooks (.githooks/):${NC} pre-commit, commit-msg"
echo -e "${BLUE}Claude Code Hooks:${NC} mandatory-sdlc-check, post-pr-creation, pre-edit-validation"
echo -e "${YELLOW}To bypass: git commit --no-verify${NC}"
