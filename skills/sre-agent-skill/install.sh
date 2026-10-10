#!/bin/sh

SKILL_NAME="sre-agent-skill"
SKILL_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)

HOME_DIR="${HOME:-$(cd ~ && pwd)}"

# Resolve Claude skills dir: honor CLAUDE_SKILLS_DIR, else $HOME/.claude/skills
CLAUDE_SKILLS_DIR="${CLAUDE_SKILLS_DIR:-$HOME_DIR/.claude/skills}"

# Target 2: Codex-compatible agent skills dir (only if `codex` is on PATH).
CODEX_SKILLS_DIR="$HOME_DIR/.agents/skills"

install_to() {
  target_dir="$1"
  label="$2"
  dest="$target_dir/$SKILL_NAME"

  mkdir -p "$target_dir"

  if [ -L "$dest" ]; then
    rm "$dest"
  elif [ -e "$dest" ]; then
    echo "  ERROR: $label -> $dest already exists and is not a symlink."
    echo "         Refusing to clobber a real directory. Remove it manually and re-run."
    return 1
  fi

  ln -s "$SKILL_DIR" "$dest"
  echo "  $label -> $dest"
  return 0
}

echo ""
echo "Installing $SKILL_NAME..."
echo "  source: $SKILL_DIR"
echo ""

# Assert sre-agent is on PATH (warn, don't fail)
if ! command -v sre-agent >/dev/null 2>&1; then
  echo "  WARNING: 'sre-agent' is not on PATH."
  echo "           The skill is installing, but it will not function until sre-agent is available."
  echo "           Install with: go install github.com/deemwar-products/ops-phoenix-agent/sre-agent/cmd/sre-agent@latest"
  echo ""
fi

fail=0

install_to "$CLAUDE_SKILLS_DIR" "Claude Code" || fail=1

if command -v codex >/dev/null 2>&1; then
  install_to "$CODEX_SKILLS_DIR" "Codex Agent" || fail=1
fi

if [ "$fail" -ne 0 ]; then
  echo ""
  echo "Install completed with errors."
  exit 1
fi

echo ""
echo "Done. Restart your agent session to pick up the new skill."
