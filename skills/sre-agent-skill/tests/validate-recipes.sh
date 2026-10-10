#!/bin/sh
# validate-recipes.sh — ensure recipe files reference real sections and cross-references are valid.

SKILL_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
PASS=0
FAIL=0

check() {
  if [ -f "$1" ]; then
    echo "  ✓ $1"
    PASS=$((PASS + 1))
  else
    echo "  ✗ $1 (missing)"
    FAIL=$((FAIL + 1))
  fi
}

echo "Validating sre-agent-skill recipe files..."
echo ""

# Required top-level files
for f in SKILL.md install.sh uninstall.sh README.md LICENSE; do
  check "$SKILL_DIR/$f"
done

# Required references
for f in workflow.md detect.md analyze.md fix.md status.md auth.md setup.md error-classification.md; do
  check "$SKILL_DIR/references/$f"
done

# Required tests
for f in all.sh validate-recipes.sh install-test.sh; do
  check "$SKILL_DIR/tests/$f"
done

# Required bin
for f in install.sh uninstall.sh; do
  check "$SKILL_DIR/bin/$f"
done

echo ""
echo "Results: $PASS passed, $FAIL failed"
if [ "$FAIL" -ne 0 ]; then
  exit 1
fi
exit 0
