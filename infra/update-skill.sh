#!/usr/bin/env bash
# Ops Phoenix — weekly skill code updater.
# Pulls the latest from the pinned ref into /opt/ops-phoenix/ without
# touching the working tree. Refuses to update if there are local changes.
set -euo pipefail

cd /opt/ops-phoenix

# Refuse if anything's dirty — agents are not developed on the VM
if ! git diff --quiet HEAD 2>/dev/null; then
 echo "local changes present in /opt/ops-phoenix, refusing to update" >&2
 exit 1
fi

# Pinned ref lives in /etc/ops-phoenix/ops-phoenix.env as OPS_PHOENIX_PINNED_REF
# Default to the installer-time branch if not set.
PINNED_REF="${OPS_PHOENIX_PINNED_REF:-origin/feat/ops-phoenix-vm-installer}"

git fetch --depth 1 origin "$PINNED_REF"
git reset --hard "origin/${PINNED_REF#origin/}"

echo "ops-phoenix skill updated to $PINNED_REF at $(date -Iseconds)"