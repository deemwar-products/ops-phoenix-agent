#!/usr/bin/env bash
# Ops Phoenix — VM installer.
# Idempotent. Run with --dry-run to preview. Reads GH_TOKEN_FILE from env.
#
# Usage:
# bash infra/setup/install-ops-phoenix.sh --dry-run
# GH_TOKEN_FILE=/path/to/gh-token.txt bash infra/setup/install-ops-phoenix.sh
#
# Exit codes:
# 0 = success (or dry-run finished)
# 1 = precondition not met (vm unreachable, secrets missing, etc.)
# 2 = partial install (some steps failed; check output)

set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
 DRY_RUN=true
 echo "=== DRY RUN — no changes will be made ==="
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ASSETS="${REPO_ROOT}/infra/setup/ops-phoenix"
VM="${OPS_PHOENIX_VM:-deemwar-dev}" # ssh alias from ~/.ssh/config
VM_USER="${OPS_PHOENIX_VM_USER:-root}"

# --- helpers ---
say() { echo "[$(date +%H:%M:%S)] $*"; }
run_remote() {
 # run on VM as opsphoenix for skill ops, root for system ops
 local remote_user="$1"; shift
 if $DRY_RUN; then
 echo " [dry-run] would run on $VM as $remote_user: $*"
 else
 ssh -o ConnectTimeout=10 -o BatchMode=yes "${VM_USER}@${VM}" "$*"
 fi
}

say "vm: $VM (user: $VM_USER)"
say "assets: $ASSETS"

# --- preflight: assets exist locally ---
for f in config.json ops-phoenix.env README \
 ops-phoenix-runner.sh update-skill.sh \
 ops-phoenix-hourly.service ops-phoenix-hourly.timer \
 ops-phoenix-update.service ops-phoenix-update.timer \
 ops-phoenix-logrotate; do
 [[ -f "${ASSETS}/${f}" ]] || { echo "missing asset: ${ASSETS}/${f}" >&2; exit 1; }
done

# --- preflight: VM reachable ---
say "checking vm reachability..."
if ! ssh -o ConnectTimeout=5 -o BatchMode=yes "${VM_USER}@${VM}" true 2>/dev/null; then
 echo "cannot reach $VM via ssh. check your ssh config." >&2
 exit 1
fi
say "vm reachable."

# --- preflight: GH_TOKEN_FILE if not dry-run ---
GH_TOKEN_CONTENT=""
if ! $DRY_RUN; then
 if [[ -z "${GH_TOKEN_FILE:-}" ]]; then
 echo "GH_TOKEN_FILE not set. generate the fine-grained PAT, save to a file, and re-run with:"
 echo " GH_TOKEN_FILE=/path/to/gh-token.txt bash $0"
 exit 1
 fi
 [[ -f "$GH_TOKEN_FILE" ]] || { echo "GH_TOKEN_FILE=$GH_TOKEN_FILE does not exist" >&2; exit 1; }
 GH_TOKEN_CONTENT="$(cat "$GH_TOKEN_FILE")"
 [[ -n "$GH_TOKEN_CONTENT" ]] || { echo "GH_TOKEN_FILE is empty" >&2; exit 1; }
 # token format check — must start with github_pat_ (fine-grained) or ghp_ (classic)
 if [[ ! "$GH_TOKEN_CONTENT" =~ ^(github_pat_|ghp_) ]]; then
 echo "token in GH_TOKEN_FILE does not start with github_pat_ or ghp_. is this a real GitHub PAT?" >&2
 exit 1
 fi
fi

# --- 1. user + dirs (idempotent) ---
say "[1/8] creating opsphoenix user + dirs..."
run_remote root bash <<'REMOTE'
set -e
id opsphoenix &>/dev/null || useradd --system \
 --home /var/lib/ops-phoenix --shell /usr/sbin/nologin opsphoenix
install -d -o opsphoenix -g opsphoenix -m 0750 /opt/ops-phoenix
install -d -o opsphoenix -g opsphoenix -m 0750 /var/lib/ops-phoenix/state
install -d -o opsphoenix -g opsphoenix -m 0750 /var/log/ops-phoenix/runs
install -d -o root -g opsphoenix -m 0750 /etc/ops-phoenix
install -d -o opsphoenix -g opsphoenix -m 0750 /var/lib/ops-phoenix
touch /var/lock/ops-phoenix.lock
chown root:opsphoenix /var/lock/ops-phoenix.lock
chmod 0644 /var/lock/ops-phoenix.lock
REMOTE

