#!/bin/sh
# install-test.sh — verify install/uninstall are idempotent and safe.

SKILL_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
HOME_DIR="${HOME:-$(cd ~ && pwd)}"
CLAUDE_SKILLS_DIR="${CLAUDE_SKILLS_DIR:-$HOME_DIR/.claude/skills}"
DEST="$CLAUDE_SKILLS_DIR/sre-agent-skill"

PASS=0
FAIL=0

pass() { echo "  ✓ $1"; PASS=$((PASS + 1)); }
fail() { echo "  ✗ $1"; FAIL=$((FAIL + 1)); }

# Test 1: install script exists and is executable
if [ -x "$SKILL_DIR/install.sh" ]; then
  pass "install.sh is executable"
else
  fail "install.sh is not executable"
fi

# Test 2: uninstall script exists and is executable
if [ -x "$SKILL_DIR/uninstall.sh" ]; then
  pass "uninstall.sh is executable"
else
  fail "uninstall.sh is not executable"
fi

# Test 3: install creates symlink
mkdir -p "$CLAUDE_SKILLS_DIR"
sh "$SKILL_DIR/install.sh" >/dev/null 2>&1
if [ -L "$DEST" ]; then
  pass "install creates symlink at $DEST"
else
  fail "install did not create symlink at $DEST"
fi

# Test 4: symlink points to correct target
TARGET=$(readlink "$DEST" 2>/dev/null)
if [ "$TARGET" = "$SKILL_DIR" ]; then
  pass "symlink points to correct target"
else
  fail "symlink points to $TARGET, expected $SKILL_DIR"
fi

# Test 5: uninstall removes symlink
sh "$SKILL_DIR/uninstall.sh" >/dev/null 2>&1
if [ ! -e "$DEST" ]; then
  pass "uninstall removes symlink"
else
  fail "uninstall did not remove symlink"
fi

# Test 6: re-install works after uninstall
sh "$SKILL_DIR/install.sh" >/dev/null 2>&1
if [ -L "$DEST" ]; then
  pass "re-install works after uninstall"
else
  fail "re-install did not create symlink"
fi

# Cleanup
rm -f "$DEST"

echo ""
echo "Results: $PASS passed, $FAIL failed"
if [ "$FAIL" -ne 0 ]; then
  exit 1
fi
exit 0
