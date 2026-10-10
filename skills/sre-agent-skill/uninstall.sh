#!/bin/sh

SKILL_NAME="sre-agent-skill"
HOME_DIR="${HOME:-$(cd ~ && pwd)}"
CLAUDE_SKILLS_DIR="${CLAUDE_SKILLS_DIR:-$HOME_DIR/.claude/skills}"
CODEX_SKILLS_DIR="$HOME_DIR/.agents/skills"

uninstall_from() {
  target_dir="$1"
  label="$2"
  dest="$target_dir/$SKILL_NAME"

  if [ -L "$dest" ]; then
    rm "$dest"
    echo "  removed $label -> $dest"
  elif [ -e "$dest" ]; then
    echo "  ERROR: $label -> $dest exists but is not a symlink."
    echo "         Remove it manually."
    return 1
  else
    echo "  $label -> $dest (not present, skipping)"
  fi
  return 0
}

echo ""
echo "Uninstalling $SKILL_NAME..."
fail=0

uninstall_from "$CLAUDE_SKILLS_DIR" "Claude Code" || fail=1
uninstall_from "$CODEX_SKILLS_DIR" "Codex Agent" || fail=1

if [ "$fail" -ne 0 ]; then
  echo ""
  echo "Uninstall completed with errors."
  exit 1
fi

echo ""
echo "Done."
