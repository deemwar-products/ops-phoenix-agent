#!/bin/sh
# all.sh — run all skill tests.

SKILL_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)
FAIL=0

echo "Running sre-agent-skill test suite..."
echo ""

echo "== validate-recipes.sh =="
sh "$SKILL_DIR/tests/validate-recipes.sh" || FAIL=1

echo ""
echo "== install-test.sh =="
sh "$SKILL_DIR/tests/install-test.sh" || FAIL=1

echo ""
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED: $FAIL test suite(s) failed"
  exit 1
fi
echo "All tests passed."
exit 0
