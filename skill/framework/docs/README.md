# Ops Phoenix Framework - Documentation Index

> **Self-Healing SRE Agent for Automated Error Detection and Fix**

---

## 📚 Documentation Index

| Document | Description |
|----------|-------------|
| **[ARCHITECTURE.md](./ARCHITECTURE.md)** | System architecture, components, data flow |
| **[QUICKSTART.md](./QUICKSTART.md)** | Quick start guide, common commands |
| **[FAQ.md](./FAQ.md)** | Frequently asked questions, troubleshooting |
| **[DIAGRAMS.md](./DIAGRAMS.md)** | Visual architecture diagrams (colorful) |

---

## 🔗 Quick Links

### Getting Started
- [Quick Start Guide](./QUICKSTART.md#quick-start) - 5 minute setup
- [First Run](./QUICKSTART.md#first-time-setup) - Configuration wizard

### Configuration
- [Config File Structure](./QUICKSTART.md#configuration)
- [Environment Variables](./QUICKSTART.md#environment-variables)
- [Observability Providers](./ARCHITECTURE.md#observability-providers)

### Usage
- [Dry Run](./QUICKSTART.md#common-commands) - Detect only
- [Full Cycle](./QUICKSTART.md#common-commands) - Detect + Fix + Deploy
- [VM Setup](./QUICKSTART.md#vm-setup-permanent) - Permanent deployment

### Troubleshooting
- [Common Issues](./FAQ.md#troubleshooting)
- [Debug Mode](./FAQ.md#debug-mode)
- [Reset Everything](./FAQ.md#reset-everything)

---

## 🦅 Framework Overview

```
ops-phoenix/
├── framework/
│   ├── ops_phoenix.py          # Main agent
│   ├── config_wizard.py         # Setup wizard
│   ├── adapters/                # Provider adapters
│   └── docs/                   # Documentation
│       ├── ARCHITECTURE.md     # Full architecture
│       ├── QUICKSTART.md       # Quick start
│       ├── FAQ.md              # Troubleshooting
│       └── DIAGRAMS.md         # Visual diagrams
└── runtime state
    ├── ~/.ops-phoenix/history.json   # Full run history
    └── ~/.ops-phoenix/memory.md      # Current lifecycle + repeated issue states
```

---

## 🚀 Quick Start

```bash
# Navigate to framework
cd infra/skills/reqsume-ops-phoenix/framework

# Set environment variables
export GRAFANA_URL="https://your-grafana.com"
export GITHUB_REPO="your-org/your-repo"
export ANTHROPIC_API_KEY="sk-ant-..."

# Run dry run
python3 ops_phoenix.py --dry-run

# Run full cycle
python3 ops_phoenix.py --env prod --full-cycle
```

---

## 📊 Architecture Summary

```
User/Env Vars
     │
     ▼
Config (~/ops-phoenix/config.json)
     │
     ├──▶ Observability Adapters (Grafana, Loki)
     ├──▶ GitHub Adapters (CLI, API)
     └──▶ CI/CD Adapters (Actions, GitLab, Jenkins)
     │
     ▼
Ops Phoenix Agent
     │
     ├──▶ 1. DETECT (query logs)
     ├──▶ 2. ANALYZE (Claude AI)
     ├──▶ 3. ISSUE (create GitHub issue)
     ├──▶ 4. FIX (generate code)
     ├──▶ 5. PR (create PR)
     ├──▶ 6. DEPLOY (merge + trigger)
     ├──▶ 7. MONITOR (wait for completion)
     └──▶ 8. AUTO-FIX (retry if failed)
```

---

## 🔧 Common Commands

| Command | Description |
|---------|-------------|
| `python3 ops_phoenix.py` | Interactive mode (asks if incomplete) |
| `python3 ops_phoenix.py --setup` | Run configuration wizard |
| `python3 ops_phoenix.py --dry-run` | Detect errors only |
| `python3 ops_phoenix.py --full-cycle` | Full workflow |
| `python3 ops_phoenix.py --config` | Show current config |
| `python3 ops_phoenix.py --window 24h` | Check last 24 hours |

---

## 🎯 Feature Status

| Feature | Status |
|---------|--------|
| Grafana Cloud adapter | 🟢 Production |
| Loki Direct adapter | 🟢 Production |
| GitHub Actions adapter | 🟢 Production |
| AI Analysis (Claude) | 🟢 Production |
| Auto-fix on failure | 🟢 Production |
| GitLab CI adapter | 🟡 Planned |
| Jenkins adapter | 🟡 Planned |
| Web UI | 🟡 Planned |

---

## 📞 Support

- Create GitHub issue
- Check [FAQ.md](./FAQ.md)
- Review [Troubleshooting](./FAQ.md#troubleshooting)