# Step 00: Pre-flight Checks

Run this before any log querying or analysis. Stop on first failure.

## 1. Repo Guard

```bash
git rev-parse --show-toplevel
git remote get-url origin
```

Must end with `/reqsume` (e.g., `muthuishere/reqsume`).

**Fail with:** "Must be run from a /reqsume checkout."

## 2. Required Binaries

Check these are on PATH:

| Binary | Install |
|--------|---------|
| `curl` | Built into macOS |
| `jq` | `brew install jq` |
| `ssh` | Built into macOS |
| `gh` | `brew install gh` |
| `python3` | Built into macOS |

Check with: `command -v <binary>` — stop if any missing.

## 3. Vault Files (Production)

```bash
# Regression env (for Loki access)
test -f infra/vault/production/regressionprod.env || echo "MISSING"

# SSH key for prod servers
test -f infra/vault/production/keys/reqsume_prod || echo "MISSING"

# Grafana API token (for Loki via Grafana proxy)
test -f infra/observability/Archive/keys/grafana-api-token || echo "MISSING"
```

If missing → "Run `vsync pull production` to hydrate the vault."

## 4. Environment Variables

```bash
# Claude API key (for analysis)
echo "ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:+set}"

# GitHub auth
gh auth status 2>&1 | head -1
```

**Fail if:**
- `ANTHROPIC_API_KEY` not set → "Set ANTHROPIC_API_KEY in your shell env"
- `gh auth` not logged in → "Run `gh auth login` first"

## 5. Sample Verification

```bash
# Quick test: can we reach Loki via Grafana proxy?
TOKEN=$(cat infra/observability/Archive/keys/grafana-api-token)
curl -sS -H "Authorization: Bearer ${TOKEN}" \
  "https://observability.deemwar.com/api/health" | head -1
```

Should return Grafana health JSON. **Fail with:** "Cannot reach Grafana API."

## Summary

Print on success:
```
preflight OK — repo OK, binaries OK, vault hydrated, Grafana accessible
```

Then proceed to step-01.