# --- 2. vendored skill ---
say "[2/8] vendoring skill into /opt/ops-phoenix..."
# We assume the user has already pushed the branch. We clone from
# origin/<current-branch> (passed via OPS_PHOENIX_PINNED_REF).
PINNED_REF="${OPS_PHOENIX_PINNED_REF:-origin/feat/ops-phoenix-vm-installer}"
if $DRY_RUN; then
 echo " [dry-run] would clone ${PINNED_REF} into /opt/ops-phoenix"
 echo " [dry-run] would copy skill subdir (infra/skills/reqsume-ops-phoenix) to /opt/ops-phoenix"
else
 run_remote root bash -s <<REMOTE
set -e
if [[ ! -d /opt/ops-phoenix/.git ]]; then
 git clone --depth 1 --branch "${PINNED_REF#origin/}" \
 https://github.com/deemwar-products/reqsume.git /opt/ops-phoenix
else
 cd /opt/ops-phoenix
 git fetch --depth 1 origin "${PINNED_REF#origin/}"
 git reset --hard "origin/${PINNED_REF#origin/}"
fi
chown -R opsphoenix:opsphoenix /opt/ops-phoenix
REMOTE
 # Copy the skill subdir from the clone
 run_remote root bash -s <<REMOTE
set -e
# The skill lives in infra/skills/reqsume-ops-phoenix/ in the repo.
# We just need that subtree as the working tree.
cd /opt/ops-phoenix
# The clone has the whole repo; we want the skill subdir as the root.
# Cleanest: re-clone with sparse-checkout.
rm -rf .git
git init -q
git remote add origin https://github.com/deemwar-products/reqsume.git
git config core.sparseCheckout true
echo "infra/skills/reqsume-ops-phoenix/*" > .git/info/sparse-checkout
git fetch --depth 1 origin "${PINNED_REF#origin/}"
git checkout "${PINNED_REF#origin/}" -- infra/skills/reqsume-ops-phoenix
# Move the subtree up so /opt/ops-phoenix/framework/... etc work as spec expects
rm -rf .git .gitignore
mv infra/skills/reqsume-ops-phoenix/* .
mv infra/skills/reqsume-ops-phoenix/.[!.]* . 2>/dev/null || true
rmdir infra/skills/reqsume-ops-phoenix 2>/dev/null || true
git init -q # for the update-skill.sh to use
git add -A
git -c user.email=opsphoenix@reqsume.local -c user.name=opsphoenix commit -q -m "sparse checkout: skill only"
chown -R opsphoenix:opsphoenix /opt/ops-phoenix
REMOTE
fi

# --- 3. python venv + deps ---
say "[3/8] python venv + skill deps..."
if $DRY_RUN; then
 echo " [dry-run] would create /opt/ops-phoenix/.venv and pip install anthropic"
else
 run_remote root bash <<'REMOTE'
set -e
apt-get install -y python3-venv python3-pip >/dev/null
if [[ ! -d /opt/ops-phoenix/.venv ]]; then
 sudo -u opsphoenix python3 -m venv /opt/ops-phoenix/.venv
 sudo -u opsphoenix /opt/ops-phoenix/.venv/bin/pip install --quiet --upgrade pip
 sudo -u opsphoenix /opt/ops-phoenix/.venv/bin/pip install --quiet anthropic
fi
REMOTE
fi

# --- 4. drop config + env + README ---
say "[4/8] dropping config + env + README into /etc/ops-phoenix..."
scp_to_vm() {
 local src="$1" dst="$2"
 if $DRY_RUN; then
 echo " [dry-run] would scp ${src} -> ${VM}:${dst}"
 else
 scp -o BatchMode=yes "$src" "${VM_USER}@${VM}:${dst}"
 fi
}
scp_to_vm "${ASSETS}/config.json" /etc/ops-phoenix/config.json
scp_to_vm "${ASSETS}/ops-phoenix.env" /etc/ops-phoenix/ops-phoenix.env
scp_to_vm "${ASSETS}/README" /etc/ops-phoenix/README
if ! $DRY_RUN; then
 run_remote root bash <<'REMOTE'
chown root:opsphoenix /etc/ops-phoenix/config.json /etc/ops-phoenix/ops-phoenix.env /etc/ops-phoenix/README
chmod 0640 /etc/ops-phoenix/config.json /etc/ops-phoenix/ops-phoenix.env
chmod 0644 /etc/ops-phoenix/README
REMOTE
fi

# --- 5. secret file (ANTHROPIC_* + merge with existing anthropic-key) ---
say "[5/8] setting up /etc/ops-phoenix/secret..."
if $DRY_RUN; then
 echo " [dry-run] would migrate /etc/ops-phoenix/anthropic-key -> /etc/ops-phoenix/secret"
 echo " [dry-run] would append ANTHROPIC_BASE_URL=https://api3.claudestore.store"
 echo " [dry-run] would set perms to 0600 root:opsphoenix"
else
 run_remote root bash <<'REMOTE'
set -e
# If anthropic-key exists and secret doesn't, migrate
if [[ -f /etc/ops-phoenix/anthropic-key ]] && [[ ! -f /etc/ops-phoenix/secret ]]; then
 mv /etc/ops-phoenix/anthropic-key /etc/ops-phoenix/secret
fi
# Append the base URL if it's not already there
if ! grep -q '^ANTHROPIC_BASE_URL=' /etc/ops-phoenix/secret 2>/dev/null; then
 echo 'ANTHROPIC_BASE_URL=https://api3.claudestore.store' >> /etc/ops-phoenix/secret
fi
chown root:opsphoenix /etc/ops-phoenix/secret
chmod 0600 /etc/ops-phoenix/secret
REMOTE
fi

# --- 6. gh-token ---
say "[6/8] dropping gh-token..."
if $DRY_RUN; then
 echo " [dry-run] would write \$GH_TOKEN_FILE to /etc/ops-phoenix/gh-token (mode 0600 root:opsphoenix)"
 echo " (token NOT shown; install requires GH_TOKEN_FILE env var)"
else
 # write token via stdin to avoid ssh argv
 printf '%s' "$GH_TOKEN_CONTENT" | ssh -o BatchMode=yes "${VM_USER}@${VM}" \
 "cat > /etc/ops-phoenix/gh-token && chown root:opsphoenix /etc/ops-phoenix/gh-token && chmod 0600 /etc/ops-phoenix/gh-token"
fi

# --- 7. wrapper + update script ---
say "[7/8] installing runner wrapper + update-skill.sh..."
scp_to_vm "${ASSETS}/ops-phoenix-runner.sh" /opt/ops-phoenix/bin/ops-phoenix-runner.sh
scp_to_vm "${ASSETS}/update-skill.sh" /opt/ops-phoenix/bin/update-skill.sh
if ! $DRY_RUN; then
 run_remote root bash <<'REMOTE'
install -d -o opsphoenix -g opsphoenix -m 0750 /opt/ops-phoenix/bin
mv /opt/ops-phoenix/ops-phoenix-runner.sh /opt/ops-phoenix/bin/ops-phoenix-runner.sh
mv /opt/ops-phoenix/update-skill.sh /opt/ops-phoenix/bin/update-skill.sh
chown root:opsphoenix /opt/ops-phoenix/bin/ops-phoenix-runner.sh /opt/ops-phoenix/bin/update-skill.sh
chmod 0750 /opt/ops-phoenix/bin/ops-phoenix-runner.sh /opt/ops-phoenix/bin/update-skill.sh
REMOTE
fi

# --- 8. systemd + logrotate ---
say "[8/8] installing systemd units + logrotate..."
scp_to_vm "${ASSETS}/ops-phoenix-hourly.service" /etc/systemd/system/ops-phoenix-hourly.service
scp_to_vm "${ASSETS}/ops-phoenix-hourly.timer" /etc/systemd/system/ops-phoenix-hourly.timer
scp_to_vm "${ASSETS}/ops-phoenix-update.service" /etc/systemd/system/ops-phoenix-update.service
scp_to_vm "${ASSETS}/ops-phoenix-update.timer" /etc/systemd/system/ops-phoenix-update.timer
scp_to_vm "${ASSETS}/ops-phoenix-logrotate" /etc/logrotate.d/ops-phoenix
if ! $DRY_RUN; then
 run_remote root bash <<'REMOTE'
chmod 0644 /etc/systemd/system/ops-phoenix-*.{service,timer}
chmod 0644 /etc/logrotate.d/ops-phoenix
systemctl daemon-reload
systemctl enable ops-phoenix-update.timer
systemctl enable --now ops-phoenix-hourly.timer
REMOTE
fi

say "done."
$DRY_RUN && say "re-run without --dry-run (and with GH_TOKEN_FILE) to apply." || say "systemd timer enabled. next fire within ~90 min."