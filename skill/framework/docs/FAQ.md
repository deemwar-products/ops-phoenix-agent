# Ops Phoenix - FAQ & Troubleshooting

## Frequently Asked Questions

### General

**Q: What is Ops Phoenix?**
> Ops Phoenix is an SRE automation framework that monitors logs, detects errors, analyzes with AI, and automatically creates PRs to fix issues.

**Q: What observability providers are supported?**
> Currently: Grafana Cloud, Loki Direct, Grafana Self-Hosted
> Planned: Elasticsearch, AWS CloudWatch, Datadog

**Q: What CI/CD systems are supported?**
> Currently: GitHub Actions
> Planned: GitLab CI, Jenkins, ArgoCD

**Q: Does it work without Claude API key?**
> Yes, but it will skip the AI analysis step. You'll only get error detection without root cause analysis.

---

### Configuration

**Q: Where is the config stored?**
> `~/.ops-phoenix/config.json`

**Q: Can I use environment variables instead of interactive setup?**
> Yes! Set these env vars:
> - `GRAFANA_URL` or `LOKI_URL`
> - `GITHUB_REPO`
> - `ANTHROPIC_API_KEY`

**Q: How do I change the configuration?**
> Run `python3 ops_phoenix.py --setup` to re-run the wizard

**Q: Can I use different configs for different projects?**
> Currently no - config is per-machine. Future versions will support project-specific configs.

---

### Deployment

**Q: What's the difference between `--env dev` and `--env prod`?**
> - `dev`: Creates PR but does NOT auto-merge or deploy
> - `prod`: Auto-merges and triggers deployment

**Q: What happens if the deployment fails?**
> Ops Phoenix will:
> 1. Analyze the failure
> 2. Attempt auto-fix
> 3. Retry deployment (up to `max_retries` times)
> 4. If still failing, create a manual issue

**Q: How long does it wait for workflow completion?**
> Default: 300 seconds (5 minutes)
> Configure with `monitor_timeout_seconds` in config

---

### Security

**Q: Where are API keys stored?**
> - Grafana token: File path (not the token itself)
> - GitHub token: Environment variable (`GITHUB_TOKEN`)
> - Claude key: Environment variable (`ANTHROPIC_API_KEY`)

**Q: Are secrets committed to git?**
> No! Config file is in `~/.ops-phoenix/` (home directory, not repo)

**Q: Can I use different tokens per project?**
> Set `token_env_var` in config to point to different env vars

---

## Troubleshooting

### Connection Issues

#### "Config not found"

```
Error: No configuration found
```

**Solution:**
```bash
python3 ops_phoenix.py --setup
```

---

#### "Grafana token not found"

```
FileNotFoundError: Grafana token not found: /path/to/token
```

**Solution:**
1. Check the path is correct
2. Ensure the file exists
3. Use `~` for home directory (e.g., `~/keys/grafana-token`)

---

#### "GitHub not authenticated"

```
Warning: GitHub: Token not found in environment
```

**Solution:**
```bash
# Check current auth
gh auth status

# If not logged in
gh auth login
```

---

### Error Detection Issues

#### "No errors found" but I know there are errors

**Possible causes:**
1. Time window too short - try `--window 24h`
2. Container patterns don't match - check config
3. Error patterns don't match - verify Loki query

**Solution:**
```bash
# Check config
python3 ops_phoenix.py --config

# Use longer time window
python3 ops_phoenix.py --window 7d --dry-run
```

---

#### "Query failed" or timeout

**Possible causes:**
1. Grafana/Loki is down
2. Network issue
3. Token expired

**Solution:**
```bash
# Test connection manually
curl -H "Authorization: Bearer $(cat ~/Downloads/keys/grafana-token)" \
  https://your-grafana.com/api/health
```

---

### GitHub Issues

#### "Repository not found"

```
gh: Repository not found: owner/repo
```

**Solution:**
1. Verify `GITHUB_REPO` is correct
2. Ensure you have access to the repo
3. Check token has correct scopes

---

#### "Merge failed"

```
Merge failed: PR already merged
```

**Solution:**
This is usually fine - another process already merged. Check if your PR exists.

---

### Deployment Issues

#### "Workflow not found"

```
Workflow trigger failed: workflow not found
```

**Solution:**
1. Check `workflow_name` in config matches exactly
2. Verify workflow exists in `.github/workflows/`

---

#### "Workflow timeout"

```
Workflow monitoring timed out
```

**Solution:**
1. Increase `monitor_timeout_seconds` in config
2. Check if workflow is actually running in GitHub Actions

---

### Claude AI Issues

#### "API key not set"

```
Warning: ANTHROPIC_API_KEY not set, skipping analysis
```

**Solution:**
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

---

#### "Claude API error"

```
Claude API error: 429 Rate Limit Exceeded
```

**Solution:**
1. Wait and retry
2. Check your Claude API quota
3. Reduce frequency of runs

---

## Debug Mode

Enable verbose logging:

```bash
export DEBUG=1
python3 ops_phoenix.py --dry-run
```

---

## Reset Everything

```bash
# Remove config
rm ~/.ops-phoenix/config.json

# Remove history
rm ~/.ops-phoenix/history.log

# Start fresh
python3 ops_phoenix.py --setup
```

---

## Get Help

If you're stuck:

1. Check this FAQ
2. Run `python3 ops_phoenix.py --help`
3. Check GitHub issues
4. Create a new issue with:
   - Config (redacted)
   - Error message
   - Steps to reproduce